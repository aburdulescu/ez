package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/aburdulescu/ez/cmn"
	"github.com/aburdulescu/ez/swp"
)

var chunkPool = sync.Pool{
	New: func() interface{} {
		buf := new(bytes.Buffer)
		buf.Grow(cmn.ChunkSize)
		return buf
	},
}

func AllocChunk() *bytes.Buffer {
	return chunkPool.Get().(*bytes.Buffer)
}

func ReleaseChunk(b *bytes.Buffer) {
	b.Reset()
	chunkPool.Put(b)
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
	if err := swp.Send(c.conn, swp.Connect{id}); err != nil {
		log.Println(err)
		return err
	}
	rsp, cleanup, err := swp.Recv(c.conn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer cleanup()
	rspType := rsp.Type()
	if rspType != swp.ACK {
		return fmt.Errorf("unexpected response: %s", rspType)
	}
	return nil
}

func (c SeederClient) Disconnect() error {
	if c.conn == nil {
		return fmt.Errorf("client was not initialized properly")
	}
	if err := swp.Send(c.conn, swp.Disconnect{}); err != nil {
		log.Println(err)
		return err
	}
	rsp, cleanup, err := swp.Recv(c.conn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer cleanup()
	rspType := rsp.Type()
	if rspType != swp.DISCONNECT {
		return fmt.Errorf("unexpected response: %s", rspType)
	}
	return nil
}

func (c SeederClient) Getchunk(index uint64) (*bytes.Buffer, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client was not initialized properly")
	}
	if err := swp.Send(c.conn, swp.Getchunk{index}); err != nil {
		log.Println(err)
		return nil, err
	}
	rsp, cleanup, err := swp.Recv(c.conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cleanup()
	rspType := rsp.Type()
	if rspType != swp.CHUNKINFO {
		return nil, fmt.Errorf("unexpected response: %s", rspType)
	}
	chunkinfoMsg := rsp.(swp.Chunkinfo)
	npieces := chunkinfoMsg.NPieces
	buf := AllocChunk()
	for i := uint64(0); i < npieces; i++ {
		rsp, cleanup, err := swp.Recv(c.conn)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		pieceMsg := rsp.(swp.Piece)
		if _, err := buf.Write(pieceMsg.Piece); err != nil {
			cleanup()
			return nil, err
		}
		cleanup()
	}
	calcChecksum := cmn.NewChecksum(buf.Bytes())
	checksum := cmn.Checksum(chunkinfoMsg.Checksum)
	if calcChecksum != checksum {
		// TODO: don't return err, retry download from other peer(or maybe the same peer?)
		return nil, fmt.Errorf("checksum of chunk %d differs from checksum provided by peer", index)
	}
	return buf, nil
}
