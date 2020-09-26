package ezs

import (
	"io"
	"log"
)

type MsgBuffer struct {
	buf []byte
}

func NewMsgBuffer(size int) *MsgBuffer {
	return &MsgBuffer{make([]byte, size)}
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
		buf = buf[nread:cap(b.buf)]
	}
}
