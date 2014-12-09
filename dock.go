package penny

import (
	proto "./proto"
	"github.com/coreos/go-etcd/etcd"
	capn "github.com/glycerine/go-capnproto"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"io"
	"time"
)

type Dock struct {
	remote      map[string]Remote
	local       map[string]Entry
	gmq         chan proto.Msg
	client      *etcd.Client
	harbor_addr *net.TCPAddr
	harbor      *net.TCPListener
}

type Remote struct {
	Addr string `json:"Addr"`
	mutex sync.Mutex `json:"-"`
	conn *net.TCPConn `json:"-"`
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
	go d.accept()

	return &d, nil
}

func (d *Dock) GetRemotes() map[string]Remote {
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

	key := strings.Join([]string{"/services/", name}, "")
	println("key",key)
	if _ , ok := r[key]; ok {
		if r[key].Addr != d.harbor_addr.String() {
			panic("service exist in remote")
		}
	} else {
		remote := new(Remote)
		remote.Addr = d.harbor_addr.String()
		b, err := json.Marshal(remote)
		if err != nil {
			panic(err)
		}

		_, err = d.client.Create(key, string(b), 0)
		if err != nil {
			panic(err)
		}
	}

	serv := Entry{mq: make(chan proto.Msg, mqlen), items: make([]reflect.Value, instance_num),locks:make([]sync.Mutex,instance_num), item_type: t, name: name}

	d.local[key] = serv
	go serv.start()
}

func (d *Dock) Dispatch() {
	for {
		m := <-d.gmq
		if m.Pass() > 100 {
			panic(errors.New("pass > 100"))
		}

		call := m.Dest()
		if serv, ok := d.local[call]; ok {
			println("send to local")
			serv.mq <- m
			continue
		}
		if remote, ok := d.remote[call]; ok {
			println("send to remote")

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

			continue

		}
		println("send to nowhere")
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

	err := conn.SetKeepAlivePeriod(time.Minute * 10)
	if err != nil {
		panic(err)
	}

	buf := make([]byte,1024*10)
	capn_buf := bytes.NewBuffer(make([]byte, 1024*10))

	for {

		i := 0
		for {
			size,err := conn.Read(buf[i:])
			if err == io.EOF {
				break
			} else if err != nil {
				println(err.Error())
				goto END
			} else {
				i += size
			}
		}
		if i == 0 {
			goto END
		}
		println("read size",i)

		cont := bytes.NewReader(buf[:i])
		seg, err := capn.ReadFromStream(cont, capn_buf)
		if err != nil {
			panic(err)
		}
		msg := proto.ReadRootMsg(seg)
		// since readRootMsg doesn't return error,we need test it
		if len(msg.Method()) == 0 {
			panic(errors.New("field method is empty"))
		}
		println("msg detail",msg.From(),msg.Dest())

		d.gmq <- msg
	}
	END:
	println("conn close")
	return
}
