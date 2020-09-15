package main

import (
	"net"
	"testing"

	"github.com/aburdulescu/go-ez/ezs"
	"google.golang.org/protobuf/proto"
)

func TestConnect(t *testing.T) {
	requests := []*ezs.Request{
		&ezs.Request{
			Type:    ezs.RequestType_CONNECT,
			Payload: &ezs.Request_Id{[]byte("id")},
		},
		&ezs.Request{
			Type:    ezs.RequestType_DISCONNECT,
			Payload: &ezs.Request_Dummy{},
		},
		&ezs.Request{
			Type:    ezs.RequestType_GETCHUNK,
			Payload: &ezs.Request_Index{42},
		},
		&ezs.Request{
			Type:    ezs.RequestType_GETPIECE,
			Payload: &ezs.Request_Index{42},
		},
	}
	for _, req := range requests {
		rsp, err := sendReq(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%v: %v", req.GetType(), rsp.GetType())
	}
}

func sendReq(req *ezs.Request) (*ezs.Response, error) {
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	writeBuf, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(writeBuf); err != nil {
		return nil, err
	}
	readBuf := make([]byte, 8192)
	n, err := conn.Read(readBuf)
	if err != nil {
		return nil, err
	}
	rsp := &ezs.Response{}
	if err := proto.Unmarshal(readBuf[:n], rsp); err != nil {
		return nil, err
	}
	return rsp, nil
}
