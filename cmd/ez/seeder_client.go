package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezs"
	"github.com/aburdulescu/ez/hash"
	"google.golang.org/protobuf/proto"
)

var chunkPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, chunks.CHUNK_SIZE)
	},
}

func AllocChunk() []byte {
	return chunkPool.Get().([]byte)
}

func ReleaseChunk(b []byte) {
	chunkPool.Put(b[:0])
}

type SeederClient struct {
	conn net.Conn
}

func DialSeederClient(addr string) (*SeederClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c := &SeederClient{conn: conn}
	return c, nil
}

func (c SeederClient) Close() {
	if c.conn == nil {
		return
	}
	c.conn.Close()
}

func (c SeederClient) Connect(id string) error {
	if c.conn == nil {
		return fmt.Errorf("client was not initialized properly")
	}
	req := ezs.Request{
		Type:    ezs.RequestType_CONNECT,
		Payload: &ezs.Request_Id{id},
	}
	if err := c.Send(req); err != nil {
		log.Println(err)
		return err
	}
	rsp, err := c.Recv()
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

func (c SeederClient) Disconnect() error {
	if c.conn == nil {
		return fmt.Errorf("client was not initialized properly")
	}
	req := ezs.Request{
		Type:    ezs.RequestType_DISCONNECT,
		Payload: &ezs.Request_Dummy{},
	}
	if err := c.Send(req); err != nil {
		log.Println(err)
		return err
	}
	rsp, err := c.Recv()
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

func (c SeederClient) Getchunk(index uint64) (*bytes.Buffer, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client was not initialized properly")
	}
	req := ezs.Request{
		Type:    ezs.RequestType_GETCHUNK,
		Payload: &ezs.Request_Index{index},
	}
	if err := c.Send(req); err != nil {
		log.Println(err)
		return nil, err
	}
	rsp, err := c.Recv()
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
	buf := bytes.NewBuffer(AllocChunk())
	for i := uint64(0); i < npieces; i++ {
		b, err := ezs.Read(c.conn)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		rsp := ezs.Piece{}
		if err := proto.Unmarshal(b, &rsp); err != nil {
			log.Println(err)
			return nil, err
		}
		if _, err := buf.Write(rsp.GetPiece()); err != nil {
			return nil, err
		}
		ezs.ReleaseMsg(b)
	}
	calcChunkHash := hash.FromChunk(buf.Bytes())
	chunkHash := hash.Hash(chunkhashMsg.GetHash())
	if !calcChunkHash.Equals(chunkHash) {
		// TODO: don't return err, retry download from other peer(or maybe the same peer?)
		return nil, fmt.Errorf("hash of chunk %d differs from hash provided by peer", index)
	}
	return buf, nil
}

func (c SeederClient) Send(req ezs.Request) error {
	if c.conn == nil {
		return fmt.Errorf("client was not initialized properly")
	}
	b, err := proto.Marshal(&req)
	if err != nil {
		log.Println(err)
		return err
	}
	if err := ezs.Write(c.conn, b); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c SeederClient) Recv() (*ezs.Response, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client was not initialized properly")
	}
	b, err := ezs.Read(c.conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rsp := &ezs.Response{}
	if err := proto.Unmarshal(b, rsp); err != nil {
		log.Println(err)
		return nil, err
	}
	ezs.ReleaseMsg(b)
	return rsp, nil
}
