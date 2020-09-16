package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/aburdulescu/go-ez/ezs"
	"github.com/aburdulescu/go-ez/hash"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	conn net.Conn
}

func (c *Client) Dial(addr string) error {
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c Client) Close() {
	c.conn.Close()
}

func (c Client) Connect(id string) error {
	req := &ezs.Request{
		Type:    ezs.RequestType_CONNECT,
		Payload: &ezs.Request_Id{id},
	}
	rsp, err := c.send(req)
	if err != nil {
		return err
	}
	rspType := rsp.GetType()
	if rspType != ezs.ResponseType_ACK {
		return fmt.Errorf("unexpected response: %s", rspType)
	}
	return nil
}

type GetchunkPart struct {
	piece []byte
	err   error
}

func (c Client) Getchunk(index uint64) (hash.Hash, chan GetchunkPart, error) {
	req := &ezs.Request{
		Type:    ezs.RequestType_GETCHUNK,
		Payload: &ezs.Request_Index{index},
	}
	rsp, err := c.send(req)
	if err != nil {
		return nil, nil, err
	}
	rspType := rsp.GetType()
	if rspType != ezs.ResponseType_CHUNKHASH {
		return nil, nil, fmt.Errorf("unexpected response: %s", rspType)
	}
	ch := make(chan GetchunkPart)
	go c.handleGetchunk(ch)
	return hash.Hash(rsp.GetHash()), ch, nil
}

func (c Client) handleGetchunk(ch chan GetchunkPart) {
	defer close(ch)
	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, c.conn)
	if err != nil {
		ch <- GetchunkPart{nil, err}
		return
	}
	log.Printf("n=%d", n)
	return

	rsp := &ezs.Response{}
	if err := proto.Unmarshal(buf.Bytes(), rsp); err != nil {
		ch <- GetchunkPart{nil, err}
		return

	}

	rspType := rsp.GetType()
	switch rspType {
	case ezs.ResponseType_CHUNKEND:
		return
	case ezs.ResponseType_PIECE:
		ch <- GetchunkPart{rsp.GetPiece(), nil}
	default:
		ch <- GetchunkPart{nil, fmt.Errorf("unexpected response %v", rspType)}
		return
	}
}

func (c Client) send(req *ezs.Request) (*ezs.Response, error) {
	writeBuf, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	if _, err := c.conn.Write(writeBuf); err != nil {
		return nil, err
	}
	readBuf := make([]byte, 8192)
	n, err := c.conn.Read(readBuf)
	if err != nil {
		return nil, err
	}
	rsp := &ezs.Response{}
	if err := proto.Unmarshal(readBuf[:n], rsp); err != nil {
		return nil, err
	}
	return rsp, nil
}
