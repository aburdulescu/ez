package hash

import (
	"crypto/sha256"
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

type Checksum uint64

const ChecksumSize = 8

func NewChecksum(data []byte) Checksum {
	return Checksum(xxh3.Hash(data))
}

type ID []byte

const Alg = "sha256"

func NewID(checksums []Checksum) ID {
	h := sha256.New()
	b := make([]byte, ChecksumSize)
	for i := range checksums {
		binary.BigEndian.PutUint64(b, uint64(checksums[i]))
		h.Write(b)
	}
	return h.Sum(nil)
}

func (id ID) String() string {
	return hex.EncodeToString(id)
}

func (l ID) Equals(r ID) bool {
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}
