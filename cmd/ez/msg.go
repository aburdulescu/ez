package main

import (
	"io"
)

type MsgBuffer struct {
	buf []byte
}

func NewMsgBuffer(size int) MsgBuffer {
	return MsgBuffer{make([]byte, size)}
}

func (p MsgBuffer) Bytes() []byte {
	return p.buf
}

func (p MsgBuffer) ReadFrom(r io.Reader) (int, error) {
	nread := 0
	b := p.buf
	for {
		n, err := r.Read(b)
		nread += n
		if err == io.EOF {
			return nread, nil
		}
		if err != nil {
			return nread, err
		}
		if nread == len(p.buf) {
			return nread, nil
		}
		b = b[nread:cap(p.buf)]
	}
}
