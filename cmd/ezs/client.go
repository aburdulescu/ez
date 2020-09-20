package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/aburdulescu/go-ez/chunks"
	"github.com/aburdulescu/go-ez/ezs"
	"github.com/aburdulescu/go-ez/ezt"
	"github.com/aburdulescu/go-ez/hash"
	badger "github.com/dgraph-io/badger/v2"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	id   string
	conn net.Conn
	db   *badger.DB
}

func (c Client) run() {
	defer c.conn.Close()
	remAddr := c.conn.RemoteAddr().String()
	b := make([]byte, 8192)
	for {
		req, err := c.recv(b)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			continue
		}
		reqType := req.GetType()
		switch reqType {
		case ezs.RequestType_CONNECT:
			c.id = req.GetId()
			if err := c.handleConnect(); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_DISCONNECT:
			c.id = ""
			if err := c.handleDisconnect(); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_GETCHUNK:
			if err := c.handleGetchunk(req.GetIndex()); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_GETPIECE:
			if err := c.handleGetpiece(req.GetIndex()); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		default:
			log.Printf("%s: error: unknown request type %v\n", remAddr, reqType)
		}
	}
}

func (c Client) send(rsp *ezs.Response) error {
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(b); err != nil {
		return err
	}
	return nil
}

func (c Client) recv(b []byte) (*ezs.Request, error) {
	n, err := c.conn.Read(b)
	if err != nil {
		return nil, err
	}
	req := &ezs.Request{}
	if err := proto.Unmarshal(b[:n], req); err != nil {
		return nil, err
	}
	return req, nil
}

func (c Client) handleConnect() error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	if err := c.send(rsp); err != nil {
		return err
	}
	return nil
}

func (c Client) handleDisconnect() error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	if err := c.send(rsp); err != nil {
		return err
	}
	return nil
}

func WritePbMsg(c net.Conn, msg []byte) error {
	const msgMaxSize = (1 << 16) - 1
	if len(msg) > msgMaxSize {
		return fmt.Errorf("msg len too big")
	}
	b := make([]byte, 2+len(msg))
	binary.LittleEndian.PutUint16(b, uint16(len(msg)))
	for i := 0; i < len(msg); i++ {
		b[i+2] = msg[i]
	}
	buf := bytes.NewBuffer(b)
	n, err := io.Copy(c, buf)
	if err != nil {
		log.Println(buf.Len(), n, err)
		return err
	}
	return nil
}

func (c Client) handleGetchunk(index uint64) error {
	chunkHashes, err := c.getChunkHashes()
	if err != nil {
		log.Println(err)
		return err
	}
	ifile, err := c.getIFile()
	if err != nil {
		log.Println(err)
		return err
	}
	chunk, err := readChunk(ifile, index)
	if err != nil {
		log.Println(err)
		return err
	}
	npieces := uint64(len(chunk) / chunks.PIECE_SIZE)
	remainder := uint64(0)
	if len(chunk)%chunks.PIECE_SIZE != 0 {
		npieces++
		remainder = 1
	}
	rsp := &ezs.Response{
		Type: ezs.ResponseType_CHUNKHASH,
		Payload: &ezs.Response_Chunkhash{
			&ezs.Chunkhash{
				Hash:    []byte(chunkHashes[index]),
				Npieces: npieces,
			},
		},
	}
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if err := WritePbMsg(c.conn, b); err != nil {
		return err
	}
	for i := uint64(0); i < npieces-remainder; i++ {
		piece := chunk[i*chunks.PIECE_SIZE : (i+1)*chunks.PIECE_SIZE]
		rsp := &ezs.Piece{Piece: piece}
		b, err := proto.Marshal(rsp)
		if err != nil {
			return err
		}
		if err := WritePbMsg(c.conn, b); err != nil {
			log.Println(err)
			return err
		}
	}
	if remainder != 0 {
		piece := chunk[(len(chunk)-1)*chunks.PIECE_SIZE:]
		rsp := &ezs.Piece{Piece: piece}
		b, err := proto.Marshal(rsp)
		if err != nil {
			return err
		}
		if err := WritePbMsg(c.conn, b); err != nil {
			log.Println(err)
			return err
		}
	}
	chunkPool.Put(chunk)
	return nil
}

var chunkPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, chunks.CHUNK_SIZE)
	},
}

func readChunk(ifile ezt.IFile, i uint64) ([]byte, error) {
	f, err := os.Open(filepath.Join(ifile.Dir, ifile.Name))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer f.Close()
	r := io.NewSectionReader(f, int64(chunks.CHUNK_SIZE*i), chunks.CHUNK_SIZE)
	b := chunkPool.Get().([]byte)
	n, err := r.Read(b)
	if err != io.EOF && err != nil {
		log.Println(n, err)
		return nil, err
	}
	return b[:n], nil
}

func (c Client) handleGetpiece(index uint64) error {
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_PIECE,
		Payload: &ezs.Response_Piece{[]byte("piece")},
	}
	if err := c.send(rsp); err != nil {
		return err
	}
	return nil
}

func (c Client) getIFile() (ezt.IFile, error) {
	var i ezt.IFile
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(string(c.id) + ".ifile"))
		if err != nil {
			return err
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(v, &i); err != nil {
			return err
		}
		return nil
	})
	return i, err
}

func (c Client) getChunkHashes() ([]hash.Hash, error) {
	var hashes []hash.Hash
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(string(c.id) + ".chunks"))
		if err != nil {
			return err
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(v, &hashes); err != nil {
			return err
		}
		return nil
	})
	return hashes, err
}
