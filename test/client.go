package main

import (
	"time"
	proto "../proto"
	capn "github.com/glycerine/go-capnproto"
	"net"
)

func main() {

	addr,err := net.ResolveTCPAddr("tcp","192.168.28.147:5501")
	if err != nil {
		panic(err)
	}
	conn,err := net.DialTCP("tcp",nil,addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	conn.SetKeepAlivePeriod(time.Minute * 10);

	seg := capn.NewBuffer(nil)
	sendMsg := proto.NewRootMsg(seg)
	sendMsg.SetFrom("penny test")
	sendMsg.SetDest("/services/slua")
	sendMsg.SetPass(0)
	sendMsg.SetMethod("test_method")
	params := proto.NewParamList(seg,0)
	sendMsg.SetParams(params)

	size,err := seg.WriteTo(conn)
	if err != nil {
		panic(err)
	}
	println("write bytes",size)
}
