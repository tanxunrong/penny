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
	instances []interface{}
	mutex sync.Mutex
	id uint16
}

func NewCenter() Center {
	return Center{storage:make(map[string]reflect.Type,32),instances:make([]interface{},32)}
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
	if c.id >= 0xFFFF-1 {
		panic("over max id")
	}
	c.id++
	c.instances[c.id] = reflect.New(c.storage[name]).Interface()
}

var defaultCenter Center

func Init() {
	defaultCenter = NewCenter()
	slog := new(Slog)
	slua := new(Slua)
	defaultCenter.addService("slog",reflect.TypeOf(*slog))
	defaultCenter.addService("slua",reflect.TypeOf(*slua))
}
