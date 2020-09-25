package chunks

import (
	"bytes"
	"io"
	"math"
	"os"

	"github.com/aburdulescu/ez/hash"
)

const PIECE_SIZE = 8 << 10
const CHUNK_SIZE = PIECE_SIZE << 10

func FromFile(f *os.File, size int64) ([]hash.Hash, error) {
	remainder := size % CHUNK_SIZE
	leftover := int64(0)
	if remainder != 0 {
		leftover = int64(1)
	}
	nchunks := int64(math.Ceil(float64(size) / float64(CHUNK_SIZE)))
	chunks := make([]hash.Hash, nchunks)
	buf := new(bytes.Buffer)
	buf.Grow(CHUNK_SIZE)
	for i := int64(0); i < nchunks-leftover; i++ {
		_, err := io.CopyN(buf, f, CHUNK_SIZE)
		if err != nil {
			return nil, err
		}
		chunks[i] = hash.FromChunk(buf.Bytes())
		buf.Reset()
	}
	if leftover == 1 {
		_, err := io.CopyN(buf, f, remainder)
		if err != nil {
			return nil, err
		}
		chunks[len(chunks)-1] = hash.FromChunk(buf.Bytes())
		buf.Reset()
	}
	return chunks, nil
}
