package server

import (
	"encoding/json"
	"gorpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

// Server负责：
// 1、启动服务、监听端口
// 2、接收请求连接
// 3、处理请求
// 4、返回响应内容
type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func StartServer() {
	// 随便使用一个端口
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Println("Rpc server listen err: ", err.Error())
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Rpc server accept err: ", err.Error())
			continue
		}
		go DefaultServer.Accept(conn)
	}
}

// |<-- option -->|<-- header -->|<-- body -->|<-- header -->|<-- body -->|...|
func (s *Server) Accept(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	// 读取第一段内容
	var opt Option
	err := json.NewDecoder(conn).Decode(&opt)
	if err != nil {
		log.Println("Rpc server option decode err: ", err.Error())
		return
	}

	// 对比magic number是否正确
	if opt.MagicNumber != MagicNumber {
		log.Println("Rpc server option err, not a valid magic number: ", opt.MagicNumber)
		return
	}

	// 是否支持该解码器
	f, ok := codec.NewCodecFuncMap[opt.CodecType]
	if !ok {
		log.Println("Rpc server option err, unknown codec type: ", opt.CodecType)
		return
	}
	s.handle(f(conn))
}

func (s *Server) handle(cc codec.Codec) {
	// 处理请求是并发的，但响应只能一个个写
	sending := new(sync.Mutex)

	// 一次连接，可以请求n次，这里要等n次请求都处理完之后才能关闭连接
	// 通过group实现
	group := new(sync.WaitGroup)

	for {
		// @todo 怎么判断是否正常？什么情况下退出？
		req, err := s.readRequest(cc)
		if err != nil {
			continue
		}

		if req == nil {
			break
		}

		// 请求的rpc服务和方法
		// 业务逻辑处理

		group.Add(1)

		// 响应
		s.sendResponse(cc, req, sending)
	}
	group.Wait()
	_ = cc.Close()
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("Rpc server handle err, failed to read header: ", err.Error())
		}
		return nil, err
	}

	req := &request{
		header: &h,
		argv:   reflect.Value{},
		replyv: reflect.Value{},
	}
	// @todo 现在还不知道body是什么结构
	req.argv = reflect.New(reflect.TypeOf(""))
	if err := cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("Rpc server handle err, failed to read body: ", err.Error())
		return nil, err
	}
	return req, nil
}

func (s *Server) sendResponse(cc codec.Codec, req *request, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	// @todo 这里要加锁
	err := cc.Write(req.header, req.replyv.Interface())
	if err != nil {
		log.Println("Rpc server handle err, failed send response: ", err.Error())
	}
}
