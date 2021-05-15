package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorpc/codec"
	"gorpc/server"
	"log"
	"net"
	"sync"
)

// Client 一个client实例，表示与服务器建立的一个connection
type Client struct {
	// 此连接使用的编码器
	cc codec.Codec

	opt *server.Option

	header codec.Header

	// 请求序号：每注册一次请求，都会自增1
	seq uint64

	// 正在处理的请求
	pending map[uint64]*Call

	// 保证pending写入的并发安全
	sending sync.Mutex

	// 保证client的并发安全
	mu sync.Mutex

	// client已关闭：通常是用户主动关闭
	Closing bool
	// client已关闭：通常是由错误引起的异常关闭
	Shutdown bool
}

var ErrClientClosed = errors.New("client has been closed")

// Close 关闭此客户端实例
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Closing {
		return ErrClientClosed
	}

	c.Closing = true
	return nil
}

// IsAvailable 判断当前client实例是否可用
func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Closing || c.Shutdown {
		return false
	}
	return true
}

func NewClient(network string, address string, opt *server.Option) (*Client, error) {
	// 可以不指定option，此时使用默认的option实例
	if opt == nil {
		opt = server.DefaultOption
	}

	f, ok := codec.NewCodecFuncMap[opt.CodecType]
	if !ok {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client codec error:", err)
		return nil, err
	}

	// 与服务端建立连接
	conn, err := net.Dial(network, address)
	if err != nil {
		log.Println("Rpc client connect to server failed:", err)
		return nil, err
	}

	// 发送option，告诉服务端使用什么编码器
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		// 如果发送错误，关闭连接
		log.Println("Rpc client option json encode err:", err)
		_ = conn.Close()
		return nil, err
	}

	// client准备完毕，可以开始发送rpc请求
	client := &Client{
		cc:      f(conn),
		opt:     opt,
		seq:     1, // 正常的请求编号从1开始
		pending: make(map[uint64]*Call),
	}

	// 启动协程等待服务器返回数据
	go client.receive()

	return client, nil
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 这里不能调用IsAvailable方法，会出现死锁
	if c.Closing || c.Shutdown {
		return 0, ErrClientClosed
	}

	// 给call编号，后续根据编号写入服务器返回内容
	call.Seq = c.seq
	c.pending[c.seq] = call
	// 请求编号自增1，给下一次调用准备
	c.seq++

	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()

	call, ok := c.pending[seq]
	if ok {
		delete(c.pending, seq)
	}
	return call
}

// 服务端或客户端发生错误时调用
// 取消所有请求，并关闭client
func (c *Client) shutdown(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Shutdown = true
	for _, call := range c.pending {
		call.Err = err
		call.done()
	}
}

// 接收服务器的响应
func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}

		// 成功读取到Header，说明服务器已经开始返回数据
		// 先从请求队列中移除
		// 这里获取到的call有几种情况：
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			// 1、call为nil：可能是请求发送不完整，或因为什么原因被取消了，服务器实际上有处理该请求
			// 这种情况下，直接丢弃服务器返回的数据
			err = c.cc.ReadBody(nil)
		case h.Err != "":
			// 2、服务器返回了错误，此时也应该丢弃返回的数据
			// 但是请求是完整的，需要显式结束call，让调用方知道结果和错误信息
			call.Err = fmt.Errorf(h.Err)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			// 3、一切正常
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Err = errors.New("Rpc client Failed to read body: " + err.Error())
			}
			call.done()
		}
	}

	// @todo receive协程只有一个，这里出了问题退出之后怎么重启新的receive协程？
	c.shutdown(err)
}

func (c *Client) send(call *Call) {
	// 这里加锁保证一个请求要完整地发送出去
	c.sending.Lock()
	defer c.sending.Unlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Err = err
		call.done()
		return
	}

	// 所有请求都共用一个header实例
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = call.Seq
	c.header.Err = ""

	// 开始发送请求
	// 这里发送完就退出，不等待结果
	// 调用方需要通过监听call.Done通道去读取服务器返回的结果
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Err = err
			call.done()
		}
	}
}

// Go 方法是异步请求方法，发出请求后不会等待服务器返回结果，直接返回一个call实例
// 调用方需要通过监听通道自定等待调用结果
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		// 非缓冲通道会阻塞写入：当通道中有数据时，需要等到该数据被取出后才能写入，效率不好
		// 所以这里强制要求使用缓冲通道
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)

	return call
}

// Call 方法是同步请求的方法，发出请求后会阻塞，直到服务器返回结果
// 最终返回调用的错误状态
func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <- c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Err
}