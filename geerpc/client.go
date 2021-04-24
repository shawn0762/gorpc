package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"sync"
)

// Call represents an active RPC.
// Call表示一次rpc调用
type Call struct {
	Seq           uint64
	ServiceMethod string      // format "<service>.<method>"
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // Strobes when call is complete.
}

// 异步调用结束时，调用此方法通知调用方
func (call *Call) done() {
	// 调用方会监听Done这个通道，这样有结果时就能收到响应的call实例
	call.Done <- call
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	cc  codec.Codec
	opt *Option

	// 保证请求有序发出，防止多个请求的报文混在一起
	sending sync.Mutex // protect following
	header  codec.Header
	mu      sync.Mutex // protect following

	// 用于给发送的请求编号，每个请求拥有唯一编号
	seq uint64

	// 未处理完的请求
	pending map[uint64]*Call

	// client的状态
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

// 保证Client必须实现io.Closer接口
var _ io.Closer = (*Client)(nil)

// 自定义一个错误，相当于一个错误类型，用于判断
var ErrShutdown = errors.New("connection is shut down")

// Close the connection
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		// 正在关闭的client不能再次关闭
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

// client是否可用
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	// closing说明用户主动关闭了client
	// shutdown说明发生了异常
	return !client.shutdown && !client.closing
}

func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		// 如果客户端正在关闭或已经关闭，不能再用此客户端实例发出rpc请求
		return 0, ErrShutdown
	}

	call.Seq = client.seq
	client.pending[call.Seq] = call

	// 这样client的编号就是最新的没有被使用的
	// 下一次发起一个rpc调用时，就可以直接使用这个编号
	client.seq++
	return call.Seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()

	// 这里可能不存在这个call，此时call是nil
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

// 接受rpc服务端返回的响应
func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}

		// 能读取到header，说明本地调用已经完成，从client中移除call实例
		// @todo 这里可能是nil
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			// it usually means that Write partially failed
			// and call was already removed.
			// @todo 这里不是很理解，直接丢弃body部分？
			// call不存在，可能是请求没有完整发送，或者因为其他原因被取消，但是服务端仍旧处理了
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			// header中的Error非空，表示服务端发生了错误
			call.Error = fmt.Errorf(h.Error)
			// 直接丢弃body部分
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			// 将body读取到call实例
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// error occurs, so terminateCalls pending calls
	// @todo 关闭全部请求？
	client.terminateCalls(err)
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	// 使用指定的编码器
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	// 发送option给服务端，告诉服务器使用什么编码器
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	// 通过协程等待读取服务端响应的信息
	// @todo 如果出了问题？怎么知道client还能不能用？
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	// if opts is nil or pass nil as parameter
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial connects to an RPC server at the specified network address
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

// 发送rpc请求
func (client *Client) send(call *Call) {
	// make sure that the client will send a complete request
	client.sending.Lock()
	defer client.sending.Unlock()

	// register this call.
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// prepare request header
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// encode and send the request
	// 发送请求后直接返回，不等待结果
	// receive协程会等待结果并将结果写入call实例
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		// call may be nil, it usually means that Write partially failed,
		// client has received the response and handled
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go invokes the function asynchronously.
// It returns the Call structure representing the invocation.
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		// @todo 这里为什么是10
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		// 如果容量为0，则是非缓冲通道，这样会出现阻塞现象：写入一个值后，要等到这个值被取出才能写入下一个值
		// 所以为了提高效率，这里要求chan必须为缓冲通道
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	// 这里发出请求后就直接返回，异步等待结果
	client.send(call)
	return call
}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	// 这里会堵塞，直到请求返回结果后才能接收到call实例
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}