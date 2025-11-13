package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/index"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// TestGetCommitFullHash tests retrieving a commit by full hash
func TestGetCommitFullHash(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create a test commit
	commitHash := createTestCommitForHistory(t, repo, "test.txt", "content\n", "Test commit", nil)

	// Test full hash lookup
	commit, retrievedHash, err := repo.GetCommit(commitHash.String())
	if err != nil {
		t.Fatalf("Failed to get commit: %v", err)
	}

	if !retrievedHash.Equals(commitHash) {
		t.Errorf("Retrieved hash = %s, want %s", retrievedHash, commitHash)
	}

	if strings.TrimSpace(commit.Message) != "Test commit" {
		t.Errorf("Commit message = %q, want Test commit", commit.Message)
	}
}

// TestGetCommitAbbreviatedHash tests retrieving a commit by abbreviated hash
func TestGetCommitAbbreviatedHash(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create a test commit
	commitHash := createTestCommitForHistory(t, repo, "test.txt", "content\n", "Test commit", nil)

	// Test abbreviated hash lookup (first 7 characters)
	shortHash := commitHash.String()[:7]
	commit, retrievedHash, err := repo.GetCommit(shortHash)
	if err != nil {
		t.Fatalf("Failed to get commit with abbreviated hash: %v", err)
	}

	if !retrievedHash.Equals(commitHash) {
		t.Errorf("Retrieved hash = %s, want %s", retrievedHash, commitHash)
	}

	if strings.TrimSpace(commit.Message) != "Test commit" {
		t.Errorf("Commit message = %q, want Test commit", commit.Message)
	}
}

// TestGetAncestors tests getting all ancestors of a commit
func TestGetAncestors(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create a chain of commits
	commit1 := createTestCommitForHistory(t, repo, "file1.txt", "content1\n", "Commit 1", nil)
	commit2 := createTestCommitForHistory(t, repo, "file2.txt", "content2\n", "Commit 2", []hash.Hash{commit1})
	commit3 := createTestCommitForHistory(t, repo, "file3.txt", "content3\n", "Commit 3", []hash.Hash{commit2})

	// Get ancestors of commit3
	ancestors, err := repo.GetAncestors(commit3)
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	// Should have 2 ancestors (commit2 and commit1)
	if len(ancestors) != 2 {
		t.Errorf("Expected 2 ancestors, got %d", len(ancestors))
	}

	// Check that ancestors include commit1 and commit2
	ancestorMap := make(map[string]bool)
	for _, a := range ancestors {
		ancestorMap[a.String()] = true
	}

	if !ancestorMap[commit1.String()] {
		t.Error("Ancestors should include commit1")
	}
	if !ancestorMap[commit2.String()] {
		t.Error("Ancestors should include commit2")
	}
}

// TestIsAncestor tests the ancestor relationship
func TestIsAncestor(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create a chain of commits
	commit1 := createTestCommitForHistory(t, repo, "file1.txt", "content1\n", "Commit 1", nil)
	commit2 := createTestCommitForHistory(t, repo, "file2.txt", "content2\n", "Commit 2", []hash.Hash{commit1})
	commit3 := createTestCommitForHistory(t, repo, "file3.txt", "content3\n", "Commit 3", []hash.Hash{commit2})

	// Test: commit1 is ancestor of commit3
	isAncestor, err := repo.IsAncestor(commit1, commit3)
	if err != nil {
		t.Fatalf("Failed to check ancestor: %v", err)
	}
	if !isAncestor {
		t.Error("commit1 should be ancestor of commit3")
	}

	// Test: commit3 is NOT ancestor of commit1
	isAncestor, err = repo.IsAncestor(commit3, commit1)
	if err != nil {
		t.Fatalf("Failed to check ancestor: %v", err)
	}
	if isAncestor {
		t.Error("commit3 should not be ancestor of commit1")
	}

	// Test: commit is ancestor of itself
	isAncestor, err = repo.IsAncestor(commit2, commit2)
	if err != nil {
		t.Fatalf("Failed to check ancestor: %v", err)
	}
	if !isAncestor {
		t.Error("commit should be ancestor of itself")
	}
}

// TestLogBasic tests basic log functionality
func TestLogBasic(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create commits
	commit1 := createTestCommitForHistory(t, repo, "file1.txt", "content1\n", "Commit 1", nil)
	commit2 := createTestCommitForHistory(t, repo, "file2.txt", "content2\n", "Commit 2", []hash.Hash{commit1})
	commit3 := createTestCommitForHistory(t, repo, "file3.txt", "content3\n", "Commit 3", []hash.Hash{commit2})

	// Create main branch
	if err := repo.CreateBranch("main", commit3); err != nil {
		t.Fatalf("Failed to create main branch: %v", err)
	}
	repo.SetHEAD("ref: refs/heads/main")

	// Get log
	opts := DefaultLogOptions()
	entries, err := repo.Log("", opts)
	if err != nil {
		t.Fatalf("Failed to get log: %v", err)
	}

	// Should have all 3 commits
	if len(entries) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(entries))
	}
}

// TestLogMaxCount tests log with max count limit
func TestLogMaxCount(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create commits
	commit1 := createTestCommitForHistory(t, repo, "file1.txt", "content1\n", "Commit 1", nil)
	commit2 := createTestCommitForHistory(t, repo, "file2.txt", "content2\n", "Commit 2", []hash.Hash{commit1})
	commit3 := createTestCommitForHistory(t, repo, "file3.txt", "content3\n", "Commit 3", []hash.Hash{commit2})

	// Create main branch
	if err := repo.CreateBranch("main", commit3); err != nil {
		t.Fatalf("Failed to create main branch: %v", err)
	}
	repo.SetHEAD("ref: refs/heads/main")

	// Get log with limit
	opts := DefaultLogOptions()
	opts.MaxCount = 2
	entries, err := repo.Log("", opts)
	if err != nil {
		t.Fatalf("Failed to get log: %v", err)
	}

	// Should have only 2 commits
	if len(entries) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(entries))
	}
}

// TestLogAuthorFilter tests log filtering by author
func TestLogAuthorFilter(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Setup object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create commits (all use same author in our test helper)
	commit1 := createTestCommitForHistory(t, repo, "file1.txt", "content1\n", "Commit 1", nil)

	// Create main branch
	if err := repo.CreateBranch("main", commit1); err != nil {
		t.Fatalf("Failed to create main branch: %v", err)
	}
	repo.SetHEAD("ref: refs/heads/main")

	// Get log with author filter
	opts := DefaultLogOptions()
	opts.Author = "Test User"
	entries, err := repo.Log("", opts)
	if err != nil {
		t.Fatalf("Failed to get log: %v", err)
	}

	// Should match
	if len(entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(entries))
	}

	// Try non-matching author
	opts.Author = "Nonexistent User"
	entries, err = repo.Log("", opts)
	if err != nil {
		t.Fatalf("Failed to get log: %v", err)
	}

	// Should not match
	if len(entries) != 0 {
		t.Errorf("Expected 0 log entries, got %d", len(entries))
	}
}

// TestFormatLogEntry tests log entry formatting
func TestFormatLogEntry(t *testing.T) {
	// Create a test commit
	author := object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	commit := object.NewCommit()
	commit.Message = "Test commit message"
	commit.Author = author
	commit.Committer = author

	h, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	entry := &LogEntry{
		Commit:  commit,
		Hash:    h,
		Refs:    []string{"main", "HEAD"},
		Parents: []hash.Hash{},
	}

	// Test oneline format
	oneline := FormatLogEntry(entry, LogFormatOneline)
	if !contains(oneline, "1234567") {
		t.Error("Oneline format should contain short hash")
	}
	if !contains(oneline, "Test commit message") {
		t.Error("Oneline format should contain message")
	}

	// Test full format
	full := FormatLogEntry(entry, LogFormatFull)
	if !contains(full, "Test User") {
		t.Error("Full format should contain author name")
	}
	if !contains(full, "test@example.com") {
		t.Error("Full format should contain author email")
	}
}

// Helper functions

func createTestCommitForHistory(t *testing.T, repo *Repository, filename, content, message string, parents []hash.Hash) hash.Hash {
	t.Helper()

	// Write file
	filePath := filepath.Join(repo.Path, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		t.Fatalf("Failed to load index: %v", err)
	}

	// Add file to index
	addOpts := index.AddOptions{Force: false, UpdateOnly: false}
	if err := idx.Add(repo.Path, []string{filename}, addOpts); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	// Write blobs
	if err := idx.WriteBlobs(repo.Path, repo.ObjectDB); err != nil {
		t.Fatalf("Failed to write blobs: %v", err)
	}

	// Create commit
	author := index.DefaultSignature("Test User", "test@example.com")
	commitOpts := index.CommitOptions{
		Message:   message,
		Author:    author,
		Committer: author,
		Parents:   parents,
	}

	commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
	if err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	// Save index
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("Failed to save index: %v", err)
	}

	return commitHash
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
