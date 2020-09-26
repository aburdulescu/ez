package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezs"
	"github.com/aburdulescu/ez/ezt"
	"github.com/aburdulescu/ez/hash"
	badger "github.com/dgraph-io/badger/v2"
	"google.golang.org/protobuf/proto"
)

var chunkPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, chunks.CHUNK_SIZE)
	},
}

type Client struct {
	id   string
	conn net.Conn
	db   *badger.DB
}

func (c Client) run() {
	defer c.conn.Close()
	remAddr := c.conn.RemoteAddr().String()
	for {
		req, err := c.Recv()
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

func (c Client) Send(rsp ezs.Response) error {
	b, err := proto.Marshal(&rsp)
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

func (c Client) Recv() (*ezs.Request, error) {
	buf, err := ezs.Read(c.conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req := &ezs.Request{}
	if err := proto.Unmarshal(buf.Bytes(), req); err != nil {
		log.Println(err)
		return nil, err
	}
	return req, nil
}

func (c Client) handleConnect() error {
	rsp := ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	if err := c.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c Client) handleDisconnect() error {
	rsp := ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	if err := c.Send(rsp); err != nil {
		log.Println(err)
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
	chunkBuf, n, err := readChunk(filepath.Join(ifile.Dir, ifile.Name), index)
	if err != nil {
		log.Println(err)
		return err
	}
	defer chunkPool.Put(chunkBuf)
	chunk := chunkBuf[:n]
	npieces := uint64(len(chunk) / chunks.PIECE_SIZE)
	remainder := uint64(0)
	if len(chunk)%chunks.PIECE_SIZE != 0 {
		npieces++
		remainder = 1
	}
	rsp := ezs.Response{
		Type: ezs.ResponseType_CHUNKHASH,
		Payload: &ezs.Response_Chunkhash{
			&ezs.Chunkhash{
				Hash:    []byte(chunkHashes[index]),
				Npieces: npieces,
			},
		},
	}
	if err := c.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	for i := uint64(0); i < npieces-remainder; i++ {
		piece := chunk[i*chunks.PIECE_SIZE : (i+1)*chunks.PIECE_SIZE]
		rsp := ezs.Piece{Piece: piece}
		b, err := proto.Marshal(&rsp)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := ezs.Write(c.conn, b); err != nil {
			log.Println(err)
			return err
		}
	}
	if remainder != 0 {
		piece := chunk[(len(chunk)-1)*chunks.PIECE_SIZE:]
		rsp := &ezs.Piece{Piece: piece}
		b, err := proto.Marshal(rsp)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := ezs.Write(c.conn, b); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func readChunk(path string, i uint64) ([]byte, int, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}
	defer f.Close()
	r := io.NewSectionReader(f, int64(chunks.CHUNK_SIZE*i), chunks.CHUNK_SIZE)
	b := chunkPool.Get().([]byte)
	n, err := r.Read(b)
	if err != io.EOF && err != nil {
		log.Println(n, err)
		return nil, 0, err
	}
	return b, n, nil
}

func (c Client) handleGetpiece(index uint64) error {
	rsp := ezs.Response{
		Type:    ezs.ResponseType_PIECE,
		Payload: &ezs.Response_Piece{[]byte("piece")},
	}
	if err := c.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c Client) getIFile() (ezt.IFile, error) {
	var i ezt.IFile
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(string(c.id) + ".ifile"))
		if err != nil {
			log.Println(err)
			return err
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := json.Unmarshal(v, &i); err != nil {
			log.Println(err)
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
			log.Println(err)
			return err
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := json.Unmarshal(v, &hashes); err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	return hashes, err
}
