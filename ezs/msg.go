package ezs

import (
	"io"
	"log"
	"sync"

	"github.com/aburdulescu/ez/chunks"
)

const EXTRA_PB_MSG_SIZE = 10
const POOL_BUF_SIZE = chunks.PIECE_SIZE + EXTRA_PB_MSG_SIZE

var msgPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, POOL_BUF_SIZE)
	},
}

func AllocMsg(size int) MsgBuffer {
	if size >= chunks.PIECE_SIZE {
		b := msgPool.Get().([]byte)
		return MsgBuffer{b[:size]}
	} else {
		return NewMsgBuffer(size)
	}
}

func ReleaseMsg(b []byte) {
	if len(b) >= chunks.PIECE_SIZE {
		msgPool.Put(b[:POOL_BUF_SIZE])
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
