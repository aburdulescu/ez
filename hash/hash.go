package hash

import (
	"crypto/sha1"
	"encoding/hex"
)

const HASH_ALG = "sha1" // TODO: use not-crypto hash(xxhash)?

type Hash []byte

func FromChunkHashes(chunkHashes []Hash) (Hash, error) {
	h := sha1.New()
	for i := range chunkHashes {
		if _, err := h.Write(chunkHashes[i]); err != nil {
			return nil, err
		}
	}
	return h.Sum(nil), nil
}

func FromChunk(chunk []byte) (Hash, error) {
	h := sha1.New()
	if _, err := h.Write(chunk); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func FromHex(s string) (Hash, error) {
	h, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h Hash) Equals(other Hash) bool {
	for i := range h {
		if h[i] != other[i] {
			return false
		}
	}
	return true
}
