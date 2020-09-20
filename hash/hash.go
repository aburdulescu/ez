package hash

import (
	"strconv"

	"github.com/zeebo/xxh3"
)

const HASH_ALG = "xxh3"

type Hash uint64

func FromChunkHashes(chunkHashes []Hash) Hash {
	var data []byte
	for i := range chunkHashes {
		data = append(data, []byte(chunkHashes[i].String())...)
	}
	return Hash(xxh3.Hash(data))
}

func FromChunk(chunk []byte) Hash {
	return Hash(xxh3.Hash(chunk))
}

func (h Hash) String() string {
	return strconv.FormatUint(uint64(h), 10)
}

func (h Hash) Equals(other Hash) bool {
	return (h == other)
}
