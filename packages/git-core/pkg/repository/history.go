package repository

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

// LogOptions contains options for log operations
type LogOptions struct {
	// MaxCount limits the number of commits to return
	MaxCount int

	// Author filters commits by author name or email
	Author string

	// Since filters commits after this date
	Since *time.Time

	// Until filters commits before this date
	Until *time.Time

	// Path filters commits that touched this path
	Path string

	// Format specifies the output format (full, oneline, etc.)
	Format LogFormat

	// Graph enables graph visualization
	Graph bool

	// All includes all branches
	All bool

	// FirstParent follows only first parent
	FirstParent bool
}

// LogFormat specifies the format for log output
type LogFormat int

const (
	// LogFormatFull shows full commit details
	LogFormatFull LogFormat = iota
	// LogFormatOneline shows one line per commit
	LogFormatOneline
	// LogFormatShort shows abbreviated commit info
	LogFormatShort
)

// DefaultLogOptions returns default log options
func DefaultLogOptions() LogOptions {
	return LogOptions{
		MaxCount:    -1, // unlimited
		Format:      LogFormatFull,
		Graph:       false,
		All:         false,
		FirstParent: false,
	}
}

// LogEntry represents a commit in the log
type LogEntry struct {
	Commit  *object.Commit
	Hash    hash.Hash
	Refs    []string // Branch/tag names pointing to this commit
	Parents []hash.Hash
}

// Log returns the commit history
func (r *Repository) Log(startRef string, opts LogOptions) ([]*LogEntry, error) {
	// Resolve starting point
	var startHash hash.Hash
	var err error

	if startRef == "" {
		// Use HEAD
		headStr, err := r.HEAD()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD: %w", err)
		}

		if headStr[:5] == "ref: " {
			refName := headStr[5:]
			startHash, err = r.ResolveRef(refName)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve HEAD: %w", err)
			}
		} else {
			startHash, err = hash.ParseHash(headStr)
			if err != nil {
				return nil, fmt.Errorf("invalid HEAD hash: %w", err)
			}
		}
	} else {
		// Try to resolve the ref
		if r.BranchExists(startRef) {
			startHash, err = r.GetBranch(startRef)
			if err != nil {
				return nil, err
			}
		} else {
			// Try as ref or hash
			startHash, err = r.ResolveRef(startRef)
			if err != nil {
				// Try as direct hash
				startHash, err = hash.ParseHash(startRef)
				if err != nil {
					return nil, fmt.Errorf("invalid ref or hash: %s", startRef)
				}
			}
		}
	}

	// Get all refs if needed
	refs := make(map[string][]string)
	if opts.All || opts.Graph {
		branches, err := r.ListBranches()
		if err == nil {
			for _, branch := range branches {
				branchHash, err := r.GetBranch(branch)
				if err == nil {
					refs[branchHash.String()] = append(refs[branchHash.String()], branch)
				}
			}
		}
	}

	// Traverse commit history
	entries, err := r.traverseCommits(startHash, opts, refs)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// traverseCommits walks the commit graph
func (r *Repository) traverseCommits(startHash hash.Hash, opts LogOptions, refs map[string][]string) ([]*LogEntry, error) {
	entries := make([]*LogEntry, 0)
	visited := make(map[string]bool)
	queue := []hash.Hash{startHash}

	for len(queue) > 0 && (opts.MaxCount < 0 || len(entries) < opts.MaxCount) {
		// Dequeue
		currentHash := queue[0]
		queue = queue[1:]

		// Skip if visited
		hashStr := currentHash.String()
		if visited[hashStr] {
			continue
		}
		visited[hashStr] = true

		// Get commit object
		commitObj, err := r.ObjectDB.Get(currentHash)
		if err != nil {
			continue // Skip if commit not found
		}

		commit, ok := commitObj.(*object.Commit)
		if !ok {
			continue
		}

		// Apply filters
		if !r.matchesFilters(commit, opts) {
			// Still traverse parents
			if !opts.FirstParent {
				queue = append(queue, commit.Parents...)
			} else if len(commit.Parents) > 0 {
				queue = append(queue, commit.Parents[0])
			}
			continue
		}

		// Create log entry
		entry := &LogEntry{
			Commit:  commit,
			Hash:    currentHash,
			Refs:    refs[hashStr],
			Parents: commit.Parents,
		}

		entries = append(entries, entry)

		// Add parents to queue
		if !opts.FirstParent {
			queue = append(queue, commit.Parents...)
		} else if len(commit.Parents) > 0 {
			queue = append(queue, commit.Parents[0])
		}
	}

	return entries, nil
}

// matchesFilters checks if a commit matches the filter criteria
func (r *Repository) matchesFilters(commit *object.Commit, opts LogOptions) bool {
	// Author filter
	if opts.Author != "" {
		authorMatch := strings.Contains(strings.ToLower(commit.Author.Name), strings.ToLower(opts.Author)) ||
			strings.Contains(strings.ToLower(commit.Author.Email), strings.ToLower(opts.Author))
		if !authorMatch {
			return false
		}
	}

	// Date filters
	if opts.Since != nil && commit.Author.When.Before(*opts.Since) {
		return false
	}
	if opts.Until != nil && commit.Author.When.After(*opts.Until) {
		return false
	}

	// Path filter (would require diff checking - simplified for now)
	// TODO: Implement path filtering with proper diff checking

	return true
}

// GetCommit retrieves a commit by hash (supports abbreviated hashes)
func (r *Repository) GetCommit(hashStr string) (*object.Commit, hash.Hash, error) {
	// Try full hash first
	if len(hashStr) == 40 || len(hashStr) == 64 {
		h, err := hash.ParseHash(hashStr)
		if err == nil {
			obj, err := r.ObjectDB.Get(h)
			if err == nil {
				if commit, ok := obj.(*object.Commit); ok {
					return commit, h, nil
				}
			}
		}
	}

	// Try abbreviated hash
	allHashes, err := r.ObjectDB.List()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list objects: %w", err)
	}

	matches := make([]hash.Hash, 0)
	for _, h := range allHashes {
		if strings.HasPrefix(h.String(), hashStr) {
			// Check if it's a commit
			obj, err := r.ObjectDB.Get(h)
			if err == nil {
				if _, ok := obj.(*object.Commit); ok {
					matches = append(matches, h)
				}
			}
		}
	}

	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("commit not found: %s", hashStr)
	}
	if len(matches) > 1 {
		return nil, nil, fmt.Errorf("ambiguous abbreviated hash: %s matches %d commits", hashStr, len(matches))
	}

	obj, err := r.ObjectDB.Get(matches[0])
	if err != nil {
		return nil, nil, err
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return nil, nil, fmt.Errorf("object is not a commit")
	}

	return commit, matches[0], nil
}

// GetAncestors returns all ancestors of a commit
func (r *Repository) GetAncestors(commitHash hash.Hash) ([]hash.Hash, error) {
	ancestors := make([]hash.Hash, 0)
	visited := make(map[string]bool)
	queue := []hash.Hash{commitHash}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		hashStr := current.String()
		if visited[hashStr] {
			continue
		}
		visited[hashStr] = true

		// Get commit
		obj, err := r.ObjectDB.Get(current)
		if err != nil {
			continue
		}

		commit, ok := obj.(*object.Commit)
		if !ok {
			continue
		}

		// Add parents to ancestors and queue
		for _, parent := range commit.Parents {
			if !visited[parent.String()] {
				ancestors = append(ancestors, parent)
				queue = append(queue, parent)
			}
		}
	}

	return ancestors, nil
}

// IsAncestor checks if commit1 is an ancestor of commit2
func (r *Repository) IsAncestor(ancestor, descendant hash.Hash) (bool, error) {
	if ancestor.Equals(descendant) {
		return true, nil
	}

	ancestors, err := r.GetAncestors(descendant)
	if err != nil {
		return false, err
	}

	ancestorStr := ancestor.String()
	for _, a := range ancestors {
		if a.String() == ancestorStr {
			return true, nil
		}
	}

	return false, nil
}

// GetCommitsBetween returns commits between two commits (from..to)
func (r *Repository) GetCommitsBetween(fromHash, toHash hash.Hash) ([]*LogEntry, error) {
	// Get all commits reachable from 'to'
	toAncestors, err := r.GetAncestors(toHash)
	if err != nil {
		return nil, err
	}

	// Get all commits reachable from 'from'
	fromAncestors, err := r.GetAncestors(fromHash)
	if err != nil {
		return nil, err
	}

	// Build set of commits to exclude (from and its ancestors)
	exclude := make(map[string]bool)
	exclude[fromHash.String()] = true
	for _, h := range fromAncestors {
		exclude[h.String()] = true
	}

	// Collect commits in 'to' that are not in 'from'
	entries := make([]*LogEntry, 0)

	// Add 'to' itself if not excluded
	if !exclude[toHash.String()] {
		obj, err := r.ObjectDB.Get(toHash)
		if err == nil {
			if commit, ok := obj.(*object.Commit); ok {
				entries = append(entries, &LogEntry{
					Commit:  commit,
					Hash:    toHash,
					Parents: commit.Parents,
				})
			}
		}
	}

	// Add ancestors of 'to' that are not in 'from'
	for _, h := range toAncestors {
		if !exclude[h.String()] {
			obj, err := r.ObjectDB.Get(h)
			if err == nil {
				if commit, ok := obj.(*object.Commit); ok {
					entries = append(entries, &LogEntry{
						Commit:  commit,
						Hash:    h,
						Parents: commit.Parents,
					})
				}
			}
		}
	}

	// Sort by commit time (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Commit.Author.When.After(entries[j].Commit.Author.When)
	})

	return entries, nil
}

// FormatLogEntry formats a log entry according to the specified format
func FormatLogEntry(entry *LogEntry, format LogFormat) string {
	switch format {
	case LogFormatOneline:
		return formatOneline(entry)
	case LogFormatShort:
		return formatShort(entry)
	default:
		return formatFull(entry)
	}
}

func formatOneline(entry *LogEntry) string {
	shortHash := entry.Hash.String()[:7]
	message := strings.Split(entry.Commit.Message, "\n")[0]

	refs := ""
	if len(entry.Refs) > 0 {
		refs = " (" + strings.Join(entry.Refs, ", ") + ")"
	}

	return fmt.Sprintf("%s%s %s", shortHash, refs, message)
}

func formatShort(entry *LogEntry) string {
	shortHash := entry.Hash.String()[:7]
	message := strings.Split(entry.Commit.Message, "\n")[0]
	author := entry.Commit.Author.Name

	refs := ""
	if len(entry.Refs) > 0 {
		refs = " (" + strings.Join(entry.Refs, ", ") + ")"
	}

	return fmt.Sprintf("commit %s%s\nAuthor: %s\n\n    %s\n", shortHash, refs, author, message)
}

func formatFull(entry *LogEntry) string {
	var sb strings.Builder

	// Hash and refs
	sb.WriteString("commit ")
	sb.WriteString(entry.Hash.String())
	if len(entry.Refs) > 0 {
		sb.WriteString(" (")
		sb.WriteString(strings.Join(entry.Refs, ", "))
		sb.WriteString(")")
	}
	sb.WriteString("\n")

	// Author
	sb.WriteString("Author: ")
	sb.WriteString(entry.Commit.Author.Name)
	sb.WriteString(" <")
	sb.WriteString(entry.Commit.Author.Email)
	sb.WriteString(">\n")

	// Date
	sb.WriteString("Date:   ")
	sb.WriteString(entry.Commit.Author.When.Format(time.RFC1123Z))
	sb.WriteString("\n\n")

	// Message (indented)
	lines := strings.Split(entry.Commit.Message, "\n")
	for _, line := range lines {
		sb.WriteString("    ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// BlameLine represents a line with its blame information
type BlameLine struct {
	LineNumber int
	Content    string
	Commit     *object.Commit
	CommitHash hash.Hash
}

// BlameOptions contains options for blame operations
type BlameOptions struct {
	// StartLine limits blame to lines starting from this line (1-indexed)
	StartLine int

	// EndLine limits blame to lines up to this line (1-indexed)
	EndLine int
}

// DefaultBlameOptions returns default blame options
func DefaultBlameOptions() BlameOptions {
	return BlameOptions{
		StartLine: 1,
		EndLine:   -1, // unlimited
	}
}

// Blame returns line-by-line history for a file
func (r *Repository) Blame(path string, commitHash hash.Hash, opts BlameOptions) ([]*BlameLine, error) {
	// Get the commit
	commitObj, err := r.ObjectDB.Get(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load commit: %w", err)
	}

	commit, ok := commitObj.(*object.Commit)
	if !ok {
		return nil, fmt.Errorf("object is not a commit")
	}

	// Get the file content at this commit
	content, err := r.getFileAtCommit(path, commit)
	if err != nil {
		return nil, err
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")

	// Apply line range filter
	startIdx := opts.StartLine - 1
	if startIdx < 0 {
		startIdx = 0
	}

	endIdx := len(lines)
	if opts.EndLine > 0 && opts.EndLine <= len(lines) {
		endIdx = opts.EndLine
	}

	// For now, simple implementation: attribute all lines to the current commit
	// A full implementation would trace back through history to find the commit
	// that introduced each line
	blameLines := make([]*BlameLine, 0, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		blameLines = append(blameLines, &BlameLine{
			LineNumber: i + 1,
			Content:    lines[i],
			Commit:     commit,
			CommitHash: commitHash,
		})
	}

	return blameLines, nil
}

// getFileAtCommit retrieves file content at a specific commit
func (r *Repository) getFileAtCommit(path string, commit *object.Commit) ([]byte, error) {
	// Get the tree
	treeObj, err := r.ObjectDB.Get(commit.Tree)
	if err != nil {
		return nil, fmt.Errorf("failed to load tree: %w", err)
	}

	tree, ok := treeObj.(*object.Tree)
	if !ok {
		return nil, fmt.Errorf("object is not a tree")
	}

	// Navigate to the file
	parts := strings.Split(path, "/")
	currentTree := tree

	for i, part := range parts {
		if part == "" {
			continue
		}

		found := false
		for _, entry := range currentTree.Entries() {
			if entry.Name == part {
				if i == len(parts)-1 {
					// Last part - should be a file
					if entry.Mode == object.ModeDir {
						return nil, fmt.Errorf("path is a directory: %s", path)
					}

					// Get blob
					blobObj, err := r.ObjectDB.Get(entry.Hash)
					if err != nil {
						return nil, fmt.Errorf("failed to load blob: %w", err)
					}

					blob, ok := blobObj.(*object.Blob)
					if !ok {
						return nil, fmt.Errorf("object is not a blob")
					}

					return blob.Content(), nil
				} else {
					// Intermediate directory
					if entry.Mode != object.ModeDir {
						return nil, fmt.Errorf("path component is not a directory: %s", part)
					}

					// Load subtree
					subtreeObj, err := r.ObjectDB.Get(entry.Hash)
					if err != nil {
						return nil, fmt.Errorf("failed to load subtree: %w", err)
					}

					subtree, ok := subtreeObj.(*object.Tree)
					if !ok {
						return nil, fmt.Errorf("object is not a tree")
					}

					currentTree = subtree
					found = true
					break
				}
			}
		}

		if !found && i < len(parts)-1 {
			return nil, fmt.Errorf("path not found: %s", path)
		}
	}

	return nil, fmt.Errorf("file not found: %s", path)
}

// FormatBlameLine formats a blame line for display
func FormatBlameLine(line *BlameLine) string {
	shortHash := line.CommitHash.String()[:8]
	author := line.Commit.Author.Name
	if len(author) > 20 {
		author = author[:17] + "..."
	}

	date := line.Commit.Author.When.Format("2006-01-02")

	return fmt.Sprintf("%s (%-20s %s %4d) %s",
		shortHash, author, date, line.LineNumber, line.Content)
}
