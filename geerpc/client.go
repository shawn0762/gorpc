package geerpc

import (
	"errors"
	"geerpc/codec"
	"io"
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
	seq     uint64

	// 未处理完的请求
	pending map[uint64]*Call

	// client的状态
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

// 保证Client必须实现io.Closer接口
var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

// Close the connection
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
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
