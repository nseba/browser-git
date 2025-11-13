package index

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// FileStatus represents the status of a file
type FileStatus int

const (
	StatusUntracked FileStatus = iota // File not in index or HEAD
	StatusUnmodified                  // File unchanged
	StatusModified                    // File modified in work tree
	StatusStaged                      // File staged (added to index)
	StatusDeleted                     // File deleted from work tree
	StatusAdded                       // File added (new in index)
	StatusRenamed                     // File renamed
	StatusConflict                    // File has merge conflict
)

// String returns the string representation of file status
func (s FileStatus) String() string {
	switch s {
	case StatusUntracked:
		return "untracked"
	case StatusUnmodified:
		return "unmodified"
	case StatusModified:
		return "modified"
	case StatusStaged:
		return "staged"
	case StatusDeleted:
		return "deleted"
	case StatusAdded:
		return "added"
	case StatusRenamed:
		return "renamed"
	case StatusConflict:
		return "conflict"
	default:
		return "unknown"
	}
}

// FileStatusEntry represents the status of a single file
type FileStatusEntry struct {
	Path         string
	Status       FileStatus
	IndexStatus  FileStatus // Status in index vs HEAD
	WorkStatus   FileStatus // Status in work tree vs index
	StagedHash   hash.Hash  // Hash in index
	WorkTreeHash hash.Hash  // Hash in work tree
}

// Status represents repository status
type Status struct {
	Untracked []string          // Untracked files
	Modified  []string          // Modified files (not staged)
	Staged    []string          // Staged files (in index)
	Deleted   []string          // Deleted files
	Added     []string          // Added files (new in index)
	Entries   []*FileStatusEntry // Detailed status entries
}

// StatusOptions contains options for status computation
type StatusOptions struct {
	IncludeUntracked bool // Include untracked files
	IncludeIgnored   bool // Include ignored files
}

// DefaultStatusOptions returns default status options
func DefaultStatusOptions() StatusOptions {
	return StatusOptions{
		IncludeUntracked: true,
		IncludeIgnored:   false,
	}
}

// GetStatus computes the status of the repository
func GetStatus(workTreePath string, idx *Index, headCommit *object.Commit, objDB object.Database, opts StatusOptions) (*Status, error) {
	status := &Status{
		Untracked: make([]string, 0),
		Modified:  make([]string, 0),
		Staged:    make([]string, 0),
		Deleted:   make([]string, 0),
		Added:     make([]string, 0),
		Entries:   make([]*FileStatusEntry, 0),
	}

	// Load gitignore
	gitignore, err := LoadGitignore(workTreePath)
	if err != nil {
		return nil, err
	}

	// Get HEAD tree entries
	headEntries := make(map[string]hash.Hash)
	if headCommit != nil {
		tree, err := objDB.Get(headCommit.Tree)
		if err != nil {
			return nil, fmt.Errorf("failed to load HEAD tree: %w", err)
		}
		treeObj, ok := tree.(*object.Tree)
		if !ok {
			return nil, fmt.Errorf("HEAD tree is not a tree object")
		}
		if err := collectTreeEntries(treeObj, "", objDB, headEntries); err != nil {
			return nil, err
		}
	}

	// Get index entries
	indexEntries := make(map[string]*Entry)
	for _, entry := range idx.Entries {
		indexEntries[entry.Path] = entry
	}

	// Get work tree files
	workTreeFiles := make(map[string]bool)
	if opts.IncludeUntracked {
		err := filepath.WalkDir(workTreePath, func(path string, d fs.DirEntry, err error) error {
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

			// Skip ignored files if not including them
			if !opts.IncludeIgnored && gitignore.Match(relPath) {
				return nil
			}

			workTreeFiles[relPath] = true
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Process files in HEAD
	for path, headHash := range headEntries {
		entry := &FileStatusEntry{Path: path}
		indexEntry, inIndex := indexEntries[path]
		_, inWorkTree := workTreeFiles[path]

		if inIndex {
			// File is in HEAD and index
			entry.StagedHash = indexEntry.Hash
			if !indexEntry.Hash.Equals(headHash) {
				// Index differs from HEAD - file is staged
				entry.IndexStatus = StatusStaged
				status.Staged = append(status.Staged, path)
			} else {
				entry.IndexStatus = StatusUnmodified
			}

			if inWorkTree {
				// Check if work tree differs from index
				modified, err := indexEntry.IsModified(workTreePath)
				if err != nil {
					return nil, err
				}
				if modified {
					entry.WorkStatus = StatusModified
					status.Modified = append(status.Modified, path)
				} else {
					entry.WorkStatus = StatusUnmodified
				}
			} else {
				// File deleted in work tree
				entry.WorkStatus = StatusDeleted
				status.Deleted = append(status.Deleted, path)
			}
		} else {
			// File is in HEAD but not in index - deleted
			entry.IndexStatus = StatusDeleted
			if !inWorkTree {
				status.Deleted = append(status.Deleted, path)
			}
		}

		status.Entries = append(status.Entries, entry)
	}

	// Process files in index but not in HEAD (new files)
	for path, indexEntry := range indexEntries {
		if _, inHead := headEntries[path]; !inHead {
			entry := &FileStatusEntry{
				Path:        path,
				StagedHash:  indexEntry.Hash,
				IndexStatus: StatusAdded,
			}
			status.Added = append(status.Added, path)

			_, inWorkTree := workTreeFiles[path]
			if inWorkTree {
				// Check if work tree differs from index
				modified, err := indexEntry.IsModified(workTreePath)
				if err != nil {
					return nil, err
				}
				if modified {
					entry.WorkStatus = StatusModified
					status.Modified = append(status.Modified, path)
				} else {
					entry.WorkStatus = StatusUnmodified
				}
			} else {
				entry.WorkStatus = StatusDeleted
				status.Deleted = append(status.Deleted, path)
			}

			status.Entries = append(status.Entries, entry)
		}
	}

	// Process untracked files (in work tree but not in index or HEAD)
	if opts.IncludeUntracked {
		for path := range workTreeFiles {
			if _, inIndex := indexEntries[path]; !inIndex {
				if _, inHead := headEntries[path]; !inHead {
					entry := &FileStatusEntry{
						Path:        path,
						IndexStatus: StatusUntracked,
						WorkStatus:  StatusUntracked,
					}
					status.Untracked = append(status.Untracked, path)
					status.Entries = append(status.Entries, entry)
				}
			}
		}
	}

	return status, nil
}

// collectTreeEntries recursively collects all entries from a tree
func collectTreeEntries(tree *object.Tree, prefix string, objDB object.Database, entries map[string]hash.Hash) error {
	treeEntries := tree.Entries()
	for _, entry := range treeEntries {
		path := entry.Name
		if prefix != "" {
			path = prefix + "/" + entry.Name
		}

		if entry.Mode == object.ModeDir {
			// Recursively process subtree
			subtreeObj, err := objDB.Get(entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to load subtree %s: %w", path, err)
			}
			subtree, ok := subtreeObj.(*object.Tree)
			if !ok {
				return fmt.Errorf("subtree %s is not a tree object", path)
			}
			if err := collectTreeEntries(subtree, path, objDB, entries); err != nil {
				return err
			}
		} else {
			// Add file entry
			entries[path] = entry.Hash
		}
	}
	return nil
}

// IsClean returns true if the repository has no changes
func (s *Status) IsClean() bool {
	return len(s.Untracked) == 0 &&
		len(s.Modified) == 0 &&
		len(s.Staged) == 0 &&
		len(s.Deleted) == 0 &&
		len(s.Added) == 0
}

// HasChanges returns true if there are any changes (staged or unstaged)
func (s *Status) HasChanges() bool {
	return !s.IsClean()
}

// HasStagedChanges returns true if there are staged changes
func (s *Status) HasStagedChanges() bool {
	return len(s.Staged) > 0 || len(s.Added) > 0
}

// HasUnstagedChanges returns true if there are unstaged changes
func (s *Status) HasUnstagedChanges() bool {
	return len(s.Modified) > 0 || len(s.Deleted) > 0
}

// Summary returns a human-readable summary of the status
func (s *Status) Summary() string {
	var sb strings.Builder

	if s.IsClean() {
		sb.WriteString("nothing to commit, working tree clean\n")
		return sb.String()
	}

	if s.HasStagedChanges() {
		sb.WriteString("Changes to be committed:\n")
		for _, path := range s.Added {
			sb.WriteString(fmt.Sprintf("  new file:   %s\n", path))
		}
		for _, path := range s.Staged {
			sb.WriteString(fmt.Sprintf("  modified:   %s\n", path))
		}
		sb.WriteString("\n")
	}

	if s.HasUnstagedChanges() {
		sb.WriteString("Changes not staged for commit:\n")
		for _, path := range s.Modified {
			sb.WriteString(fmt.Sprintf("  modified:   %s\n", path))
		}
		for _, path := range s.Deleted {
			sb.WriteString(fmt.Sprintf("  deleted:    %s\n", path))
		}
		sb.WriteString("\n")
	}

	if len(s.Untracked) > 0 {
		sb.WriteString("Untracked files:\n")
		for _, path := range s.Untracked {
			sb.WriteString(fmt.Sprintf("  %s\n", path))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
