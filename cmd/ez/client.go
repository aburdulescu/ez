package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/aburdulescu/go-ez/chunks"
	"github.com/aburdulescu/go-ez/ezs"
	"github.com/aburdulescu/go-ez/hash"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	conn net.Conn
}

func (c *Client) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
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

func (c Client) Getchunk(index uint64) (*bytes.Buffer, error) {
	req := &ezs.Request{
		Type:    ezs.RequestType_GETCHUNK,
		Payload: &ezs.Request_Index{index},
	}
	rsp, err := c.send(req, true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rspType := rsp.GetType()
	if rspType != ezs.ResponseType_CHUNKHASH {
		return nil, fmt.Errorf("unexpected response: %s", rspType)
	}
	chunkhashMsg := rsp.GetChunkhash()
	npieces := chunkhashMsg.GetNpieces()
	buf := new(bytes.Buffer)
	buf.Grow(chunks.CHUNK_SIZE)
	for i := uint64(0); i < npieces; i++ {
		msgBuf, err := readPbMsg(c.conn)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		rsp := &ezs.Piece{}
		if err := proto.Unmarshal(msgBuf.Bytes(), rsp); err != nil {
			log.Println(err)
			return nil, err
		}
		msgBuf.Release()
		if _, err := buf.Write(rsp.GetPiece()); err != nil {
			return nil, err
		}
	}
	calcChunkHash := hash.FromChunk(buf.Bytes())
	chunkHash := hash.Hash(chunkhashMsg.GetHash())
	if !calcChunkHash.Equals(chunkHash) {
		// TODO: don't return err, retry download from other peer(or maybe the same peer?)
		return nil, fmt.Errorf("hash of chunk %d differs from hash provided by peer", index)
	}
	return buf, nil
}

func getPbMsgSize(c net.Conn) (int, error) {
	b := make([]byte, 2)
	_, err := io.ReadAtLeast(c, b, 2)
	if err != nil {
		return -1, err
	}
	msgsize := binary.LittleEndian.Uint16(b)
	return int(msgsize), nil
}

func readPbMsg(c net.Conn) (*MsgBuffer, error) {
	msgsize, err := getPbMsgSize(c)
	if err != nil {
		return nil, err
	}
	src := io.LimitReader(c, int64(msgsize))
	buf := new(MsgBuffer)
	buf.Alloc(msgsize)
	n, err := buf.ReadFrom(src)
	if err != nil {
		log.Println(n, err)
		return nil, err
	}
	return buf, nil
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
		b, err := readPbMsg(c.conn)
		if err != nil {
			return nil, err
		}
		readBuf = b.Bytes()
		b.Release()
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
