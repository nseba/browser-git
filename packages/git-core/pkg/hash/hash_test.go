package hash

import (
	"testing"
)

// Test data
const (
	testString       = "hello world"
	testStringSHA1   = "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
	testStringSHA256 = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	emptyStringSHA1   = "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	emptyStringSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// TestSHA1Basic tests basic SHA-1 hashing
func TestSHA1Basic(t *testing.T) {
	hasher := NewSHA1()

	if hasher.Algorithm() != SHA1 {
		t.Errorf("Expected algorithm SHA1, got %s", hasher.Algorithm())
	}

	if hasher.Size() != 20 {
		t.Errorf("Expected size 20, got %d", hasher.Size())
	}

	hash := hasher.HashString(testString)
	if hash.String() != testStringSHA1 {
		t.Errorf("Expected hash %s, got %s", testStringSHA1, hash.String())
	}
}

// TestSHA256Basic tests basic SHA-256 hashing
func TestSHA256Basic(t *testing.T) {
	hasher := NewSHA256()

	if hasher.Algorithm() != SHA256 {
		t.Errorf("Expected algorithm SHA256, got %s", hasher.Algorithm())
	}

	if hasher.Size() != 32 {
		t.Errorf("Expected size 32, got %d", hasher.Size())
	}

	hash := hasher.HashString(testString)
	if hash.String() != testStringSHA256 {
		t.Errorf("Expected hash %s, got %s", testStringSHA256, hash.String())
	}
}

// TestEmptyString tests hashing empty strings
func TestEmptyString(t *testing.T) {
	sha1 := NewSHA1()
	sha256 := NewSHA256()

	hash1 := sha1.HashString("")
	if hash1.String() != emptyStringSHA1 {
		t.Errorf("SHA-1: Expected hash %s, got %s", emptyStringSHA1, hash1.String())
	}

	hash256 := sha256.HashString("")
	if hash256.String() != emptyStringSHA256 {
		t.Errorf("SHA-256: Expected hash %s, got %s", emptyStringSHA256, hash256.String())
	}
}

// TestHashMultiple tests hashing multiple chunks
func TestHashMultiple(t *testing.T) {
	sha1 := NewSHA1()

	// Hash as single chunk
	single := sha1.HashString(testString)

	// Hash as multiple chunks
	multiple := sha1.HashMultiple([]byte("hello"), []byte(" "), []byte("world"))

	if !single.Equals(multiple) {
		t.Errorf("Hashes don't match: single=%s, multiple=%s", single.String(), multiple.String())
	}
}

// TestHashEquality tests hash equality comparison
func TestHashEquality(t *testing.T) {
	sha1 := NewSHA1()

	hash1 := sha1.HashString(testString)
	hash2 := sha1.HashString(testString)
	hash3 := sha1.HashString("different")

	if !hash1.Equals(hash2) {
		t.Error("Equal hashes not detected as equal")
	}

	if hash1.Equals(hash3) {
		t.Error("Different hashes detected as equal")
	}
}

// TestHashValidation tests hash validation
func TestHashValidation(t *testing.T) {
	sha1 := NewSHA1()
	sha256 := NewSHA256()

	hash1 := sha1.HashString(testString)
	hash256 := sha256.HashString(testString)

	// Valid hashes
	if err := hash1.Validate(SHA1); err != nil {
		t.Errorf("SHA-1 hash validation failed: %v", err)
	}

	if err := hash256.Validate(SHA256); err != nil {
		t.Errorf("SHA-256 hash validation failed: %v", err)
	}

	// Invalid algorithm for hash
	if err := hash1.Validate(SHA256); err == nil {
		t.Error("SHA-1 hash incorrectly validated as SHA-256")
	}

	if err := hash256.Validate(SHA1); err == nil {
		t.Error("SHA-256 hash incorrectly validated as SHA-1")
	}
}

// TestParseHash tests parsing hash strings
func TestParseHash(t *testing.T) {
	tests := []struct {
		input    string
		valid    bool
		expected string
	}{
		{testStringSHA1, true, testStringSHA1},
		{testStringSHA256, true, testStringSHA256},
		{"invalid", false, ""},
		{"", false, ""},
		{"g" + testStringSHA1[1:], false, ""}, // invalid hex char
	}

	for _, tt := range tests {
		hash, err := ParseHash(tt.input)
		if tt.valid {
			if err != nil {
				t.Errorf("Failed to parse valid hash %s: %v", tt.input, err)
			}
			if hash.String() != tt.expected {
				t.Errorf("Parsed hash mismatch: expected %s, got %s", tt.expected, hash.String())
			}
		} else {
			if err == nil {
				t.Errorf("Invalid hash %s should have failed to parse", tt.input)
			}
		}
	}
}

// TestHashFactory tests the hash factory function
func TestHashFactory(t *testing.T) {
	sha1, err := NewHasher(SHA1)
	if err != nil {
		t.Fatalf("Failed to create SHA-1 hasher: %v", err)
	}

	sha256, err := NewHasher(SHA256)
	if err != nil {
		t.Fatalf("Failed to create SHA-256 hasher: %v", err)
	}

	if sha1.Algorithm() != SHA1 {
		t.Error("SHA-1 hasher has wrong algorithm")
	}

	if sha256.Algorithm() != SHA256 {
		t.Error("SHA-256 hasher has wrong algorithm")
	}

	// Invalid algorithm
	_, err = NewHasher("invalid")
	if err == nil {
		t.Error("Invalid algorithm should have returned error")
	}
}

// TestIsValidHashString tests hash string validation
func TestIsValidHashString(t *testing.T) {
	tests := []struct {
		hash  string
		algo  Algorithm
		valid bool
	}{
		{testStringSHA1, SHA1, true},
		{testStringSHA256, SHA256, true},
		{testStringSHA1, SHA256, false},      // wrong length
		{testStringSHA256, SHA1, false},      // wrong length
		{testStringSHA1[:10], SHA1, false},   // too short
		{"g" + testStringSHA1[1:], SHA1, false}, // invalid char
	}

	for _, tt := range tests {
		result := IsValidHashString(tt.hash, tt.algo)
		if result != tt.valid {
			t.Errorf("IsValidHashString(%s, %s) = %v, want %v",
				tt.hash, tt.algo, result, tt.valid)
		}
	}
}

// TestShortHash tests short hash generation
func TestShortHash(t *testing.T) {
	sha1 := NewSHA1()
	hash := sha1.HashString(testString)

	short := hash.ShortHash()
	if len(short) != 7 {
		t.Errorf("Short hash length is %d, expected 7", len(short))
	}

	if short != testStringSHA1[:7] {
		t.Errorf("Short hash is %s, expected %s", short, testStringSHA1[:7])
	}

	// Test custom length
	short10 := hash.ShortHashN(10)
	if len(short10) != 10 {
		t.Errorf("Short hash (10) length is %d, expected 10", len(short10))
	}
}

// TestHashObject tests Git object hashing
func TestHashObject(t *testing.T) {
	sha1 := NewSHA1()

	content := []byte("test content")
	hash := HashObject(sha1, "blob", content)

	// Hash should not be zero
	if hash.IsZero() {
		t.Error("Hash is zero")
	}

	// Hash should be reproducible
	hash2 := HashObject(sha1, "blob", content)
	if !hash.Equals(hash2) {
		t.Error("Hash is not reproducible")
	}

	// Different types should produce different hashes
	hashTree := HashObject(sha1, "tree", content)
	if hash.Equals(hashTree) {
		t.Error("Blob and tree hashes should be different")
	}
}

// TestHashBlob tests blob hashing convenience function
func TestHashBlob(t *testing.T) {
	sha1 := NewSHA1()

	content := []byte("blob content")
	hash := HashBlob(sha1, content)

	// Should match HashObject with "blob" type
	expected := HashObject(sha1, "blob", content)
	if !hash.Equals(expected) {
		t.Error("HashBlob doesn't match HashObject")
	}
}

// TestUniqueHashes tests hash deduplication
func TestUniqueHashes(t *testing.T) {
	sha1 := NewSHA1()

	hash1 := sha1.HashString("a")
	hash2 := sha1.HashString("b")
	hash3 := sha1.HashString("a") // duplicate

	hashes := []Hash{hash1, hash2, hash3, hash1}
	unique := UniqueHashes(hashes)

	if len(unique) != 2 {
		t.Errorf("Expected 2 unique hashes, got %d", len(unique))
	}

	if !ContainsHash(unique, hash1) || !ContainsHash(unique, hash2) {
		t.Error("Unique hashes missing expected values")
	}
}

// TestZeroHash tests zero hash creation
func TestZeroHash(t *testing.T) {
	zero1 := ZeroHash(SHA1)
	if len(zero1) != 20 {
		t.Errorf("SHA-1 zero hash length is %d, expected 20", len(zero1))
	}

	if !zero1.IsZero() {
		t.Error("Zero hash is not zero")
	}

	zero256 := ZeroHash(SHA256)
	if len(zero256) != 32 {
		t.Errorf("SHA-256 zero hash length is %d, expected 32", len(zero256))
	}
}

// BenchmarkSHA1 benchmarks SHA-1 hashing
func BenchmarkSHA1(b *testing.B) {
	hasher := NewSHA1()
	data := []byte(testString)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Hash(data)
	}
}

// BenchmarkSHA256 benchmarks SHA-256 hashing
func BenchmarkSHA256(b *testing.B) {
	hasher := NewSHA256()
	data := []byte(testString)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Hash(data)
	}
}
