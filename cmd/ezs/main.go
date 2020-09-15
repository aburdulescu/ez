package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/aburdulescu/go-ez/ezs"
	"google.golang.org/protobuf/proto"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		handleErr(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	b := make([]byte, 8192)
	n, err := conn.Read(b)
	if err == io.EOF {
		log.Printf("%s: closed the connection\n", conn.RemoteAddr())
		return
	}
	if err != nil {
		log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
	}
	req := &ezs.Request{}
	if err := proto.Unmarshal(b[:n], req); err != nil {
		log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
	}
	reqType := req.GetType()
	switch reqType {
	case ezs.RequestType_CONNECT:
		if err := handleConnect(conn, req.GetId()); err != nil {
			log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
		}
	case ezs.RequestType_DISCONNECT:
		if err := handleDisconnect(conn); err != nil {
			log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
		}

	case ezs.RequestType_GETCHUNK:
		if err := handleGetchunk(conn, req.GetIndex()); err != nil {
			log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
		}
	case ezs.RequestType_GETPIECE:
		if err := handleGetpiece(conn, req.GetIndex()); err != nil {
			log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
		}
	default:
		log.Printf("%s: error: unknown request type %v\n", conn.RemoteAddr(), reqType)
	}
}

func handleConnect(conn net.Conn, id []byte) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}

func handleDisconnect(conn net.Conn) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}

func handleGetchunk(conn net.Conn, index uint64) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_CHUNKHASH,
		Payload: &ezs.Response_Hash{[]byte("chankhash")},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}

func handleGetpiece(conn net.Conn, index uint64) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_PIECE,
		Payload: &ezs.Response_Piece{[]byte("piece")},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}
