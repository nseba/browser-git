package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nseba/browser-git/git-core/pkg/object"
)

// TestMergeFastForward tests a fast-forward merge scenario
func TestMergeFastForward(t *testing.T) {
	// Create temporary directory for test repository
	tmpDir, err := os.MkdirTemp("", "test-merge-ff-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	opts := DefaultInitOptions()
	if err := Init(tmpDir, opts); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	// Create initial commit on main
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Add and commit
	if err := addFile(repo, "file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	commit1Hash, err := createCommit(repo, "Initial commit")
	if err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Create feature branch
	if err := repo.CreateBranch("feature", commit1Hash); err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
	}

	// Switch to feature branch
	if err := switchBranch(repo, "feature"); err != nil {
		t.Fatalf("Failed to switch to feature branch: %v", err)
	}

	// Make changes on feature branch
	if err := os.WriteFile(testFile, []byte("initial content\nfeature change\n"), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	if err := addFile(repo, "file.txt"); err != nil {
		t.Fatalf("Failed to add updated file: %v", err)
	}

	commit2Hash, err := createCommit(repo, "Feature commit")
	if err != nil {
		t.Fatalf("Failed to create feature commit: %v", err)
	}

	// Switch back to main
	if err := switchBranch(repo, "main"); err != nil {
		t.Fatalf("Failed to switch back to main: %v", err)
	}

	// Merge feature into main (should be fast-forward)
	result, err := repo.Merge("feature", DefaultMergeOptions())
	if err != nil {
		t.Fatalf("Failed to merge: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected merge to succeed")
	}

	if !result.IsFastForward {
		t.Error("Expected fast-forward merge")
	}

	if result.CommitHash.String() != commit2Hash.String() {
		t.Errorf("Expected commit hash %s, got %s", commit2Hash.String(), result.CommitHash.String())
	}

	// Verify file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "initial content\nfeature change\n"
	if string(content) != expected {
		t.Errorf("Expected file content:\n%s\nGot:\n%s", expected, string(content))
	}
}

// TestMergeThreeWayNoConflict tests a three-way merge without conflicts
func TestMergeThreeWayNoConflict(t *testing.T) {
	// Create temporary directory for test repository
	tmpDir, err := os.MkdirTemp("", "test-merge-3way-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	opts := DefaultInitOptions()
	if err := Init(tmpDir, opts); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("line 1\nline 2\nline 3\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	if err := addFile(repo, "file.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	baseCommitHash, err := createCommit(repo, "Initial commit")
	if err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Create feature branch
	if err := repo.CreateBranch("feature", baseCommitHash); err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
	}

	// Make changes on main
	if err := os.WriteFile(testFile, []byte("line 1\nmodified line 2\nline 3\n"), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	if err := addFile(repo, "file.txt"); err != nil {
		t.Fatalf("Failed to add updated file: %v", err)
	}

	if _, err := createCommit(repo, "Main commit"); err != nil {
		t.Fatalf("Failed to create main commit: %v", err)
	}

	// Switch to feature branch
	if err := switchBranch(repo, "feature"); err != nil {
		t.Fatalf("Failed to switch to feature branch: %v", err)
	}

	// Make different changes on feature
	if err := os.WriteFile(testFile, []byte("line 1\nline 2\nline 3\nline 4\n"), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	if err := addFile(repo, "file.txt"); err != nil {
		t.Fatalf("Failed to add updated file: %v", err)
	}

	if _, err := createCommit(repo, "Feature commit"); err != nil {
		t.Fatalf("Failed to create feature commit: %v", err)
	}

	// Switch back to main
	if err := switchBranch(repo, "main"); err != nil {
		t.Fatalf("Failed to switch back to main: %v", err)
	}

	// Merge feature into main
	mergeOpts := DefaultMergeOptions()
	mergeOpts.AllowFastForward = false // Force three-way merge
	result, err := repo.Merge("feature", mergeOpts)
	if err != nil {
		t.Fatalf("Failed to merge: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected merge to succeed")
	}

	if result.IsFastForward {
		t.Error("Expected three-way merge, not fast-forward")
	}

	// Verify merge commit has two parents
	mergeCommit, err := repo.ObjectDB.ReadObject(result.CommitHash)
	if err != nil {
		t.Fatalf("Failed to read merge commit: %v", err)
	}

	commit, ok := mergeCommit.(*object.Commit)
	if !ok {
		t.Fatal("Expected commit object")
	}

	if len(commit.Parents) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(commit.Parents))
	}

	// Verify file content (should have both changes)
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "line 1\nmodified line 2\nline 3\nline 4\n"
	if string(content) != expected {
		t.Errorf("Expected file content:\n%s\nGot:\n%s", expected, string(content))
	}
}

// Helper functions for testing

func addFile(repo *Repository, path string) error {
	// This is a placeholder - actual implementation would use index.Add
	// For now, we'll skip this in the test
	return nil
}

func createCommit(repo *Repository, message string) (hash.Hash, error) {
	// This is a placeholder - actual implementation would create a real commit
	// For now, we'll skip this in the test
	return nil, nil
}

func switchBranch(repo *Repository, branchName string) error {
	// This is a placeholder - actual implementation would use checkout
	// For now, we'll skip this in the test
	return nil
}
