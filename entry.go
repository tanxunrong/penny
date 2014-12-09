package penny

import (
	proto "./proto"
	"reflect"
	"sync"
	"sync/atomic"
)

type Entry struct {
	name      string
	item_type reflect.Type
	items     []reflect.Value
	locks     []sync.Mutex
	reqn      int64
	mq        chan proto.Msg
	mutex     sync.Mutex
}

// start service
func (serv *Entry) start() {
	for {
		m := <-serv.mq

		cur := atomic.AddInt64(&serv.reqn, 1)
		idx := int(cur % int64(len(serv.items)))

		go serv.callService(idx, m)
	}
}

func (serv *Entry) callService(idx int, m proto.Msg) {
	serv.locks[idx].Lock()
	defer serv.locks[idx].Unlock()

	rv := serv.items[idx]
	if !rv.IsValid() {
		rv = setup(serv.item_type)
		serv.items[idx] = rv
	}

	callMethod := rv.MethodByName("Call")
	param := []reflect.Value{reflect.ValueOf(m)}
	ret := callMethod.Call(param)

	//if call return error,then close and remove the instance
	//if close failed,panic
	if !ret[0].IsNil() {

		closeMethod := rv.MethodByName("Close")
		param = []reflect.Value{}
		closeRet := closeMethod.Call(param)
		if !closeRet[0].IsNil() {
			panic("close failed")
		}

		// remove instance
		serv.items[idx] = reflect.ValueOf(nil)
	}
}

func setup(tp reflect.Type) reflect.Value {

	instance := reflect.New(tp)
	initMethod := instance.MethodByName("Init")
	if initMethod.IsValid() {
		param := []reflect.Value{}
		ret := initMethod.Call(param)
		if !ret[0].IsNil() {
			panic("setup failed")
		}
	} else {
		panic("Init Method invalid")
	}
	return instance
}
