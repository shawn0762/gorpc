package server

import "gorpc/codec"

// 客户端传递的MagicNumber必须等于这个常量
const MagicNumber = 0x3bef5c

// 客户端通过option交换协议
type Option struct {
	// 用这个表明这是一个rpc请求
	MagicNumber int
	// 使用的解码器
	CodecType codec.Type
}

// 提供一个默认的Option实例，方便使用
var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}
