package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	idx := NewIndex()

	// Add file
	opts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"test.txt"}, opts); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Verify
	if !idx.HasEntry("test.txt") {
		t.Error("expected file to be in index")
	}
}

func TestAddMultiple(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	idx := NewIndex()

	// Add files
	opts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"test1.txt", "test2.txt"}, opts); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	// Verify
	if !idx.HasEntry("test1.txt") {
		t.Error("expected test1.txt to be in index")
	}
	if !idx.HasEntry("test2.txt") {
		t.Error("expected test2.txt to be in index")
	}
}

func TestAddDirectory(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(subdir, "nested.txt")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	idx := NewIndex()

	// Add directory
	opts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"."}, opts); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	// Verify
	if !idx.HasEntry("test.txt") {
		t.Error("expected test.txt to be in index")
	}
	if !idx.HasEntry("subdir/nested.txt") {
		t.Error("expected subdir/nested.txt to be in index")
	}
}

func TestAddWithGitignore(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(tmpDir, "ignored.txt")
	gitignore := filepath.Join(tmpDir, ".gitignore")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create ignored file: %v", err)
	}
	if err := os.WriteFile(gitignore, []byte("ignored.txt\n"), 0644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	idx := NewIndex()

	// Add directory
	opts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"."}, opts); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	// Verify
	if !idx.HasEntry("test.txt") {
		t.Error("expected test.txt to be in index")
	}
	if idx.HasEntry("ignored.txt") {
		t.Error("expected ignored.txt not to be in index")
	}
}

func TestAddForce(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "ignored.txt")
	gitignore := filepath.Join(tmpDir, ".gitignore")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(gitignore, []byte("ignored.txt\n"), 0644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	idx := NewIndex()

	// Add with force
	opts := AddOptions{Force: true}
	if err := idx.Add(tmpDir, []string{"ignored.txt"}, opts); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Verify
	if !idx.HasEntry("ignored.txt") {
		t.Error("expected ignored.txt to be in index with force option")
	}
}

func TestAddUpdateOnly(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "tracked.txt")
	file2 := filepath.Join(tmpDir, "untracked.txt")

	if err := os.WriteFile(file1, []byte("tracked"), 0644); err != nil {
		t.Fatalf("failed to create tracked file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("untracked"), 0644); err != nil {
		t.Fatalf("failed to create untracked file: %v", err)
	}

	idx := NewIndex()

	// Add tracked file first
	opts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"tracked.txt"}, opts); err != nil {
		t.Fatalf("failed to add tracked file: %v", err)
	}

	// Modify tracked file
	if err := os.WriteFile(file1, []byte("modified tracked"), 0644); err != nil {
		t.Fatalf("failed to modify tracked file: %v", err)
	}

	// Add with update-only
	optsUpdate := AddOptions{UpdateOnly: true}
	if err := idx.Add(tmpDir, []string{"tracked.txt", "untracked.txt"}, optsUpdate); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	// Verify: tracked file should be updated, untracked should not be added
	if !idx.HasEntry("tracked.txt") {
		t.Error("expected tracked.txt to be in index")
	}
	if idx.HasEntry("untracked.txt") {
		t.Error("expected untracked.txt not to be in index with update-only")
	}
}

func TestAddAll(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(tmpDir, "test.md")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	idx := NewIndex()

	// Add all
	opts := AddOptions{}
	if err := idx.AddAll(tmpDir, ".", opts); err != nil {
		t.Fatalf("failed to add all: %v", err)
	}

	// Verify
	if !idx.HasEntry("test.txt") {
		t.Error("expected test.txt to be in index")
	}
	if !idx.HasEntry("test.md") {
		t.Error("expected test.md to be in index")
	}
}

func TestAddAllPattern(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(tmpDir, "test.md")
	file3 := filepath.Join(tmpDir, "other.txt")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content 3"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	idx := NewIndex()

	// Add with pattern
	opts := AddOptions{}
	if err := idx.AddAll(tmpDir, "*.txt", opts); err != nil {
		t.Fatalf("failed to add with pattern: %v", err)
	}

	// Verify
	if !idx.HasEntry("test.txt") {
		t.Error("expected test.txt to be in index")
	}
	if !idx.HasEntry("other.txt") {
		t.Error("expected other.txt to be in index")
	}
	if idx.HasEntry("test.md") {
		t.Error("expected test.md not to be in index")
	}
}

func TestAddAllRecursive(t *testing.T) {
	// Create temp directory with nested files
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(subdir, "nested.txt")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	idx := NewIndex()

	// Add with recursive pattern
	opts := AddOptions{}
	if err := idx.AddAll(tmpDir, "**/*.txt", opts); err != nil {
		t.Fatalf("failed to add with recursive pattern: %v", err)
	}

	// Verify
	if !idx.HasEntry("test.txt") {
		t.Error("expected test.txt to be in index")
	}
	if !idx.HasEntry("subdir/nested.txt") {
		t.Error("expected subdir/nested.txt to be in index")
	}
}

func TestRemove(t *testing.T) {
	idx := NewIndex()

	// Add entry
	entry := &Entry{
		Path: "test.txt",
		Mode: FileModeRegular,
	}
	idx.AddEntry(entry)

	// Remove
	if err := idx.Remove("test.txt"); err != nil {
		t.Fatalf("failed to remove: %v", err)
	}

	// Verify
	if idx.HasEntry("test.txt") {
		t.Error("expected test.txt not to be in index")
	}
}

func TestRemoveNonexistent(t *testing.T) {
	idx := NewIndex()

	// Try to remove nonexistent entry
	err := idx.Remove("nonexistent.txt")
	if err == nil {
		t.Error("expected error when removing nonexistent entry")
	}
}

func TestRemoveAll(t *testing.T) {
	idx := NewIndex()

	// Add entries
	entry1 := &Entry{Path: "test1.txt", Mode: FileModeRegular}
	entry2 := &Entry{Path: "test2.txt", Mode: FileModeRegular}
	entry3 := &Entry{Path: "other.md", Mode: FileModeRegular}

	idx.AddEntry(entry1)
	idx.AddEntry(entry2)
	idx.AddEntry(entry3)

	// Remove all matching pattern
	if err := idx.RemoveAll("*.txt"); err != nil {
		t.Fatalf("failed to remove all: %v", err)
	}

	// Verify
	if idx.HasEntry("test1.txt") {
		t.Error("expected test1.txt not to be in index")
	}
	if idx.HasEntry("test2.txt") {
		t.Error("expected test2.txt not to be in index")
	}
	if !idx.HasEntry("other.md") {
		t.Error("expected other.md to still be in index")
	}
}
