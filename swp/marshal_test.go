package swp

import (
	"encoding/binary"
	"testing"
)

type UnknownMsg struct{}

func (r UnknownMsg) Type() MsgType { return MsgType(255) }

func TestMarshal(t *testing.T) {
	t.Run("EmptyBuffer", func(t *testing.T) {
		if err := Marshal(nil, nil); err != ErrBufferTooSmall {
			t.Fatalf("expected %v, got %v", ErrBufferTooSmall, err)
		}
	})
	t.Run("UnknownMsg", func(t *testing.T) {
		b := make([]byte, 1)
		if err := Marshal(UnknownMsg{}, b); err != ErrUnknownMsg {
			t.Fatalf("expected %v, got %v", ErrUnknownMsg, err)
		}
	})
	t.Run("Disconnect", func(t *testing.T) {
		b := make([]byte, 1)
		if err := Marshal(Disconnect{}, b); err != nil {
			t.Fatal(err)
		}
		msgType := MsgType(b[0])
		if msgType != DISCONNECT {
			t.Fatal("msg type not DISCONNECT")
		}
	})
	t.Run("Ack", func(t *testing.T) {
		b := make([]byte, 1)
		if err := Marshal(Ack{}, b); err != nil {
			t.Fatal(err)
		}
		msgType := MsgType(b[0])
		if msgType != ACK {
			t.Fatal("msg type not ACK")
		}
	})
	t.Run("Connect", func(t *testing.T) {
		t.Run("BufferTooSmall", func(t *testing.T) {
			b := make([]byte, 1)
			if err := Marshal(Connect{"abc"}, b); err != ErrBufferTooSmall {
				t.Fatalf("expected %v, got %v", ErrBufferTooSmall, err)
			}
		})
		t.Run("Good", func(t *testing.T) {
			expectedId := "id0"
			b := make([]byte, 1+len(expectedId))
			if err := Marshal(Connect{expectedId}, b); err != nil {
				t.Fatal(err)
			}
			msgType := MsgType(b[0])
			if msgType != CONNECT {
				t.Fatal("msg type not CONNECT")
			}
			id := string(b[1:])
			if id != expectedId {
				t.Fatalf("expected %v, got %v", expectedId, id)
			}
		})
	})
	t.Run("Getchunk", func(t *testing.T) {
		t.Run("BufferTooSmall", func(t *testing.T) {
			b := make([]byte, 1)
			if err := Marshal(Getchunk{42}, b); err != ErrBufferTooSmall {
				t.Fatalf("expected %v, got %v", ErrBufferTooSmall, err)
			}
		})
		t.Run("Good", func(t *testing.T) {
			var expectedIndex uint64 = 42
			b := make([]byte, 1+8)
			if err := Marshal(Getchunk{expectedIndex}, b); err != nil {
				t.Fatal(err)
			}
			msgType := MsgType(b[0])
			if msgType != GETCHUNK {
				t.Fatal("msg type not GETCHUNK")
			}
			index := binary.LittleEndian.Uint64(b[1:])
			if index != expectedIndex {
				t.Fatalf("expected %v, got %v", expectedIndex, index)
			}
		})
	})
	t.Run("Piece", func(t *testing.T) {
		t.Run("BufferTooSmall", func(t *testing.T) {
			b := make([]byte, 1)
			if err := Marshal(Piece{[]byte{0, 1, 2, 3}}, b); err != ErrBufferTooSmall {
				t.Fatalf("expected %v, got %v", ErrBufferTooSmall, err)
			}
		})
		t.Run("Good", func(t *testing.T) {
			expectedPiece := []byte{0, 1, 2, 3}
			b := make([]byte, 1+len(expectedPiece))
			if err := Marshal(Piece{expectedPiece}, b); err != nil {
				t.Fatal(err)
			}
			msgType := MsgType(b[0])
			if msgType != PIECE {
				t.Fatal("msg type not PIECE")
			}
			piece := b[1:]
			if err := compareByteSlice(piece, expectedPiece); err != nil {
				t.Fatal(err)
			}
		})
	})
	t.Run("Chunkhash", func(t *testing.T) {
		t.Run("BufferTooSmall1", func(t *testing.T) {
			b := make([]byte, 1)
			if err := Marshal(Chunkhash{42, nil}, b); err != ErrBufferTooSmall {
				t.Fatalf("expected %v, got %v", ErrBufferTooSmall, err)
			}
		})
		t.Run("BufferTooSmall2", func(t *testing.T) {
			b := make([]byte, 1+8)
			if err := Marshal(Chunkhash{42, []byte{0, 1, 2, 3}}, b); err != ErrBufferTooSmall {
				t.Fatalf("expected %v, got %v", ErrBufferTooSmall, err)
			}
		})
		t.Run("Good", func(t *testing.T) {
			var expectedNPieces uint64 = 42
			expectedHash := []byte{0, 1, 2, 3}
			b := make([]byte, 1+8+len(expectedHash))
			if err := Marshal(Chunkhash{expectedNPieces, expectedHash}, b); err != nil {
				t.Fatal(err)
			}
			msgType := MsgType(b[0])
			if msgType != CHUNKHASH {
				t.Fatal("msg type not CHUNKHASH")
			}
			npieces := binary.LittleEndian.Uint64(b[1:9])
			if npieces != expectedNPieces {
				t.Fatalf("expected %v, got %v", expectedNPieces, npieces)
			}
			hash := b[9:]
			if err := compareByteSlice(hash, expectedHash); err != nil {
				t.Fatal(err)
			}
		})
	})
}
