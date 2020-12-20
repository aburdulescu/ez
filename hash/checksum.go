package checksum

import (
	"hash/crc64"
)

type Checksum uint64

const Size = crc64.Size

func New(data []byte) Checksum {
	return crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
}
