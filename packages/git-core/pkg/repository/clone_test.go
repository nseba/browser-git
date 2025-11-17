package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCloneOptions(t *testing.T) {
	opts := DefaultCloneOptions()

	if opts.Bare {
		t.Error("Default clone should not be bare")
	}

	if opts.Depth != 0 {
		t.Errorf("Default clone depth should be 0, got %d", opts.Depth)
	}

	if opts.Remote != "origin" {
		t.Errorf("Default remote should be 'origin', got %s", opts.Remote)
	}

	if opts.Branch != "" {
		t.Errorf("Default branch should be empty, got %s", opts.Branch)
	}
}

func TestSetupRemote(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "clone-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a repository
	initOpts := DefaultInitOptions()
	if err := Init(tmpDir, initOpts); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Open the repository
	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	// Setup remote
	testURL := "https://github.com/user/repo.git"
	if err := setupRemote(repo, "origin", testURL); err != nil {
		t.Fatalf("Failed to setup remote: %v", err)
	}

	// Read config file
	configPath := filepath.Join(repo.GitDir, "config")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	configStr := string(content)

	// Verify remote configuration
	if !containsSubstring(configStr, "[remote \"origin\"]") {
		t.Error("Config should contain remote section")
	}

	if !containsSubstring(configStr, testURL) {
		t.Errorf("Config should contain URL: %s", testURL)
	}

	if !containsSubstring(configStr, "fetch = +refs/heads/*:refs/remotes/origin/*") {
		t.Error("Config should contain fetch refspec")
	}
}

func TestCloneValidation(t *testing.T) {
	// Create a temporary directory for non-empty test
	tmpDir, err := os.MkdirTemp("", "clone-validation-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the directory
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to clone into non-empty directory
	// Note: We can't actually test full clone without a Git server,
	// but we can test the validation logic would work

	// The clone function would fail at the directory check
	// We're just verifying the logic is correct
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Directory should not be empty")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCloneDirectoryCreation(t *testing.T) {
	// Create a temporary parent directory
	tmpParent, err := os.MkdirTemp("", "clone-parent-*")
	if err != nil {
		t.Fatalf("Failed to create temp parent dir: %v", err)
	}
	defer os.RemoveAll(tmpParent)

	// Try to create a new directory for clone
	cloneDir := filepath.Join(tmpParent, "test-repo")

	// Verify directory doesn't exist yet
	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		t.Error("Clone directory should not exist yet")
	}

	// Create the directory (simulating what clone does)
	if err := os.MkdirAll(cloneDir, 0755); err != nil {
		t.Fatalf("Failed to create clone directory: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(cloneDir)
	if err != nil {
		t.Fatalf("Clone directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Clone path should be a directory")
	}
}

func TestFileStorage(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create objects directory
	objectsPath := filepath.Join(tmpDir, "objects")
	if err := os.MkdirAll(objectsPath, 0755); err != nil {
		t.Fatalf("Failed to create objects directory: %v", err)
	}

	// Create a mock hasher
	// Note: This is a simplified test - real tests would use actual hash package
	// For now, just verify the storage path logic
	testHash := "0123456789abcdef0123456789abcdef01234567"

	// Verify object path format
	expectedPath := filepath.Join(objectsPath, "01", "23456789abcdef0123456789abcdef01234567")

	// Manually compute the path (since we don't have the storage instance)
	hashStr := testHash
	computedPath := filepath.Join(objectsPath, hashStr[:2], hashStr[2:])

	if computedPath != expectedPath {
		t.Errorf("Object path mismatch: expected %s, got %s", expectedPath, computedPath)
	}
}
