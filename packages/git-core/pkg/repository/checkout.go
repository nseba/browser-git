package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/index"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// CheckoutOptions contains options for checkout operations
type CheckoutOptions struct {
	// Force allows checkout even with uncommitted changes
	Force bool

	// CreateBranch creates a new branch before checking out
	CreateBranch bool

	// Detach creates a detached HEAD state
	Detach bool
}

// DefaultCheckoutOptions returns default checkout options
func DefaultCheckoutOptions() CheckoutOptions {
	return CheckoutOptions{
		Force:        false,
		CreateBranch: false,
		Detach:       false,
	}
}

// Checkout checks out a branch or commit
// target can be:
// - branch name (e.g., "main", "feature")
// - commit hash (e.g., "abc123...")
// - symbolic ref (e.g., "refs/heads/main")
func (r *Repository) Checkout(target string, opts CheckoutOptions) error {
	// Load current index
	indexPath := filepath.Join(r.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Check for uncommitted changes if not forcing
	if !opts.Force {
		if err := r.checkUncommittedChanges(idx); err != nil {
			return err
		}
	}

	// Resolve target to a commit hash
	targetHash, isBranch, err := r.resolveCheckoutTarget(target)
	if err != nil {
		return err
	}

	// If creating a new branch, do it now
	if opts.CreateBranch {
		if err := r.CreateBranch(target, targetHash); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
		isBranch = true
	}

	// Get the commit object
	commitObj, err := r.ObjectDB.Get(targetHash)
	if err != nil {
		return fmt.Errorf("failed to load commit: %w", err)
	}

	commit, ok := commitObj.(*object.Commit)
	if !ok {
		return fmt.Errorf("target is not a commit")
	}

	// Update working directory and index
	if err := r.updateWorkingDirectory(commit.Tree, idx); err != nil {
		return fmt.Errorf("failed to update working directory: %w", err)
	}

	// Update HEAD
	if opts.Detach || !isBranch {
		// Detached HEAD state - point directly to commit
		if err := r.SetHEAD(targetHash.String()); err != nil {
			return fmt.Errorf("failed to update HEAD: %w", err)
		}
	} else {
		// Normal checkout - point HEAD to branch
		branchRef := fmt.Sprintf("ref: refs/heads/%s", target)
		if err := r.SetHEAD(branchRef); err != nil {
			return fmt.Errorf("failed to update HEAD: %w", err)
		}
	}

	// Save updated index
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// CheckoutFile restores a file from the index
func (r *Repository) CheckoutFile(path string) error {
	// Load index
	indexPath := filepath.Join(r.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Get entry from index
	entry, ok := idx.GetEntry(path)
	if !ok {
		return fmt.Errorf("file %s not in index", path)
	}

	// Get blob from object database
	blobObj, err := r.ObjectDB.Get(entry.Hash)
	if err != nil {
		return fmt.Errorf("failed to load blob: %w", err)
	}

	blob, ok := blobObj.(*object.Blob)
	if !ok {
		return fmt.Errorf("object is not a blob")
	}

	// Write file to working directory
	workTreePath := r.WorkTree()
	filePath := filepath.Join(workTreePath, path)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write file content
	mode := os.FileMode(entry.Mode & 0777)
	if err := os.WriteFile(filePath, blob.Content(), mode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// checkUncommittedChanges checks for uncommitted changes that would be overwritten
func (r *Repository) checkUncommittedChanges(idx *index.Index) error {
	// Get HEAD commit
	headStr, err := r.HEAD()
	if err != nil {
		// No HEAD yet (empty repository), allow checkout
		return nil
	}

	var headCommit *object.Commit
	if headStr[:5] == "ref: " {
		// Symbolic ref
		refName := headStr[5:]
		commitHash, err := r.ResolveRef(refName)
		if err != nil {
			// Branch doesn't exist yet, allow checkout
			return nil
		}

		commitObj, err := r.ObjectDB.Get(commitHash)
		if err != nil {
			return fmt.Errorf("failed to load HEAD commit: %w", err)
		}

		headCommit, _ = commitObj.(*object.Commit)
	} else {
		// Direct hash
		commitHash, err := hash.ParseHash(headStr)
		if err != nil {
			return fmt.Errorf("invalid HEAD hash: %w", err)
		}

		commitObj, err := r.ObjectDB.Get(commitHash)
		if err != nil {
			return fmt.Errorf("failed to load HEAD commit: %w", err)
		}

		headCommit, _ = commitObj.(*object.Commit)
	}

	if headCommit == nil {
		return nil
	}

	// Get status
	workTreePath := r.WorkTree()
	statusOpts := index.DefaultStatusOptions()
	status, err := index.GetStatus(workTreePath, idx, headCommit, r.ObjectDB, statusOpts)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Check for uncommitted changes
	if len(status.Modified) > 0 || len(status.Deleted) > 0 || len(status.Added) > 0 || len(status.Staged) > 0 {
		return fmt.Errorf("uncommitted changes would be overwritten by checkout; use --force to override")
	}

	return nil
}

// resolveCheckoutTarget resolves a checkout target to a commit hash
// Returns (hash, isBranch, error)
func (r *Repository) resolveCheckoutTarget(target string) (hash.Hash, bool, error) {
	// Try as branch name first
	if r.BranchExists(target) {
		h, err := r.GetBranch(target)
		return h, true, err
	}

	// Try as full ref (refs/heads/*, refs/tags/*)
	if len(target) >= 5 && target[:5] == "refs/" {
		h, err := r.ResolveRef(target)
		if err == nil {
			// Check if it's a branch ref
			isBranch := len(target) >= 11 && target[:11] == "refs/heads/"
			return h, isBranch, nil
		}
	}

	// Try as commit hash
	h, err := hash.ParseHash(target)
	if err != nil {
		return nil, false, fmt.Errorf("invalid target: not a branch, ref, or valid hash")
	}

	// Verify the commit exists
	if !r.ObjectDB.Has(h) {
		return nil, false, fmt.Errorf("commit %s not found", h.String())
	}

	return h, false, nil
}

// updateWorkingDirectory updates the working directory and index from a tree
func (r *Repository) updateWorkingDirectory(treeHash hash.Hash, idx *index.Index) error {
	workTreePath := r.WorkTree()

	// Get the tree object
	treeObj, err := r.ObjectDB.Get(treeHash)
	if err != nil {
		return fmt.Errorf("failed to load tree: %w", err)
	}

	tree, ok := treeObj.(*object.Tree)
	if !ok {
		return fmt.Errorf("object is not a tree")
	}

	// Collect all files in the target tree
	targetFiles := make(map[string]struct {
		hash hash.Hash
		mode object.FileMode
	})

	if err := r.collectTreeFiles(tree, "", targetFiles); err != nil {
		return err
	}

	// Remove files that are in index but not in target tree
	for _, entry := range idx.Entries {
		if _, exists := targetFiles[entry.Path]; !exists {
			// File should be removed
			filePath := filepath.Join(workTreePath, entry.Path)
			os.Remove(filePath) // Ignore errors
		}
	}

	// Clear index and rebuild from tree
	idx.Entries = make([]*index.Entry, 0)

	// Write all files from target tree
	for path, file := range targetFiles {
		// Get blob
		blobObj, err := r.ObjectDB.Get(file.hash)
		if err != nil {
			return fmt.Errorf("failed to load blob for %s: %w", path, err)
		}

		blob, ok := blobObj.(*object.Blob)
		if !ok {
			return fmt.Errorf("object is not a blob: %s", path)
		}

		// Write file
		filePath := filepath.Join(workTreePath, path)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}

		// Write content
		mode := os.FileMode(file.mode & 0777)
		if err := os.WriteFile(filePath, blob.Content(), mode); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}

		// Get file info for index entry
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", path, err)
		}

		// Create index entry
		entry := &index.Entry{
			MTime: fileInfo.ModTime(),
			CTime: fileInfo.ModTime(),
			Size:  uint32(fileInfo.Size()),
			Mode:  uint32(file.mode),
			Hash:  file.hash,
			Path:  path,
		}

		idx.AddEntry(entry)
	}

	return nil
}

// collectTreeFiles recursively collects all files in a tree
func (r *Repository) collectTreeFiles(tree *object.Tree, prefix string, files map[string]struct {
	hash hash.Hash
	mode object.FileMode
}) error {
	for _, entry := range tree.Entries() {
		path := entry.Name
		if prefix != "" {
			path = prefix + "/" + entry.Name
		}

		if entry.Mode == object.ModeDir {
			// Recursively process subtree
			subtreeObj, err := r.ObjectDB.Get(entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to load subtree %s: %w", path, err)
			}

			subtree, ok := subtreeObj.(*object.Tree)
			if !ok {
				return fmt.Errorf("subtree %s is not a tree object", path)
			}

			if err := r.collectTreeFiles(subtree, path, files); err != nil {
				return err
			}
		} else {
			// Add file
			files[path] = struct {
				hash hash.Hash
				mode object.FileMode
			}{
				hash: entry.Hash,
				mode: entry.Mode,
			}
		}
	}

	return nil
}
