/*Package surpc 提供结构体与服务的映射，通过反射实现。
 */
package surpc

import (
	"go/token"
	"log"
	"reflect"
	"sync/atomic"
)

// rpc调用的方法必须为可导出，且包含两个参数，包含一个返回参数
// 第一个参数为请求参数，第二个参数为接收返回结果的参数
// 方法的返回参数是error
type methodType struct {
	// 方法本身
	method reflect.Method
	// 第一个参数类型
	ArgType reflect.Type
	// 第二个参数类型
	ReplyType reflect.Type
	// 调用次数
	numCalls uint64
}

// NumCalls 调用次数
func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// 构造请求参数
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

// 构造返回参数实例
func (m *methodType) newReplyv() reflect.Value {
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
	// 映射结构体的名称
	name string
	// 结构体类型
	typ reflect.Type
	// 结构体本身
	rcvr reflect.Value
	// 存储映射的结构体所有符合条件（已注册）的方法(可导出，且参数满足条件)
	method map[string]*methodType
}

func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.typ = reflect.TypeOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	if !token.IsExported(s.name) {
		log.Fatalf("%s is not a service name", s.name)
	}

	s.registerMethods()

	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		// 第一个参数是结构体本身，第二个是入参，第三个是接收参数
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		// 返回值只有一个，且为error
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

// 判断是否为可导出类型或为内建类型
func isExportedOrBuiltinType(typ reflect.Type) bool {
	return token.IsExported(typ.Name()) || typ.PkgPath() == ""
}

// 通过反射值调用方法
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
