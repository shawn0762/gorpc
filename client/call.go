package client

// 一个call表示一个rpc请求
type Call struct {
	// 格式：服务名.方法名
	ServiceMethod string
	// 参数
	Args interface{}
	// 返回值
	Reply interface{}
	// 表示请求编号，每一次call实例都会有一个唯一的编号
	// 由于多个rpc调用是并发的，当服务器返回结果时需要根据编号找到对应的call实例
	// 这个编号由client生成
	Seq uint64

	// 请求结束时，会向这个通道写入该call实例
	// 调用方通过监听此通道来等待调用结果
	Done chan *Call

	// 如果调用发生错误，会将错误信息写入Err
	Err error
}

func (c *Call) done() {
	c.Done <- c
}

//func NewCall(serviceMethod string, args, reply interface{}) *Call {
//	return &Call{
//		ServiceMethod: serviceMethod,
//		Args:          args,
//		Reply:         reply,
//		Done:          make,
//	}
//}
