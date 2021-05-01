package server

import "reflect"

// 一个method代表一个可以被客户端调用的方法
type methodType struct {
	// 方法体本身
	method reflect.Method
	// 第一个参数
	argType reflect.Type
	// 第二个参数
	replyType reflect.Type
	// 调用次数
	numCalls uint64
}

// newArg 根据arg属性实例化一个对象
func (m methodType) newArgType() reflect.Value {
	var argValue reflect.Value
	// 首先要知道type
	if m.argType.Kind() == reflect.Ptr {
		// 如果argType是指针，则要通过Elem方法获得所指向的真正的类型
		// New方法返回的是一个指针Value，即所需要返回的类型
		argValue = reflect.New(m.argType.Elem())
	} else {
		// 其他就是值类型，可以直接New
		// 但New返回的是指针Value，而argType是值类型，所以argValue必须是一个值类型
		// 所以New完还需要通过Elem方法获取相应的值才能返回
		argValue = reflect.New(m.argType)
	}
	return argValue
}

func (m methodType) newReply() reflect.Value {
	// 因为需要将相应内容写入第二个参数，供调用方接收
	// 所以第二个参数必须是指针
	// 如果不是指针调用Elem方法时会直接报错
	replyVal := reflect.New(m.replyType.Elem())

	switch m.replyType.Elem().Kind() {
	case reflect.Map:
		// 创建一个map：reflect.New出来的map类型Value只是零值，并没有初始化
		// map必须要初始化之后才能使用
		newMap := reflect.MakeMap(m.replyType.Elem())
		replyVal.Set(newMap)
	case reflect.Slice:
		// 创建一个slice：原因同map
		newSlice := reflect.MakeSlice(m.replyType.Elem(), 0, 0)
		replyVal.Set(newSlice)
	}
	return replyVal
}
