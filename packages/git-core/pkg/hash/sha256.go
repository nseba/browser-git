package hash

import (
	"crypto/sha256"
	"hash"
)

// SHA256Hasher implements the Hasher interface using SHA-256
type SHA256Hasher struct{}

// NewSHA256 creates a new SHA-256 hasher
func NewSHA256() Hasher {
	return &SHA256Hasher{}
}

// Algorithm returns the SHA-256 algorithm identifier
func (h *SHA256Hasher) Algorithm() Algorithm {
	return SHA256
}

// Size returns the size of SHA-256 hashes in bytes (32 bytes)
func (h *SHA256Hasher) Size() int {
	return sha256.Size
}

// Hash computes the SHA-256 hash of the given data
func (h *SHA256Hasher) Hash(data []byte) Hash {
	sum := sha256.Sum256(data)
	return Hash(sum[:])
}

// HashString computes the SHA-256 hash of a string
func (h *SHA256Hasher) HashString(s string) Hash {
	return h.Hash([]byte(s))
}

// HashMultiple computes the SHA-256 hash of multiple data chunks
func (h *SHA256Hasher) HashMultiple(chunks ...[]byte) Hash {
	hasher := sha256.New()
	for _, chunk := range chunks {
		hasher.Write(chunk)
	}
	return Hash(hasher.Sum(nil))
}

// New returns a new hash.Hash for incremental SHA-256 hashing
func (h *SHA256Hasher) New() hash.Hash {
	return sha256.New()
}
