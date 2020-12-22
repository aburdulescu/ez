package main

import (
	"bytes"
	"io"
	"math"
	"os"

	"github.com/aburdulescu/ez/cmn"
)

func ProcessFile(f *os.File, size int64) ([]cmn.Checksum, error) {
	remainder := size % cmn.ChunkSize
	leftover := int64(0)
	if remainder != 0 {
		leftover = int64(1)
	}
	nchunks := int64(math.Ceil(float64(size) / float64(cmn.ChunkSize)))
	checksums := make([]cmn.Checksum, nchunks)
	buf := new(bytes.Buffer)
	buf.Grow(cmn.ChunkSize)
	for i := int64(0); i < nchunks-leftover; i++ {
		_, err := io.CopyN(buf, f, cmn.ChunkSize)
		if err != nil {
			return nil, err
		}
		checksums[i] = cmn.NewChecksum(buf.Bytes())
		buf.Reset()
	}
	if leftover == 1 {
		_, err := io.CopyN(buf, f, remainder)
		if err != nil {
			return nil, err
		}
		checksums[len(checksums)-1] = cmn.NewChecksum(buf.Bytes())
		buf.Reset()
	}
	return checksums, nil
}
