package swp

import (
	"encoding/binary"
	"errors"
)

var ErrUnknownMsg = errors.New("unknown msg")
var ErrEmptyInput = errors.New("empty input")
var ErrEmptyPayload = errors.New("empty payload")
var ErrPayloadTooSmall = errors.New("payload too small")

// b0 = MsgType
// b1, ... = Payload
// Call Type method on returned Msg and convert it to appropriate impl
// Ex:
//   if msg.Type == CONNECT {
//      connect := msg.(Connect)
//      fmt.Println(connect.Id)
//   }
func Unmarshal(b []byte) (Msg, error) {
	if b == nil {
		return nil, ErrEmptyInput
	}
	msgType := MsgType(b[0])
	payload := b[1:]
	switch msgType {
	case CONNECT:
		if len(payload) == 0 {
			return nil, ErrEmptyPayload
		}
		id := string(payload)
		return Connect{id}, nil
	case DISCONNECT:
		return Disconnect{}, nil
	case GETCHUNK:
		if len(payload) == 0 {
			return nil, ErrEmptyPayload
		}
		index := binary.LittleEndian.Uint64(payload)
		return Getchunk{index}, nil
	case ACK:
		return Ack{}, nil
	case CHUNKINFO:
		if len(payload) < 8 {
			return nil, ErrPayloadTooSmall
		}
		npieces := binary.LittleEndian.Uint64(payload[:8])
		if len(payload[8:]) < 8 {
			return nil, ErrPayloadTooSmall
		}
		checksum := binary.LittleEndian.Uint64(payload[8:])
		return Chunkinfo{NPieces: npieces, Checksum: checksum}, nil
	case PIECE:
		if len(payload) == 0 {
			return nil, ErrEmptyPayload
		}
		piece := payload
		return Piece{Piece: piece}, nil
	default:
		return nil, ErrUnknownMsg
	}
}
