package geerpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

// 对外提供rpc调用服务的每一个方法，都会
// 抽象成一个methodType
type methodType struct {
	// 方法本身
	method    reflect.Method
	// 第一个参数
	ArgType   reflect.Type
	// 第二个参数
	ReplyType reflect.Type
	// 调用次数
	numCalls  uint64
}

// NumCalls 记录该方法被调用的次数
func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	// arg may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		// Elem方法需要输入值为interface或者指针
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	// reply must be a pointer type
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}


type service struct {
	name   string
	// 所注册的rpc服务结构体类型
	typ    reflect.Type

	// 所注册的rpc服务实例
	rcvr   reflect.Value
	// rpc实例对外暴露的方法
	method map[string]*methodType
}

// newService 将一个rpc服务，注册成service
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		// 所注册的rpc服务结构体，必须是包外可见的
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

// registerMethods 获取rpc服务结构体的方法，供客户端调用
// 可供调用的条件：
// 1、方法是包外可见的
// 2、返回值只能有一个，且必须是error类型
// 3、必须是三个入参，第二、三个必须是包外可见的自定义类型，或者是内建类型
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}

		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	// Call方法的参数数组，第一个元素必须是方法所属的实例本身
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}