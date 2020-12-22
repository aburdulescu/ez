package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"github.com/aburdulescu/ez/cmn"
)

const idAlg = "sha256"

func NewID(checksums []cmn.Checksum) string {
	h := sha256.New()
	b := make([]byte, cmn.ChecksumSize)
	for i := range checksums {
		binary.BigEndian.PutUint64(b, uint64(checksums[i]))
		h.Write(b)
	}
	digest := h.Sum(nil)
	id := idAlg + "-" + hex.EncodeToString(digest)
	return id
}
