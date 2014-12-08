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
	"encoding/json"
	"strings"
	proto "./proto"
	capn "github.com/glycerine/go-capnproto"
)

type Config struct {
	Machines []string
	Log      string
	Addr     string
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

type Remote struct {
	Addr string `json:"Addr"`
}

type Dock struct {
	remote  *map[string]Remote
	storage map[string]Entry
	gmq     chan proto.Msg
	client  *etcd.Client
	harbor  *net.TCPAddr
	handle  *net.TCPListener
}

func NewDock(conf *Config) (*Dock, error) {
	file, err := os.OpenFile(conf.Log,os.O_WRONLY | os.O_APPEND,0666)
	if err != nil {
		panic(err)
	}
	etcd.SetLogger(log.New(file, "[test]", log.LstdFlags|log.Lshortfile))

	client := etcd.NewClient(conf.Machines)
	res, err := client.Get("services", true, false)
	if err != nil {
		if res, err = client.CreateDir("services", 0); err != nil {
			panic(err)
		}
	}
	if !res.Node.Dir {
		return nil, errors.New("service is not a dir")
	}

	addr, err := net.ResolveTCPAddr("tcp", conf.Addr)
	if err != nil {
		return nil, err
	}
	handle, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	d := &Dock{storage: make(map[string]Entry, 32), gmq: make(chan proto.Msg, 100), client: client, harbor: addr, handle: handle }
	r := d.GetRemotes()
	d.remote = &r
	return d,nil
}

func (d *Dock) GetRemotes() map[string]Remote {
	res, err := d.client.Get("services", true, true)
	if err != nil {
		panic(err)
	}
	count := len(res.Node.Nodes)
	remote := make(map[string]Remote,count)
	for i:=0;i<count;i++ {
		name := res.Node.Nodes[i].Key
		value := res.Node.Nodes[i].Value
		var r Remote
		decoder := json.NewDecoder(strings.NewReader(value))
		if err := decoder.Decode(&r); err != nil {
			panic(err)
		}
		remote[name] = r
	}
	return remote
}

//TODO two suggested length is two verbose.
func (d *Dock) AddService(name string, t reflect.Type, mqlen, instance_num int) {
	if _, ok := d.storage[name]; ok {
		panic("service exists")
	}

	r := d.GetRemotes()
	if _,ok := r[name]; ok {
		panic("service exist in remote")
	}
	remote := Remote{Addr:d.harbor.String()}
	b,err := json.Marshal(remote)
	if err != nil {
		panic(err)
	}

	key := strings.Join([]string{"services/",name},"")
	_,err = d.client.Create(key,string(b),10)
	if err != nil {
		panic(err)
	}

	r = d.GetRemotes()
	d.remote = &r

	d.storage[name] = Entry{mq: make(chan proto.Msg, mqlen), items: make([]reflect.Value, instance_num), item_type: t, name: name}

	go d.start(name)
}

// start service
func (d *Dock) start(name string) {
	serv, ok := d.storage[name]
	if !ok {
		panic("service not exists")
	}
	for {
		//TODO better load balance
		cur := atomic.AddInt64(&serv.reqn, 1)
		idx := cur % int64(len(serv.items))
		rv := serv.items[idx]
		if rv.IsValid() {
			rv = d.setup(name)
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

func (d *Dock) setup(name string) reflect.Value {
	if cell, ok := d.storage[name]; !ok {
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

func (d *Dock) dispatch() {
	for {
		m := <-d.gmq
		if serv, ok := d.storage[m.From()]; ok {
			serv.mq <- m
		}
		//TODO have no service named m.dest
	}
}

func (d *Dock) accept() {
	for {
		conn,err := d.handle.AcceptTCP()
		if err != nil {
			//TODO log the handing error
			continue
		}
		go d.readMsg(conn)
	}
}

// Read msg from other dock instance.
func (d *Dock) readMsg(conn *net.TCPConn) {
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
		d.gmq <- msg
	}
}

func (d *Dock) Run() {
	go DefaultDock.accept()
	go DefaultDock.dispatch()
}

var DefaultDock Dock

func init() {
	conf, err := parse("./conf.toml")
	if err != nil {
		panic(err)
	}
	DefaultDock, err := NewDock(conf)
	if err != nil {
		panic(err)
	}
	slog := new(Slog)
	DefaultDock.AddService("slog", reflect.TypeOf(*slog), 10, 1)
	DefaultDock.Run()
}
