package main

import (
	"net"
	"testing"

	"github.com/aburdulescu/go-ez/ezs"
	"google.golang.org/protobuf/proto"
)

func TestConnect(t *testing.T) {
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	req := &ezs.Request{
		Type:    ezs.RequestType_CONNECT,
		Payload: &ezs.Request_Id{[]byte("10101010")},
	}
	writeBuf, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write(writeBuf); err != nil {
		t.Fatal(err)
	}
	readBuf := make([]byte, 8192)
	n, err := conn.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	rsp := &ezs.Response{}
	if err := proto.Unmarshal(readBuf[:n], rsp); err != nil {
		t.Fatal(err)
	}
	t.Log(rsp.GetType())
}
