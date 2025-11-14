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

// ConflictResolutionStrategy represents how to resolve a conflict
type ConflictResolutionStrategy string

const (
	// AcceptOurs resolves by accepting our version
	AcceptOurs ConflictResolutionStrategy = "ours"
	// AcceptTheirs resolves by accepting their version
	AcceptTheirs ConflictResolutionStrategy = "theirs"
	// AcceptManual resolves using manually edited content
	AcceptManual ConflictResolutionStrategy = "manual"
)

// ConflictState stores the state of an ongoing merge with conflicts
type ConflictState struct {
	// Conflicts is the list of unresolved conflicts
	Conflicts []merge.Conflict
	// OurCommit is the commit we're merging into
	OurCommit hash.Hash
	// TheirCommit is the commit we're merging from
	TheirCommit hash.Hash
	// MergeBase is the common ancestor commit
	MergeBase hash.Hash
	// BranchName is the name of the branch being merged
	BranchName string
}

// GetConflicts retrieves the current conflicts from the repository
func (r *Repository) GetConflicts() (*ConflictState, error) {
	// Check if MERGE_HEAD exists
	mergeHeadPath := filepath.Join(r.GitDir, "MERGE_HEAD")
	if _, err := os.Stat(mergeHeadPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no merge in progress")
	}

	// Read MERGE_HEAD
	mergeHeadData, err := os.ReadFile(mergeHeadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MERGE_HEAD: %w", err)
	}

	// Trim whitespace (including newlines)
	mergeHeadStr := string(mergeHeadData)
	if len(mergeHeadStr) > 0 && mergeHeadStr[len(mergeHeadStr)-1] == '\n' {
		mergeHeadStr = mergeHeadStr[:len(mergeHeadStr)-1]
	}

	theirCommit, err := hash.ParseHash(mergeHeadStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MERGE_HEAD: %w", err)
	}

	// Get current HEAD
	ourCommit, err := r.ResolveHEAD()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve HEAD: %w", err)
	}

	// Find merge base
	mergeBase, err := merge.FindMergeBase(r.ObjectDB, ourCommit, theirCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %w", err)
	}

	// Read branch name from MERGE_MSG if available
	branchName := "MERGE"
	mergeMsgPath := filepath.Join(r.GitDir, "MERGE_MSG")
	if msgData, err := os.ReadFile(mergeMsgPath); err == nil {
		// Extract branch name from merge message
		// Format: "Merge branch 'branch-name'"
		msg := string(msgData)
		if len(msg) > 14 && msg[:14] == "Merge branch '" {
			endIdx := 14
			for endIdx < len(msg) && msg[endIdx] != '\'' {
				endIdx++
			}
			if endIdx < len(msg) {
				branchName = msg[14:endIdx]
			}
		}
	}

	// Read conflicts from MERGE_CONFLICTS file
	conflictsPath := filepath.Join(r.GitDir, "MERGE_CONFLICTS")
	conflictsData, err := os.ReadFile(conflictsPath)
	if err != nil {
		// If no conflicts file, try to re-compute conflicts
		return &ConflictState{
			Conflicts:   []merge.Conflict{},
			OurCommit:   ourCommit,
			TheirCommit: theirCommit,
			MergeBase:   mergeBase,
			BranchName:  branchName,
		}, nil
	}

	// Parse conflicts (simple format: one path per line)
	conflicts := make([]merge.Conflict, 0)
	lines := splitLines(conflictsData)
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Try to read the conflict markers from the file
		filePath := filepath.Join(r.WorkTree(), line)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Check if file has conflict markers
		if hasConflictMarkers(content) {
			conflicts = append(conflicts, merge.Conflict{
				Path: line,
				Type: merge.ContentConflict,
			})
		}
	}

	return &ConflictState{
		Conflicts:   conflicts,
		OurCommit:   ourCommit,
		TheirCommit: theirCommit,
		MergeBase:   mergeBase,
		BranchName:  branchName,
	}, nil
}

// ResolveConflict resolves a conflict using the specified strategy
func (r *Repository) ResolveConflict(path string, strategy ConflictResolutionStrategy, manualContent []byte) error {
	// Check if merge is in progress
	state, err := r.GetConflicts()
	if err != nil {
		return fmt.Errorf("no merge in progress: %w", err)
	}

	// Find the conflict for this path
	conflictIdx := -1
	for i, c := range state.Conflicts {
		if c.Path == path {
			conflictIdx = i
			break
		}
	}

	if conflictIdx == -1 {
		return fmt.Errorf("no conflict found for path: %s", path)
	}

	conflict := state.Conflicts[conflictIdx]

	// Determine the resolved content
	var resolvedContent []byte
	switch strategy {
	case AcceptOurs:
		resolvedContent = conflict.Ours
	case AcceptTheirs:
		resolvedContent = conflict.Theirs
	case AcceptManual:
		if manualContent == nil {
			return fmt.Errorf("manual content required for manual resolution")
		}
		resolvedContent = manualContent
	default:
		return fmt.Errorf("unknown resolution strategy: %s", strategy)
	}

	// Write resolved content to working directory
	filePath := filepath.Join(r.WorkTree(), path)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(filePath, resolvedContent, 0644); err != nil {
		return fmt.Errorf("failed to write resolved file: %w", err)
	}

	// Create blob for resolved content
	blob := object.NewBlob(resolvedContent)
	blobHash, err := r.ObjectDB.Put(blob)
	if err != nil {
		return fmt.Errorf("failed to store blob: %w", err)
	}

	// Update index with resolved content
	indexPath := filepath.Join(r.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		idx = index.NewIndex()
	}

	// Add resolved file to index
	entry := &index.Entry{
		Path:      path,
		Hash:      blobHash,
		Mode:      index.FileModeRegular,
		MTime:     time.Now(),
		CTime:     time.Now(),
		StageFlag: 0, // Stage 0 = resolved
	}
	idx.AddEntry(entry)

	// Save index
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Remove conflict from list
	state.Conflicts = append(state.Conflicts[:conflictIdx], state.Conflicts[conflictIdx+1:]...)

	// Update MERGE_CONFLICTS file
	if err := r.saveConflictsState(state); err != nil {
		return fmt.Errorf("failed to save conflicts state: %w", err)
	}

	// If all conflicts resolved, clean up merge state
	if len(state.Conflicts) == 0 {
		if err := r.cleanupMergeState(); err != nil {
			return fmt.Errorf("failed to cleanup merge state: %w", err)
		}
	}

	return nil
}

// ContinueMerge continues a merge after all conflicts have been resolved
func (r *Repository) ContinueMerge(message string) (hash.Hash, error) {
	// Check if merge is in progress
	state, err := r.GetConflicts()
	if err != nil {
		return nil, fmt.Errorf("no merge in progress: %w", err)
	}

	// Check if all conflicts are resolved
	if len(state.Conflicts) > 0 {
		return nil, fmt.Errorf("cannot continue merge: %d conflicts remaining", len(state.Conflicts))
	}

	// Load index to get the tree
	indexPath := filepath.Join(r.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	// Build tree from index
	treeHash, err := buildTreeFromIndex(r, idx)
	if err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	// Create merge commit
	commit := object.NewCommit()
	commit.Tree = treeHash
	commit.AddParent(state.OurCommit)
	commit.AddParent(state.TheirCommit)

	userName, userEmail := r.Config.GetUser()
	commit.Author = object.Signature{
		Name:  userName,
		Email: userEmail,
		When:  time.Now(),
	}
	commit.Committer = commit.Author

	if message == "" {
		message = fmt.Sprintf("Merge branch '%s'", state.BranchName)
	}
	commit.Message = message

	// Store commit
	commitHash, err := r.ObjectDB.Put(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to store commit: %w", err)
	}

	// Update HEAD
	currentBranch, err := r.CurrentBranch()
	if err == nil {
		branchRef := fmt.Sprintf("refs/heads/%s", currentBranch)
		if err := r.UpdateRef(branchRef, commitHash); err != nil {
			return nil, fmt.Errorf("failed to update branch: %w", err)
		}
	} else {
		if err := r.SetHEAD(commitHash.String()); err != nil {
			return nil, fmt.Errorf("failed to update HEAD: %w", err)
		}
	}

	// Clean up merge state
	if err := r.cleanupMergeState(); err != nil {
		return nil, fmt.Errorf("failed to cleanup merge state: %w", err)
	}

	return commitHash, nil
}

// AbortMerge aborts an ongoing merge and returns to the pre-merge state
func (r *Repository) AbortMerge() error {
	// Check if merge is in progress
	_, err := r.GetConflicts()
	if err != nil {
		return fmt.Errorf("no merge in progress: %w", err)
	}

	// Reset to HEAD
	headHash, err := r.ResolveHEAD()
	if err != nil {
		return fmt.Errorf("failed to resolve HEAD: %w", err)
	}

	// Load HEAD commit
	obj, err := r.ObjectDB.Get(headHash)
	if err != nil {
		return fmt.Errorf("failed to load HEAD commit: %w", err)
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return fmt.Errorf("HEAD is not a commit")
	}

	// Checkout HEAD tree
	if err := r.checkoutTree(commit.Tree); err != nil {
		return fmt.Errorf("failed to checkout HEAD: %w", err)
	}

	// Clean up merge state
	if err := r.cleanupMergeState(); err != nil {
		return fmt.Errorf("failed to cleanup merge state: %w", err)
	}

	return nil
}

// Helper functions

func (r *Repository) saveConflictsState(state *ConflictState) error {
	conflictsPath := filepath.Join(r.GitDir, "MERGE_CONFLICTS")

	if len(state.Conflicts) == 0 {
		// Remove file if no conflicts
		os.Remove(conflictsPath)
		return nil
	}

	// Write conflict paths
	var content string
	for _, c := range state.Conflicts {
		content += c.Path + "\n"
	}

	return os.WriteFile(conflictsPath, []byte(content), 0644)
}

func (r *Repository) cleanupMergeState() error {
	// Remove merge-related files
	filesToRemove := []string{
		"MERGE_HEAD",
		"MERGE_MSG",
		"MERGE_CONFLICTS",
		"MERGE_MODE",
	}

	for _, file := range filesToRemove {
		path := filepath.Join(r.GitDir, file)
		os.Remove(path) // Ignore errors
	}

	return nil
}

func splitLines(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}
	content := string(data)
	if content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	if content == "" {
		return []string{}
	}
	return splitString(content, "\n")
}

func splitString(s, sep string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func hasConflictMarkers(content []byte) bool {
	s := string(content)
	return containsSubstr(s, "<<<<<<<") && containsSubstr(s, ">>>>>>>") && containsSubstr(s, "=======")
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func buildTreeFromIndex(r *Repository, idx *index.Index) (hash.Hash, error) {
	// This is a simplified version - a full implementation would handle
	// directory trees properly
	tree := object.NewTree()

	for _, entry := range idx.Entries {
		tree.AddEntryWithMode(object.FileMode(entry.Mode), entry.Path, entry.Hash)
	}

	tree.Sort()

	treeHash, err := r.ObjectDB.Put(tree)
	if err != nil {
		return nil, err
	}

	return treeHash, nil
}
