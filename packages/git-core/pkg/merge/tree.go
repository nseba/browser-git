package merge

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// TreeMerger handles merging of Git trees (directory structures)
type TreeMerger struct {
	db     object.Database
	hasher hash.Hasher
}

// NewTreeMerger creates a new tree merger
func NewTreeMerger(db object.Database, hasher hash.Hasher) *TreeMerger {
	return &TreeMerger{
		db:     db,
		hasher: hasher,
	}
}

// MergeTrees merges three trees (base, ours, theirs) and returns merged tree hash and conflicts
func (tm *TreeMerger) MergeTrees(
	baseTreeHash hash.Hash,
	ourTreeHash hash.Hash,
	theirTreeHash hash.Hash,
	path string,
) (hash.Hash, []Conflict, error) {
	// Load all three trees
	baseTree, err := tm.loadTreeOrNil(baseTreeHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load base tree: %w", err)
	}

	ourTree, err := tm.loadTreeOrNil(ourTreeHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load our tree: %w", err)
	}

	theirTree, err := tm.loadTreeOrNil(theirTreeHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load their tree: %w", err)
	}

	// Collect all entries from all trees
	allEntries := tm.collectAllEntries(baseTree, ourTree, theirTree)

	// Merge each entry
	mergedTree := object.NewTree()
	var conflicts []Conflict

	for name, entries := range allEntries {
		entryPath := filepath.Join(path, name)

		conflict, mergedEntry, err := tm.mergeEntry(
			entries.base,
			entries.ours,
			entries.theirs,
			entryPath,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to merge entry %s: %w", name, err)
		}

		if conflict != nil {
			conflicts = append(conflicts, *conflict)
		} else if mergedEntry != nil {
			mergedTree.AddEntry(*mergedEntry)
		}
		// If both are nil, the entry was deleted on both sides (no-op)
	}

	// If there are conflicts, return them without storing the tree
	if len(conflicts) > 0 {
		return nil, conflicts, nil
	}

	// Sort tree entries (required by Git format)
	mergedTree.Sort()

	// Store the merged tree
	treeHash, err := tm.db.Put(mergedTree)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write merged tree: %w", err)
	}

	return treeHash, nil, nil
}

// entrySet holds the entries from base, ours, and theirs for a given path
type entrySet struct {
	base   *object.TreeEntry
	ours   *object.TreeEntry
	theirs *object.TreeEntry
}

// collectAllEntries collects all entries from all three trees
func (tm *TreeMerger) collectAllEntries(
	baseTree, ourTree, theirTree *object.Tree,
) map[string]entrySet {
	entries := make(map[string]entrySet)

	// Collect from base
	if baseTree != nil {
		baseEntries := baseTree.Entries()
		for i := range baseEntries {
			entry := &baseEntries[i]
			set := entries[entry.Name]
			set.base = entry
			entries[entry.Name] = set
		}
	}

	// Collect from ours
	if ourTree != nil {
		ourEntries := ourTree.Entries()
		for i := range ourEntries {
			entry := &ourEntries[i]
			set := entries[entry.Name]
			set.ours = entry
			entries[entry.Name] = set
		}
	}

	// Collect from theirs
	if theirTree != nil {
		theirEntries := theirTree.Entries()
		for i := range theirEntries {
			entry := &theirEntries[i]
			set := entries[entry.Name]
			set.theirs = entry
			entries[entry.Name] = set
		}
	}

	return entries
}

// mergeEntry merges a single tree entry
func (tm *TreeMerger) mergeEntry(
	base, ours, theirs *object.TreeEntry,
	path string,
) (*Conflict, *object.TreeEntry, error) {
	// Case 1: Entry unchanged on both sides
	if ours != nil && theirs != nil && tm.entriesEqual(ours, theirs) {
		return nil, ours, nil
	}

	// Case 2: Entry only changed on our side
	if theirs != nil && base != nil && tm.entriesEqual(theirs, base) {
		if ours != nil {
			return nil, ours, nil
		}
		// We deleted it
		return nil, nil, nil
	}

	// Case 3: Entry only changed on their side
	if ours != nil && base != nil && tm.entriesEqual(ours, base) {
		if theirs != nil {
			return nil, theirs, nil
		}
		// They deleted it
		return nil, nil, nil
	}

	// Case 4: Entry added on both sides with same content
	if base == nil && ours != nil && theirs != nil && tm.entriesEqual(ours, theirs) {
		return nil, ours, nil
	}

	// Case 5: Both sides are directories - recurse
	if ours != nil && theirs != nil &&
		ours.Mode == object.ModeDir && theirs.Mode == object.ModeDir {

		var baseHash hash.Hash
		if base != nil && base.Mode == object.ModeDir {
			baseHash = base.Hash
		}

		mergedHash, conflicts, err := tm.MergeTrees(baseHash, ours.Hash, theirs.Hash, path)
		if err != nil {
			return nil, nil, err
		}

		if len(conflicts) > 0 {
			// Propagate conflicts up
			return &conflicts[0], nil, nil
		}

		return nil, &object.TreeEntry{
			Mode: object.ModeDir,
			Name: ours.Name,
			Hash: mergedHash,
		}, nil
	}

	// Case 6: Conflict - both sides modified differently or incompatible types
	return tm.createConflict(base, ours, theirs, path)
}

// createConflict creates a conflict object for an entry
func (tm *TreeMerger) createConflict(
	base, ours, theirs *object.TreeEntry,
	path string,
) (*Conflict, *object.TreeEntry, error) {
	conflict := &Conflict{
		Path: path,
		Type: ContentConflict,
	}

	// Load content from each side
	var err error

	if base != nil && base.Mode != object.ModeDir {
		conflict.Base, err = loadBlobContent(tm.db, base.Hash)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load base content: %w", err)
		}
	}

	if ours != nil && ours.Mode != object.ModeDir {
		conflict.Ours, err = loadBlobContent(tm.db, ours.Hash)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load our content: %w", err)
		}
	}

	if theirs != nil && theirs.Mode != object.ModeDir {
		conflict.Theirs, err = loadBlobContent(tm.db, theirs.Hash)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load their content: %w", err)
		}
	}

	// Determine conflict type
	conflict.IsBinary = isBinaryContent(conflict.Base) ||
		isBinaryContent(conflict.Ours) ||
		isBinaryContent(conflict.Theirs)

	// Set change types for metadata
	ourChangeType := "modified"
	theirChangeType := "modified"

	if conflict.IsBinary {
		conflict.Type = BinaryConflict
	} else if base == nil {
		conflict.Type = AddConflict
		ourChangeType = "added"
		theirChangeType = "added"
	} else if ours == nil || theirs == nil {
		conflict.Type = DeleteConflict
		if ours == nil {
			ourChangeType = "deleted"
		}
		if theirs == nil {
			theirChangeType = "deleted"
		}
	}

	// Add metadata
	conflict.Metadata = &ConflictMetadata{
		StartLine:       1,
		EndLine:         countLines(conflict.Base),
		OurChangeType:   ourChangeType,
		TheirChangeType: theirChangeType,
	}

	return conflict, nil, nil
}

// countLines counts the number of lines in content
func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	count := 1
	for _, b := range content {
		if b == '\n' {
			count++
		}
	}
	return count
}

// entriesEqual checks if two tree entries are equal
func (tm *TreeMerger) entriesEqual(e1, e2 *object.TreeEntry) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}

	return e1.Mode == e2.Mode &&
		e1.Name == e2.Name &&
		e1.Hash.String() == e2.Hash.String()
}

// loadTreeOrNil loads a tree or returns nil if hash is nil
func (tm *TreeMerger) loadTreeOrNil(h hash.Hash) (*object.Tree, error) {
	if h == nil || h.String() == "" || isZeroHash(h) {
		return nil, nil
	}

	return loadTree(tm.db, h)
}

// isZeroHash checks if a hash is all zeros
func isZeroHash(h hash.Hash) bool {
	if h == nil {
		return true
	}

	bytes := h.Bytes()
	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}
	return true
}

// FlattenTreePaths returns all file paths in a tree (recursively)
func FlattenTreePaths(db object.Database, treeHash hash.Hash, prefix string) (map[string]hash.Hash, error) {
	paths := make(map[string]hash.Hash)

	tree, err := loadTree(db, treeHash)
	if err != nil {
		return nil, err
	}

	entries := tree.Entries()
	for _, entry := range entries {
		entryPath := filepath.Join(prefix, entry.Name)

		if entry.Mode == object.ModeDir {
			// Recursively flatten subdirectory
			subPaths, err := FlattenTreePaths(db, entry.Hash, entryPath)
			if err != nil {
				return nil, err
			}

			for path, hash := range subPaths {
				paths[path] = hash
			}
		} else {
			// Regular file
			paths[entryPath] = entry.Hash
		}
	}

	return paths, nil
}

// BuildTreeFromPaths builds a tree from a flat map of paths to blob hashes
func BuildTreeFromPaths(
	db object.Database,
	hasher hash.Hasher,
	paths map[string]hash.Hash,
) (hash.Hash, error) {
	// Group paths by directory
	dirContents := make(map[string]map[string]hash.Hash)

	for path, blobHash := range paths {
		dir := filepath.Dir(path)
		base := filepath.Base(path)

		if dir == "." {
			dir = ""
		}

		if dirContents[dir] == nil {
			dirContents[dir] = make(map[string]hash.Hash)
		}
		dirContents[dir][base] = blobHash
	}

	// Build trees bottom-up
	return buildTreeRecursive(db, hasher, dirContents, "")
}

// buildTreeRecursive builds a tree recursively
func buildTreeRecursive(
	db object.Database,
	hasher hash.Hasher,
	dirContents map[string]map[string]hash.Hash,
	currentDir string,
) (hash.Hash, error) {
	tree := object.NewTree()

	contents := dirContents[currentDir]
	if contents == nil {
		contents = make(map[string]hash.Hash)
	}

	for name, hash := range contents {
		// Check if this is a subdirectory
		subDir := filepath.Join(currentDir, name)
		if dirContents[subDir] != nil {
			// Build subdirectory tree
			subTreeHash, err := buildTreeRecursive(db, hasher, dirContents, subDir)
			if err != nil {
				return nil, err
			}

			tree.AddEntryWithMode(object.ModeDir, name, subTreeHash)
		} else {
			// Regular file
			tree.AddEntryWithMode(object.ModeRegular, name, hash)
		}
	}

	tree.Sort()

	treeHash, err := db.Put(tree)
	if err != nil {
		return nil, err
	}

	return treeHash, nil
}

// NormalizePathSeparators normalizes path separators to forward slashes
func NormalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
