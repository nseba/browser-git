package index

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// CommitOptions contains options for creating a commit
type CommitOptions struct {
	Message   string
	Author    object.Signature
	Committer object.Signature
	Parents   []hash.Hash
}

// BuildTree builds a tree object from the index entries
func (idx *Index) BuildTree(hasher hash.Hasher, objDB object.Database) (hash.Hash, error) {
	if len(idx.Entries) == 0 {
		// Empty tree
		tree := object.NewTree()
		if err := tree.ComputeHash(hasher); err != nil {
			return nil, err
		}
		if objDB != nil {
			if _, err := objDB.Put(tree); err != nil {
				return nil, err
			}
		}
		return tree.Hash(), nil
	}

	// Build tree structure from flat index
	return idx.buildTreeRecursive(hasher, objDB, "", idx.Entries)
}

// buildTreeRecursive builds a tree recursively for a directory
func (idx *Index) buildTreeRecursive(hasher hash.Hasher, objDB object.Database, prefix string, entries []*Entry) (hash.Hash, error) {
	tree := object.NewTree()

	// Group entries by first path component
	groups := make(map[string][]*Entry)
	files := make([]*Entry, 0)

	for _, entry := range entries {
		relPath := entry.Path
		if prefix != "" {
			if !strings.HasPrefix(entry.Path, prefix+"/") {
				continue
			}
			relPath = strings.TrimPrefix(entry.Path, prefix+"/")
		}

		// Split path
		parts := strings.Split(relPath, "/")
		if len(parts) == 1 {
			// File in current directory
			files = append(files, entry)
		} else {
			// File in subdirectory
			subdir := parts[0]
			if groups[subdir] == nil {
				groups[subdir] = make([]*Entry, 0)
			}
			groups[subdir] = append(groups[subdir], entry)
		}
	}

	// Add files to tree
	for _, entry := range files {
		name := filepath.Base(entry.Path)
		mode := convertModeToTreeMode(entry.Mode)
		tree.AddEntryWithMode(mode, name, entry.Hash)
	}

	// Process subdirectories
	subdirs := make([]string, 0, len(groups))
	for subdir := range groups {
		subdirs = append(subdirs, subdir)
	}
	sort.Strings(subdirs)

	for _, subdir := range subdirs {
		subdirEntries := groups[subdir]
		subdirPrefix := subdir
		if prefix != "" {
			subdirPrefix = prefix + "/" + subdir
		}

		// Build subtree
		subtreeHash, err := idx.buildTreeRecursive(hasher, objDB, subdirPrefix, subdirEntries)
		if err != nil {
			return nil, fmt.Errorf("failed to build tree for %s: %w", subdirPrefix, err)
		}

		// Add subtree to current tree
		tree.AddEntryWithMode(object.ModeDir, subdir, subtreeHash)
	}

	// Compute tree hash
	if err := tree.ComputeHash(hasher); err != nil {
		return nil, err
	}

	// Store tree in object database
	if objDB != nil {
		if _, err := objDB.Put(tree); err != nil {
			return nil, fmt.Errorf("failed to store tree: %w", err)
		}
	}

	return tree.Hash(), nil
}

// convertModeToTreeMode converts index file mode to tree entry mode
func convertModeToTreeMode(mode uint32) object.FileMode {
	switch mode {
	case FileModeRegular:
		return object.ModeRegular
	case FileModeExecutable:
		return object.ModeExecutable
	case FileModeSymlink:
		return object.ModeSymlink
	case FileModeGitlink:
		return object.ModeGitlink
	default:
		return object.ModeRegular
	}
}

// CreateCommit creates a commit object from the index
func (idx *Index) CreateCommit(hasher hash.Hasher, objDB object.Database, opts CommitOptions) (hash.Hash, error) {
	// Build tree from index
	treeHash, err := idx.BuildTree(hasher, objDB)
	if err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	// Create commit object
	commit := object.NewCommit()
	commit.Tree = treeHash
	commit.Parents = opts.Parents
	commit.Author = opts.Author
	commit.Committer = opts.Committer
	commit.Message = opts.Message

	// Ensure message ends with newline
	if !strings.HasSuffix(commit.Message, "\n") {
		commit.Message += "\n"
	}

	// Compute commit hash
	if err := commit.ComputeHash(hasher); err != nil {
		return nil, fmt.Errorf("failed to compute commit hash: %w", err)
	}

	// Store commit in object database
	if objDB != nil {
		if _, err := objDB.Put(commit); err != nil {
			return nil, fmt.Errorf("failed to store commit: %w", err)
		}
	}

	return commit.Hash(), nil
}

// WriteBlobs writes all blob objects from the index to the object database
func (idx *Index) WriteBlobs(workTreePath string, objDB object.Database) error {
	for _, entry := range idx.Entries {
		// Check if blob already exists
		if objDB.Has(entry.Hash) {
			continue
		}

		// Read file content
		fullPath := filepath.Join(workTreePath, entry.Path)
		content, err := readFileContent(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", entry.Path, err)
		}

		// Create blob
		blob := object.NewBlob(content)
		blob.SetHash(entry.Hash)

		// Store blob
		if _, err := objDB.Put(blob); err != nil {
			return fmt.Errorf("failed to store blob for %s: %w", entry.Path, err)
		}
	}

	return nil
}

// readFileContent reads file content, handling different file types
func readFileContent(path string) ([]byte, error) {
	info, err := filepath.EvalSymlinks(path)
	if err == nil {
		// If it's a symlink, read the link target
		if info != path {
			target, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil, err
			}
			return []byte(target), nil
		}
	}

	// Regular file
	return os.ReadFile(path)
}

// GetParentCommit gets the parent commit hash from HEAD
func GetParentCommit(repo interface{ HEAD() (string, error); ResolveRef(string) (hash.Hash, error) }) ([]hash.Hash, error) {
	// Get HEAD
	head, err := repo.HEAD()
	if err != nil {
		return nil, err
	}

	// Check if HEAD is a symbolic ref
	if strings.HasPrefix(head, "ref: ") {
		ref := strings.TrimPrefix(head, "ref: ")

		// Try to resolve the ref
		commitHash, err := repo.ResolveRef(ref)
		if err != nil {
			// Branch exists but has no commits yet (initial commit)
			return nil, nil
		}

		return []hash.Hash{commitHash}, nil
	}

	// Direct hash (detached HEAD)
	commitHash, err := hash.ParseHash(head)
	if err != nil {
		return nil, err
	}

	return []hash.Hash{commitHash}, nil
}

// DefaultSignature creates a signature with the given name and email
func DefaultSignature(name, email string) object.Signature {
	return object.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}
