package index

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

func TestNewIndex(t *testing.T) {
	idx := NewIndex()
	if idx.Version != 2 {
		t.Errorf("expected version 2, got %d", idx.Version)
	}
	if len(idx.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(idx.Entries))
	}
}

func TestAddEntry(t *testing.T) {
	idx := NewIndex()

	entry := &Entry{
		Path:  "test.txt",
		Hash:  hash.NewHash([]byte("1234567890123456789012345678901234567890")[:20]),
		Mode:  FileModeRegular,
		Size:  100,
		MTime: time.Now(),
	}

	idx.AddEntry(entry)

	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(idx.Entries))
	}

	if idx.Entries[0].Path != "test.txt" {
		t.Errorf("expected path test.txt, got %s", idx.Entries[0].Path)
	}
}

func TestAddEntryReplaces(t *testing.T) {
	idx := NewIndex()

	// Add first entry
	entry1 := &Entry{
		Path:  "test.txt",
		Hash:  hash.NewHash([]byte("1111111111111111111111111111111111111111")[:20]),
		Mode:  FileModeRegular,
		Size:  100,
		MTime: time.Now(),
	}
	idx.AddEntry(entry1)

	// Add second entry with same path
	entry2 := &Entry{
		Path:  "test.txt",
		Hash:  hash.NewHash([]byte("2222222222222222222222222222222222222222")[:20]),
		Mode:  FileModeRegular,
		Size:  200,
		MTime: time.Now(),
	}
	idx.AddEntry(entry2)

	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(idx.Entries))
	}

	if idx.Entries[0].Size != 200 {
		t.Errorf("expected size 200, got %d", idx.Entries[0].Size)
	}
}

func TestRemoveEntry(t *testing.T) {
	idx := NewIndex()

	entry := &Entry{
		Path:  "test.txt",
		Hash:  hash.NewHash([]byte("1234567890123456789012345678901234567890")[:20]),
		Mode:  FileModeRegular,
		Size:  100,
		MTime: time.Now(),
	}
	idx.AddEntry(entry)

	if !idx.RemoveEntry("test.txt") {
		t.Error("expected RemoveEntry to return true")
	}

	if len(idx.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(idx.Entries))
	}

	if idx.RemoveEntry("nonexistent.txt") {
		t.Error("expected RemoveEntry to return false for nonexistent entry")
	}
}

func TestGetEntry(t *testing.T) {
	idx := NewIndex()

	entry := &Entry{
		Path:  "test.txt",
		Hash:  hash.NewHash([]byte("1234567890123456789012345678901234567890")[:20]),
		Mode:  FileModeRegular,
		Size:  100,
		MTime: time.Now(),
	}
	idx.AddEntry(entry)

	found, ok := idx.GetEntry("test.txt")
	if !ok {
		t.Error("expected to find entry")
	}
	if found.Path != "test.txt" {
		t.Errorf("expected path test.txt, got %s", found.Path)
	}

	_, ok = idx.GetEntry("nonexistent.txt")
	if ok {
		t.Error("expected not to find nonexistent entry")
	}
}

func TestHasEntry(t *testing.T) {
	idx := NewIndex()

	entry := &Entry{
		Path:  "test.txt",
		Hash:  hash.NewHash([]byte("1234567890123456789012345678901234567890")[:20]),
		Mode:  FileModeRegular,
		Size:  100,
		MTime: time.Now(),
	}
	idx.AddEntry(entry)

	if !idx.HasEntry("test.txt") {
		t.Error("expected HasEntry to return true")
	}

	if idx.HasEntry("nonexistent.txt") {
		t.Error("expected HasEntry to return false for nonexistent entry")
	}
}

func TestSort(t *testing.T) {
	idx := NewIndex()

	// Add entries in reverse order
	entry1 := &Entry{Path: "c.txt", Hash: hash.NewHash(make([]byte, 20)), Mode: FileModeRegular}
	entry2 := &Entry{Path: "b.txt", Hash: hash.NewHash(make([]byte, 20)), Mode: FileModeRegular}
	entry3 := &Entry{Path: "a.txt", Hash: hash.NewHash(make([]byte, 20)), Mode: FileModeRegular}

	idx.Entries = append(idx.Entries, entry1, entry2, entry3)
	idx.Sort()

	expectedOrder := []string{"a.txt", "b.txt", "c.txt"}
	for i, expected := range expectedOrder {
		if idx.Entries[i].Path != expected {
			t.Errorf("expected entry %d to be %s, got %s", i, expected, idx.Entries[i].Path)
		}
	}
}

func TestSerializeDeserialize(t *testing.T) {
	idx := NewIndex()

	// Add some entries
	hasher := hash.NewSHA1()
	entry1 := &Entry{
		Path:  "test1.txt",
		Hash:  hash.HashBlob(hasher, []byte("test content 1")),
		Mode:  FileModeRegular,
		Size:  14,
		MTime: time.Now().Truncate(time.Second), // Truncate to avoid precision issues
	}
	entry2 := &Entry{
		Path:  "test2.txt",
		Hash:  hash.HashBlob(hasher, []byte("test content 2")),
		Mode:  FileModeExecutable,
		Size:  14,
		MTime: time.Now().Truncate(time.Second),
	}

	idx.AddEntry(entry1)
	idx.AddEntry(entry2)

	// Serialize
	var buf bytes.Buffer
	if err := idx.Serialize(&buf); err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize
	deserialized, err := Deserialize(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Verify
	if deserialized.Version != idx.Version {
		t.Errorf("expected version %d, got %d", idx.Version, deserialized.Version)
	}

	if len(deserialized.Entries) != len(idx.Entries) {
		t.Fatalf("expected %d entries, got %d", len(idx.Entries), len(deserialized.Entries))
	}

	for i := range idx.Entries {
		if deserialized.Entries[i].Path != idx.Entries[i].Path {
			t.Errorf("entry %d: expected path %s, got %s", i, idx.Entries[i].Path, deserialized.Entries[i].Path)
		}
		if !deserialized.Entries[i].Hash.Equals(idx.Entries[i].Hash) {
			t.Errorf("entry %d: hash mismatch", i)
		}
		if deserialized.Entries[i].Mode != idx.Entries[i].Mode {
			t.Errorf("entry %d: expected mode %o, got %o", i, idx.Entries[i].Mode, deserialized.Entries[i].Mode)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index")

	idx := NewIndex()

	// Add entry
	hasher := hash.NewSHA1()
	entry := &Entry{
		Path:  "test.txt",
		Hash:  hash.HashBlob(hasher, []byte("test content")),
		Mode:  FileModeRegular,
		Size:  12,
		MTime: time.Now().Truncate(time.Second),
	}
	idx.AddEntry(entry)

	// Save
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load
	loaded, err := Load(indexPath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}

	if loaded.Entries[0].Path != "test.txt" {
		t.Errorf("expected path test.txt, got %s", loaded.Entries[0].Path)
	}
}

func TestLoadNonexistent(t *testing.T) {
	// Try to load from nonexistent path
	idx, err := Load("/nonexistent/index")
	if err != nil {
		t.Fatalf("expected no error for nonexistent index, got %v", err)
	}

	// Should return empty index
	if len(idx.Entries) != 0 {
		t.Errorf("expected empty index, got %d entries", len(idx.Entries))
	}
}

func TestNewEntryFromFile(t *testing.T) {
	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create entry from file
	entry, err := NewEntryFromFile("test.txt", tmpDir)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Verify
	if entry.Path != "test.txt" {
		t.Errorf("expected path test.txt, got %s", entry.Path)
	}

	if entry.Size != uint32(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), entry.Size)
	}

	if entry.Mode != FileModeRegular {
		t.Errorf("expected mode %o, got %o", FileModeRegular, entry.Mode)
	}

	// Verify hash
	hasher := hash.NewSHA1()
	expectedHash := hash.HashBlob(hasher, content)
	if !entry.Hash.Equals(expectedHash) {
		t.Error("hash mismatch")
	}
}

func TestNewEntryFromExecutableFile(t *testing.T) {
	// Create temp directory and executable file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.sh")
	content := []byte("#!/bin/bash\necho test")

	if err := os.WriteFile(testFile, content, 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create entry from file
	entry, err := NewEntryFromFile("test.sh", tmpDir)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Verify
	if entry.Mode != FileModeExecutable {
		t.Errorf("expected mode %o, got %o", FileModeExecutable, entry.Mode)
	}
}

func TestIsModified(t *testing.T) {
	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create entry from file
	entry, err := NewEntryFromFile("test.txt", tmpDir)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// File should not be modified initially
	modified, err := entry.IsModified(tmpDir)
	if err != nil {
		t.Fatalf("failed to check if modified: %v", err)
	}
	if modified {
		t.Error("expected file not to be modified")
	}

	// Modify file
	time.Sleep(10 * time.Millisecond) // Ensure different mtime
	newContent := []byte("modified content")
	if err := os.WriteFile(testFile, newContent, 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// File should now be modified
	modified, err = entry.IsModified(tmpDir)
	if err != nil {
		t.Fatalf("failed to check if modified: %v", err)
	}
	if !modified {
		t.Error("expected file to be modified")
	}
}

func TestIsModifiedDeleted(t *testing.T) {
	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create entry from file
	entry, err := NewEntryFromFile("test.txt", tmpDir)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Delete file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	// File should be marked as modified (deleted)
	modified, err := entry.IsModified(tmpDir)
	if err != nil {
		t.Fatalf("failed to check if modified: %v", err)
	}
	if !modified {
		t.Error("expected deleted file to be marked as modified")
	}
}

func TestClear(t *testing.T) {
	idx := NewIndex()

	// Add some entries
	entry1 := &Entry{Path: "test1.txt", Hash: hash.NewHash(make([]byte, 20)), Mode: FileModeRegular}
	entry2 := &Entry{Path: "test2.txt", Hash: hash.NewHash(make([]byte, 20)), Mode: FileModeRegular}

	idx.AddEntry(entry1)
	idx.AddEntry(entry2)

	if len(idx.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(idx.Entries))
	}

	// Clear
	idx.Clear()

	if len(idx.Entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(idx.Entries))
	}
}

func TestEntryCount(t *testing.T) {
	idx := NewIndex()

	if idx.EntryCount() != 0 {
		t.Errorf("expected count 0, got %d", idx.EntryCount())
	}

	entry := &Entry{Path: "test.txt", Hash: hash.NewHash(make([]byte, 20)), Mode: FileModeRegular}
	idx.AddEntry(entry)

	if idx.EntryCount() != 1 {
		t.Errorf("expected count 1, got %d", idx.EntryCount())
	}
}
