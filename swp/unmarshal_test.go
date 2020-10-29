package swp

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	t.Run("EmptyInput", func(t *testing.T) {
		_, err := Unmarshal(nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UnknownMsg", func(t *testing.T) {
		input := []byte{255}
		_, err := Unmarshal(input)
		if err != ErrUnknownMsg {
			t.Fatalf("expected %v, got %v", err, ErrUnknownMsg)
		}
	})
	t.Run("Connect", func(t *testing.T) {
		t.Run("EmptyPayload", func(t *testing.T) {
			input := []byte{byte(CONNECT)}
			_, err := Unmarshal(input)
			if err != ErrEmptyPayload {
				t.Fatalf("expected %v, got %v", err, ErrEmptyPayload)
			}
		})
		t.Run("Good", func(t *testing.T) {
			expected := "id0"
			var input []byte
			input = append(input, byte(CONNECT))
			input = append(input, []byte(expected)...)
			msg, err := Unmarshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if msg.Type() != CONNECT {
				t.Fatal("msg type not CONNECT")
			}
			realMsg := msg.(Connect)
			if realMsg.Id != expected {
				t.Fatalf("expected %v, got %v", expected, realMsg.Id)
			}
		})
	})
	t.Run("Disconnect", func(t *testing.T) {
		t.Run("Good", func(t *testing.T) {
			input := []byte{byte(DISCONNECT)}
			msg, err := Unmarshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if msg.Type() != DISCONNECT {
				t.Fatal("msg type not DISCONNECT")
			}
		})
	})
	t.Run("Getchunk", func(t *testing.T) {
		t.Run("EmptyPayload", func(t *testing.T) {
			input := []byte{byte(GETCHUNK)}
			_, err := Unmarshal(input)
			if err != ErrEmptyPayload {
				t.Fatalf("expected %v, got %v", err, ErrEmptyPayload)
			}
		})
		t.Run("Good", func(t *testing.T) {
			var expected uint64 = 42
			var input []byte
			input = append(input, byte(GETCHUNK))
			indexBuf := make([]byte, 8)
			binary.LittleEndian.PutUint64(indexBuf, expected)
			input = append(input, indexBuf...)
			msg, err := Unmarshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if msg.Type() != GETCHUNK {
				t.Fatal("msg type not GETCHUNK")
			}
			realMsg := msg.(Getchunk)
			if realMsg.Index != expected {
				t.Fatalf("expected %v, got %v", expected, realMsg.Index)
			}
		})
	})
	t.Run("Ack", func(t *testing.T) {
		t.Run("Good", func(t *testing.T) {
			input := []byte{byte(ACK)}
			msg, err := Unmarshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if msg.Type() != ACK {
				t.Fatal("msg type not ACK")
			}
		})
	})
	t.Run("Piece", func(t *testing.T) {
		t.Run("EmptyPayload", func(t *testing.T) {
			input := []byte{byte(PIECE)}
			_, err := Unmarshal(input)
			if err != ErrEmptyPayload {
				t.Fatalf("expected %v, got %v", err, ErrEmptyPayload)
			}
		})
		t.Run("Good", func(t *testing.T) {
			expected := []byte{0, 1, 2, 3, 4, 5}
			var input []byte
			input = append(input, byte(PIECE))
			input = append(input, expected...)
			msg, err := Unmarshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if msg.Type() != PIECE {
				t.Fatal("msg type not PIECE")
			}
			realMsg := msg.(Piece)
			if err := compareByteSlice(realMsg.Piece, expected); err != nil {
				t.Fatal(err)
			}
		})
	})
	t.Run("Chunkhash", func(t *testing.T) {
		t.Run("MissingNPieces", func(t *testing.T) {
			input := []byte{byte(CHUNKHASH), 0, 0}
			_, err := Unmarshal(input)
			if err != ErrPayloadTooSmall {
				t.Fatalf("expected %v, got %v", err, ErrPayloadTooSmall)
			}
		})
		t.Run("MissingHash", func(t *testing.T) {
			input := []byte{byte(CHUNKHASH), 0, 0, 0, 0, 0, 0, 0, 0}
			_, err := Unmarshal(input)
			if err != ErrEmptyPayload {
				t.Fatalf("expected %v, got %v", err, ErrEmptyPayload)
			}
		})
		t.Run("Good", func(t *testing.T) {
			expectedHash := []byte{0, 1, 2, 3, 4, 5}
			var expectedNPieces uint64 = 42
			npiecesBuf := make([]byte, 8)
			binary.LittleEndian.PutUint64(npiecesBuf, 42)
			var input []byte
			input = append(input, byte(CHUNKHASH))
			input = append(input, npiecesBuf...)
			input = append(input, expectedHash...)
			msg, err := Unmarshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if msg.Type() != CHUNKHASH {
				t.Fatal("msg type not CHUNKHASH")
			}
			realMsg := msg.(Chunkhash)
			if realMsg.NPieces != expectedNPieces {
				t.Fatalf("expected %v, got %v", expectedNPieces, realMsg.NPieces)
			}
			if err := compareByteSlice(realMsg.Hash, expectedHash); err != nil {
				t.Fatal(err)
			}
		})
	})
}

func compareByteSlice(l, r []byte) error {
	if len(l) != len(r) {
		return fmt.Errorf("length mismatch: %d != %d", len(l), len(r))
	}
	for i := range l {
		if l[i] != r[i] {
			return fmt.Errorf("mismatch at index %d: %v != %v", i, l[i], r[i])
		}
	}
	return nil
}
