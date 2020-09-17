package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/aburdulescu/go-ez/stream/rpc"
	"google.golang.org/protobuf/proto"
)

const PIECE_SIZE = 8192
const CHUNK_SIZE = 1024 * PIECE_SIZE

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	client := Client{}
	if err := client.Dial(":28080"); err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	buf := new(bytes.Buffer)
	nchunks := (100 << 20) / CHUNK_SIZE // TODO: check remainder
	if (100<<20)%CHUNK_SIZE != 0 {
		nchunks++
	}
	for i := 0; i < nchunks; i++ {
		ch, err := client.SendRequest(uint64(i + 1))
		if err != nil {
			log.Fatal(err)
		}
		for part := range ch {
			if part.err != nil {
				log.Fatal(part.err)
			}
			if _, err := buf.Write(part.piece); err != nil {
				log.Fatal(err)
			}
		}
		log.Println(i, buf.Len())
	}
	f, err := os.Create("f100MB")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, buf); err != nil {
		log.Fatal(err)
	}
}

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

type ResponsePart struct {
	piece []byte
	err   error
}

func (c Client) SendRequest(index uint64) (chan ResponsePart, error) {
	req := &rpc.Request{Index: index}
	rsp, err := c.send(req)
	if err != nil {
		return nil, err
	}
	ch := make(chan ResponsePart)
	go c.handleResponse(ch, rsp.GetNpieces())
	return ch, nil
}

func (c Client) handleResponse(ch chan ResponsePart, npieces uint64) {
	defer close(ch)
	for i := uint64(0); i < npieces; i++ {
		b, err := ReadPbMsg(c.conn)
		if err != nil {
			log.Println(err)
			ch <- ResponsePart{nil, err}
			return
		}
		rsp := &rpc.Piece{}
		if err := proto.Unmarshal(b, rsp); err != nil {
			log.Println(err)
			ch <- ResponsePart{nil, err}
			return
		}
		ch <- ResponsePart{rsp.GetPiece(), nil}
	}
}

func (c Client) send(req *rpc.Request) (*rpc.Response, error) {
	writeBuf, err := proto.Marshal(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if _, err := c.conn.Write(writeBuf); err != nil {
		log.Println(err)
		return nil, err
	}
	readBuf, err := ReadPbMsg(c.conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rsp := &rpc.Response{}
	if err := proto.Unmarshal(readBuf, rsp); err != nil {
		log.Println(err)
		return nil, err
	}
	return rsp, nil
}

func ReadPbMsg(c net.Conn) ([]byte, error) {
	b := make([]byte, 2)
	n, err := io.ReadAtLeast(c, b, 2)
	if err != nil {
		return nil, err
	}
	msgsize := binary.LittleEndian.Uint16(b)
	dataBuf := make([]byte, msgsize)
	n, err = c.Read(dataBuf)
	if err != nil {
		return nil, err
	}
	if uint16(n) != msgsize {
		return nil, fmt.Errorf("read less than expected: n=%d, msgsize=%d", n, msgsize)
	}
	return dataBuf, nil
}
