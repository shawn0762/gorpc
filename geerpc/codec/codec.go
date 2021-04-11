package codec

import "io"

// rpc请求头
type Header struct {
	// 服务方法
	ServiceMethod string
	// 序号：不同请求不同序号
	Seq uint64
	// 错误信息
	Error string
}

// 定义编/解码抽象接口
// 所有实现此接口的编/解码器都可以替换掉默认的编/解码器
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// 定义一个函数类型，
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	// 定义两种编码类型
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

// 定义一个全局变量，用来注册不同类型编码器的构造函数
var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	// 初始化全局变量NewCodecFuncMap
	NewCodecFuncMap = make(map[Type]NewCodecFunc)

	// 默认使用go自带的Gob解码器
	NewCodecFuncMap[GobType] = NewGobCodec
}
