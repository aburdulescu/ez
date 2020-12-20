package id

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

type ID []byte

const Alg = "sha256"

func New(checksums []Checksum) ID {
	h := sha256.New()
	b := make(Checksum, checksum.Size)
	for i := range checksums {
		binary.BigEndian.PutUint64(b, checksums[i])
		h.Write(b)
	}
	return id
}

func (id ID) String() string {
	return hex.EncodeToString(id)
}

func (l ID) Equals(r ID) bool {
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}
