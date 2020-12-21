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

func BenchmarkChecksum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewChecksum(chunk)
	}
}

var checksums = []Checksum{1, 2, 3, 4, 5}

func BenchmarkNewID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewID(checksums)
	}
}
