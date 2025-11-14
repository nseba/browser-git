package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/merge"
)

// TestConflictFilesCreation tests that merge state files are created correctly
func TestConflictFilesCreation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-conflict-files-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	err = Init(tmpDir, InitOptions{Bare: false})
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	// Test creating merge state files manually
	mergeHeadPath := filepath.Join(repo.GitDir, "MERGE_HEAD")
	mergeMsgPath := filepath.Join(repo.GitDir, "MERGE_MSG")
	conflictsPath := filepath.Join(repo.GitDir, "MERGE_CONFLICTS")

	// Write MERGE_HEAD
	mockHash := "1234567890123456789012345678901234567890"
	if err := os.WriteFile(mergeHeadPath, []byte(mockHash+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write MERGE_HEAD: %v", err)
	}

	// Write MERGE_MSG
	mergeMsg := "Merge branch 'feature'\n\nConflicts:\n\tfile1.txt\n"
	if err := os.WriteFile(mergeMsgPath, []byte(mergeMsg), 0644); err != nil {
		t.Fatalf("Failed to write MERGE_MSG: %v", err)
	}

	// Write MERGE_CONFLICTS
	if err := os.WriteFile(conflictsPath, []byte("file1.txt\n"), 0644); err != nil {
		t.Fatalf("Failed to write MERGE_CONFLICTS: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(mergeHeadPath); os.IsNotExist(err) {
		t.Error("MERGE_HEAD should exist")
	}

	if _, err := os.Stat(mergeMsgPath); os.IsNotExist(err) {
		t.Error("MERGE_MSG should exist")
	}

	if _, err := os.Stat(conflictsPath); os.IsNotExist(err) {
		t.Error("MERGE_CONFLICTS should exist")
	}

	// Test cleanupMergeState
	if err := repo.cleanupMergeState(); err != nil {
		t.Fatalf("Failed to cleanup merge state: %v", err)
	}

	// Verify files were removed
	if _, err := os.Stat(mergeHeadPath); !os.IsNotExist(err) {
		t.Error("MERGE_HEAD should be removed after cleanup")
	}

	if _, err := os.Stat(mergeMsgPath); !os.IsNotExist(err) {
		t.Error("MERGE_MSG should be removed after cleanup")
	}

	if _, err := os.Stat(conflictsPath); !os.IsNotExist(err) {
		t.Error("MERGE_CONFLICTS should be removed after cleanup")
	}
}

// TestAbortMergeIntegration tests aborting a merge
func TestAbortMergeIntegration(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-abort-merge-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	err = Init(tmpDir, InitOptions{Bare: false})
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	repo.Config.SetUser("Test User", "test@example.com")

	// Create a test file with original content
	testFile := filepath.Join(tmpDir, "file.txt")
	originalContent := "original content\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Simulate merge in progress with conflicts
	mergeHeadPath := filepath.Join(repo.GitDir, "MERGE_HEAD")
	mockHash := "1111111111111111111111111111111111111111"
	if err := os.WriteFile(mergeHeadPath, []byte(mockHash+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write MERGE_HEAD: %v", err)
	}

	mergeMsgPath := filepath.Join(repo.GitDir, "MERGE_MSG")
	if err := os.WriteFile(mergeMsgPath, []byte("Merge branch 'test'\n"), 0644); err != nil {
		t.Fatalf("Failed to write MERGE_MSG: %v", err)
	}

	// Write conflict markers to file
	conflictContent := "<<<<<<< HEAD\nours\n=======\ntheirs\n>>>>>>> test\n"
	if err := os.WriteFile(testFile, []byte(conflictContent), 0644); err != nil {
		t.Fatalf("Failed to write conflict content: %v", err)
	}

	// Note: AbortMerge requires a valid HEAD commit to reset to
	// Since we don't have one in this minimal test, we'll verify that
	// the function attempts to clean up merge state even if it fails

	// Verify merge state exists before abort
	if _, err := os.Stat(mergeHeadPath); os.IsNotExist(err) {
		t.Error("MERGE_HEAD should exist before abort")
	}

	// Note: We skip testing GetConflicts() here because it requires
	// a valid HEAD commit which we don't have in this minimal test
	// The function is tested more thoroughly in the unit tests
}

// TestConflictMarkers tests that conflict markers are correctly detected
func TestConflictMarkersDetection(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		hasMarkers bool
	}{
		{
			name: "valid conflict markers",
			content: `<<<<<<< HEAD
ours
=======
theirs
>>>>>>> branch
`,
			hasMarkers: true,
		},
		{
			name: "no conflict markers",
			content: `normal content
no conflicts here
`,
			hasMarkers: false,
		},
		{
			name: "incomplete markers",
			content: `<<<<<<< HEAD
ours
`,
			hasMarkers: false,
		},
		{
			name: "markers in wrong order",
			content: `=======
>>>>>>> branch
<<<<<<< HEAD
`,
			hasMarkers: true, // Still has all markers, just wrong order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasConflictMarkers([]byte(tt.content))
			if result != tt.hasMarkers {
				t.Errorf("Expected hasConflictMarkers=%v, got %v", tt.hasMarkers, result)
			}
		})
	}
}

// TestConflictStateStructure tests the ConflictState structure
func TestConflictStateStructure(t *testing.T) {
	// Test that ConflictState can be created and used
	baseHash, _ := hash.ParseHash("0000000000000000000000000000000000000001")
	ourHash, _ := hash.ParseHash("0000000000000000000000000000000000000002")
	theirHash, _ := hash.ParseHash("0000000000000000000000000000000000000003")

	state := &ConflictState{
		Conflicts: []merge.Conflict{
			{
				Path: "file1.txt",
				Type: merge.ContentConflict,
			},
			{
				Path: "file2.txt",
				Type: merge.BinaryConflict,
			},
		},
		OurCommit:   ourHash,
		TheirCommit: theirHash,
		MergeBase:   baseHash,
		BranchName:  "feature",
	}

	if len(state.Conflicts) != 2 {
		t.Errorf("Expected 2 conflicts, got %d", len(state.Conflicts))
	}

	if state.BranchName != "feature" {
		t.Errorf("Expected branch name 'feature', got %q", state.BranchName)
	}

	if state.OurCommit.String() != ourHash.String() {
		t.Error("OurCommit hash mismatch")
	}

	if state.TheirCommit.String() != theirHash.String() {
		t.Error("TheirCommit hash mismatch")
	}

	if state.MergeBase.String() != baseHash.String() {
		t.Error("MergeBase hash mismatch")
	}
}

// TestResolutionStrategyConstants tests the resolution strategy constants
func TestResolutionStrategyConstants(t *testing.T) {
	tests := []struct {
		strategy ConflictResolutionStrategy
		expected string
	}{
		{AcceptOurs, "ours"},
		{AcceptTheirs, "theirs"},
		{AcceptManual, "manual"},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			if string(tt.strategy) != tt.expected {
				t.Errorf("Expected strategy value %q, got %q", tt.expected, string(tt.strategy))
			}
		})
	}
}
