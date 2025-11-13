package object

import (
	"bytes"
	"testing"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// TestBlobBasic tests basic blob functionality
func TestBlobBasic(t *testing.T) {
	content := []byte("hello world")
	blob := NewBlob(content)

	if blob.Type() != BlobType {
		t.Errorf("Expected type %s, got %s", BlobType, blob.Type())
	}

	if !bytes.Equal(blob.Content(), content) {
		t.Errorf("Content mismatch")
	}

	if blob.Size() != int64(len(content)) {
		t.Errorf("Size mismatch: expected %d, got %d", len(content), blob.Size())
	}
}

// TestBlobSerialization tests blob serialization and deserialization
func TestBlobSerialization(t *testing.T) {
	content := []byte("test content\nline 2")
	blob := NewBlob(content)

	// Serialize
	data, err := blob.Bytes()
	if err != nil {
		t.Fatalf("Failed to serialize blob: %v", err)
	}

	// Parse
	parsed, err := ParseObjectWithHeader(data)
	if err != nil {
		t.Fatalf("Failed to parse blob: %v", err)
	}

	parsedBlob, ok := parsed.(*Blob)
	if !ok {
		t.Fatalf("Parsed object is not a blob")
	}

	if !parsedBlob.Equals(blob) {
		t.Errorf("Parsed blob does not match original")
	}
}

// TestBlobHash tests blob hashing
func TestBlobHash(t *testing.T) {
	content := []byte("hello")
	blob := NewBlob(content)

	hasher := hash.NewSHA1()
	if err := blob.ComputeHash(hasher); err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	// Verify hash is not zero
	if blob.Hash().IsZero() {
		t.Error("Hash is zero")
	}

	// Verify hash is reproducible
	blob2 := NewBlob(content)
	if err := blob2.ComputeHash(hasher); err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	if !blob.Hash().Equals(blob2.Hash()) {
		t.Error("Hashes don't match for identical content")
	}
}

// TestTreeBasic tests basic tree functionality
func TestTreeBasic(t *testing.T) {
	tree := NewTree()

	if tree.Type() != TreeType {
		t.Errorf("Expected type %s, got %s", TreeType, tree.Type())
	}

	// Add entries
	h1 := hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	h2 := hash.MustParseHash("b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")

	tree.AddEntryWithMode(ModeRegular, "file1.txt", h1)
	tree.AddEntryWithMode(ModeDir, "subdir", h2)

	entries := tree.Entries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

// TestTreeSerialization tests tree serialization and deserialization
func TestTreeSerialization(t *testing.T) {
	tree := NewTree()
	h1 := hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	h2 := hash.MustParseHash("3aae6c35c94fcfb415dbe95f408b9ce91ee846ed")

	tree.AddEntryWithMode(ModeRegular, "file.txt", h1)
	tree.AddEntryWithMode(ModeExecutable, "script.sh", h2)

	// Serialize
	data, err := tree.Bytes()
	if err != nil {
		t.Fatalf("Failed to serialize tree: %v", err)
	}

	// Parse
	parsed, err := ParseObjectWithHeader(data)
	if err != nil {
		t.Fatalf("Failed to parse tree: %v", err)
	}

	parsedTree, ok := parsed.(*Tree)
	if !ok {
		t.Fatalf("Parsed object is not a tree")
	}

	if len(parsedTree.Entries()) != len(tree.Entries()) {
		t.Errorf("Entry count mismatch: expected %d, got %d",
			len(tree.Entries()), len(parsedTree.Entries()))
	}
}

// TestTreeSorting tests tree entry sorting
func TestTreeSorting(t *testing.T) {
	tree := NewTree()
	h := hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")

	// Add entries in wrong order
	tree.AddEntryWithMode(ModeRegular, "zebra.txt", h)
	tree.AddEntryWithMode(ModeRegular, "apple.txt", h)
	tree.AddEntryWithMode(ModeDir, "dir", h)

	tree.Sort()

	entries := tree.Entries()
	if entries[0].Name != "apple.txt" {
		t.Errorf("First entry should be apple.txt, got %s", entries[0].Name)
	}
}

// TestCommitBasic tests basic commit functionality
func TestCommitBasic(t *testing.T) {
	commit := NewCommit()

	if commit.Type() != CommitType {
		t.Errorf("Expected type %s, got %s", CommitType, commit.Type())
	}

	// Set fields
	commit.Tree = hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	commit.Author = Signature{
		Name:  "Test Author",
		Email: "author@example.com",
		When:  time.Unix(1234567890, 0).UTC(),
	}
	commit.Committer = Signature{
		Name:  "Test Committer",
		Email: "committer@example.com",
		When:  time.Unix(1234567890, 0).UTC(),
	}
	commit.Message = "Test commit\n"

	if !commit.IsRoot() {
		t.Error("Commit with no parents should be root")
	}

	// Add parent
	parent := hash.MustParseHash("3aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	commit.AddParent(parent)

	if commit.IsRoot() {
		t.Error("Commit with parent should not be root")
	}

	if commit.IsMerge() {
		t.Error("Commit with one parent should not be merge")
	}
}

// TestCommitSerialization tests commit serialization and deserialization
func TestCommitSerialization(t *testing.T) {
	commit := NewCommit()
	commit.Tree = hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	commit.AddParent(hash.MustParseHash("3aae6c35c94fcfb415dbe95f408b9ce91ee846ed"))
	commit.Author = Signature{
		Name:  "Test Author",
		Email: "test@example.com",
		When:  time.Unix(1234567890, 0).UTC(),
	}
	commit.Committer = commit.Author
	commit.Message = "Test commit message\n"

	// Serialize
	data, err := commit.Bytes()
	if err != nil {
		t.Fatalf("Failed to serialize commit: %v", err)
	}

	// Parse
	parsed, err := ParseObjectWithHeader(data)
	if err != nil {
		t.Fatalf("Failed to parse commit: %v", err)
	}

	parsedCommit, ok := parsed.(*Commit)
	if !ok {
		t.Fatalf("Parsed object is not a commit")
	}

	if !parsedCommit.Tree.Equals(commit.Tree) {
		t.Error("Tree hash mismatch")
	}

	if len(parsedCommit.Parents) != len(commit.Parents) {
		t.Errorf("Parent count mismatch")
	}

	if parsedCommit.Message != commit.Message {
		t.Errorf("Message mismatch: expected %q, got %q",
			commit.Message, parsedCommit.Message)
	}
}

// TestSignatureFormat tests signature formatting and parsing
func TestSignatureFormat(t *testing.T) {
	sig := Signature{
		Name:  "John Doe",
		Email: "john@example.com",
		When:  time.Unix(1234567890, 0).In(time.FixedZone("", -7*3600)),
	}

	formatted := sig.Format()
	parsed, err := ParseSignature(formatted)
	if err != nil {
		t.Fatalf("Failed to parse signature: %v", err)
	}

	if parsed.Name != sig.Name {
		t.Errorf("Name mismatch: expected %q, got %q", sig.Name, parsed.Name)
	}

	if parsed.Email != sig.Email {
		t.Errorf("Email mismatch: expected %q, got %q", sig.Email, parsed.Email)
	}

	if parsed.When.Unix() != sig.When.Unix() {
		t.Errorf("Timestamp mismatch: expected %d, got %d",
			sig.When.Unix(), parsed.When.Unix())
	}
}

// TestTagBasic tests basic tag functionality
func TestTagBasic(t *testing.T) {
	tag := NewTag()

	if tag.Type() != TagType {
		t.Errorf("Expected type %s, got %s", TagType, tag.Type())
	}

	tag.Target = hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	tag.TargetType = CommitType
	tag.Name = "v1.0.0"
	tag.Tagger = Signature{
		Name:  "Tagger",
		Email: "tagger@example.com",
		When:  time.Unix(1234567890, 0).UTC(),
	}
	tag.Message = "Release v1.0.0\n"

	if tag.IsLightweight() {
		t.Error("Tag objects are never lightweight")
	}
}

// TestTagSerialization tests tag serialization and deserialization
func TestTagSerialization(t *testing.T) {
	tag := NewTag()
	tag.Target = hash.MustParseHash("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	tag.TargetType = CommitType
	tag.Name = "v1.0.0"
	tag.Tagger = Signature{
		Name:  "Test Tagger",
		Email: "tagger@example.com",
		When:  time.Unix(1234567890, 0).UTC(),
	}
	tag.Message = "Release version 1.0.0\n"

	// Serialize
	data, err := tag.Bytes()
	if err != nil {
		t.Fatalf("Failed to serialize tag: %v", err)
	}

	// Parse
	parsed, err := ParseObjectWithHeader(data)
	if err != nil {
		t.Fatalf("Failed to parse tag: %v", err)
	}

	parsedTag, ok := parsed.(*Tag)
	if !ok {
		t.Fatalf("Parsed object is not a tag")
	}

	if !parsedTag.Target.Equals(tag.Target) {
		t.Error("Target hash mismatch")
	}

	if parsedTag.TargetType != tag.TargetType {
		t.Errorf("Target type mismatch: expected %s, got %s",
			tag.TargetType, parsedTag.TargetType)
	}

	if parsedTag.Name != tag.Name {
		t.Errorf("Name mismatch: expected %q, got %q", tag.Name, parsedTag.Name)
	}
}

// TestCompression tests object compression and decompression
func TestCompression(t *testing.T) {
	data := []byte("test data for compression")

	compressed, err := Compress(data)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}

	if len(compressed) == 0 {
		t.Error("Compressed data is empty")
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if !bytes.Equal(data, decompressed) {
		t.Error("Decompressed data doesn't match original")
	}
}

// TestParseType tests type parsing
func TestParseType(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		want  Type
	}{
		{"blob", true, BlobType},
		{"tree", true, TreeType},
		{"commit", true, CommitType},
		{"tag", true, TagType},
		{"invalid", false, ""},
	}

	for _, tt := range tests {
		got, err := ParseType(tt.input)
		if tt.valid {
			if err != nil {
				t.Errorf("ParseType(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		} else {
			if err == nil {
				t.Errorf("ParseType(%q) should have returned error", tt.input)
			}
		}
	}
}

// TestFileModeValidation tests file mode validation
func TestFileModeValidation(t *testing.T) {
	validModes := []FileMode{ModeDir, ModeRegular, ModeExecutable, ModeSymlink, ModeGitlink}
	for _, mode := range validModes {
		if !IsValidMode(mode) {
			t.Errorf("Mode %o should be valid", mode)
		}
	}

	invalidMode := FileMode(0777777)
	if IsValidMode(invalidMode) {
		t.Errorf("Mode %o should be invalid", invalidMode)
	}
}
