package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/index"
	"github.com/nseba/browser-git/git-core/pkg/merge"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// MergeOptions contains options for merge operations
type MergeOptions struct {
	// AllowFastForward enables fast-forward merges when possible
	AllowFastForward bool
	// CommitMessage is the message for the merge commit (if creating one)
	CommitMessage string
	// Author is the author signature for the merge commit
	Author *object.Signature
	// Committer is the committer signature for the merge commit
	Committer *object.Signature
}

// DefaultMergeOptions returns default merge options
func DefaultMergeOptions() *MergeOptions {
	return &MergeOptions{
		AllowFastForward: true,
	}
}

// Merge merges a branch into the current branch
func (r *Repository) Merge(branchName string, opts *MergeOptions) (*merge.MergeResult, error) {
	if opts == nil {
		opts = DefaultMergeOptions()
	}

	// Get current HEAD
	currentCommitHash, err := r.ResolveHEAD()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve HEAD: %w", err)
	}

	// Get the branch to merge
	branchRef := fmt.Sprintf("refs/heads/%s", branchName)
	branchCommitHash, err := r.ResolveRef(branchRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve branch %s: %w", branchName, err)
	}

	// Check if we can do a fast-forward merge
	if opts.AllowFastForward {
		canFF, err := merge.CanFastForward(r.ObjectDB, currentCommitHash, branchCommitHash)
		if err != nil {
			return nil, fmt.Errorf("failed to check fast-forward: %w", err)
		}

		if canFF {
			// Perform fast-forward merge
			return r.fastForwardMerge(branchCommitHash, branchName)
		}
	}

	// Find merge base
	mergeBaseHash, err := merge.FindMergeBase(r.ObjectDB, currentCommitHash, branchCommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %w", err)
	}

	// Perform three-way merge
	result, err := merge.ThreeWayMerge(
		r.ObjectDB,
		r.Hasher,
		mergeBaseHash,
		currentCommitHash,
		branchCommitHash,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to perform three-way merge: %w", err)
	}

	// If there are conflicts, return them
	if !result.Success {
		return result, nil
	}

	// Create merge commit
	commitHash, err := r.createMergeCommit(
		result.TreeHash,
		currentCommitHash,
		branchCommitHash,
		branchName,
		opts,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create merge commit: %w", err)
	}

	result.CommitHash = commitHash

	// Update HEAD to point to new commit
	currentBranch, err := r.CurrentBranch()
	if err == nil {
		// Update branch ref
		branchRef := fmt.Sprintf("refs/heads/%s", currentBranch)
		if err := r.UpdateRef(branchRef, commitHash); err != nil {
			return nil, fmt.Errorf("failed to update branch ref: %w", err)
		}
	} else {
		// Detached HEAD - update HEAD directly
		if err := r.SetHEAD(commitHash.String()); err != nil {
			return nil, fmt.Errorf("failed to update HEAD: %w", err)
		}
	}

	// Update working directory
	if err := r.checkoutTree(result.TreeHash); err != nil {
		return nil, fmt.Errorf("failed to update working directory: %w", err)
	}

	return result, nil
}

// fastForwardMerge performs a fast-forward merge
func (r *Repository) fastForwardMerge(targetCommitHash hash.Hash, branchName string) (*merge.MergeResult, error) {
	// Update current branch to point to target commit
	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("cannot fast-forward in detached HEAD state")
	}

	branchRef := fmt.Sprintf("refs/heads/%s", currentBranch)
	if err := r.UpdateRef(branchRef, targetCommitHash); err != nil {
		return nil, fmt.Errorf("failed to update branch ref: %w", err)
	}

	// Load target commit to get its tree
	commit, err := r.ObjectDB.Get(targetCommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load target commit: %w", err)
	}

	targetCommit, ok := commit.(*object.Commit)
	if !ok {
		return nil, fmt.Errorf("object is not a commit")
	}

	// Update working directory
	if err := r.checkoutTree(targetCommit.Tree); err != nil {
		return nil, fmt.Errorf("failed to update working directory: %w", err)
	}

	return &merge.MergeResult{
		Success:       true,
		IsFastForward: true,
		CommitHash:    targetCommitHash,
		TreeHash:      targetCommit.Tree,
	}, nil
}

// createMergeCommit creates a merge commit with two parents
func (r *Repository) createMergeCommit(
	treeHash hash.Hash,
	parent1 hash.Hash,
	parent2 hash.Hash,
	branchName string,
	opts *MergeOptions,
) (hash.Hash, error) {
	// Create commit object
	commit := object.NewCommit()
	commit.Tree = treeHash
	commit.AddParent(parent1)
	commit.AddParent(parent2)

	// Set author and committer
	if opts.Author != nil {
		commit.Author = *opts.Author
	} else {
		// Use config or default
		userName, userEmail := r.Config.GetUser()
		commit.Author = object.Signature{
			Name:  userName,
			Email: userEmail,
			When:  time.Now(),
		}
	}

	if opts.Committer != nil {
		commit.Committer = *opts.Committer
	} else {
		commit.Committer = commit.Author
	}

	// Set commit message
	if opts.CommitMessage != "" {
		commit.Message = opts.CommitMessage
	} else {
		commit.Message = fmt.Sprintf("Merge branch '%s'", branchName)
	}

	// Compute hash
	if err := commit.ComputeHash(r.Hasher); err != nil {
		return nil, fmt.Errorf("failed to compute commit hash: %w", err)
	}

	// Store commit
	commitHash, err := r.ObjectDB.Put(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to write commit: %w", err)
	}

	return commitHash, nil
}

// ResolveHEAD resolves HEAD to a commit hash
func (r *Repository) ResolveHEAD() (hash.Hash, error) {
	head, err := r.HEAD()
	if err != nil {
		return nil, err
	}

	// If HEAD is a symbolic ref, resolve it
	const prefix = "ref: "
	if len(head) > len(prefix) && head[:len(prefix)] == prefix {
		refName := head[len(prefix):]
		return r.ResolveRef(refName)
	}

	// HEAD is a direct hash
	return hash.ParseHash(head)
}

// checkoutTree checks out a tree to the working directory
func (r *Repository) checkoutTree(treeHash hash.Hash) error {
	// Load the tree
	obj, err := r.ObjectDB.Get(treeHash)
	if err != nil {
		return fmt.Errorf("failed to load tree: %w", err)
	}

	tree, ok := obj.(*object.Tree)
	if !ok {
		return fmt.Errorf("object is not a tree")
	}

	// Update index
	indexPath := filepath.Join(r.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		// If index doesn't exist, create a new one
		idx = index.NewIndex()
	}

	// Clear index
	idx.Clear()

	// Write tree contents to working directory and update index
	if err := r.checkoutTreeRecursive(tree, "", idx); err != nil {
		return err
	}

	// Save index
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// checkoutTreeRecursive recursively checks out a tree
func (r *Repository) checkoutTreeRecursive(tree *object.Tree, prefix string, idx *index.Index) error {
	entries := tree.Entries()
	for _, entry := range entries {
		path := prefix + entry.Name

		if entry.Mode == object.ModeDir {
			// Create directory
			if !r.IsBare() {
				dirPath := filepath.Join(r.WorkTree(), path)
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", path, err)
				}
			}

			// Load subtree
			obj, err := r.ObjectDB.Get(entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to load subtree: %w", err)
			}

			subtree, ok := obj.(*object.Tree)
			if !ok {
				return fmt.Errorf("object is not a tree")
			}

			// Recurse
			if err := r.checkoutTreeRecursive(subtree, path+"/", idx); err != nil {
				return err
			}
		} else {
			// Write file
			if !r.IsBare() {
				obj, err := r.ObjectDB.Get(entry.Hash)
				if err != nil {
					return fmt.Errorf("failed to load blob: %w", err)
				}

				blob, ok := obj.(*object.Blob)
				if !ok {
					return fmt.Errorf("object is not a blob")
				}

				filePath := filepath.Join(r.WorkTree(), path)
				// Ensure parent directory exists
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					return fmt.Errorf("failed to create parent directory: %w", err)
				}
				if err := os.WriteFile(filePath, blob.Content(), os.FileMode(entry.Mode)); err != nil {
					return fmt.Errorf("failed to write file %s: %w", path, err)
				}
			}

			// Add to index
			indexEntry := &index.Entry{
				Path:  path,
				Hash:  entry.Hash,
				Mode:  uint32(entry.Mode),
				MTime: time.Now(),
				CTime: time.Now(),
			}
			idx.AddEntry(indexEntry)
		}
	}

	return nil
}
