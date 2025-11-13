package merge

import (
	"fmt"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// FindMergeBase finds the common ancestor (merge base) between two commits
// using a simple algorithm that traverses both commit histories
func FindMergeBase(db object.Database, commit1Hash, commit2Hash hash.Hash) (hash.Hash, error) {
	// Handle special case: if commits are the same
	if commit1Hash.String() == commit2Hash.String() {
		return commit1Hash, nil
	}

	// Build ancestry set for commit1
	commit1Ancestors, err := getAncestors(db, commit1Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors of commit1: %w", err)
	}

	// Traverse commit2 history until we find a commit in commit1's ancestry
	mergeBase, err := findFirstCommonAncestor(db, commit2Hash, commit1Ancestors)
	if err != nil {
		return nil, fmt.Errorf("failed to find common ancestor: %w", err)
	}

	if mergeBase == nil {
		return nil, fmt.Errorf("no common ancestor found")
	}

	return mergeBase, nil
}

// getAncestors returns a set of all ancestors of a commit (including itself)
func getAncestors(db object.Database, commitHash hash.Hash) (map[string]bool, error) {
	ancestors := make(map[string]bool)
	queue := []hash.Hash{commitHash}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		currentStr := current.String()

		// Skip if already visited
		if visited[currentStr] {
			continue
		}
		visited[currentStr] = true
		ancestors[currentStr] = true

		// Load commit
		commit, err := loadCommit(db, current)
		if err != nil {
			return nil, fmt.Errorf("failed to load commit %s: %w", currentStr, err)
		}

		// Enqueue parents
		for _, parent := range commit.Parents {
			if !visited[parent.String()] {
				queue = append(queue, parent)
			}
		}
	}

	return ancestors, nil
}

// findFirstCommonAncestor traverses from startCommit and finds the first commit
// that exists in the ancestors set
func findFirstCommonAncestor(
	db object.Database,
	startCommit hash.Hash,
	ancestors map[string]bool,
) (hash.Hash, error) {
	queue := []hash.Hash{startCommit}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		currentStr := current.String()

		// Skip if already visited
		if visited[currentStr] {
			continue
		}
		visited[currentStr] = true

		// Check if this commit is in the ancestor set
		if ancestors[currentStr] {
			return current, nil
		}

		// Load commit
		commit, err := loadCommit(db, current)
		if err != nil {
			return nil, fmt.Errorf("failed to load commit %s: %w", currentStr, err)
		}

		// Enqueue parents
		for _, parent := range commit.Parents {
			if !visited[parent.String()] {
				queue = append(queue, parent)
			}
		}
	}

	return nil, nil
}

// CanFastForward checks if we can do a fast-forward merge from 'from' to 'to'
// A fast-forward is possible if 'from' is an ancestor of 'to'
func CanFastForward(db object.Database, from, to hash.Hash) (bool, error) {
	// If they're the same, it's trivially a fast-forward (nothing to do)
	if from.String() == to.String() {
		return true, nil
	}

	// Check if 'from' is an ancestor of 'to'
	ancestors, err := getAncestors(db, to)
	if err != nil {
		return false, fmt.Errorf("failed to get ancestors: %w", err)
	}

	return ancestors[from.String()], nil
}

// IsAncestor checks if 'ancestor' is an ancestor of 'descendant'
func IsAncestor(db object.Database, ancestor, descendant hash.Hash) (bool, error) {
	if ancestor.String() == descendant.String() {
		return true, nil
	}

	ancestors, err := getAncestors(db, descendant)
	if err != nil {
		return false, err
	}

	return ancestors[ancestor.String()], nil
}
