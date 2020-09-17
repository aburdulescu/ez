package main

import (
	"bytes"
	"encoding/binary"
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
	rsp, err := c.send(req, false)
	if err != nil {
		log.Println(err)
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
	rsp, err := c.send(req, true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	rspType := rsp.GetType()
	if rspType != ezs.ResponseType_CHUNKHASH {
		return nil, nil, fmt.Errorf("unexpected response: %s", rspType)
	}
	chunkhashMsg := rsp.GetChunkhash()
	ch := make(chan GetchunkPart)
	go c.handleGetchunk(chunkhashMsg.GetNpieces(), ch)
	return hash.Hash(chunkhashMsg.GetHash()), ch, nil
}

func ReadPbMsg(c net.Conn) ([]byte, error) {
	b := make([]byte, 2)
	_, err := io.ReadAtLeast(c, b, 2)
	if err != nil {
		return nil, err
	}
	msgsize := binary.LittleEndian.Uint16(b)
	buf := new(bytes.Buffer)
	n, err := io.CopyN(buf, c, int64(msgsize))
	if err != nil {
		log.Println(n, err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c Client) handleGetchunk(npieces uint64, ch chan GetchunkPart) {
	defer close(ch)
	for i := uint64(0); i < npieces; i++ {
		b, err := ReadPbMsg(c.conn)
		if err != nil {
			log.Println(err) // TODO: sometimes the read fails because nread < size specified in header
			ch <- GetchunkPart{nil, err}
			return
		}
		rsp := &ezs.Piece{}
		if err := proto.Unmarshal(b, rsp); err != nil {
			log.Println(err)
			ch <- GetchunkPart{nil, err}
			return
		}
		ch <- GetchunkPart{rsp.GetPiece(), nil}
	}
}

func (c Client) send(req *ezs.Request, streaming bool) (*ezs.Response, error) {
	writeBuf, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	if _, err := c.conn.Write(writeBuf); err != nil {
		return nil, err
	}
	var readBuf []byte
	if streaming {
		b, err := ReadPbMsg(c.conn)
		if err != nil {
			return nil, err
		}
		readBuf = b
	} else {
		b := make([]byte, 8192)
		n, err := c.conn.Read(b)
		if err != nil {
			return nil, err
		}
		readBuf = b[:n]
	}
	rsp := &ezs.Response{}
	if err := proto.Unmarshal(readBuf, rsp); err != nil {
		log.Println(err)
		return nil, err
	}
	return rsp, nil
}
