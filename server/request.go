package server

import (
	"gorpc/codec"
	"reflect"
)

type request struct {
	header *codec.Header

	// 请求参数和响应参数
	argv, replyv reflect.Value
}
