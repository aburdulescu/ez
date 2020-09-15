package main

import (
	"encoding/json"
	"io"
	"log"
	"net"

	"github.com/aburdulescu/go-ez/ezs"
	"github.com/aburdulescu/go-ez/ezt"
	"github.com/aburdulescu/go-ez/hash"
	badger "github.com/dgraph-io/badger/v2"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	id   string
	conn net.Conn
	db   badger.DB
}

func (c Client) run(db *badger.DB) {
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

func (c Client) handleGetchunk(index uint64) error {
	chunkHashes, err := c.getChunkHashes()
	if err != nil {
		return err
	}
	rsp := &ezs.Response{
		Type:    ezs.ResponseType_CHUNKHASH,
		Payload: &ezs.Response_Hash{[]byte(chunkHashes[index])},
	}
	if err := c.send(rsp); err != nil {
		return err
	}
	return nil
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
