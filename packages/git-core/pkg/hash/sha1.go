package hash

import (
	"crypto/sha1"
	"hash"
)

// SHA1Hasher implements the Hasher interface using SHA-1
type SHA1Hasher struct{}

// NewSHA1 creates a new SHA-1 hasher
func NewSHA1() Hasher {
	return &SHA1Hasher{}
}

// Algorithm returns the SHA-1 algorithm identifier
func (h *SHA1Hasher) Algorithm() Algorithm {
	return SHA1
}

// Size returns the size of SHA-1 hashes in bytes (20 bytes)
func (h *SHA1Hasher) Size() int {
	return sha1.Size
}

// Hash computes the SHA-1 hash of the given data
func (h *SHA1Hasher) Hash(data []byte) Hash {
	sum := sha1.Sum(data)
	return Hash(sum[:])
}

// HashString computes the SHA-1 hash of a string
func (h *SHA1Hasher) HashString(s string) Hash {
	return h.Hash([]byte(s))
}

// HashMultiple computes the SHA-1 hash of multiple data chunks
func (h *SHA1Hasher) HashMultiple(chunks ...[]byte) Hash {
	hasher := sha1.New()
	for _, chunk := range chunks {
		hasher.Write(chunk)
	}
	return Hash(hasher.Sum(nil))
}

// New returns a new hash.Hash for incremental SHA-1 hashing
func (h *SHA1Hasher) New() hash.Hash {
	return sha1.New()
}
