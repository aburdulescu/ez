package main

import (
	"io"
	"os"
	"sync"
	"testing"

	"github.com/aburdulescu/ez/chunks"
)

func BenchmarkReadChunk(b *testing.B) {
	indexes := []int{0, 1, 2, 3}
	for i := 0; i < b.N; i++ {
		f, err := os.Open("file")
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()
		for _, index := range indexes {
			r := io.NewSectionReader(f, int64(chunks.CHUNK_SIZE*index), chunks.CHUNK_SIZE)
			buf := make([]byte, chunks.CHUNK_SIZE)
			_, err = r.Read(buf)
			if err != io.EOF && err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkReadChunkWithPool(b *testing.B) {
	pool := sync.Pool{
		New: func() interface{} {
			return make([]byte, chunks.CHUNK_SIZE)
		},
	}
	indexes := []int{0, 1, 2, 3}
	for i := 0; i < b.N; i++ {
		f, err := os.Open("file")
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()
		for _, index := range indexes {
			r := io.NewSectionReader(f, int64(chunks.CHUNK_SIZE*index), chunks.CHUNK_SIZE)
			buf := pool.Get().([]byte)
			_, err = r.Read(buf)
			if err != io.EOF && err != nil {
				b.Fatal(err)
			}
			pool.Put(buf)
		}
	}
}
