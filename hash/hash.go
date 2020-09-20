package hash

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/zeebo/xxh3"
)

const (
	ALG  = "xxh3"
	SIZE = 8
)

type Hash []byte

func FromChunkHashes(chunkHashes []Hash) Hash {
	var data []byte
	for i := range chunkHashes {
		data = append(data, chunkHashes[i]...)
	}
	return fromUint64(xxh3.Hash(data))
}

func fromUint64(u uint64) []byte {
	h := make(Hash, SIZE)
	binary.BigEndian.PutUint64(h, u)
	return h
}

func FromChunk(chunk []byte) Hash {
	return fromUint64(xxh3.Hash(chunk))
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func (h Hash) Equals(other Hash) bool {
	for i := range h {
		if h[i] != other[i] {
			return false
		}
	}
	return true
}
