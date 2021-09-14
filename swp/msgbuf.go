package swp

import (
	"io"
	"log"
	"sync"

	"github.com/aburdulescu/ez/cmn"
)

const EXTRA_MEMORY = 16
const POOL_BUF_SIZE = cmn.PieceSize + EXTRA_MEMORY

var msgbufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, POOL_BUF_SIZE)
	},
}

func AllocMsgbuf(size int) MsgBuffer {
	if size >= cmn.PieceSize {
		b := msgbufPool.Get().([]byte)
		return MsgBuffer{b[:size]}
	} else {
		return NewMsgBuffer(size)
	}
}

func ReleaseMsgbuf(b []byte) {
	if len(b) >= cmn.PieceSize {
		msgbufPool.Put(b[:POOL_BUF_SIZE]) // TODO: PERF: this causes GC activity
	}
}

type MsgBuffer struct {
	buf []byte
}

func NewMsgBuffer(size int) MsgBuffer {
	return MsgBuffer{make([]byte, size)}
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
			log.Println(err)
			return nread, err
		}
		if nread == len(b.buf) {
			return nread, nil
		}
		buf = b.buf[nread:]
	}
}
