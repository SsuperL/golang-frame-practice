package surpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int
type Args struct{ Arg1, Arg2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Arg1 + args.Arg2
	return nil
}

func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Arg1 + args.Arg2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertsion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	service := newService(&foo)
	mType := service.method["Sum"]
	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Arg1: 1, Arg2: 3}))
	err := service.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == uint64(1), "failed to call Foo.Sum")
}
