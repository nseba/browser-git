package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// Repository represents a Git repository
type Repository struct {
	// Path is the working directory path (for non-bare repos) or repository path (for bare repos)
	Path string

	// GitDir is the .git directory path
	GitDir string

	// Config is the repository configuration
	Config *Config

	// Hasher is the hash algorithm used by this repository
	Hasher hash.Hasher

	// ObjectDB is the object database
	ObjectDB object.Database
}

// Open opens an existing repository at the specified path
func Open(path string) (*Repository, error) {
	// Find the repository
	repoPath, err := FindRepository(path)
	if err != nil {
		return nil, err
	}

	// Get .git directory
	gitDir, err := GetGitDir(repoPath)
	if err != nil {
		return nil, err
	}

	// Load config
	config, err := LoadConfigFromRepo(gitDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Determine hash algorithm
	algo := config.GetHashAlgorithm()
	hasher, err := hash.NewHasher(hash.Algorithm(algo))
	if err != nil {
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	repo := &Repository{
		Path:   repoPath,
		GitDir: gitDir,
		Config: config,
		Hasher: hasher,
	}

	return repo, nil
}

// Create creates a new repository at the specified path
func Create(path string, opts InitOptions) (*Repository, error) {
	// Initialize the repository
	if err := Init(path, opts); err != nil {
		return nil, err
	}

	// Open the newly created repository
	return Open(path)
}

// HEAD returns the current HEAD reference
func (r *Repository) HEAD() (string, error) {
	content, err := ReadFile(r.GitDir, "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	headStr := string(content)
	// Remove trailing newline
	if len(headStr) > 0 && headStr[len(headStr)-1] == '\n' {
		headStr = headStr[:len(headStr)-1]
	}

	return headStr, nil
}

// SetHEAD sets the HEAD reference
func (r *Repository) SetHEAD(ref string) error {
	content := []byte(ref + "\n")
	return WriteFileInRepo(r.GitDir, "HEAD", content, 0644)
}

// CurrentBranch returns the name of the current branch
func (r *Repository) CurrentBranch() (string, error) {
	head, err := r.HEAD()
	if err != nil {
		return "", err
	}

	// Parse symbolic ref: "ref: refs/heads/main"
	const prefix = "ref: refs/heads/"
	if len(head) > len(prefix) && head[:len(prefix)] == prefix {
		return head[len(prefix):], nil
	}

	// Detached HEAD
	return "", fmt.Errorf("HEAD is detached")
}

// ResolveRef resolves a reference to a hash
func (r *Repository) ResolveRef(ref string) (hash.Hash, error) {
	// If ref starts with "refs/", read the ref file
	if len(ref) >= 5 && ref[:5] == "refs/" {
		content, err := ReadFile(r.GitDir, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to read ref %s: %w", ref, err)
		}

		hashStr := string(content)
		// Remove trailing newline
		if len(hashStr) > 0 && hashStr[len(hashStr)-1] == '\n' {
			hashStr = hashStr[:len(hashStr)-1]
		}

		return hash.ParseHash(hashStr)
	}

	// Try to parse as a hash
	return hash.ParseHash(ref)
}

// UpdateRef updates a reference to point to a hash
func (r *Repository) UpdateRef(ref string, h hash.Hash) error {
	if len(ref) < 5 || ref[:5] != "refs/" {
		return fmt.Errorf("invalid ref: must start with refs/")
	}

	content := []byte(h.String() + "\n")
	return WriteFileInRepo(r.GitDir, ref, content, 0644)
}

// BranchExists checks if a branch exists
func (r *Repository) BranchExists(name string) bool {
	ref := fmt.Sprintf("refs/heads/%s", name)
	_, err := ReadFile(r.GitDir, ref)
	return err == nil
}

// GetBranch gets the hash that a branch points to
func (r *Repository) GetBranch(name string) (hash.Hash, error) {
	ref := fmt.Sprintf("refs/heads/%s", name)
	return r.ResolveRef(ref)
}

// CreateBranch creates a new branch pointing to a hash
func (r *Repository) CreateBranch(name string, h hash.Hash) error {
	if r.BranchExists(name) {
		return fmt.Errorf("branch %s already exists", name)
	}

	ref := fmt.Sprintf("refs/heads/%s", name)
	return r.UpdateRef(ref, h)
}

// DeleteBranch deletes a branch
func (r *Repository) DeleteBranch(name string) error {
	// Don't allow deleting the current branch
	currentBranch, err := r.CurrentBranch()
	if err == nil && currentBranch == name {
		return fmt.Errorf("cannot delete the currently checked out branch")
	}

	ref := fmt.Sprintf("refs/heads/%s", name)
	refPath := filepath.Join(r.GitDir, ref)
	return removeFile(refPath)
}

// ListBranches lists all branches
func (r *Repository) ListBranches() ([]string, error) {
	entries, err := ListDirectory(r.GitDir, "refs/heads")
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			branches = append(branches, entry.Name())
		}
	}

	return branches, nil
}

// IsBare returns whether this is a bare repository
func (r *Repository) IsBare() bool {
	return r.Config.IsBare()
}

// WorkTree returns the working tree path (returns empty string for bare repos)
func (r *Repository) WorkTree() string {
	if r.IsBare() {
		return ""
	}
	return r.Path
}

// ObjectsPath returns the path to the objects directory
func (r *Repository) ObjectsPath() string {
	return filepath.Join(r.GitDir, "objects")
}

// RefsPath returns the path to the refs directory
func (r *Repository) RefsPath() string {
	return filepath.Join(r.GitDir, "refs")
}

// removeFile removes a file
func removeFile(path string) error {
	return os.Remove(path)
}
