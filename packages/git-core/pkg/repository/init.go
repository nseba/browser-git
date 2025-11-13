package repository

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InitOptions contains options for initializing a repository
type InitOptions struct {
	// Bare indicates if this should be a bare repository
	Bare bool
	// InitialBranch is the name of the initial branch (default: "main")
	InitialBranch string
	// HashAlgorithm is the hash algorithm to use ("sha1" or "sha256", default: "sha1")
	HashAlgorithm string
}

// DefaultInitOptions returns default initialization options
func DefaultInitOptions() InitOptions {
	return InitOptions{
		Bare:          false,
		InitialBranch: "main",
		HashAlgorithm: "sha1",
	}
}

// Init initializes a new Git repository at the specified path
func Init(path string, opts InitOptions) error {
	// Determine .git directory location
	gitDir := path
	if !opts.Bare {
		gitDir = filepath.Join(path, ".git")
	}

	// Check if repository already exists
	if _, err := os.Stat(gitDir); err == nil {
		return fmt.Errorf("repository already exists at %s", gitDir)
	}

	// Create .git directory structure
	if err := createGitDirectories(gitDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Create HEAD file
	if err := createHEAD(gitDir, opts.InitialBranch); err != nil {
		return fmt.Errorf("failed to create HEAD: %w", err)
	}

	// Create config file
	if err := createConfig(gitDir, opts); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Create description file
	if err := createDescription(gitDir); err != nil {
		return fmt.Errorf("failed to create description: %w", err)
	}

	return nil
}

// createGitDirectories creates the standard .git directory structure
func createGitDirectories(gitDir string) error {
	// Main .git directory
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return err
	}

	// Standard subdirectories
	dirs := []string{
		"objects",
		"objects/info",
		"objects/pack",
		"refs",
		"refs/heads",
		"refs/tags",
		"hooks",
		"info",
		"branches", // Legacy, but kept for compatibility
	}

	for _, dir := range dirs {
		path := filepath.Join(gitDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// createHEAD creates the HEAD file pointing to the initial branch
func createHEAD(gitDir string, initialBranch string) error {
	headPath := filepath.Join(gitDir, "HEAD")
	content := fmt.Sprintf("ref: refs/heads/%s\n", initialBranch)

	return writeFile(headPath, []byte(content), 0644)
}

// createConfig creates the initial config file
func createConfig(gitDir string, opts InitOptions) error {
	configPath := filepath.Join(gitDir, "config")

	// Build config content
	config := "[core]\n"
	config += "\trepositoryformatversion = 0\n"
	config += "\tfilemode = true\n"
	config += fmt.Sprintf("\tbare = %t\n", opts.Bare)

	// Add hash algorithm extension if using SHA-256
	if opts.HashAlgorithm == "sha256" {
		config += "\n[extensions]\n"
		config += "\tobjectFormat = sha256\n"
	}

	return writeFile(configPath, []byte(config), 0644)
}

// createDescription creates the description file
func createDescription(gitDir string) error {
	descPath := filepath.Join(gitDir, "description")
	content := "Unnamed repository; edit this file 'description' to name the repository.\n"

	return writeFile(descPath, []byte(content), 0644)
}

// writeFile writes content to a file with the specified permissions
func writeFile(path string, content []byte, perm os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	return err
}

// IsRepository checks if a directory contains a Git repository
func IsRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FindRepository searches for a Git repository starting from path and walking up
func FindRepository(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		if IsRepository(current) {
			return current, nil
		}

		// Check if this is a bare repository
		if isBareRepository(current) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			return "", fmt.Errorf("not a git repository (or any of the parent directories)")
		}
		current = parent
	}
}

// isBareRepository checks if a directory is a bare repository
func isBareRepository(path string) bool {
	// A bare repository has HEAD and refs directly in the directory
	headPath := filepath.Join(path, "HEAD")
	refsPath := filepath.Join(path, "refs")

	headInfo, headErr := os.Stat(headPath)
	refsInfo, refsErr := os.Stat(refsPath)

	return headErr == nil && !headInfo.IsDir() &&
		refsErr == nil && refsInfo.IsDir()
}

// GetGitDir returns the .git directory path for a repository
func GetGitDir(repoPath string) (string, error) {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	if err == nil && info.IsDir() {
		return gitDir, nil
	}

	// Check if this is a bare repository
	if isBareRepository(repoPath) {
		return repoPath, nil
	}

	return "", fmt.Errorf("not a git repository: %s", repoPath)
}

// ReadFile reads a file from the repository
func ReadFile(gitDir string, relativePath string) ([]byte, error) {
	path := filepath.Join(gitDir, relativePath)
	return os.ReadFile(path)
}

// WriteFileInRepo writes a file to the repository
func WriteFileInRepo(gitDir string, relativePath string, content []byte, perm os.FileMode) error {
	path := filepath.Join(gitDir, relativePath)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return writeFile(path, content, perm)
}

// ListDirectory lists files in a directory within the repository
func ListDirectory(gitDir string, relativePath string) ([]os.DirEntry, error) {
	path := filepath.Join(gitDir, relativePath)
	return os.ReadDir(path)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
