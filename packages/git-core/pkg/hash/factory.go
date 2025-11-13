package hash

import "fmt"

// DefaultAlgorithm is the default hash algorithm (SHA-1 for compatibility)
const DefaultAlgorithm = SHA1

// NewHasher creates a new hasher for the specified algorithm
func NewHasher(algo Algorithm) (Hasher, error) {
	switch algo {
	case SHA1:
		return NewSHA1(), nil
	case SHA256:
		return NewSHA256(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algo)
	}
}

// NewDefaultHasher creates a new hasher using the default algorithm (SHA-1)
func NewDefaultHasher() Hasher {
	return NewSHA1()
}

// GetHasher is a convenience function that returns a hasher or panics on error
// Use NewHasher for error handling
func GetHasher(algo Algorithm) Hasher {
	hasher, err := NewHasher(algo)
	if err != nil {
		panic(err)
	}
	return hasher
}

// ParseAlgorithm parses an algorithm string
func ParseAlgorithm(s string) (Algorithm, error) {
	algo := Algorithm(s)
	switch algo {
	case SHA1, SHA256:
		return algo, nil
	default:
		return "", fmt.Errorf("unknown algorithm: %s", s)
	}
}

// IsValidAlgorithm checks if an algorithm is valid
func IsValidAlgorithm(algo Algorithm) bool {
	return algo == SHA1 || algo == SHA256
}
