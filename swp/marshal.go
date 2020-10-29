package swp

import (
	"encoding/binary"
	"errors"
)

var ErrBufferTooSmall = errors.New("buffer too small")

// len(b) must be enough for the message
func Marshal(msg Msg, b []byte) error {
	if len(b) < 1 {
		return ErrBufferTooSmall
	}
	b[0] = byte(msg.Type())
	switch msg.Type() {
	case CONNECT:
		realMsg := msg.(Connect)
		id := []byte(realMsg.Id)
		if len(b[1:]) < len(id) {
			return ErrBufferTooSmall
		}
		copy(b[1:], id)
		return nil
	case DISCONNECT:
		return nil
	case GETCHUNK:
		if len(b[1:]) < 8 {
			return ErrBufferTooSmall
		}
		realMsg := msg.(Getchunk)
		binary.LittleEndian.PutUint64(b[1:], realMsg.Index)
		return nil
	case ACK:
		return nil
	case CHUNKHASH:
		if len(b[1:]) < 8 {
			return ErrBufferTooSmall
		}
		realMsg := msg.(Chunkhash)
		binary.LittleEndian.PutUint64(b[1:9], realMsg.NPieces)
		if len(b[9:]) < len(realMsg.Hash) {
			return ErrBufferTooSmall
		}
		copy(b[9:], realMsg.Hash)
		return nil
	case PIECE:
		realMsg := msg.(Piece)
		if len(b[1:]) < len(realMsg.Piece) {
			return ErrBufferTooSmall
		}
		copy(b[1:], realMsg.Piece)
		return nil
	default:
		return ErrUnknownMsg
	}
}
