package hash

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

// HashObject hashes a Git object with the given type and content
// Format: "<type> <size>\0<content>"
func HashObject(hasher Hasher, objectType string, content []byte) Hash {
	header := fmt.Sprintf("%s %d\x00", objectType, len(content))
	return hasher.HashMultiple([]byte(header), content)
}

// HashBlob hashes blob content
func HashBlob(hasher Hasher, content []byte) Hash {
	return HashObject(hasher, "blob", content)
}

// HashTree hashes tree content
func HashTree(hasher Hasher, content []byte) Hash {
	return HashObject(hasher, "tree", content)
}

// HashCommit hashes commit content
func HashCommit(hasher Hasher, content []byte) Hash {
	return HashObject(hasher, "commit", content)
}

// HashTag hashes tag content
func HashTag(hasher Hasher, content []byte) Hash {
	return HashObject(hasher, "tag", content)
}

// CompareHashes compares two hashes lexicographically
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func CompareHashes(a, b Hash) int {
	return bytes.Compare(a, b)
}

// IsValidHashString checks if a string is a valid hex-encoded hash
func IsValidHashString(s string, algo Algorithm) bool {
	// Check length
	expectedLen := 0
	switch algo {
	case SHA1:
		expectedLen = 40 // 20 bytes * 2 hex chars
	case SHA256:
		expectedLen = 64 // 32 bytes * 2 hex chars
	default:
		return false
	}

	if len(s) != expectedLen {
		return false
	}

	// Check if all characters are valid hex
	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// ShortHash returns a shortened version of the hash (first 7 characters)
func (h Hash) ShortHash() string {
	s := h.String()
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

// ShortHashN returns a shortened version of the hash with n characters
func (h Hash) ShortHashN(n int) string {
	s := h.String()
	if len(s) > n {
		return s[:n]
	}
	return s
}

// ParseHashWithAlgo parses a hash string and validates it against an algorithm
func ParseHashWithAlgo(s string, algo Algorithm) (Hash, error) {
	if !IsValidHashString(s, algo) {
		return nil, fmt.Errorf("invalid hash string for algorithm %s", algo)
	}

	return ParseHash(s)
}

// MustParseHash parses a hash string or panics on error
func MustParseHash(s string) Hash {
	h, err := ParseHash(s)
	if err != nil {
		panic(err)
	}
	return h
}

// HashesEqual checks if two hash slices are equal
func HashesEqual(a, b []Hash) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].Equals(b[i]) {
			return false
		}
	}

	return true
}

// ContainsHash checks if a slice contains a specific hash
func ContainsHash(hashes []Hash, target Hash) bool {
	for _, h := range hashes {
		if h.Equals(target) {
			return true
		}
	}
	return false
}

// UniqueHashes returns a deduplicated slice of hashes
func UniqueHashes(hashes []Hash) []Hash {
	seen := make(map[string]bool)
	result := make([]Hash, 0, len(hashes))

	for _, h := range hashes {
		key := h.String()
		if !seen[key] {
			seen[key] = true
			result = append(result, h)
		}
	}

	return result
}

// FormatHash formats a hash with optional prefix and color
func FormatHash(h Hash, short bool) string {
	if short {
		return h.ShortHash()
	}
	return h.String()
}

// ParseHashPrefix finds a hash that starts with the given prefix
func ParseHashPrefix(prefix string, candidates []Hash) (Hash, error) {
	prefix = strings.ToLower(prefix)
	matches := make([]Hash, 0)

	for _, h := range candidates {
		if strings.HasPrefix(h.String(), prefix) {
			matches = append(matches, h)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no hash found with prefix: %s", prefix)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous hash prefix: %s (matches %d hashes)", prefix, len(matches))
	}

	return matches[0], nil
}

// EncodeHex encodes bytes to hex string
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex decodes hex string to bytes
func DecodeHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// MustDecodeHex decodes hex string to bytes or panics on error
func MustDecodeHex(s string) []byte {
	data, err := DecodeHex(s)
	if err != nil {
		panic(err)
	}
	return data
}
