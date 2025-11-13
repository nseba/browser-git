package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/index"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// MemoryStorage is a simple in-memory storage for testing
type MemoryStorage struct {
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

func (m *MemoryStorage) Read(h hash.Hash) ([]byte, error) {
	data, ok := m.data[h.String()]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", h.String())
	}
	return data, nil
}

func (m *MemoryStorage) Write(h hash.Hash, data []byte) error {
	m.data[h.String()] = data
	return nil
}

func (m *MemoryStorage) Has(h hash.Hash) bool {
	_, ok := m.data[h.String()]
	return ok
}

func (m *MemoryStorage) Delete(h hash.Hash) error {
	delete(m.data, h.String())
	return nil
}

func (m *MemoryStorage) List() ([]hash.Hash, error) {
	hashes := make([]hash.Hash, 0, len(m.data))
	for hashStr := range m.data {
		h, err := hash.ParseHash(hashStr)
		if err != nil {
			continue
		}
		hashes = append(hashes, h)
	}
	return hashes, nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

// TestBranchLifecycleIntegration tests the complete lifecycle of branches
// from initialization through creation, switching, renaming, and deletion
func TestBranchLifecycleIntegration(t *testing.T) {
	// Setup: Create a temporary directory
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "branch-lifecycle-test")

	// Step 1: Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Verify initial state
	branches, err := repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if len(branches) != 0 {
		t.Errorf("Expected 0 branches initially, got %d", len(branches))
	}

	// Step 2: Create first commit on main branch
	// Create a test file
	testFilePath := filepath.Join(repoPath, "test.txt")
	testContent := []byte("Hello, World!\n")
	if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		t.Fatalf("Failed to load index: %v", err)
	}

	// Add file to index
	addOpts := index.AddOptions{Force: false, UpdateOnly: false}
	if err := idx.Add(repoPath, []string{"test.txt"}, addOpts); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	// Save index
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("Failed to save index: %v", err)
	}

	// Write blobs
	if err := idx.WriteBlobs(repoPath, repo.ObjectDB); err != nil {
		t.Fatalf("Failed to write blobs: %v", err)
	}

	// Create commit
	author := index.DefaultSignature("Test User", "test@example.com")
	commitOpts := index.CommitOptions{
		Message:   "Initial commit",
		Author:    author,
		Committer: author,
		Parents:   nil,
	}

	commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
	if err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	// Update main branch
	if err := repo.CreateBranch("main", commitHash); err != nil {
		t.Fatalf("Failed to create main branch: %v", err)
	}

	// Set HEAD to main
	if err := repo.SetHEAD("ref: refs/heads/main"); err != nil {
		t.Fatalf("Failed to set HEAD: %v", err)
	}

	// Verify main branch exists
	if !repo.BranchExists("main") {
		t.Error("main branch should exist after creation")
	}

	// Verify current branch is main
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if currentBranch != "main" {
		t.Errorf("Current branch = %s, want main", currentBranch)
	}

	// Step 3: Create feature branch from main
	if err := repo.CreateBranch("feature", commitHash); err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
	}

	// Verify feature branch exists
	if !repo.BranchExists("feature") {
		t.Error("feature branch should exist after creation")
	}

	// Verify both branches exist
	branches, err = repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("Expected 2 branches, got %d", len(branches))
	}

	// Step 4: Switch to feature branch (simulate checkout)
	if err := repo.SetHEAD("ref: refs/heads/feature"); err != nil {
		t.Fatalf("Failed to switch to feature branch: %v", err)
	}

	// Verify current branch is feature
	currentBranch, err = repo.CurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if currentBranch != "feature" {
		t.Errorf("Current branch = %s, want feature", currentBranch)
	}

	// Step 5: Rename feature branch to development
	if err := repo.RenameBranch("feature", "development"); err != nil {
		t.Fatalf("Failed to rename branch: %v", err)
	}

	// Verify old name doesn't exist
	if repo.BranchExists("feature") {
		t.Error("feature branch should not exist after rename")
	}

	// Verify new name exists
	if !repo.BranchExists("development") {
		t.Error("development branch should exist after rename")
	}

	// Verify current branch is updated
	currentBranch, err = repo.CurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if currentBranch != "development" {
		t.Errorf("Current branch = %s, want development", currentBranch)
	}

	// Step 6: Create another branch for deletion test
	if err := repo.CreateBranch("temp", commitHash); err != nil {
		t.Fatalf("Failed to create temp branch: %v", err)
	}

	// Verify 3 branches exist now
	branches, err = repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if len(branches) != 3 {
		t.Errorf("Expected 3 branches, got %d", len(branches))
	}

	// Step 7: Delete temp branch (not current branch)
	if err := repo.DeleteBranch("temp"); err != nil {
		t.Fatalf("Failed to delete temp branch: %v", err)
	}

	// Verify temp branch is deleted
	if repo.BranchExists("temp") {
		t.Error("temp branch should not exist after deletion")
	}

	// Verify 2 branches remain
	branches, err = repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("Expected 2 branches, got %d", len(branches))
	}

	// Step 8: Try to delete current branch (should fail)
	err = repo.DeleteBranch("development")
	if err == nil {
		t.Error("Expected error when deleting current branch")
	}

	// Step 9: Switch back to main
	if err := repo.SetHEAD("ref: refs/heads/main"); err != nil {
		t.Fatalf("Failed to switch to main: %v", err)
	}

	// Step 10: Now delete development branch (should succeed)
	if err := repo.DeleteBranch("development"); err != nil {
		t.Fatalf("Failed to delete development branch: %v", err)
	}

	// Verify only main branch remains
	branches, err = repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if len(branches) != 1 {
		t.Errorf("Expected 1 branch, got %d", len(branches))
	}
	if branches[0] != "main" {
		t.Errorf("Expected main branch, got %s", branches[0])
	}

	// Step 11: Verify all branch operations maintained hash integrity
	mainHash, err := repo.GetBranch("main")
	if err != nil {
		t.Fatalf("Failed to get main branch hash: %v", err)
	}
	if !mainHash.Equals(commitHash) {
		t.Errorf("main branch hash = %s, want %s", mainHash, commitHash)
	}
}

// TestMultipleBranchesIntegration tests working with multiple branches
func TestMultipleBranchesIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "multi-branch-test")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create a commit hash for testing
	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")

	// Create multiple branches
	branchNames := []string{"main", "develop", "feature-1", "feature-2", "hotfix"}
	for _, name := range branchNames {
		if err := repo.CreateBranch(name, testHash); err != nil {
			t.Fatalf("Failed to create branch %s: %v", name, err)
		}
	}

	// List all branches
	branches, err := repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	if len(branches) != len(branchNames) {
		t.Errorf("Expected %d branches, got %d", len(branchNames), len(branches))
	}

	// Verify all expected branches exist
	branchMap := make(map[string]bool)
	for _, branch := range branches {
		branchMap[branch] = true
	}

	for _, expected := range branchNames {
		if !branchMap[expected] {
			t.Errorf("Expected branch %s not found", expected)
		}
	}

	// Verify each branch points to the same hash
	for _, name := range branchNames {
		branchHash, err := repo.GetBranch(name)
		if err != nil {
			t.Fatalf("Failed to get branch %s: %v", name, err)
		}
		if !branchHash.Equals(testHash) {
			t.Errorf("Branch %s hash = %s, want %s", name, branchHash, testHash)
		}
	}

	// Delete all but one branch
	repo.SetHEAD("ref: refs/heads/main")
	for _, name := range []string{"develop", "feature-1", "feature-2", "hotfix"} {
		if err := repo.DeleteBranch(name); err != nil {
			t.Fatalf("Failed to delete branch %s: %v", name, err)
		}
	}

	// Verify only main branch remains
	branches, err = repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	if len(branches) != 1 || branches[0] != "main" {
		t.Errorf("Expected only main branch, got %v", branches)
	}
}

// TestBranchWithCommitsIntegration tests branches with actual commits
func TestBranchWithCommitsIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "branch-commits-test")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Helper function to create a commit
	createCommit := func(filename, content, message string, parents []hash.Hash) hash.Hash {
		// Write file
		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Load/create index
		indexPath := filepath.Join(repo.GitDir, "index")
		idx, err := index.Load(indexPath)
		if err != nil {
			t.Fatalf("Failed to load index: %v", err)
		}

		// Add file to index
		addOpts := index.AddOptions{Force: false, UpdateOnly: false}
		if err := idx.Add(repoPath, []string{filename}, addOpts); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		// Save index
		if err := idx.Save(indexPath); err != nil {
			t.Fatalf("Failed to save index: %v", err)
		}

		// Write blobs
		if err := idx.WriteBlobs(repoPath, repo.ObjectDB); err != nil {
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

		return commitHash
	}

	// Create first commit
	commit1 := createCommit("file1.txt", "content 1\n", "First commit", nil)

	// Create main branch at first commit
	if err := repo.CreateBranch("main", commit1); err != nil {
		t.Fatalf("Failed to create main branch: %v", err)
	}
	repo.SetHEAD("ref: refs/heads/main")

	// Create second commit on main
	commit2 := createCommit("file2.txt", "content 2\n", "Second commit", []hash.Hash{commit1})
	if err := repo.UpdateRef("refs/heads/main", commit2); err != nil {
		t.Fatalf("Failed to update main: %v", err)
	}

	// Create feature branch from first commit
	if err := repo.CreateBranch("feature", commit1); err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
	}

	// Verify main points to commit2
	mainHash, err := repo.GetBranch("main")
	if err != nil {
		t.Fatalf("Failed to get main: %v", err)
	}
	if !mainHash.Equals(commit2) {
		t.Errorf("main hash = %s, want %s", mainHash, commit2)
	}

	// Verify feature points to commit1
	featureHash, err := repo.GetBranch("feature")
	if err != nil {
		t.Fatalf("Failed to get feature: %v", err)
	}
	if !featureHash.Equals(commit1) {
		t.Errorf("feature hash = %s, want %s", featureHash, commit1)
	}

	// Switch to feature and create a commit
	repo.SetHEAD("ref: refs/heads/feature")
	commit3 := createCommit("file3.txt", "content 3\n", "Third commit on feature", []hash.Hash{commit1})
	if err := repo.UpdateRef("refs/heads/feature", commit3); err != nil {
		t.Fatalf("Failed to update feature: %v", err)
	}

	// Verify feature now points to commit3
	featureHash, err = repo.GetBranch("feature")
	if err != nil {
		t.Fatalf("Failed to get feature: %v", err)
	}
	if !featureHash.Equals(commit3) {
		t.Errorf("feature hash = %s, want %s", featureHash, commit3)
	}

	// Verify main still points to commit2
	mainHash, err = repo.GetBranch("main")
	if err != nil {
		t.Fatalf("Failed to get main: %v", err)
	}
	if !mainHash.Equals(commit2) {
		t.Errorf("main hash = %s, want %s", mainHash, commit2)
	}
}
