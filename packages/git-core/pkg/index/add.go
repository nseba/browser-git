package index

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AddOptions contains options for adding files to the index
type AddOptions struct {
	// Force adds files even if they're ignored
	Force bool
	// UpdateOnly only updates already tracked files
	UpdateOnly bool
}

// Add adds files to the index
func (idx *Index) Add(workTreePath string, paths []string, opts AddOptions) error {
	gitignore, err := LoadGitignore(workTreePath)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if err := idx.addPath(workTreePath, path, gitignore, opts); err != nil {
			return err
		}
	}

	return nil
}

// addPath adds a single path (file or directory) to the index
func (idx *Index) addPath(workTreePath string, path string, gitignore *Gitignore, opts AddOptions) error {
	fullPath := filepath.Join(workTreePath, path)

	info, err := os.Lstat(fullPath)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if info.IsDir() {
		// Add directory recursively
		return idx.addDirectory(workTreePath, path, gitignore, opts)
	}

	// Check if file should be ignored
	if !opts.Force && gitignore.Match(path) {
		return nil // Skip ignored files
	}

	// Check if update-only mode
	if opts.UpdateOnly && !idx.HasEntry(path) {
		return nil // Skip untracked files in update-only mode
	}

	// Add single file
	entry, err := NewEntryFromFile(path, workTreePath)
	if err != nil {
		return fmt.Errorf("failed to create entry for %s: %w", path, err)
	}

	idx.AddEntry(entry)
	return nil
}

// addDirectory adds all files in a directory recursively
func (idx *Index) addDirectory(workTreePath string, dir string, gitignore *Gitignore, opts AddOptions) error {
	fullPath := filepath.Join(workTreePath, dir)

	return filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip directories themselves
		if d.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(workTreePath, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Check if file should be ignored
		if !opts.Force && gitignore.Match(relPath) {
			return nil
		}

		// Check if update-only mode
		if opts.UpdateOnly && !idx.HasEntry(relPath) {
			return nil
		}

		// Add file
		entry, err := NewEntryFromFile(relPath, workTreePath)
		if err != nil {
			return fmt.Errorf("failed to create entry for %s: %w", relPath, err)
		}

		idx.AddEntry(entry)
		return nil
	})
}

// AddAll adds all files matching the pattern to the index
// Supports glob patterns like "*.txt", "src/**/*.go", etc.
func (idx *Index) AddAll(workTreePath string, pattern string, opts AddOptions) error {
	gitignore, err := LoadGitignore(workTreePath)
	if err != nil {
		return err
	}

	// If pattern is ".", add everything
	if pattern == "." || pattern == "" {
		return idx.addDirectory(workTreePath, ".", gitignore, opts)
	}

	// Find matching files
	matches, err := findMatches(workTreePath, pattern)
	if err != nil {
		return err
	}

	// Add each match
	for _, match := range matches {
		if err := idx.addPath(workTreePath, match, gitignore, opts); err != nil {
			return err
		}
	}

	return nil
}

// findMatches finds all files matching a glob pattern
func findMatches(workTreePath string, pattern string) ([]string, error) {
	var matches []string

	// Handle ** patterns by walking the directory
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid glob pattern: %s", pattern)
		}

		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		startPath := workTreePath
		if prefix != "" {
			startPath = filepath.Join(workTreePath, prefix)
		}

		err := filepath.WalkDir(startPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip .git directory
			if d.IsDir() && d.Name() == ".git" {
				return filepath.SkipDir
			}

			if d.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(workTreePath, path)
			if err != nil {
				return err
			}

			relPath = filepath.ToSlash(relPath)

			// Check if matches suffix pattern
			if suffix == "" || matchPattern(filepath.Base(relPath), suffix) {
				matches = append(matches, relPath)
			}

			return nil
		})

		return matches, err
	}

	// Use filepath.Glob for simple patterns
	pattern = filepath.Join(workTreePath, pattern)
	globMatches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range globMatches {
		relPath, err := filepath.Rel(workTreePath, match)
		if err != nil {
			continue
		}
		matches = append(matches, filepath.ToSlash(relPath))
	}

	return matches, nil
}

// matchPattern checks if a filename matches a pattern
func matchPattern(name, pattern string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}

// Remove removes a path from the index
func (idx *Index) Remove(path string) error {
	if !idx.RemoveEntry(path) {
		return fmt.Errorf("path not in index: %s", path)
	}
	return nil
}

// RemoveAll removes all files matching a pattern from the index
func (idx *Index) RemoveAll(pattern string) error {
	toRemove := make([]string, 0)

	for _, entry := range idx.Entries {
		if matchPattern(entry.Path, pattern) {
			toRemove = append(toRemove, entry.Path)
		}
	}

	for _, path := range toRemove {
		idx.RemoveEntry(path)
	}

	return nil
}
