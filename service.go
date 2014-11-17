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

type Center struct {
	storage map[string]reflect.Type
	instances []reflect.Value
	mutex sync.Mutex
	id uint16
}

func NewCenter() Center {
	num := 32
	return Center{storage:make(map[string]reflect.Type,num),instances:make([]reflect.Value,num)}
}

func (c *Center) addService(name string,t reflect.Type) {
	if _,ok := c.storage[name]; ok {
		panic("service exists")
	}
	c.storage[name] = t
}

func (c *Center) setup(name string) {
	if _,ok := c.storage[name]; !ok {
		panic("service not exists")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.id >= 0xFF-1 {
		panic("over max id")
	}
	c.id++
	c.instances[c.id] = reflect.New(c.storage[name])
	initMethod := c.instances[c.id].MethodByName("Init")
	if initMethod.IsValid() {
		param := make([]reflect.Value,0)
		ret := initMethod.Call(param)
		if !ret[0].IsNil() {
			panic("setup failed")
		}
	} else {
		panic("Init Method invalid")
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
