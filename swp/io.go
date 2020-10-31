package swp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

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

func read(r io.Reader) ([]byte, error) {
	msgsize, err := msgSize(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	src := io.LimitReader(r, int64(msgsize))
	buf := AllocMsgbuf(msgsize)
	n, err := buf.ReadFrom(src)
	if err != nil {
		log.Println(n, err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func write(w io.Writer, b []byte) error {
	const msgMaxSize = (1 << 16) - 1
	if len(b) > msgMaxSize {
		return fmt.Errorf("msg len too big")
	}
	binary.LittleEndian.PutUint16(b[:2], uint16(len(b[2:])))
	buf := bytes.NewBuffer(b)
	n, err := io.Copy(w, buf)
	if err != nil {
		log.Println(buf.Len(), n, err)
		return err
	}
	return nil
}

func Send(w io.Writer, msg Msg) error {
	b := AllocMsgbuf(2 + msg.Size()).Bytes()
	defer ReleaseMsgbuf(b)
	if err := Marshal(msg, b[2:]); err != nil {
		return err
	}
	if err := write(w, b); err != nil {
		return err
	}
	return nil
}

func Recv(r io.Reader) (Msg, func(), error) {
	b, err := read(r)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	msg, err := Unmarshal(b)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	cleanup := func() { ReleaseMsgbuf(b) }
	return msg, cleanup, nil
}
