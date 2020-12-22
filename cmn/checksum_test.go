package cmn

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
