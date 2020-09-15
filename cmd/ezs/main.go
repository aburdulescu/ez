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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleClient(conn)
	}
}

type Client struct {
	id   []byte
	conn net.Conn
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	c := Client{
		conn: conn,
	}
	b := make([]byte, 8192)
	for {
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
			c.id = req.GetId()
			if err := c.handleConnect(req.GetId()); err != nil {
				log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
			}
		case ezs.RequestType_DISCONNECT:
			c.id = nil
			if err := c.handleDisconnect(); err != nil {
				log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
			}
		case ezs.RequestType_GETCHUNK:
			if err := c.handleGetchunk(req.GetIndex()); err != nil {
				log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
			}
		case ezs.RequestType_GETPIECE:
			if err := c.handleGetpiece(req.GetIndex()); err != nil {
				log.Printf("%s: error: %v\n", conn.RemoteAddr(), err)
			}
		default:
			log.Printf("%s: error: unknown request type %v\n", conn.RemoteAddr(), reqType)
		}
	}
}

func (c Client) handleConnect(id []byte) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(b); err != nil {
		return err
	}
	return nil
}

func (c Client) handleDisconnect() error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(b); err != nil {
		return err
	}
	return nil
}

func (c Client) handleGetchunk(index uint64) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_CHUNKHASH,
		Payload: &ezs.Response_Hash{[]byte("chankhash")},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(b); err != nil {
		return err
	}
	return nil
}

func (c Client) handleGetpiece(index uint64) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_PIECE,
		Payload: &ezs.Response_Piece{[]byte("piece")},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(b); err != nil {
		return err
	}
	return nil
}
