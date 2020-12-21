package hash

import (
	"math/rand"
	"testing"
	"time"
)

func makeRand(size uint64) []byte {
	b := make([]byte, size)
	time.Now().UTC().UnixNano()
	rand.Read(b)
	return b
}

var chunk = makeRand(16 << 20)

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FromChunk(chunk)
	}
}

func BenchmarkChecksum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewChecksum(chunk)
	}
}

var chunkHashes = []Hash{
	Hash("0000"),
	Hash("0001"),
	Hash("0002"),
	Hash("0003"),
	Hash("0004"),
}

var checksums = []Checksum{1, 2, 3, 4, 5}

func BenchmarkNewIDOld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FromChunkHashes(chunkHashes)
	}
}

func BenchmarkNewID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewID(checksums)
	}
}
