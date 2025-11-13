package hash

import (
	"encoding/hex"
	"fmt"
	"hash"
)

// Algorithm represents the hash algorithm type
type Algorithm string

const (
	// SHA1 represents the SHA-1 hash algorithm
	SHA1 Algorithm = "sha1"
	// SHA256 represents the SHA-256 hash algorithm
	SHA256 Algorithm = "sha256"
)

// Hash represents a Git object hash
type Hash []byte

// Hasher is the interface for Git hash algorithms
type Hasher interface {
	// Algorithm returns the hash algorithm type
	Algorithm() Algorithm

	// Size returns the size of the hash in bytes
	Size() int

	// Hash computes the hash of the given data
	Hash(data []byte) Hash

	// HashString computes the hash of a string
	HashString(s string) Hash

	// HashMultiple computes the hash of multiple data chunks
	HashMultiple(chunks ...[]byte) Hash

	// New returns a new hash.Hash for incremental hashing
	New() hash.Hash
}

// String returns the hex-encoded string representation of the hash
func (h Hash) String() string {
	return hex.EncodeToString(h)
}

// Bytes returns the hash as a byte slice
func (h Hash) Bytes() []byte {
	return []byte(h)
}

// Equals compares two hashes for equality
func (h Hash) Equals(other Hash) bool {
	if len(h) != len(other) {
		return false
	}
	for i := range h {
		if h[i] != other[i] {
			return false
		}
	}
	return true
}

// IsZero returns true if the hash is empty (all zeros)
func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

// Validate checks if the hash has the correct size for its algorithm
func (h Hash) Validate(algo Algorithm) error {
	expectedSize := 0
	switch algo {
	case SHA1:
		expectedSize = 20
	case SHA256:
		expectedSize = 32
	default:
		return fmt.Errorf("unknown algorithm: %s", algo)
	}

	if len(h) != expectedSize {
		return fmt.Errorf("invalid hash size: expected %d bytes for %s, got %d",
			expectedSize, algo, len(h))
	}

	return nil
}

// ParseHash parses a hex-encoded hash string
func ParseHash(s string) (Hash, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("invalid hash string: empty string")
	}

	if len(s)%2 != 0 {
		return nil, fmt.Errorf("invalid hash string: odd length")
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hash string: %w", err)
	}

	return Hash(bytes), nil
}

// NewHash creates a new hash from bytes
func NewHash(bytes []byte) Hash {
	h := make(Hash, len(bytes))
	copy(h, bytes)
	return h
}

// ZeroHash creates a zero hash of the specified size
func ZeroHash(algo Algorithm) Hash {
	size := 0
	switch algo {
	case SHA1:
		size = 20
	case SHA256:
		size = 32
	default:
		size = 20 // default to SHA-1
	}
	return make(Hash, size)
}
