package ezs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type ClientIf interface {
	Send(req *Request) error
	Recv() (*Response, error)
}

type ServerIf interface {
	Send(rsp *Response) error
	Recv() (*Request, error)
}

func msgSize(r io.Reader) (int, error) {
	b := make([]byte, 2)
	_, err := io.ReadAtLeast(r, b, 2)
	if err != nil {
		log.Println(err)
		return -1, err
	}
	size := binary.LittleEndian.Uint16(b)
	return int(size), nil
}

func Read(r io.Reader) (*MsgBuffer, error) {
	msgsize, err := msgSize(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	src := io.LimitReader(r, int64(msgsize))
	buf := NewMsgBuffer(msgsize)
	n, err := buf.ReadFrom(src)
	if err != nil {
		log.Println(n, err)
		return nil, err
	}
	return buf, nil
}

func Write(w io.Writer, msg []byte) error {
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
	n, err := io.Copy(w, buf)
	if err != nil {
		log.Println(buf.Len(), n, err)
		return err
	}
	return nil
}
