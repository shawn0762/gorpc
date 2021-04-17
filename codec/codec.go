package codec

import (
	"io"
)

// 抽象Header
type Header struct {
	// 服务方法名
	ServiceMethod string
	// 请求序号，一个请求一个唯一序号？
	Seq uint64
	// 发生错误时的错误说明
	Err string
}

// 定义编码器的接口规范
// 方便使用不同编码方式进行编码

type Codec interface {
	io.Closer // 需要关闭？

	// 读取请求头内容，写入Header实例
	ReadHeader(*Header) error

	// 读取请求体内容，写入Header实例
	ReadBody(interface{}) error

	// 写入响应内容：响应头、响应体
	Write(*Header, interface{}) error
}

type Type string

// 将所有支持的编码器类型注册到常量
const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

// 所有编码器都要接收tcp连接，然后从中读取、反编码请求体的内容
// 这里将所有编码器都必须实现的构造方法抽象为函数类型
type NewCodecFunc func(conn io.ReadWriteCloser) Codec

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)

	// 注册Gob编码器
	NewCodecFuncMap[GobType] = NewGobCodec
	// 注册Json编码器
	NewCodecFuncMap[JsonType] = NewJsonCodec
}
