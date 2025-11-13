package merge

import (
	"fmt"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// ConflictType represents the type of conflict
type ConflictType int

const (
	// ContentConflict indicates both sides modified the same content
	ContentConflict ConflictType = iota
	// BinaryConflict indicates a conflict in a binary file
	BinaryConflict
	// DeleteConflict indicates one side deleted while other modified
	DeleteConflict
	// AddConflict indicates both sides added different content at same path
	AddConflict
)

// Conflict represents a merge conflict
type Conflict struct {
	// Path is the file path with the conflict
	Path string
	// Type is the conflict type
	Type ConflictType
	// Base is the content from the common ancestor (may be nil)
	Base []byte
	// Ours is the content from the current branch
	Ours []byte
	// Theirs is the content from the branch being merged
	Theirs []byte
	// IsBinary indicates if this is a binary file conflict
	IsBinary bool
}

// MergeResult represents the result of a merge operation
type MergeResult struct {
	// Success indicates if the merge completed without conflicts
	Success bool
	// TreeHash is the hash of the merged tree (only set if Success is true)
	TreeHash hash.Hash
	// Conflicts is the list of conflicts (only set if Success is false)
	Conflicts []Conflict
	// IsFastForward indicates if this was a fast-forward merge
	IsFastForward bool
	// CommitHash is the target commit hash (for fast-forward merges)
	CommitHash hash.Hash
}

// FileChange represents a change to a file in a merge
type FileChange struct {
	Path    string
	BaseOID hash.Hash // nil if file didn't exist in base
	OurOID  hash.Hash // nil if file doesn't exist in ours
	TheirOID hash.Hash // nil if file doesn't exist in theirs
}

// ThreeWayMerge performs a three-way merge between two commits
// It finds the merge base and merges the changes from both sides
func ThreeWayMerge(
	db object.Database,
	hasher hash.Hasher,
	baseCommitHash hash.Hash,
	ourCommitHash hash.Hash,
	theirCommitHash hash.Hash,
) (*MergeResult, error) {
	// Load the commits
	baseCommit, err := loadCommit(db, baseCommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load base commit: %w", err)
	}

	ourCommit, err := loadCommit(db, ourCommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load our commit: %w", err)
	}

	theirCommit, err := loadCommit(db, theirCommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load their commit: %w", err)
	}

	// Use TreeMerger to merge the trees
	merger := NewTreeMerger(db, hasher)
	mergedTreeHash, conflicts, err := merger.MergeTrees(
		baseCommit.Tree,
		ourCommit.Tree,
		theirCommit.Tree,
		"",
	)

	if err != nil {
		return nil, fmt.Errorf("failed to merge trees: %w", err)
	}

	// If there are conflicts, return them
	if len(conflicts) > 0 {
		return &MergeResult{
			Success:   false,
			Conflicts: conflicts,
		}, nil
	}

	return &MergeResult{
		Success:  true,
		TreeHash: mergedTreeHash,
	}, nil
}


// Helper functions

func loadCommit(db object.Database, h hash.Hash) (*object.Commit, error) {
	obj, err := db.Get(h)
	if err != nil {
		return nil, err
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return nil, fmt.Errorf("object is not a commit")
	}

	return commit, nil
}

func loadTree(db object.Database, h hash.Hash) (*object.Tree, error) {
	obj, err := db.Get(h)
	if err != nil {
		return nil, err
	}

	tree, ok := obj.(*object.Tree)
	if !ok {
		return nil, fmt.Errorf("object is not a tree")
	}

	return tree, nil
}

func loadBlobContent(db object.Database, h hash.Hash) ([]byte, error) {
	obj, err := db.Get(h)
	if err != nil {
		return nil, err
	}

	blob, ok := obj.(*object.Blob)
	if !ok {
		return nil, fmt.Errorf("object is not a blob")
	}

	return blob.Content(), nil
}


func isBinaryContent(content []byte) bool {
	if len(content) == 0 {
		return false
	}

	// Check first 8000 bytes for null bytes (common binary indicator)
	checkLen := len(content)
	if checkLen > 8000 {
		checkLen = 8000
	}

	for i := 0; i < checkLen; i++ {
		if content[i] == 0 {
			return true
		}
	}

	return false
}

// GenerateConflictMarkers generates Git-style conflict markers for a conflict
func GenerateConflictMarkers(conflict Conflict) string {
	if conflict.IsBinary {
		return fmt.Sprintf("Binary file conflict in %s\n", conflict.Path)
	}

	var buf strings.Builder

	buf.WriteString("<<<<<<< HEAD\n")
	buf.Write(conflict.Ours)
	buf.WriteString("\n=======\n")
	buf.Write(conflict.Theirs)
	buf.WriteString("\n>>>>>>> MERGE\n")

	return buf.String()
}
