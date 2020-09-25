package main

import (
	"io"
	"sync"

	"github.com/aburdulescu/ez/chunks"
)

var msgBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, chunks.PIECE_SIZE+3)
	},
}

type MsgBuffer struct {
	buf []byte
}

func (b *MsgBuffer) Alloc(size int) {
	if size == (chunks.PIECE_SIZE + 3) {
		b.buf = msgBufPool.Get().([]byte)
	} else {
		b.buf = make([]byte, size)
	}
}

func (b *MsgBuffer) Release() {
	if len(b.buf) == (chunks.PIECE_SIZE + 3) {
		msgBufPool.Put(b.buf)
	} else {
		b.buf = nil
	}
}

func (b MsgBuffer) Bytes() []byte {
	return b.buf
}

func (b MsgBuffer) ReadFrom(r io.Reader) (int, error) {
	nread := 0
	buf := b.buf
	for {
		n, err := r.Read(buf)
		nread += n
		if err == io.EOF {
			return nread, nil
		}
		if err != nil {
			return nread, err
		}
		if nread == len(b.buf) {
			return nread, nil
		}
		buf = buf[nread:cap(b.buf)]
	}
}
