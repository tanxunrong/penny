package penny

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/coreos/go-etcd/etcd"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"bytes"
	"time"
	proto "./proto"
	capn "github.com/glycerine/go-capnproto"
)

type Config struct {
	machines []string
	log      string
	addr     string
}

func parse(file string) (*Config, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var conf Config
	if _, parse_err := toml.Decode(string(content), &conf); parse_err != nil {
		return nil, parse_err
	}
	return &conf, nil
}

type Service interface {
	Name() string
	Init() error
	Call(m proto.Msg) error
	Close()
}

type Entry struct {
	name      string
	item_type reflect.Type
	items     []reflect.Value
	reqn      int64
	mq        chan proto.Msg
	mutex     sync.Mutex
}

type Dock struct {
	storage map[string]Entry
	gmq     chan proto.Msg
	client  *etcd.Client
	harbor  *net.TCPAddr
	handle  *net.TCPListener
}

func NewDock(conf *Config) (*Dock, error) {
	file, err := os.Open(conf.log)
	if err != nil {
		return nil, err
	}
	etcd.SetLogger(log.New(file, "[test]", log.LstdFlags|log.Lshortfile))

	client := etcd.NewClient(conf.machines)
	res, err := client.Get("services", true, true)
	if err != nil {
		if _, cerr := client.CreateDir("services", 0); cerr != nil {
			return nil, err
		}
		res, err = client.Get("services", true, true)
		if err != nil {
			return nil, err
		}
	}
	if !res.Node.Dir {
		return nil, errors.New("service is not a dir")
	}

	addr, err := net.ResolveTCPAddr("tcp", conf.addr)
	if err != nil {
		return nil, err
	}
	handle, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Dock{storage: make(map[string]Entry, 32), gmq: make(chan proto.Msg, 100), client: client, harbor: addr, handle: handle }, nil
}

//TODO two suggested length is two verbose.
func (c *Dock) AddService(name string, t reflect.Type, mqlen, instance_num int) {
	if _, ok := c.storage[name]; ok {
		panic("service exists")
	}

	c.storage[name] = Entry{mq: make(chan proto.Msg, mqlen), items: make([]reflect.Value, instance_num), item_type: t, name: name}
	go c.start(name)
}

// start service
func (c *Dock) start(name string) {
	serv, ok := c.storage[name]
	if !ok {
		panic("service not exists")
	}
	for {
		//TODO better load balance
		cur := atomic.AddInt64(&serv.reqn, 1)
		idx := cur % int64(len(serv.items))
		rv := serv.items[idx]
		if rv.IsNil() {
			rv = c.setup(name)
			serv.items[idx] = rv
		}

		m := <-serv.mq
		go callService(serv, idx, rv, m)
	}
}

func callService(serv Entry, idx int64, rv reflect.Value, m proto.Msg) {
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
		serv.items[int(idx)] = reflect.ValueOf(nil)
	}
}

func (c *Dock) setup(name string) reflect.Value {
	if cell, ok := c.storage[name]; !ok {
		panic("service not exists")
	} else {
		cell.mutex.Lock()
		defer cell.mutex.Unlock()

		instance := reflect.New(cell.item_type)
		initMethod := instance.MethodByName("Init")
		if initMethod.IsValid() {
			param := make([]reflect.Value, 0)
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

func (c *Dock) dispatch() {
	for {
		m := <-c.gmq
		if serv, ok := c.storage[m.From()]; ok {
			serv.mq <- m
		}
		//TODO have no service named m.dest
	}
}

func (c *Dock) accept() {
	for {
		conn,err := c.handle.AcceptTCP()
		if err != nil {
			//TODO log the handing error
			continue
		}
		go c.readMsg(conn)
	}
}

// Read msg from other dock instance.
func (c *Dock) readMsg(conn *net.TCPConn) {
	defer conn.Close()

	err := conn.SetKeepAlivePeriod(time.Hour)
	if err != nil {
		panic(err)
	}

	// 10k read buffer
	buf := bytes.NewBuffer(make([]byte,1024*10))
	for {
		seg,err := capn.ReadFromStream(conn,buf)
		if err != nil {
			panic(err)
		}
		msg := proto.ReadRootMsg(seg)
		// since readRootMsg doesn't return error,we need test it
		if len(msg.Method()) == 0 {
			continue
		}

		// inc pass by 1.so we know if some pkg transfer how many times
		msg.SetPass( msg.Pass() + 1 )
		c.gmq <- msg
	}
}

func (c *Dock) Run() {
	go defaultDock.accept()
	go defaultDock.dispatch()
}

var defaultDock Dock

func init() {
	conf, err := parse("./conf.toml")
	if err != nil {
		panic(err)
	}
	defaultDock, err := NewDock(conf)
	if err != nil {
		panic(err)
	}
	slog := new(Slog)
	slua := new(Slua)
	defaultDock.AddService("slog", reflect.TypeOf(*slog), 10, 1)
	defaultDock.AddService("slua", reflect.TypeOf(*slua), 100, 10)
	defaultDock.Run()
}
