package penny

import (
	"sync"
	"sync/atomic"
	"reflect"
)

type Service interface {
	Name() string
	Init() error
	Call(m Msg) error
	Close()
}

type Msg struct {
	source string
	dest string
	data []byte
}

type Cell struct {
	name string
	rtype reflect.Type
	rvalue []reflect.Value
	reqn int64
	mq chan Msg
	mutex sync.Mutex
}

type Center struct {
	storage map[string]Cell
	gmq chan Msg
}

func NewCenter() Center {
	return Center{storage:make(map[string]Cell,32),gmq:make(chan Msg,100)}
}

//TODO two suggested length is two verbose.
func (c *Center) addService(name string,t reflect.Type,mqlen,instance_num int) {
	if _,ok := c.storage[name]; ok {
		panic("service exists")
	}
	c.storage[name] = Cell{mq:make(chan Msg,mqlen),rvalue:make([]reflect.Value,instance_num),rtype:t,name:name}
	go c.run(name)
}

// run service
func (c *Center) run(name string) {
	serv,ok := c.storage[name]
	if !ok {
		panic("service not exists")
	}
	for ;; {
		//TODO better load balance
		cur := atomic.AddInt64(&serv.reqn,1)
		idx := cur % int64(len(serv.rvalue))
		rv := serv.rvalue[idx]
		if rv.IsNil() {
			rv = c.setup(name)
			serv.rvalue[idx] = rv
		}

		m := <-serv.mq
		go callService(serv,idx,rv,m)
	}
}

func callService(serv Cell,idx int64,rv reflect.Value,m Msg) {
	//TODO mutex while calling the service instance
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
		//TODO nicer remove instance
		serv.rvalue[int(idx)] = reflect.ValueOf(nil)
	}
}

func (c *Center) setup(name string) reflect.Value {
	if cell,ok := c.storage[name]; !ok {
		panic("service not exists")
	} else {
		cell.mutex.Lock()
		defer cell.mutex.Unlock()

		instance := reflect.New(cell.rtype)
		initMethod := instance.MethodByName("Init")
		if initMethod.IsValid() {
			param := make([]reflect.Value,0)
			ret := initMethod.Call(param)
			if !ret[0].IsNil() {
				panic("setup failed")
			}
		} else {
			panic("Init Method invalid")
		}
		return instance
	}
}

func (c *Center) dispatch() {
	for ;; {
		m := <-c.gmq
		if serv,ok := c.storage[m.dest]; ok {
			serv.mq <-m
		}
		//TODO have no service named m.dest
	}
}

var defaultCenter Center

func init() {
	defaultCenter = NewCenter()
	slog := new(Slog)
	slua := new(Slua)
	defaultCenter.addService("slog",reflect.TypeOf(*slog),10,1)
	defaultCenter.addService("slua",reflect.TypeOf(*slua),100,10)
}
