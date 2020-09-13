package main

import (
	"crypto/sha1"
	"encoding/hex"
)

const HASH_ALG = "sha1"

type Hash []byte

func NewHash(chunks []Hash) (Hash, error) {
	h := sha1.New()
	for i := range chunks {
		if _, err := h.Write(chunks[i]); err != nil {
			return nil, err
		}
	}
	return h.Sum(nil), nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func NewHashFromString(s string) (Hash, error) {
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
