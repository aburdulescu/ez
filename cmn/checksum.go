package cmn

import "hash/crc64"

type Checksum uint64

const ChecksumSize = crc64.Size

func NewChecksum(data []byte) Checksum {
	return Checksum(crc64.Checksum(data, crc64.MakeTable(crc64.ECMA)))
}
