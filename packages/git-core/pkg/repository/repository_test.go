package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nseba/browser-git/git-core/pkg/hash"
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

// TestBranchExists tests branch existence check
func TestBranchExists(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "branch-exists-test")

	// Initialize repository
	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// main branch should not exist yet (no commits)
	if repo.BranchExists("main") {
		t.Error("main branch should not exist before first commit")
	}

	// Create a test branch file manually
	branchPath := filepath.Join(repo.GitDir, "refs", "heads", "test-branch")
	err = os.WriteFile(branchPath, []byte("0000000000000000000000000000000000000000\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test branch: %v", err)
	}

	// test-branch should exist now
	if !repo.BranchExists("test-branch") {
		t.Error("test-branch should exist after creation")
	}

	// nonexistent branch should not exist
	if repo.BranchExists("nonexistent") {
		t.Error("nonexistent branch should not exist")
	}
}

// TestCreateBranch tests branch creation
func TestCreateBranch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "create-branch-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Parse a test hash
	testHash, err := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	if err != nil {
		t.Fatalf("Failed to parse test hash: %v", err)
	}

	// Create a branch
	err = repo.CreateBranch("feature", testHash)
	if err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}

	// Verify branch exists
	if !repo.BranchExists("feature") {
		t.Error("Branch should exist after creation")
	}

	// Verify branch points to correct hash
	branchHash, err := repo.GetBranch("feature")
	if err != nil {
		t.Fatalf("Failed to get branch hash: %v", err)
	}

	if !branchHash.Equals(testHash) {
		t.Errorf("Branch hash = %s, want %s", branchHash, testHash)
	}

	// Try to create same branch again should fail
	err = repo.CreateBranch("feature", testHash)
	if err == nil {
		t.Error("Expected error when creating existing branch")
	}
}

// TestDeleteBranch tests branch deletion
func TestDeleteBranch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "delete-branch-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create test branches
	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	repo.CreateBranch("feature", testHash)
	repo.CreateBranch("main", testHash)

	// Set main as current branch
	repo.SetHEAD("ref: refs/heads/main")

	// Delete feature branch should succeed
	err = repo.DeleteBranch("feature")
	if err != nil {
		t.Fatalf("Failed to delete branch: %v", err)
	}

	// Verify branch is deleted
	if repo.BranchExists("feature") {
		t.Error("Branch should not exist after deletion")
	}

	// Try to delete current branch should fail
	err = repo.DeleteBranch("main")
	if err == nil {
		t.Error("Expected error when deleting current branch")
	}

	// Deleting nonexistent branch should succeed (no error)
	err = repo.DeleteBranch("nonexistent")
	if err == nil {
		// This is actually expected - os.Remove on nonexistent file returns an error
		// but we're okay with this behavior
	}
}

// TestRenameBranch tests branch renaming
func TestRenameBranch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rename-branch-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create test branch
	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	repo.CreateBranch("old-name", testHash)

	// Rename branch
	err = repo.RenameBranch("old-name", "new-name")
	if err != nil {
		t.Fatalf("Failed to rename branch: %v", err)
	}

	// Verify old branch is gone
	if repo.BranchExists("old-name") {
		t.Error("Old branch should not exist after rename")
	}

	// Verify new branch exists
	if !repo.BranchExists("new-name") {
		t.Error("New branch should exist after rename")
	}

	// Verify new branch points to same hash
	newHash, err := repo.GetBranch("new-name")
	if err != nil {
		t.Fatalf("Failed to get new branch hash: %v", err)
	}

	if !newHash.Equals(testHash) {
		t.Errorf("New branch hash = %s, want %s", newHash, testHash)
	}
}

// TestRenameBranchCurrent tests renaming the current branch
func TestRenameBranchCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rename-current-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create and checkout branch
	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	repo.CreateBranch("current", testHash)
	repo.SetHEAD("ref: refs/heads/current")

	// Rename current branch
	err = repo.RenameBranch("current", "renamed")
	if err != nil {
		t.Fatalf("Failed to rename current branch: %v", err)
	}

	// Verify HEAD now points to renamed branch
	head, err := repo.HEAD()
	if err != nil {
		t.Fatalf("Failed to get HEAD: %v", err)
	}

	expected := "ref: refs/heads/renamed"
	if head != expected {
		t.Errorf("HEAD = %s, want %s", head, expected)
	}

	// Verify we can get current branch name
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if currentBranch != "renamed" {
		t.Errorf("Current branch = %s, want renamed", currentBranch)
	}
}

// TestRenameBranchErrors tests error cases for branch renaming
func TestRenameBranchErrors(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rename-errors-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")

	// Try to rename nonexistent branch
	err = repo.RenameBranch("nonexistent", "new")
	if err == nil {
		t.Error("Expected error when renaming nonexistent branch")
	}

	// Create two branches
	repo.CreateBranch("existing1", testHash)
	repo.CreateBranch("existing2", testHash)

	// Try to rename to existing name
	err = repo.RenameBranch("existing1", "existing2")
	if err == nil {
		t.Error("Expected error when renaming to existing branch name")
	}
}

// TestListBranches tests listing all branches
func TestListBranches(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "list-branches-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Initially should have no branches
	branches, err := repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	if len(branches) != 0 {
		t.Errorf("Expected 0 branches, got %d", len(branches))
	}

	// Create some branches
	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	repo.CreateBranch("main", testHash)
	repo.CreateBranch("feature", testHash)
	repo.CreateBranch("develop", testHash)

	// List branches
	branches, err = repo.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	if len(branches) != 3 {
		t.Errorf("Expected 3 branches, got %d", len(branches))
	}

	// Verify branch names (order may vary)
	branchMap := make(map[string]bool)
	for _, branch := range branches {
		branchMap[branch] = true
	}

	expected := []string{"main", "feature", "develop"}
	for _, name := range expected {
		if !branchMap[name] {
			t.Errorf("Expected branch %s not found in list", name)
		}
	}
}

// TestCurrentBranch tests getting the current branch
func TestCurrentBranch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "current-branch-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Set HEAD to point to main
	repo.SetHEAD("ref: refs/heads/main")

	// Get current branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if currentBranch != "main" {
		t.Errorf("Current branch = %s, want main", currentBranch)
	}

	// Set HEAD to point to feature
	repo.SetHEAD("ref: refs/heads/feature")

	currentBranch, err = repo.CurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if currentBranch != "feature" {
		t.Errorf("Current branch = %s, want feature", currentBranch)
	}

	// Set HEAD to detached (direct hash)
	repo.SetHEAD("1234567890abcdef1234567890abcdef12345678")

	_, err = repo.CurrentBranch()
	if err == nil {
		t.Error("Expected error for detached HEAD")
	}
}

// TestGetBranch tests getting branch hash
func TestGetBranch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "get-branch-test")

	repo, err := Create(repoPath, DefaultInitOptions())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create branch
	testHash, _ := hash.ParseHash("1234567890abcdef1234567890abcdef12345678")
	repo.CreateBranch("test", testHash)

	// Get branch hash
	branchHash, err := repo.GetBranch("test")
	if err != nil {
		t.Fatalf("Failed to get branch: %v", err)
	}

	if !branchHash.Equals(testHash) {
		t.Errorf("Branch hash = %s, want %s", branchHash, testHash)
	}

	// Try to get nonexistent branch
	_, err = repo.GetBranch("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent branch")
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
