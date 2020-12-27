package cmn

import (
	xxh "github.com/cespare/xxhash/v2"
)

type Checksum uint64

const ChecksumSize = 8

func NewChecksum(data []byte) Checksum {
	return Checksum(xxh.Sum64(data))
}
