package repository

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/auth"
	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
	"github.com/nseba/browser-git/git-core/pkg/protocol"
)

// FetchOptions contains options for fetching from a remote
type FetchOptions struct {
	// Remote is the name of the remote to fetch from (default: "origin")
	Remote string
	// RefSpecs are the refspecs to fetch (default: fetch all branches)
	RefSpecs []string
	// Prune removes remote tracking branches that no longer exist on remote
	Prune bool
	// Force allows non-fast-forward updates
	Force bool
	// Depth for shallow fetch (0 for full fetch)
	Depth int
	// AuthProvider is the authentication provider to use
	AuthProvider auth.AuthProvider
	// ProgressCallback is called with progress updates
	ProgressCallback func(message string)
}

// DefaultFetchOptions returns default fetch options
func DefaultFetchOptions() FetchOptions {
	return FetchOptions{
		Remote:   "origin",
		RefSpecs: []string{},
		Prune:    false,
		Force:    false,
		Depth:    0,
	}
}

// FetchResult contains information about the fetch operation
type FetchResult struct {
	// UpdatedRefs contains the refs that were updated
	UpdatedRefs map[string]RefUpdate
	// PrunedRefs contains the refs that were pruned
	PrunedRefs []string
	// ObjectCount is the number of objects fetched
	ObjectCount int
}

// RefUpdate describes an update to a reference
type RefUpdate struct {
	// RefName is the name of the reference
	RefName string
	// OldHash is the previous hash (empty if new)
	OldHash string
	// NewHash is the new hash (empty if deleted)
	NewHash string
	// Forced indicates if this was a forced update
	Forced bool
}

// Fetch fetches objects and refs from a remote repository
func (r *Repository) Fetch(opts FetchOptions) (*FetchResult, error) {
	// Get remote URL from config
	remoteURL, err := r.Config.GetRemoteURL(opts.Remote)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Progress callback helper
	progress := func(msg string) {
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(msg)
		}
	}

	progress(fmt.Sprintf("Fetching from %s...", opts.Remote))

	// Create protocol client
	client := protocol.NewClient()

	// Set authentication if provided
	if opts.AuthProvider != nil {
		client.SetAuthProvider(opts.AuthProvider)
	}

	// Perform discovery to get remote references
	progress("Discovering remote references...")
	discovery, err := client.Discover(remoteURL, protocol.UploadPackService)
	if err != nil {
		return nil, fmt.Errorf("failed to discover remote: %w", err)
	}

	// Get fetch refspecs from config if not provided
	refspecs := opts.RefSpecs
	if len(refspecs) == 0 {
		refspecs, err = r.Config.GetFetchRefSpecs(opts.Remote)
		if err != nil {
			// Use default refspec
			refspecs = []string{fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", opts.Remote)}
		}
	}

	// Calculate which refs to update
	refsToUpdate, err := r.calculateRefUpdates(discovery, refspecs, opts.Remote, opts.Force)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ref updates: %w", err)
	}

	// If no refs to update, we're up to date
	if len(refsToUpdate) == 0 {
		progress("Already up to date")
		return &FetchResult{
			UpdatedRefs: make(map[string]RefUpdate),
			PrunedRefs:  []string{},
			ObjectCount: 0,
		}, nil
	}

	// Collect objects we want
	wants := []string{}
	for _, update := range refsToUpdate {
		if update.NewHash != "" && update.NewHash != update.OldHash {
			wants = append(wants, update.NewHash)
		}
	}

	// Collect objects we already have
	haves, err := r.getAllLocalCommits()
	if err != nil {
		return nil, fmt.Errorf("failed to get local commits: %w", err)
	}

	// If we want objects we already have, filter them out
	filteredWants := []string{}
	for _, want := range wants {
		if !stringSliceContains(haves, want) {
			filteredWants = append(filteredWants, want)
		}
	}

	// If no new objects to fetch, just update refs
	var objectCount int
	if len(filteredWants) > 0 {
		// Build capabilities
		capabilities := protocol.BuildCapabilities()

		// Fetch packfile from remote
		progress("Receiving objects...")
		uploadPackClient := protocol.NewUploadPackClient(client, remoteURL)
		packfileData, err := uploadPackClient.FetchPackfile(filteredWants, haves, capabilities)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch packfile: %w", err)
		}

		progress(fmt.Sprintf("Received %d bytes", len(packfileData)))

		// Unpack objects from packfile
		progress("Unpacking objects...")
		count, err := r.unpackPackfile(packfileData)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack objects: %w", err)
		}
		objectCount = count
		progress(fmt.Sprintf("Unpacked %d objects", objectCount))
	}

	// Update remote tracking branches
	progress("Updating remote tracking branches...")
	updatedRefs := make(map[string]RefUpdate)
	for _, update := range refsToUpdate {
		if err := r.updateRef(update); err != nil {
			return nil, fmt.Errorf("failed to update ref %s: %w", update.RefName, err)
		}
		updatedRefs[update.RefName] = update
	}

	// Prune remote tracking branches if requested
	prunedRefs := []string{}
	if opts.Prune {
		pruned, err := r.pruneRemoteRefs(discovery, opts.Remote)
		if err != nil {
			return nil, fmt.Errorf("failed to prune refs: %w", err)
		}
		prunedRefs = pruned
		if len(prunedRefs) > 0 {
			progress(fmt.Sprintf("Pruned %d refs", len(prunedRefs)))
		}
	}

	progress("Done!")
	return &FetchResult{
		UpdatedRefs: updatedRefs,
		PrunedRefs:  prunedRefs,
		ObjectCount: objectCount,
	}, nil
}

// calculateRefUpdates determines which refs need to be updated based on refspecs
func (r *Repository) calculateRefUpdates(discovery *protocol.DiscoveryResponse, refspecs []string, remote string, force bool) ([]RefUpdate, error) {
	updates := []RefUpdate{}

	for _, refspec := range refspecs {
		// Parse refspec
		src, dst, isForce := parseRefSpec(refspec)

		// Allow force updates if refspec has + or force option is set
		allowForce := isForce || force

		// Match source pattern against remote refs
		for _, ref := range discovery.References {
			if matchesPattern(ref.Name, src) {
				// Calculate destination ref name
				dstRef := calculateDestRef(ref.Name, src, dst, remote)

				// Get current value of destination ref
				oldHash := ""
				if currentHash, err := r.GetRef(dstRef); err == nil {
					oldHash = currentHash.String()
				}

				// Check if this is a fast-forward or forced update
				if oldHash != "" && oldHash != ref.Hash {
					if !allowForce {
						// Check if it's a fast-forward
						isFF, err := r.isAncestor(oldHash, ref.Hash)
						if err != nil || !isFF {
							// Skip non-fast-forward updates unless forced
							continue
						}
					}
				}

				// Add update
				updates = append(updates, RefUpdate{
					RefName: dstRef,
					OldHash: oldHash,
					NewHash: ref.Hash,
					Forced:  allowForce && oldHash != "" && oldHash != ref.Hash,
				})
			}
		}
	}

	return updates, nil
}

// parseRefSpec parses a refspec string
// Returns: source, destination, force
func parseRefSpec(refspec string) (string, string, bool) {
	force := false
	if strings.HasPrefix(refspec, "+") {
		force = true
		refspec = refspec[1:]
	}

	parts := strings.SplitN(refspec, ":", 2)
	if len(parts) == 1 {
		return parts[0], parts[0], force
	}
	return parts[0], parts[1], force
}

// matchesPattern checks if a ref name matches a pattern
func matchesPattern(refName, pattern string) bool {
	// Handle wildcard patterns
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(refName, prefix)
	}
	return refName == pattern
}

// calculateDestRef calculates the destination ref name
func calculateDestRef(refName, srcPattern, dstPattern, remote string) string {
	// If pattern has wildcard, substitute
	if strings.HasSuffix(srcPattern, "*") && strings.HasSuffix(dstPattern, "*") {
		prefix := strings.TrimSuffix(srcPattern, "*")
		suffix := strings.TrimPrefix(refName, prefix)
		dstPrefix := strings.TrimSuffix(dstPattern, "*")
		return dstPrefix + suffix
	}
	return dstPattern
}

// getAllLocalCommits returns hashes of all local commits
func (r *Repository) getAllLocalCommits() ([]string, error) {
	commits := []string{}

	// Get all local branches
	refs, err := r.ListRefs("refs/heads/")
	if err != nil {
		return commits, nil // No local refs yet
	}

	// Add all remote tracking branches
	remoteRefs, err := r.ListRefs("refs/remotes/")
	if err == nil {
		refs = append(refs, remoteRefs...)
	}

	// Collect unique commit hashes
	seen := make(map[string]bool)
	for _, ref := range refs {
		h, err := r.GetRef(ref)
		if err != nil {
			continue
		}
		hashStr := h.String()
		if !seen[hashStr] {
			commits = append(commits, hashStr)
			seen[hashStr] = true
		}
	}

	return commits, nil
}

// unpackPackfile unpacks objects from a packfile and returns the count
func (r *Repository) unpackPackfile(packfileData []byte) (int, error) {
	// Ensure object database is initialized
	if r.ObjectDB == nil {
		storage, err := createObjectStorage(r)
		if err != nil {
			return 0, fmt.Errorf("failed to create object storage: %w", err)
		}
		r.ObjectDB = object.NewObjectDatabase(storage, r.Hasher)
	}

	// Use the same unpackPackfile logic from clone
	if err := unpackPackfile(r, packfileData); err != nil {
		return 0, err
	}

	// Parse packfile to get object count
	reader := protocol.NewPackfileReader(bytes.NewReader(packfileData))
	packfile, err := reader.ReadPackfile()
	if err != nil {
		return 0, fmt.Errorf("failed to read packfile: %w", err)
	}

	return len(packfile.Objects), nil
}

// updateRef updates a reference
func (r *Repository) updateRef(update RefUpdate) error {
	// Parse new hash
	h, err := hash.ParseHash(update.NewHash)
	if err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}

	// Update the ref
	if err := r.UpdateRef(update.RefName, h); err != nil {
		return err
	}

	return nil
}

// pruneRemoteRefs removes remote tracking branches that no longer exist on remote
func (r *Repository) pruneRemoteRefs(discovery *protocol.DiscoveryResponse, remote string) ([]string, error) {
	pruned := []string{}

	// Get all remote tracking branches
	remotePrefix := fmt.Sprintf("refs/remotes/%s/", remote)
	localRemoteRefs, err := r.ListRefs(remotePrefix)
	if err != nil {
		return pruned, nil // No remote refs
	}

	// Build set of remote branch names
	remoteBranches := make(map[string]bool)
	for _, ref := range discovery.References {
		if strings.HasPrefix(ref.Name, "refs/heads/") {
			branchName := strings.TrimPrefix(ref.Name, "refs/heads/")
			remoteBranches[branchName] = true
		}
	}

	// Check each local remote ref
	for _, localRef := range localRemoteRefs {
		branchName := strings.TrimPrefix(localRef, remotePrefix)
		if !remoteBranches[branchName] {
			// This branch no longer exists on remote, prune it
			if err := r.DeleteRef(localRef); err != nil {
				return pruned, fmt.Errorf("failed to delete ref %s: %w", localRef, err)
			}
			pruned = append(pruned, localRef)
		}
	}

	return pruned, nil
}

// isAncestor checks if commit 'ancestor' is an ancestor of commit 'descendant'
func (r *Repository) isAncestor(ancestor, descendant string) (bool, error) {
	// Parse hashes
	ancestorHash, err := hash.ParseHash(ancestor)
	if err != nil {
		return false, err
	}

	descendantHash, err := hash.ParseHash(descendant)
	if err != nil {
		return false, err
	}

	// If they're the same, ancestor check is true
	if ancestorHash.Equals(descendantHash) {
		return true, nil
	}

	// Walk commit history from descendant to see if we reach ancestor
	visited := make(map[string]bool)
	toVisit := []hash.Hash{descendantHash}

	for len(toVisit) > 0 {
		current := toVisit[0]
		toVisit = toVisit[1:]

		if visited[current.String()] {
			continue
		}
		visited[current.String()] = true

		// Check if this is the ancestor
		if current.Equals(ancestorHash) {
			return true, nil
		}

		// Get commit object
		obj, err := r.ObjectDB.Get(current)
		if err != nil {
			continue // Skip missing commits
		}

		commit, ok := obj.(*object.Commit)
		if !ok {
			continue
		}

		// Add parents to visit
		for _, parent := range commit.Parents {
			toVisit = append(toVisit, parent)
		}
	}

	return false, nil
}

// PullOptions contains options for pulling from a remote
type PullOptions struct {
	// Remote is the name of the remote to pull from (default: "origin")
	Remote string
	// Branch is the specific branch to pull (empty for current branch's upstream)
	Branch string
	// Rebase indicates whether to rebase instead of merge
	Rebase bool
	// FastForwardOnly only allows fast-forward merges
	FastForwardOnly bool
	// Force allows non-fast-forward updates during fetch
	Force bool
	// AuthProvider is the authentication provider to use
	AuthProvider auth.AuthProvider
	// ProgressCallback is called with progress updates
	ProgressCallback func(message string)
}

// DefaultPullOptions returns default pull options
func DefaultPullOptions() PullOptions {
	return PullOptions{
		Remote:          "origin",
		Branch:          "",
		Rebase:          false,
		FastForwardOnly: false,
		Force:           false,
	}
}

// PullResult contains information about the pull operation
type PullResult struct {
	// FetchResult contains the fetch operation result
	FetchResult *FetchResult
	// MergeResult contains the merge operation result (nil if already up to date)
	MergeResult interface{}
	// FastForward indicates if this was a fast-forward update
	FastForward bool
	// AlreadyUpToDate indicates if there was nothing to pull
	AlreadyUpToDate bool
}

// Pull fetches from remote and integrates changes into current branch
func (r *Repository) Pull(opts PullOptions) (*PullResult, error) {
	// Progress callback helper
	progress := func(msg string) {
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(msg)
		}
	}

	// Get current branch
	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Determine the branch to pull
	pullBranch := opts.Branch
	if pullBranch == "" {
		// Use current branch's upstream
		upstream, err := r.Config.GetBranchUpstream(currentBranch)
		if err != nil {
			// Default to same branch name
			pullBranch = currentBranch
		} else {
			// Extract branch name from upstream (e.g., "refs/remotes/origin/main" -> "main")
			pullBranch = strings.TrimPrefix(upstream, fmt.Sprintf("refs/remotes/%s/", opts.Remote))
		}
	}

	// Get current HEAD commit
	currentCommit, err := r.ResolveHEAD()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve HEAD: %w", err)
	}

	// Fetch from remote
	progress(fmt.Sprintf("Pulling %s from %s...", pullBranch, opts.Remote))
	fetchOpts := FetchOptions{
		Remote:           opts.Remote,
		RefSpecs:         []string{fmt.Sprintf("refs/heads/%s:refs/remotes/%s/%s", pullBranch, opts.Remote, pullBranch)},
		Force:            opts.Force,
		AuthProvider:     opts.AuthProvider,
		ProgressCallback: opts.ProgressCallback,
	}

	fetchResult, err := r.Fetch(fetchOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	// Get the fetched remote branch ref
	remoteBranchRef := fmt.Sprintf("refs/remotes/%s/%s", opts.Remote, pullBranch)
	remoteBranchHash, err := r.GetRef(remoteBranchRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote branch: %w", err)
	}

	// Check if already up to date
	if currentCommit.Equals(remoteBranchHash) {
		progress("Already up to date")
		return &PullResult{
			FetchResult:     fetchResult,
			MergeResult:     nil,
			FastForward:     false,
			AlreadyUpToDate: true,
		}, nil
	}

	// Check if we can fast-forward
	canFF, err := r.isAncestor(currentCommit.String(), remoteBranchHash.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check fast-forward: %w", err)
	}

	if canFF {
		// Perform fast-forward
		progress("Fast-forwarding...")
		if err := r.fastForward(remoteBranchHash); err != nil {
			return nil, fmt.Errorf("failed to fast-forward: %w", err)
		}

		return &PullResult{
			FetchResult:     fetchResult,
			MergeResult:     nil,
			FastForward:     true,
			AlreadyUpToDate: false,
		}, nil
	}

	// If fast-forward only mode, fail
	if opts.FastForwardOnly {
		return nil, fmt.Errorf("cannot fast-forward; refusing to merge (use --no-ff to override)")
	}

	// Perform merge or rebase
	if opts.Rebase {
		// TODO: Implement rebase
		return nil, fmt.Errorf("rebase not yet implemented")
	} else {
		// Perform merge
		progress("Merging changes...")
		mergeOpts := DefaultMergeOptions()
		mergeOpts.AllowFastForward = false // We already checked for fast-forward
		mergeOpts.CommitMessage = fmt.Sprintf("Merge branch '%s' of %s into %s", pullBranch, opts.Remote, currentBranch)

		// Instead of merging by branch name, we need to merge the remote branch
		// First, create a temporary local branch pointing to the remote branch
		tempBranch := fmt.Sprintf("PULL_HEAD_%s", remoteBranchHash.String()[:8])
		if err := r.UpdateRef(fmt.Sprintf("refs/heads/%s", tempBranch), remoteBranchHash); err != nil {
			return nil, fmt.Errorf("failed to create temp branch: %w", err)
		}
		defer r.DeleteRef(fmt.Sprintf("refs/heads/%s", tempBranch))

		mergeResult, err := r.Merge(tempBranch, mergeOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to merge: %w", err)
		}

		return &PullResult{
			FetchResult:     fetchResult,
			MergeResult:     mergeResult,
			FastForward:     false,
			AlreadyUpToDate: false,
		}, nil
	}
}

// fastForward performs a fast-forward update to the specified hash
func (r *Repository) fastForward(newHash hash.Hash) error {
	// Get current branch
	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Update the branch ref
	branchRef := fmt.Sprintf("refs/heads/%s", currentBranch)
	if err := r.UpdateRef(branchRef, newHash); err != nil {
		return fmt.Errorf("failed to update branch ref: %w", err)
	}

	// Update working directory
	obj, err := r.ObjectDB.Get(newHash)
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return fmt.Errorf("expected commit, got %T", obj)
	}

	// Get tree object
	treeObj, err := r.ObjectDB.Get(commit.Tree)
	if err != nil {
		return fmt.Errorf("failed to get tree: %w", err)
	}

	tree, ok := treeObj.(*object.Tree)
	if !ok {
		return fmt.Errorf("expected tree, got %T", treeObj)
	}

	// Checkout tree to working directory
	if err := checkoutTree(r, tree, r.Path); err != nil {
		return fmt.Errorf("failed to checkout tree: %w", err)
	}

	return nil
}
