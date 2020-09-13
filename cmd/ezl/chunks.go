package main

import (
	"crypto/sha1"
	"io"
	"math"
	"os"
)

const PIECE_SIZE = 8 << 10
const CHUNK_SIZE = PIECE_SIZE << 10

func ChunksFromFile(f *os.File, size int64) ([]Hash, error) {
	remainder := size % CHUNK_SIZE
	leftover := int64(0)
	if remainder != 0 {
		leftover = int64(1)
	}
	nchunks := int64(math.Ceil(float64(size) / float64(CHUNK_SIZE)))
	chunks := make([]Hash, nchunks)
	h := sha1.New()
	for i := int64(0); i < nchunks-leftover; i++ {
		_, err := io.CopyN(h, f, CHUNK_SIZE)
		if err != nil {
			return nil, err
		}
		chunks[i] = h.Sum(nil)
		h.Reset()
	}
	if leftover == 1 {
		_, err := io.CopyN(h, f, remainder)
		if err != nil {
			return nil, err
		}
		chunks[len(chunks)-1] = h.Sum(nil)
		h.Reset()
	}
	return chunks, nil
}
