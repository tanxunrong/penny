package penny

import (
	"sync"
	"reflect"
)

type Service interface {
	Name() string
	Init() error
	Call(m Msg) error
	Close()
}

type Msg struct {
	source uint32
	session int
	data []byte
}

type Cell struct {
	name string
	rtype reflect.Type
	rvalue []reflect.Value
	mq chan Msg
	mutex sync.Mutex
}

type Center struct {
	storage map[string]Cell
}

func NewCenter() Center {
	num := 32
	return Center{storage:make(map[string]Cell,num)}
}

func (c *Center) addService(name string,t reflect.Type) {
	if _,ok := c.storage[name]; ok {
		panic("service exists")
	}
	c.storage[name] = Cell{mq:make(chan Msg,10),rvalue:make([]reflect.Value,10),rtype:t,name:name}
}

func (c *Center) setup(name string) {
	if cell,ok := c.storage[name]; !ok {
		panic("service not exists")
	} else {
		cell.mutex.Lock()
		defer cell.mutex.Lock()

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
		cell.rvalue = append(cell.rvalue,instance)
	}
}

var defaultCenter Center

func init() {
	defaultCenter = NewCenter()
	slog := new(Slog)
	slua := new(Slua)
	defaultCenter.addService("slog",reflect.TypeOf(*slog))
	defaultCenter.addService("slua",reflect.TypeOf(*slua))
	defaultCenter.setup("slog")
	defaultCenter.setup("slua")
}
