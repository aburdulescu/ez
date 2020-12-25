package main

import (
	"testing"

	"github.com/aburdulescu/ez/cmn"
)

var checksums = []cmn.Checksum{1, 2, 3, 4, 5}

func BenchmarkNewID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewID(checksums)
	}
}
