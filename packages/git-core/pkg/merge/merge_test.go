package merge

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// mockDatabase is a simple in-memory object database for testing
type mockDatabase struct {
	objects map[string]object.Object
}

func newMockDatabase() *mockDatabase {
	return &mockDatabase{
		objects: make(map[string]object.Object),
	}
}

func (db *mockDatabase) Get(h hash.Hash) (object.Object, error) {
	obj, ok := db.objects[h.String()]
	if !ok {
		return nil, fmt.Errorf("object not found")
	}
	return obj, nil
}

func (db *mockDatabase) Put(obj object.Object) (hash.Hash, error) {
	// Get serialized bytes to compute hash
	var buf bytes.Buffer
	if err := obj.SerializeWithHeader(&buf); err != nil {
		return nil, err
	}

	// For testing, we'll use a simple hasher
	hasher, _ := hash.NewHasher(hash.SHA1)
	h := hasher.Hash(buf.Bytes())
	obj.SetHash(h)

	db.objects[h.String()] = obj
	return h, nil
}

func (db *mockDatabase) Has(h hash.Hash) bool {
	_, ok := db.objects[h.String()]
	return ok
}

func (db *mockDatabase) Delete(h hash.Hash) error {
	delete(db.objects, h.String())
	return nil
}

func (db *mockDatabase) List() ([]hash.Hash, error) {
	hashes := make([]hash.Hash, 0, len(db.objects))
	for hashStr := range db.objects {
		h, _ := hash.ParseHash(hashStr)
		hashes = append(hashes, h)
	}
	return hashes, nil
}

func (db *mockDatabase) Close() error {
	return nil
}

// Test helper functions
func createTestCommit(db *mockDatabase, hasher hash.Hasher, treeHash hash.Hash, parents []hash.Hash, message string) (*object.Commit, error) {
	commit := object.NewCommit()
	commit.Tree = treeHash
	commit.Parents = parents
	commit.Author = object.Signature{
		Name:  "Test Author",
		Email: "test@example.com",
		When:  time.Now(),
	}
	commit.Committer = commit.Author
	commit.Message = message

	if _, err := db.Put(commit); err != nil {
		return nil, err
	}

	return commit, nil
}

func createTestBlob(db *mockDatabase, hasher hash.Hasher, content []byte) (*object.Blob, error) {
	blob := object.NewBlob(content)

	if _, err := db.Put(blob); err != nil {
		return nil, err
	}

	return blob, nil
}

func createTestTree(db *mockDatabase, hasher hash.Hasher, entries []object.TreeEntry) (*object.Tree, error) {
	tree := object.NewTree()
	for _, entry := range entries {
		tree.AddEntry(entry)
	}

	tree.Sort()

	if _, err := db.Put(tree); err != nil {
		return nil, err
	}

	return tree, nil
}

// TestFindMergeBase tests finding the common ancestor
func TestFindMergeBase(t *testing.T) {
	db := newMockDatabase()
	hasher, _ := hash.NewHasher(hash.SHA1)

	// Create a simple commit history:
	//   A (base)
	//  / \
	// B   C
	// |   |
	// D   E

	// Create tree (empty for this test)
	tree, err := createTestTree(db, hasher, []object.TreeEntry{})
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Create commit A (base)
	commitA, err := createTestCommit(db, hasher, tree.Hash(), nil, "Commit A")
	if err != nil {
		t.Fatalf("Failed to create commit A: %v", err)
	}

	// Create commit B (child of A)
	commitB, err := createTestCommit(db, hasher, tree.Hash(), []hash.Hash{commitA.Hash()}, "Commit B")
	if err != nil {
		t.Fatalf("Failed to create commit B: %v", err)
	}

	// Create commit C (child of A)
	commitC, err := createTestCommit(db, hasher, tree.Hash(), []hash.Hash{commitA.Hash()}, "Commit C")
	if err != nil {
		t.Fatalf("Failed to create commit C: %v", err)
	}

	// Create commit D (child of B)
	commitD, err := createTestCommit(db, hasher, tree.Hash(), []hash.Hash{commitB.Hash()}, "Commit D")
	if err != nil {
		t.Fatalf("Failed to create commit D: %v", err)
	}

	// Create commit E (child of C)
	commitE, err := createTestCommit(db, hasher, tree.Hash(), []hash.Hash{commitC.Hash()}, "Commit E")
	if err != nil {
		t.Fatalf("Failed to create commit E: %v", err)
	}

	// Find merge base between D and E
	mergeBase, err := FindMergeBase(db, commitD.Hash(), commitE.Hash())
	if err != nil {
		t.Fatalf("Failed to find merge base: %v", err)
	}

	// The merge base should be commit A
	if mergeBase.String() != commitA.Hash().String() {
		t.Errorf("Expected merge base %s, got %s", commitA.Hash().String(), mergeBase.String())
	}
}

// TestCanFastForward tests fast-forward detection
func TestCanFastForward(t *testing.T) {
	db := newMockDatabase()
	hasher, _ := hash.NewHasher(hash.SHA1)

	// Create a linear history: A -> B -> C
	tree, err := createTestTree(db, hasher, []object.TreeEntry{})
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	commitA, err := createTestCommit(db, hasher, tree.Hash(), nil, "Commit A")
	if err != nil {
		t.Fatalf("Failed to create commit A: %v", err)
	}

	commitB, err := createTestCommit(db, hasher, tree.Hash(), []hash.Hash{commitA.Hash()}, "Commit B")
	if err != nil {
		t.Fatalf("Failed to create commit B: %v", err)
	}

	commitC, err := createTestCommit(db, hasher, tree.Hash(), []hash.Hash{commitB.Hash()}, "Commit C")
	if err != nil {
		t.Fatalf("Failed to create commit C: %v", err)
	}

	// A -> C should be fast-forwardable
	canFF, err := CanFastForward(db, commitA.Hash(), commitC.Hash())
	if err != nil {
		t.Fatalf("Failed to check fast-forward: %v", err)
	}
	if !canFF {
		t.Error("Expected to be able to fast-forward from A to C")
	}

	// C -> A should NOT be fast-forwardable
	canFF, err = CanFastForward(db, commitC.Hash(), commitA.Hash())
	if err != nil {
		t.Fatalf("Failed to check fast-forward: %v", err)
	}
	if canFF {
		t.Error("Should not be able to fast-forward from C to A")
	}
}

// TestMergeContentNoConflict tests content merging without conflicts
func TestMergeContentNoConflict(t *testing.T) {
	base := []byte("line 1\nline 2\nline 3\n")
	ours := []byte("line 1\nmodified line 2\nline 3\n")
	theirs := []byte("line 1\nline 2\nline 3\nline 4\n")

	merged, hasConflict, err := MergeContent(base, ours, theirs)
	if err != nil {
		t.Fatalf("Failed to merge content: %v", err)
	}

	if hasConflict {
		t.Error("Expected no conflict, but got conflict")
	}

	expected := "line 1\nmodified line 2\nline 3\nline 4\n"
	if string(merged) != expected {
		t.Errorf("Expected merged content:\n%s\nGot:\n%s", expected, string(merged))
	}
}

// TestMergeContentWithConflict tests content merging with conflicts
func TestMergeContentWithConflict(t *testing.T) {
	base := []byte("line 1\nline 2\nline 3\n")
	ours := []byte("line 1\nour change\nline 3\n")
	theirs := []byte("line 1\ntheir change\nline 3\n")

	merged, hasConflict, err := MergeContent(base, ours, theirs)
	if err != nil {
		t.Fatalf("Failed to merge content: %v", err)
	}

	if !hasConflict {
		t.Error("Expected conflict, but got no conflict")
	}

	// Check that merged content contains conflict markers
	mergedStr := string(merged)
	if !contains(mergedStr, "<<<<<<<") || !contains(mergedStr, ">>>>>>>") || !contains(mergedStr, "=======") {
		t.Error("Expected conflict markers in merged content")
	}
}

// TestBinaryContentDetection tests binary content detection
func TestBinaryContentDetection(t *testing.T) {
	// Text content
	textContent := []byte("This is text content\nwith multiple lines\n")
	if isBinaryContent(textContent) {
		t.Error("Expected text content to be detected as non-binary")
	}

	// Binary content (contains null bytes)
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	if !isBinaryContent(binaryContent) {
		t.Error("Expected binary content to be detected as binary")
	}

	// Empty content
	emptyContent := []byte{}
	if isBinaryContent(emptyContent) {
		t.Error("Expected empty content to be detected as non-binary")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
