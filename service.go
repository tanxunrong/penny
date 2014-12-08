package penny

import (
	proto "./proto"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/coreos/go-etcd/etcd"
	capn "github.com/glycerine/go-capnproto"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// any service should implement this interface.
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
	locks	  []sync.Mutex
	reqn      int64
	mq        chan proto.Msg
	mutex     sync.Mutex
}

type Remote struct {
	Addr string `json:"Addr"`
	mutex sync.Mutex `json:"-"`
	conn *net.TCPConn `json:"-"`
}

type Dock struct {
	remote      map[string]Remote
	local       map[string]Entry
	gmq         chan proto.Msg
	client      *etcd.Client
	harbor_addr *net.TCPAddr
	harbor      *net.TCPListener
}

func NewDock(conf *Config) (*Dock, error) {
	// init the logger
	file, err := os.OpenFile(conf.Log, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	etcd.SetLogger(log.New(file, "[test]", log.LstdFlags|log.Lshortfile))

	// init the etcd client
	client := etcd.NewClient(conf.Machines)
	// create services dir
	res, err := client.Get("services", true, false)
	if err != nil {
		if res, err = client.CreateDir("services", 0); err != nil {
			return nil, err
		}
	}
	if !res.Node.Dir {
		return nil, errors.New("service is not a dir")
	}
	//TODO remove this check
	// 100 services is too many
	if len(res.Node.Nodes) > 100 {
		return nil, errors.New("remote services more than 100")
	}

	// bind local addr
	addr, err := net.ResolveTCPAddr("tcp", conf.Addr)
	if err != nil {
		return nil, err
	}
	harbor, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	d := Dock{ remote: make(map[string]Remote, 100), local: make(map[string]Entry, 32), gmq: make(chan proto.Msg, 100), client: client, harbor_addr: addr, harbor: harbor}
	return &d, nil
}

func (d *Dock) GetRemotes() map[string]Remote {
	println("client %v",d.client)
	if d.client == nil {
		panic(errors.New("client empty"))
	}

	res, err := d.client.Get("services", true, false)
	if err != nil {
		panic(err)
	}
	count := len(res.Node.Nodes)
	remote := make(map[string]Remote, count)
	for i := 0; i < count; i++ {
		name := res.Node.Nodes[i].Key
		value := res.Node.Nodes[i].Value
		r := new(Remote)
		decoder := json.NewDecoder(strings.NewReader(value))
		if err := decoder.Decode(r); err != nil {
			panic(err)
		}
		remote[name] = *r
	}
	return remote
}

//TODO two suggested length is two verbose.
func (d *Dock) AddService(name string, t reflect.Type, mqlen, instance_num int) {
	if _, ok := d.local[name]; ok {
		panic("service exists")
	}

	r := d.GetRemotes()
	if _, ok := r[name]; ok {
		panic("service exist in remote")
	}
	remote := new(Remote)
	remote.Addr = d.harbor_addr.String()
	b, err := json.Marshal(remote)
	if err != nil {
		panic(err)
	}

	key := strings.Join([]string{"services/", name}, "")
	_, err = d.client.Create(key, string(b), 0)
	if err != nil {
		panic(err)
	}

	d.remote = d.GetRemotes()

	serv := Entry{mq: make(chan proto.Msg, mqlen), items: make([]reflect.Value, instance_num),locks:make([]sync.Mutex,instance_num), item_type: t, name: name}

	d.local[name] = serv
	go serv.start()
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

func (d *Dock) dispatch() {
	for {
		m := <-d.gmq
		call := m.From()
		if serv, ok := d.local[call]; ok {
			serv.mq <- m
		}
		if remote, ok := d.remote[call]; ok {

			remote.mutex.Lock()
			if remote.conn == nil {
				addr,err := net.ResolveTCPAddr("tcp",remote.Addr)
				if err != nil {
					panic(err)
				}
				remote.conn,err = net.DialTCP("tcp",nil,addr)
				if err != nil {
					panic(err)
				}
				remote.conn.SetKeepAlivePeriod(time.Minute * 10)
			}
			remote.mutex.Unlock()

			seg := capn.NewBuffer(nil)
			sendMsg := proto.NewRootMsg(seg)
			sendMsg.SetFrom(m.From())
			sendMsg.SetDest(m.Dest())
			sendMsg.SetPass(m.Pass())
			sendMsg.SetMethod(m.Method())
			sendMsg.SetParams(m.Params())

			if _,err := seg.WriteTo(remote.conn);err != nil {
				panic(err)
			}

		}
	}
}

func (d *Dock) accept() {
	for {
		conn, err := d.harbor.AcceptTCP()
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
	buf := bytes.NewBuffer(make([]byte, 1024*10))
	for {
		seg, err := capn.ReadFromStream(conn, buf)
		if err != nil {
			panic(err)
		}
		msg := proto.ReadRootMsg(seg)
		// since readRootMsg doesn't return error,we need test it
		if len(msg.Method()) == 0 {
			panic(errors.New("field method is empty"))
		}

		// inc pass by 1.so we know if some pkg transfer how many times
		msg.SetPass(msg.Pass() + 1)
		d.gmq <- msg
	}
}

// the default dock.
// usually one dock is enough.
var DefaultDock Dock

// start the default dock.
// path is the config file.
func Run(path string) {
	conf := ParseConfig(path)
	DefaultDock, err := NewDock(conf)
	if err != nil {
		panic(err)
	}

	go DefaultDock.accept()
	go DefaultDock.dispatch()
}
