package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"github.com/zeebo/xxh3"
)

type Checksum uint64 // TODO: move to separate package or rename this one or to cmn package

const ChecksumSize = 8

func NewChecksum(data []byte) Checksum {
	return Checksum(xxh3.Hash(data))
}

const IDAlg = "sha256"

func NewID(checksums []Checksum) string { // TODO: move to cmd/ezs
	h := sha256.New()
	b := make([]byte, ChecksumSize)
	for i := range checksums {
		binary.BigEndian.PutUint64(b, uint64(checksums[i]))
		h.Write(b)
	}
	digest := h.Sum(nil)
	id := IDAlg + "-" + hex.EncodeToString(digest)
	return id
}
