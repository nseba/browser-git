package repository

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInit tests repository initialization
func TestInit(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	opts := DefaultInitOptions()
	err := Init(repoPath, opts)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Verify .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if !dirExists(gitDir) {
		t.Errorf(".git directory not created")
	}

	// Verify standard directories
	dirs := []string{
		"objects",
		"objects/info",
		"objects/pack",
		"refs",
		"refs/heads",
		"refs/tags",
		"hooks",
		"info",
	}

	for _, dir := range dirs {
		path := filepath.Join(gitDir, dir)
		if !dirExists(path) {
			t.Errorf("Directory %s not created", dir)
		}
	}

	// Verify HEAD file
	headPath := filepath.Join(gitDir, "HEAD")
	if !fileExists(headPath) {
		t.Errorf("HEAD file not created")
	} else {
		content, err := os.ReadFile(headPath)
		if err != nil {
			t.Errorf("Failed to read HEAD: %v", err)
		}
		expected := "ref: refs/heads/main\n"
		if string(content) != expected {
			t.Errorf("HEAD content = %q, want %q", string(content), expected)
		}
	}

	// Verify config file
	configPath := filepath.Join(gitDir, "config")
	if !fileExists(configPath) {
		t.Errorf("config file not created")
	}
}

// TestInitBare tests bare repository initialization
func TestInitBare(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "bare-repo.git")

	opts := InitOptions{
		Bare:          true,
		InitialBranch: "main",
		HashAlgorithm: "sha1",
	}

	err := Init(repoPath, opts)
	if err != nil {
		t.Fatalf("Failed to initialize bare repository: %v", err)
	}

	// Bare repos don't have .git subdirectory
	gitDir := repoPath

	// Verify HEAD exists directly in repo dir
	headPath := filepath.Join(gitDir, "HEAD")
	if !fileExists(headPath) {
		t.Errorf("HEAD file not created in bare repository")
	}

	// Verify config marks as bare
	config, err := LoadConfig(filepath.Join(gitDir, "config"))
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !config.IsBare() {
		t.Error("Config should indicate bare repository")
	}
}

// TestInitCustomBranch tests initialization with custom initial branch
func TestInitCustomBranch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "custom-branch-repo")

	opts := InitOptions{
		Bare:          false,
		InitialBranch: "develop",
		HashAlgorithm: "sha1",
	}

	err := Init(repoPath, opts)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	gitDir := filepath.Join(repoPath, ".git")
	headPath := filepath.Join(gitDir, "HEAD")

	content, err := os.ReadFile(headPath)
	if err != nil {
		t.Fatalf("Failed to read HEAD: %v", err)
	}

	expected := "ref: refs/heads/develop\n"
	if string(content) != expected {
		t.Errorf("HEAD content = %q, want %q", string(content), expected)
	}
}

// TestInitSHA256 tests initialization with SHA-256
func TestInitSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "sha256-repo")

	opts := InitOptions{
		Bare:          false,
		InitialBranch: "main",
		HashAlgorithm: "sha256",
	}

	err := Init(repoPath, opts)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	gitDir := filepath.Join(repoPath, ".git")
	config, err := LoadConfig(filepath.Join(gitDir, "config"))
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	algo := config.GetHashAlgorithm()
	if algo != "sha256" {
		t.Errorf("Hash algorithm = %s, want sha256", algo)
	}
}

// TestInitAlreadyExists tests that Init fails if repository already exists
func TestInitAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "existing-repo")

	// Initialize once
	opts := DefaultInitOptions()
	err := Init(repoPath, opts)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Try to initialize again
	err = Init(repoPath, opts)
	if err == nil {
		t.Error("Expected error when initializing existing repository")
	}
}

// TestIsRepository tests repository detection
func TestIsRepository(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "is-repo-test")

	// Should not be a repository initially
	if IsRepository(repoPath) {
		t.Error("Path should not be detected as repository before init")
	}

	// Initialize repository
	err := Init(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Should be detected as repository now
	if !IsRepository(repoPath) {
		t.Error("Path should be detected as repository after init")
	}
}

// TestFindRepository tests finding repository in parent directories
func TestFindRepository(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "find-repo-test")

	// Initialize repository
	err := Init(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(repoPath, "sub", "nested")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Find repository from subdirectory
	found, err := FindRepository(subDir)
	if err != nil {
		t.Fatalf("Failed to find repository: %v", err)
	}

	if found != repoPath {
		t.Errorf("Found repository at %s, want %s", found, repoPath)
	}
}

// TestOpen tests opening a repository
func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "open-test")

	// Initialize repository
	err := Init(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Open repository
	repo, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	if repo.Path != repoPath {
		t.Errorf("Repository path = %s, want %s", repo.Path, repoPath)
	}

	gitDir := filepath.Join(repoPath, ".git")
	if repo.GitDir != gitDir {
		t.Errorf("GitDir = %s, want %s", repo.GitDir, gitDir)
	}

	if repo.Config == nil {
		t.Error("Config should not be nil")
	}

	if repo.Hasher == nil {
		t.Error("Hasher should not be nil")
	}
}

// TestCreate tests creating and opening a repository
func TestCreate(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "create-test")

	opts := DefaultInitOptions()
	repo, err := Create(repoPath, opts)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	if repo == nil {
		t.Fatal("Repository should not be nil")
	}

	if !dirExists(filepath.Join(repoPath, ".git")) {
		t.Error(".git directory should exist")
	}
}

// Helper functions

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
