package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// fileStorage implements object.Storage using filesystem
type fileStorage struct {
	objectsPath string
	hasher      hash.Hasher
}

// newFileStorage creates a new file-based storage
func newFileStorage(objectsPath string, hasher hash.Hasher) *fileStorage {
	return &fileStorage{
		objectsPath: objectsPath,
		hasher:      hasher,
	}
}

// Read reads compressed object data for the given hash
func (fs *fileStorage) Read(h hash.Hash) ([]byte, error) {
	path := fs.objectPath(h)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object %s not found", h.String())
		}
		return nil, err
	}
	return data, nil
}

// Write writes compressed object data with the given hash
func (fs *fileStorage) Write(h hash.Hash, data []byte) error {
	path := fs.objectPath(h)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0444); err != nil {
		return fmt.Errorf("failed to write object: %w", err)
	}

	return nil
}

// Has checks if an object exists
func (fs *fileStorage) Has(h hash.Hash) bool {
	path := fs.objectPath(h)
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes an object
func (fs *fileStorage) Delete(h hash.Hash) error {
	path := fs.objectPath(h)
	return os.Remove(path)
}

// List returns all object hashes
func (fs *fileStorage) List() ([]hash.Hash, error) {
	hashes := []hash.Hash{}

	// Walk through all subdirectories in objects/
	err := filepath.Walk(fs.objectsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, info, pack
		if info.IsDir() {
			name := info.Name()
			if name == "info" || name == "pack" {
				return filepath.SkipDir
			}
			return nil
		}

		// Parse hash from path
		// Path format: objects/ab/cdef...
		rel, err := filepath.Rel(fs.objectsPath, path)
		if err != nil {
			return nil
		}

		// Skip if not in expected format
		if !strings.Contains(rel, string(filepath.Separator)) {
			return nil
		}

		// Reconstruct hash
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) != 2 || len(parts[0]) != 2 {
			return nil
		}

		hashStr := parts[0] + parts[1]
		h, err := hash.ParseHash(hashStr)
		if err != nil {
			return nil // Skip invalid hashes
		}

		hashes = append(hashes, h)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// Close closes the storage
func (fs *fileStorage) Close() error {
	// No cleanup needed for file storage
	return nil
}

// objectPath returns the filesystem path for an object hash
// Objects are stored as: objects/ab/cdef1234...
func (fs *fileStorage) objectPath(h hash.Hash) string {
	hashStr := h.String()
	if len(hashStr) < 2 {
		return filepath.Join(fs.objectsPath, hashStr)
	}
	return filepath.Join(fs.objectsPath, hashStr[:2], hashStr[2:])
}
