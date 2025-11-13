package index

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
	"github.com/nseba/browser-git/git-core/pkg/repository"
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

// TestCompleteAddCommitWorkflow tests the entire add-commit workflow
func TestCompleteAddCommitWorkflow(t *testing.T) {
	// Create temp directory for repository
	tmpDir := t.TempDir()

	// Initialize repository
	opts := repository.DefaultInitOptions()
	if err := repository.Init(tmpDir, opts); err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	// Create test files
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Open repository
	repo, err := repository.Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Set user config
	repo.Config.SetUser("Test User", "test@example.com")

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	if err := repo.Config.Save(filepath.Join(repo.GitDir, "config")); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := Load(indexPath)
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	// Add files to index
	addOpts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"test1.txt", "test2.txt"}, addOpts); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	// Verify files are in index
	if !idx.HasEntry("test1.txt") {
		t.Error("expected test1.txt to be in index")
	}
	if !idx.HasEntry("test2.txt") {
		t.Error("expected test2.txt to be in index")
	}

	// Save index
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// Write blobs to object database
	if err := idx.WriteBlobs(tmpDir, repo.ObjectDB); err != nil {
		t.Fatalf("failed to write blobs: %v", err)
	}

	// Create commit
	commitOpts := CommitOptions{
		Message:   "Initial commit",
		Author:    DefaultSignature("Test User", "test@example.com"),
		Committer: DefaultSignature("Test User", "test@example.com"),
		Parents:   nil, // Initial commit has no parents
	}

	commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
	if err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}

	// Update HEAD
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	if err := repo.UpdateRef("refs/heads/"+currentBranch, commitHash); err != nil {
		t.Fatalf("failed to update branch: %v", err)
	}

	// Verify commit was created
	obj, err := repo.ObjectDB.Get(commitHash)
	if err != nil {
		t.Fatalf("failed to get commit object: %v", err)
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		t.Fatal("expected commit object")
	}

	if commit.Message != "Initial commit\n" {
		t.Errorf("expected message 'Initial commit\\n', got '%s'", commit.Message)
	}

	if commit.Author.Name != "Test User" {
		t.Errorf("expected author name 'Test User', got '%s'", commit.Author.Name)
	}
}

// TestMultipleCommits tests creating multiple commits
func TestMultipleCommits(t *testing.T) {
	// Create temp directory for repository
	tmpDir := t.TempDir()

	// Initialize repository
	opts := repository.DefaultInitOptions()
	if err := repository.Init(tmpDir, opts); err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	// Open repository
	repo, err := repository.Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Set user config
	repo.Config.SetUser("Test User", "test@example.com")

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Helper function to create a commit
	createCommit := func(fileName, content, message string, parent hash.Hash) hash.Hash {
		// Create file
		filePath := filepath.Join(tmpDir, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		// Load index
		indexPath := filepath.Join(repo.GitDir, "index")
		idx, err := Load(indexPath)
		if err != nil {
			t.Fatalf("failed to load index: %v", err)
		}

		// Add file to index
		addOpts := AddOptions{}
		if err := idx.Add(tmpDir, []string{fileName}, addOpts); err != nil {
			t.Fatalf("failed to add file: %v", err)
		}

		// Save index
		if err := idx.Save(indexPath); err != nil {
			t.Fatalf("failed to save index: %v", err)
		}

		// Write blobs
		if err := idx.WriteBlobs(tmpDir, repo.ObjectDB); err != nil {
			t.Fatalf("failed to write blobs: %v", err)
		}

		// Create commit
		parents := []hash.Hash{}
		if parent != nil {
			parents = append(parents, parent)
		}

		commitOpts := CommitOptions{
			Message:   message,
			Author:    DefaultSignature("Test User", "test@example.com"),
			Committer: DefaultSignature("Test User", "test@example.com"),
			Parents:   parents,
		}

		commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
		if err != nil {
			t.Fatalf("failed to create commit: %v", err)
		}

		// Update HEAD
		currentBranch, _ := repo.CurrentBranch()
		if err := repo.UpdateRef("refs/heads/"+currentBranch, commitHash); err != nil {
			t.Fatalf("failed to update branch: %v", err)
		}

		return commitHash
	}

	// Create first commit
	commit1Hash := createCommit("file1.txt", "content 1", "First commit", nil)

	// Create second commit
	commit2Hash := createCommit("file2.txt", "content 2", "Second commit", commit1Hash)

	// Create third commit
	commit3Hash := createCommit("file3.txt", "content 3", "Third commit", commit2Hash)

	// Verify commits are linked
	commit3, _ := repo.ObjectDB.Get(commit3Hash)
	commit3Obj := commit3.(*object.Commit)
	if len(commit3Obj.Parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(commit3Obj.Parents))
	}
	if !commit3Obj.Parents[0].Equals(commit2Hash) {
		t.Error("commit3 parent should be commit2")
	}

	commit2, _ := repo.ObjectDB.Get(commit2Hash)
	commit2Obj := commit2.(*object.Commit)
	if len(commit2Obj.Parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(commit2Obj.Parents))
	}
	if !commit2Obj.Parents[0].Equals(commit1Hash) {
		t.Error("commit2 parent should be commit1")
	}

	commit1, _ := repo.ObjectDB.Get(commit1Hash)
	commit1Obj := commit1.(*object.Commit)
	if len(commit1Obj.Parents) != 0 {
		t.Errorf("expected 0 parents for initial commit, got %d", len(commit1Obj.Parents))
	}
}

// TestAddCommitWithGitignore tests add-commit workflow with .gitignore
func TestAddCommitWithGitignore(t *testing.T) {
	// Create temp directory for repository
	tmpDir := t.TempDir()

	// Initialize repository
	opts := repository.DefaultInitOptions()
	if err := repository.Init(tmpDir, opts); err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	// Create test files and .gitignore
	file1 := filepath.Join(tmpDir, "tracked.txt")
	file2 := filepath.Join(tmpDir, "ignored.txt")
	gitignore := filepath.Join(tmpDir, ".gitignore")

	if err := os.WriteFile(file1, []byte("tracked content"), 0644); err != nil {
		t.Fatalf("failed to create tracked file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("ignored content"), 0644); err != nil {
		t.Fatalf("failed to create ignored file: %v", err)
	}
	if err := os.WriteFile(gitignore, []byte("ignored.txt\n"), 0644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	// Open repository
	repo, err := repository.Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Set user config
	repo.Config.SetUser("Test User", "test@example.com")

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := Load(indexPath)
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	// Add all files
	addOpts := AddOptions{}
	if err := idx.Add(tmpDir, []string{"."}, addOpts); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	// Verify only tracked files are in index
	if !idx.HasEntry("tracked.txt") {
		t.Error("expected tracked.txt to be in index")
	}
	if idx.HasEntry("ignored.txt") {
		t.Error("expected ignored.txt not to be in index")
	}

	// .gitignore itself should be in index
	if !idx.HasEntry(".gitignore") {
		t.Error("expected .gitignore to be in index")
	}

	// Save index
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// Write blobs
	if err := idx.WriteBlobs(tmpDir, repo.ObjectDB); err != nil {
		t.Fatalf("failed to write blobs: %v", err)
	}

	// Create commit
	commitOpts := CommitOptions{
		Message:   "Add files with gitignore",
		Author:    DefaultSignature("Test User", "test@example.com"),
		Committer: DefaultSignature("Test User", "test@example.com"),
		Parents:   nil,
	}

	commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
	if err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}

	// Verify commit tree
	commit, _ := repo.ObjectDB.Get(commitHash)
	commitObj := commit.(*object.Commit)

	tree, _ := repo.ObjectDB.Get(commitObj.Tree)
	treeObj := tree.(*object.Tree)
	entries := treeObj.Entries()

	// Should have 2 entries: tracked.txt and .gitignore
	if len(entries) != 2 {
		t.Errorf("expected 2 tree entries, got %d", len(entries))
	}
}

// TestModifyAndCommit tests modifying files and committing changes
func TestModifyAndCommit(t *testing.T) {
	// Create temp directory for repository
	tmpDir := t.TempDir()

	// Initialize repository
	opts := repository.DefaultInitOptions()
	if err := repository.Init(tmpDir, opts); err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	// Open repository
	repo, err := repository.Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Set user config
	repo.Config.SetUser("Test User", "test@example.com")

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Create initial file and commit
	file1 := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(file1, []byte("initial content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	indexPath := filepath.Join(repo.GitDir, "index")
	idx, _ := Load(indexPath)

	idx.Add(tmpDir, []string{"test.txt"}, AddOptions{})
	idx.Save(indexPath)
	idx.WriteBlobs(tmpDir, repo.ObjectDB)

	commit1Opts := CommitOptions{
		Message:   "Initial commit",
		Author:    DefaultSignature("Test User", "test@example.com"),
		Committer: DefaultSignature("Test User", "test@example.com"),
		Parents:   nil,
	}
	commit1Hash, _ := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commit1Opts)
	currentBranch, _ := repo.CurrentBranch()
	repo.UpdateRef("refs/heads/"+currentBranch, commit1Hash)

	// Modify file
	time.Sleep(10 * time.Millisecond) // Ensure different mtime
	if err := os.WriteFile(file1, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Check that file is modified
	entry, _ := idx.GetEntry("test.txt")
	isModified, _ := entry.IsModified(tmpDir)
	if !isModified {
		t.Error("expected file to be marked as modified")
	}

	// Add modified file
	idx.Add(tmpDir, []string{"test.txt"}, AddOptions{})
	idx.Save(indexPath)
	idx.WriteBlobs(tmpDir, repo.ObjectDB)

	// Create second commit
	commit2Opts := CommitOptions{
		Message:   "Modify test.txt",
		Author:    DefaultSignature("Test User", "test@example.com"),
		Committer: DefaultSignature("Test User", "test@example.com"),
		Parents:   []hash.Hash{commit1Hash},
	}
	commit2Hash, _ := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commit2Opts)
	repo.UpdateRef("refs/heads/"+currentBranch, commit2Hash)

	// Verify both commits exist
	commit1, _ := repo.ObjectDB.Get(commit1Hash)
	commit2, _ := repo.ObjectDB.Get(commit2Hash)

	if commit1 == nil || commit2 == nil {
		t.Error("expected both commits to exist")
	}

	// Verify commit2 has different tree than commit1
	commit1Obj := commit1.(*object.Commit)
	commit2Obj := commit2.(*object.Commit)

	if commit1Obj.Tree.Equals(commit2Obj.Tree) {
		t.Error("expected different trees for commits")
	}
}

// TestAddCommitWithSubdirectories tests add-commit with nested directories
func TestAddCommitWithSubdirectories(t *testing.T) {
	// Create temp directory for repository
	tmpDir := t.TempDir()

	// Initialize repository
	opts := repository.DefaultInitOptions()
	if err := repository.Init(tmpDir, opts); err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	// Create nested directory structure
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tmpDir, "root.txt")
	file2 := filepath.Join(subdir, "nested.txt")

	if err := os.WriteFile(file1, []byte("root content"), 0644); err != nil {
		t.Fatalf("failed to create root file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("nested content"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	// Open repository
	repo, err := repository.Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Set user config
	repo.Config.SetUser("Test User", "test@example.com")

	// Initialize object database
	storage := NewMemoryStorage()
	repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)

	// Load index and add all files
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, _ := Load(indexPath)

	idx.Add(tmpDir, []string{"."}, AddOptions{})

	// Verify files are in index with correct paths
	if !idx.HasEntry("root.txt") {
		t.Error("expected root.txt to be in index")
	}
	if !idx.HasEntry("subdir/nested.txt") {
		t.Error("expected subdir/nested.txt to be in index")
	}

	idx.Save(indexPath)
	idx.WriteBlobs(tmpDir, repo.ObjectDB)

	// Create commit
	commitOpts := CommitOptions{
		Message:   "Add nested files",
		Author:    DefaultSignature("Test User", "test@example.com"),
		Committer: DefaultSignature("Test User", "test@example.com"),
		Parents:   nil,
	}

	commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
	if err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}

	// Verify tree structure
	commit, _ := repo.ObjectDB.Get(commitHash)
	commitObj := commit.(*object.Commit)

	tree, _ := repo.ObjectDB.Get(commitObj.Tree)
	treeObj := tree.(*object.Tree)
	entries := treeObj.Entries()

	// Should have root.txt and subdir
	foundRoot := false
	foundSubdir := false
	for _, entry := range entries {
		if entry.Name == "root.txt" {
			foundRoot = true
		}
		if entry.Name == "subdir" && entry.Mode == object.ModeDir {
			foundSubdir = true
		}
	}

	if !foundRoot {
		t.Error("expected root.txt in root tree")
	}
	if !foundSubdir {
		t.Error("expected subdir in root tree")
	}
}
