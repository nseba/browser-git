package repository

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
	"github.com/nseba/browser-git/git-core/pkg/protocol"
)

// PushOptions contains options for pushing to a remote
type PushOptions struct {
	// Remote is the name of the remote to push to (default: "origin")
	Remote string
	// RefSpecs are the refspecs to push (e.g., "refs/heads/main:refs/heads/main")
	RefSpecs []string
	// Force allows non-fast-forward updates
	Force bool
	// AuthProvider is the authentication provider to use
	AuthProvider interface{}
	// ProgressCallback is called with progress updates
	ProgressCallback func(message string)
}

// DefaultPushOptions returns default push options
func DefaultPushOptions() PushOptions {
	return PushOptions{
		Remote:   "origin",
		RefSpecs: []string{},
		Force:    false,
	}
}

// Push pushes local commits to a remote repository
func (r *Repository) Push(opts PushOptions) error {
	// Progress callback helper
	progress := func(msg string) {
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(msg)
		}
	}

	progress("Pushing to remote...")

	// Get remote URL from config
	remoteURL, err := r.GetRemoteURL(opts.Remote)
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Create protocol client
	client := protocol.NewClient()

	// Set authentication if provided
	if opts.AuthProvider != nil {
		// TODO: Set auth provider when we have the interface defined
	}

	// Perform discovery to get remote references
	progress("Fetching remote references...")
	discovery, err := client.Discover(remoteURL, protocol.ReceivePackService)
	if err != nil {
		return fmt.Errorf("failed to discover remote: %w", err)
	}

	// Build remote references map for easy lookup
	remoteRefs := make(map[string]string) // refname -> hash
	for _, ref := range discovery.References {
		remoteRefs[ref.Name] = ref.Hash
	}

	// Determine which refs to push
	var refsToPush []refToPush
	if len(opts.RefSpecs) == 0 {
		// If no refspecs provided, push current branch
		currentBranch, err := r.CurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}

		refsToPush, err = r.buildRefsToPushForBranch(currentBranch, remoteRefs, opts.Force)
		if err != nil {
			return err
		}
	} else {
		// Parse and process refspecs
		for _, refspec := range opts.RefSpecs {
			refs, err := r.parseAndBuildRefsToPush(refspec, remoteRefs, opts.Force)
			if err != nil {
				return err
			}
			refsToPush = append(refsToPush, refs...)
		}
	}

	if len(refsToPush) == 0 {
		progress("Everything up-to-date")
		return nil
	}

	// Find commits to send
	progress("Determining commits to send...")
	commitsToSend, err := r.findCommitsToSend(refsToPush, remoteRefs)
	if err != nil {
		return fmt.Errorf("failed to find commits to send: %w", err)
	}

	if len(commitsToSend) == 0 && !hasNewOrDeletedRefs(refsToPush) {
		progress("Everything up-to-date")
		return nil
	}

	progress(fmt.Sprintf("Found %d commits to send", len(commitsToSend)))

	// Collect all objects to send
	progress("Collecting objects...")
	objectsToSend, err := r.collectObjectsForCommits(commitsToSend)
	if err != nil {
		return fmt.Errorf("failed to collect objects: %w", err)
	}

	progress(fmt.Sprintf("Collected %d objects", len(objectsToSend)))

	// Create packfile
	progress("Creating packfile...")
	packfileData, err := r.createPackfileForPush(objectsToSend)
	if err != nil {
		return fmt.Errorf("failed to create packfile: %w", err)
	}

	progress(fmt.Sprintf("Created packfile with %d bytes", len(packfileData)))

	// Build ref updates for push request
	refUpdates := make([]protocol.RefUpdate, 0, len(refsToPush))
	for _, ref := range refsToPush {
		refUpdates = append(refUpdates, protocol.RefUpdate{
			Name:    ref.remoteName,
			OldHash: ref.oldHash,
			NewHash: ref.newHash,
		})
	}

	// Create push request
	pushReq := &protocol.PushRequest{
		Updates:      refUpdates,
		Capabilities: protocol.BuildPushCapabilities(),
		Packfile:     packfileData,
		Force:        opts.Force,
		ReportStatus: true,
	}

	// Send push request
	progress("Sending packfile to remote...")
	receivePackClient := protocol.NewReceivePackClient(client, remoteURL)
	pushResp, err := receivePackClient.Push(pushReq)
	if err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	// Check response
	if pushResp.UnpackStatus != "ok" {
		return fmt.Errorf("unpack failed: %s", pushResp.UnpackStatus)
	}

	// Update remote tracking branches
	progress("Updating remote tracking branches...")
	for _, ref := range refsToPush {
		if ref.newHash != "0000000000000000000000000000000000000000" {
			// Parse hash
			h, err := hash.ParseHash(ref.newHash)
			if err != nil {
				continue
			}

			// Update remote tracking branch
			remoteBranch := ref.remoteName
			if strings.HasPrefix(remoteBranch, "refs/heads/") {
				branchName := strings.TrimPrefix(remoteBranch, "refs/heads/")
				trackingBranch := fmt.Sprintf("refs/remotes/%s/%s", opts.Remote, branchName)
				if err := r.UpdateRef(trackingBranch, h); err != nil {
					// Log error but continue
					continue
				}
			}
		}
	}

	progress("Push successful!")
	return nil
}

// refToPush represents a reference to push
type refToPush struct {
	localName  string // Local reference name
	remoteName string // Remote reference name
	oldHash    string // Current hash on remote (40 zeros for new)
	newHash    string // New hash to push (40 zeros for delete)
}

// buildRefsToPushForBranch builds refs to push for the current branch
func (r *Repository) buildRefsToPushForBranch(branchName string, remoteRefs map[string]string, force bool) ([]refToPush, error) {
	// Get local branch hash
	localRef := "refs/heads/" + branchName
	localHash, err := r.GetBranch(branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to get local branch: %w", err)
	}

	// Check if remote has this branch
	remoteRef := "refs/heads/" + branchName
	remoteHash, hasRemote := remoteRefs[remoteRef]

	if !hasRemote {
		// New branch on remote
		return []refToPush{{
			localName:  localRef,
			remoteName: remoteRef,
			oldHash:    "0000000000000000000000000000000000000000",
			newHash:    localHash.String(),
		}}, nil
	}

	// Check if it's a fast-forward
	if remoteHash == localHash.String() {
		// Already up-to-date
		return nil, nil
	}

	if !force {
		// Check if local is ahead of remote (fast-forward)
		isAncestor, err := r.isAncestor(remoteHash, localHash.String())
		if err != nil {
			return nil, fmt.Errorf("failed to check ancestry: %w", err)
		}

		if !isAncestor {
			return nil, fmt.Errorf("non-fast-forward update rejected (use --force to override)")
		}
	}

	return []refToPush{{
		localName:  localRef,
		remoteName: remoteRef,
		oldHash:    remoteHash,
		newHash:    localHash.String(),
	}}, nil
}

// parseAndBuildRefsToPush parses a refspec and builds refs to push
func (r *Repository) parseAndBuildRefsToPush(refspec string, remoteRefs map[string]string, force bool) ([]refToPush, error) {
	// Simple refspec parsing: "local:remote" or "branch" or ":remote" (delete)
	parts := strings.Split(refspec, ":")

	var localRef, remoteRef string

	if len(parts) == 1 {
		// Single ref - push to same name
		localRef = parts[0]
		remoteRef = parts[0]

		// Add refs/heads/ if not present
		if !strings.HasPrefix(localRef, "refs/") {
			localRef = "refs/heads/" + localRef
			remoteRef = "refs/heads/" + remoteRef
		}
	} else if len(parts) == 2 {
		localRef = parts[0]
		remoteRef = parts[1]

		// Handle delete (":remote")
		if localRef == "" {
			// Delete remote ref
			remoteHash, hasRemote := remoteRefs[remoteRef]
			if !hasRemote {
				return nil, fmt.Errorf("remote ref %s does not exist", remoteRef)
			}

			return []refToPush{{
				localName:  "",
				remoteName: remoteRef,
				oldHash:    remoteHash,
				newHash:    "0000000000000000000000000000000000000000",
			}}, nil
		}

		// Add refs/heads/ if not present
		if !strings.HasPrefix(localRef, "refs/") {
			localRef = "refs/heads/" + localRef
		}
		if !strings.HasPrefix(remoteRef, "refs/") {
			remoteRef = "refs/heads/" + remoteRef
		}
	} else {
		return nil, fmt.Errorf("invalid refspec: %s", refspec)
	}

	// Get local hash
	branchName := strings.TrimPrefix(localRef, "refs/heads/")
	localHash, err := r.GetBranch(branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to get local branch %s: %w", branchName, err)
	}

	// Check remote
	remoteHash, hasRemote := remoteRefs[remoteRef]

	if !hasRemote {
		// New branch
		return []refToPush{{
			localName:  localRef,
			remoteName: remoteRef,
			oldHash:    "0000000000000000000000000000000000000000",
			newHash:    localHash.String(),
		}}, nil
	}

	if remoteHash == localHash.String() {
		// Up-to-date
		return nil, nil
	}

	// Check fast-forward
	if !force {
		isAncestor, err := r.isAncestor(remoteHash, localHash.String())
		if err != nil {
			return nil, fmt.Errorf("failed to check ancestry: %w", err)
		}

		if !isAncestor {
			return nil, fmt.Errorf("non-fast-forward update rejected for %s (use --force to override)", remoteRef)
		}
	}

	return []refToPush{{
		localName:  localRef,
		remoteName: remoteRef,
		oldHash:    remoteHash,
		newHash:    localHash.String(),
	}}, nil
}

// isAncestor checks if oldHash is an ancestor of newHash
func (r *Repository) isAncestor(oldHashStr string, newHashStr string) (bool, error) {
	// Parse hashes
	oldHash, err := hash.ParseHash(oldHashStr)
	if err != nil {
		return false, err
	}

	newHash, err := hash.ParseHash(newHashStr)
	if err != nil {
		return false, err
	}

	// Walk from newHash back to find oldHash
	visited := make(map[string]bool)
	queue := []hash.Hash{newHash}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.String() == oldHash.String() {
			return true, nil
		}

		if visited[current.String()] {
			continue
		}
		visited[current.String()] = true

		// Get commit
		obj, err := r.ObjectDB.Get(current)
		if err != nil {
			continue
		}

		commit, ok := obj.(*object.Commit)
		if !ok {
			continue
		}

		// Add parents to queue
		for _, parent := range commit.Parents {
			queue = append(queue, parent)
		}
	}

	return false, nil
}

// findCommitsToSend finds all commits that need to be sent
func (r *Repository) findCommitsToSend(refs []refToPush, remoteRefs map[string]string) ([]hash.Hash, error) {
	commits := []hash.Hash{}
	seen := make(map[string]bool)

	// Collect all remote commit hashes (what they have)
	remoteHaves := make(map[string]bool)
	for _, remoteHash := range remoteRefs {
		if remoteHash != "0000000000000000000000000000000000000000" {
			remoteHaves[remoteHash] = true
		}
	}

	// For each ref, walk back from newHash until we hit remoteHash or a commit they have
	for _, ref := range refs {
		if ref.newHash == "0000000000000000000000000000000000000000" {
			// Delete - no commits to send
			continue
		}

		newHash, err := hash.ParseHash(ref.newHash)
		if err != nil {
			continue
		}

		// Walk commits
		if err := r.walkCommitsToSend(newHash, remoteHaves, seen, &commits); err != nil {
			return nil, err
		}
	}

	return commits, nil
}

// walkCommitsToSend recursively walks commits to find ones to send
func (r *Repository) walkCommitsToSend(commitHash hash.Hash, remoteHaves map[string]bool, seen map[string]bool, commits *[]hash.Hash) error {
	hashStr := commitHash.String()

	// Skip if already processed
	if seen[hashStr] {
		return nil
	}
	seen[hashStr] = true

	// Skip if remote has this commit
	if remoteHaves[hashStr] {
		return nil
	}

	// Get commit object
	obj, err := r.ObjectDB.Get(commitHash)
	if err != nil {
		return err
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return fmt.Errorf("expected commit, got %T", obj)
	}

	// Add to commits to send
	*commits = append(*commits, commitHash)

	// Recurse into parents
	for _, parent := range commit.Parents {
		if err := r.walkCommitsToSend(parent, remoteHaves, seen, commits); err != nil {
			return err
		}
	}

	return nil
}

// collectObjectsForCommits collects all objects (commits, trees, blobs) for the given commits
func (r *Repository) collectObjectsForCommits(commits []hash.Hash) ([]object.Object, error) {
	objects := []object.Object{}
	seen := make(map[string]bool)

	for _, commitHash := range commits {
		if err := r.collectObjectsRecursive(commitHash, seen, &objects); err != nil {
			return nil, err
		}
	}

	return objects, nil
}

// collectObjectsRecursive recursively collects all objects
func (r *Repository) collectObjectsRecursive(h hash.Hash, seen map[string]bool, objects *[]object.Object) error {
	hashStr := h.String()
	if seen[hashStr] {
		return nil
	}
	seen[hashStr] = true

	// Get object
	obj, err := r.ObjectDB.Get(h)
	if err != nil {
		return err
	}

	// Add to objects
	*objects = append(*objects, obj)

	// Recurse based on type
	switch o := obj.(type) {
	case *object.Commit:
		// Recurse into tree
		if err := r.collectObjectsRecursive(o.Tree, seen, objects); err != nil {
			return err
		}
		// Note: We don't recurse into parents since we already walked commits

	case *object.Tree:
		// Recurse into all entries
		for _, entry := range o.Entries() {
			if err := r.collectObjectsRecursive(entry.Hash, seen, objects); err != nil {
				return err
			}
		}

	case *object.Blob:
		// Leaf node - nothing to recurse

	case *object.Tag:
		// Recurse into target
		if err := r.collectObjectsRecursive(o.Target, seen, objects); err != nil {
			return err
		}
	}

	return nil
}

// createPackfileForPush creates a packfile with the given objects
func (r *Repository) createPackfileForPush(objects []object.Object) ([]byte, error) {
	// Convert objects to packfile objects
	packfileObjects := make([]protocol.PackfileObject, 0, len(objects))

	for _, obj := range objects {
		// Serialize object
		var buf bytes.Buffer
		if err := obj.Serialize(&buf); err != nil {
			return nil, fmt.Errorf("failed to serialize object: %w", err)
		}
		data := buf.Bytes()

		// Determine type
		var objType uint8
		switch obj.(type) {
		case *object.Commit:
			objType = protocol.ObjCommit
		case *object.Tree:
			objType = protocol.ObjTree
		case *object.Blob:
			objType = protocol.ObjBlob
		case *object.Tag:
			objType = protocol.ObjTag
		default:
			return nil, fmt.Errorf("unknown object type: %T", obj)
		}

		packfileObjects = append(packfileObjects, protocol.PackfileObject{
			Type:    objType,
			Size:    uint64(len(data)),
			Data:    data,
			IsDelta: false,
		})
	}

	// Write packfile
	var buf bytes.Buffer
	writer := protocol.NewPackfileWriter(&buf)
	if err := writer.WritePackfile(packfileObjects); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// hasNewOrDeletedRefs checks if any refs are new or deleted
func hasNewOrDeletedRefs(refs []refToPush) bool {
	for _, ref := range refs {
		if ref.oldHash == "0000000000000000000000000000000000000000" ||
			ref.newHash == "0000000000000000000000000000000000000000" {
			return true
		}
	}
	return false
}

// GetRemoteURL gets the URL for a remote from config
func (r *Repository) GetRemoteURL(remoteName string) (string, error) {
	// Look for remote.<name>.url in config
	section := fmt.Sprintf("remote.%s", remoteName)
	url, ok := r.Config.Get(section, "url")
	if !ok {
		return "", fmt.Errorf("remote '%s' not found", remoteName)
	}

	return url, nil
}
