package repository

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
	"github.com/nseba/browser-git/git-core/pkg/protocol"
)

// CloneOptions contains options for cloning a repository
type CloneOptions struct {
	// Bare indicates if this should be a bare clone
	Bare bool
	// Depth is the depth for shallow clone (0 for full clone)
	Depth int
	// Branch is the specific branch to clone (empty for default)
	Branch string
	// Remote is the name of the remote (default: "origin")
	Remote string
	// AuthProvider is the authentication provider to use
	AuthProvider interface{}
	// ProgressCallback is called with progress updates
	ProgressCallback func(message string)
}

// DefaultCloneOptions returns default clone options
func DefaultCloneOptions() CloneOptions {
	return CloneOptions{
		Bare:   false,
		Depth:  0,
		Branch: "",
		Remote: "origin",
	}
}

// Clone clones a remote repository to the specified path
func Clone(url string, path string, opts CloneOptions) (*Repository, error) {
	// Create the target directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if directory is empty
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	if len(entries) > 0 {
		return nil, fmt.Errorf("destination path '%s' already exists and is not an empty directory", path)
	}

	// Progress callback helper
	progress := func(msg string) {
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(msg)
		}
	}

	progress("Cloning into '" + path + "'...")

	// Create protocol client
	client := protocol.NewClient()

	// Set authentication if provided
	if opts.AuthProvider != nil {
		// TODO: Set auth provider when we have the interface defined
		// For now, we'll use the default no-auth provider
	}

	// Perform discovery to get remote references
	progress("Fetching remote references...")
	discovery, err := client.Discover(url, protocol.UploadPackService)
	if err != nil {
		return nil, fmt.Errorf("failed to discover remote: %w", err)
	}

	// Determine the branch to clone
	var targetBranch string
	var targetHash string

	if opts.Branch != "" {
		// Clone specific branch
		targetBranch = "refs/heads/" + opts.Branch
		ref, found := discovery.GetReference(targetBranch)
		if !found {
			return nil, fmt.Errorf("remote branch '%s' not found", opts.Branch)
		}
		targetHash = ref.Hash
	} else {
		// Use default branch from HEAD
		defaultBranch, err := discovery.GetDefaultBranch()
		if err != nil {
			return nil, fmt.Errorf("failed to get default branch: %w", err)
		}
		targetBranch = defaultBranch
		ref, found := discovery.GetReference(targetBranch)
		if !found {
			return nil, fmt.Errorf("default branch '%s' not found", targetBranch)
		}
		targetHash = ref.Hash
	}

	progress(fmt.Sprintf("Using branch '%s'...", strings.TrimPrefix(targetBranch, "refs/heads/")))

	// Collect all refs we want to fetch
	wants := []string{targetHash}

	// For a full clone, we want all branches
	if opts.Depth == 0 {
		for _, ref := range discovery.References {
			if strings.HasPrefix(ref.Name, "refs/heads/") || strings.HasPrefix(ref.Name, "refs/tags/") {
				if !stringSliceContains(wants, ref.Hash) {
					wants = append(wants, ref.Hash)
				}
			}
		}
	}

	// We don't have any objects yet (empty repository)
	haves := []string{}

	// Build capabilities
	capabilities := protocol.BuildCapabilities()

	// Fetch packfile from remote
	progress("Receiving objects...")
	uploadPackClient := protocol.NewUploadPackClient(client, url)
	packfileData, err := uploadPackClient.FetchPackfile(wants, haves, capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packfile: %w", err)
	}

	progress(fmt.Sprintf("Received %d bytes", len(packfileData)))

	// Initialize the local repository
	progress("Initializing local repository...")
	initOpts := InitOptions{
		Bare:          opts.Bare,
		InitialBranch: strings.TrimPrefix(targetBranch, "refs/heads/"),
		HashAlgorithm: "sha1", // TODO: Detect from remote
	}

	if err := Init(path, initOpts); err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Open the repository
	repo, err := Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Set up remote configuration
	progress("Setting up remote...")
	if err := setupRemote(repo, opts.Remote, url); err != nil {
		return nil, fmt.Errorf("failed to setup remote: %w", err)
	}

	// Unpack objects from packfile
	progress("Unpacking objects...")
	if err := unpackPackfile(repo, packfileData); err != nil {
		return nil, fmt.Errorf("failed to unpack objects: %w", err)
	}

	// Create remote tracking branches
	progress("Creating remote tracking branches...")
	for _, ref := range discovery.References {
		if strings.HasPrefix(ref.Name, "refs/heads/") {
			branchName := strings.TrimPrefix(ref.Name, "refs/heads/")
			remoteBranch := fmt.Sprintf("refs/remotes/%s/%s", opts.Remote, branchName)

			// Parse hash
			h, err := hash.ParseHash(ref.Hash)
			if err != nil {
				continue // Skip invalid hashes
			}

			// Create remote tracking branch
			if err := repo.UpdateRef(remoteBranch, h); err != nil {
				// Log error but continue
				continue
			}

			// Create local branch if it's the target branch
			if ref.Name == targetBranch {
				localBranch := "refs/heads/" + branchName
				if err := repo.UpdateRef(localBranch, h); err != nil {
					return nil, fmt.Errorf("failed to create local branch: %w", err)
				}
			}
		}
	}

	// Checkout the target branch (unless bare)
	if !opts.Bare {
		progress("Checking out files...")
		branchName := strings.TrimPrefix(targetBranch, "refs/heads/")
		if err := checkoutBranch(repo, branchName); err != nil {
			return nil, fmt.Errorf("failed to checkout branch: %w", err)
		}
	}

	progress("Done!")
	return repo, nil
}

// setupRemote configures the remote in the repository config
func setupRemote(repo *Repository, remoteName string, url string) error {
	// Update config file
	configPath := filepath.Join(repo.GitDir, "config")

	// Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Append remote configuration
	remoteConfig := fmt.Sprintf("\n[remote \"%s\"]\n", remoteName)
	remoteConfig += fmt.Sprintf("\turl = %s\n", url)
	remoteConfig += fmt.Sprintf("\tfetch = +refs/heads/*:refs/remotes/%s/*\n", remoteName)

	// Write updated config
	updatedContent := string(content) + remoteConfig
	if err := os.WriteFile(configPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// unpackPackfile unpacks objects from a packfile into the repository
func unpackPackfile(repo *Repository, packfileData []byte) error {
	// Parse packfile
	reader := protocol.NewPackfileReader(bytes.NewReader(packfileData))
	packfile, err := reader.ReadPackfile()
	if err != nil {
		return fmt.Errorf("failed to read packfile: %w", err)
	}

	// Create object database if not exists
	if repo.ObjectDB == nil {
		storage, err := createObjectStorage(repo)
		if err != nil {
			return fmt.Errorf("failed to create object storage: %w", err)
		}
		repo.ObjectDB = object.NewObjectDatabase(storage, repo.Hasher)
	}

	// First pass: store all non-delta objects
	// We need to do this in multiple passes to resolve deltas
	objectsByHash := make(map[string]*protocol.PackfileObject)
	resolvedObjects := make(map[string][]byte) // hash -> decompressed data

	for i := range packfile.Objects {
		obj := &packfile.Objects[i]

		if !obj.IsDelta {
			// Store regular object
			if err := storePackfileObject(repo, obj, resolvedObjects); err != nil {
				return fmt.Errorf("failed to store object %d: %w", i, err)
			}
		} else {
			// Collect delta objects for later processing
			// We need to compute their hash first
			if len(obj.BaseHash) > 0 {
				baseHashStr := fmt.Sprintf("%x", obj.BaseHash)
				objectsByHash[baseHashStr] = obj
			}
		}
	}

	// Second pass: resolve delta objects
	// We may need multiple iterations if deltas reference other deltas
	maxIterations := 10
	for iteration := 0; iteration < maxIterations; iteration++ {
		resolvedAny := false

		for i := range packfile.Objects {
			obj := &packfile.Objects[i]

			if obj.IsDelta {
				var baseData []byte
				var found bool

				if len(obj.BaseHash) > 0 {
					// REF_DELTA: find base by hash
					baseHashStr := fmt.Sprintf("%x", obj.BaseHash)
					baseData, found = resolvedObjects[baseHashStr]
				} else if obj.Offset > 0 {
					// OFS_DELTA: find base by offset
					// This is more complex, we'd need to track offsets
					// For now, skip OFS_DELTA
					continue
				}

				if found && baseData != nil {
					// Apply delta
					delta, err := protocol.ParseDelta(obj.Data)
					if err != nil {
						continue
					}

					resultData, err := protocol.ApplyDelta(baseData, delta)
					if err != nil {
						continue
					}

					// Store the resolved object
					obj.Data = resultData
					obj.IsDelta = false
					if err := storePackfileObject(repo, obj, resolvedObjects); err != nil {
						continue
					}

					resolvedAny = true
				}
			}
		}

		// If we didn't resolve any deltas in this iteration, we're done or stuck
		if !resolvedAny {
			break
		}
	}

	return nil
}

// storePackfileObject stores a single packfile object in the repository
func storePackfileObject(repo *Repository, packObj *protocol.PackfileObject, resolvedObjects map[string][]byte) error {
	// Convert packfile object type to Git object type
	var obj object.Object

	switch packObj.Type {
	case protocol.ObjCommit:
		commit, err := object.ParseCommit(packObj.Data)
		if err != nil {
			return fmt.Errorf("failed to parse commit: %w", err)
		}
		obj = commit

	case protocol.ObjTree:
		tree, err := object.ParseTree(packObj.Data)
		if err != nil {
			return fmt.Errorf("failed to parse tree: %w", err)
		}
		obj = tree

	case protocol.ObjBlob:
		blob := object.NewBlob(packObj.Data)
		obj = blob

	case protocol.ObjTag:
		tag, err := object.ParseTag(packObj.Data)
		if err != nil {
			return fmt.Errorf("failed to parse tag: %w", err)
		}
		obj = tag

	default:
		return fmt.Errorf("unsupported object type: %d", packObj.Type)
	}

	// Store object in database
	h, err := repo.ObjectDB.Put(obj)
	if err != nil {
		return fmt.Errorf("failed to store object: %w", err)
	}

	// Store resolved object data for delta resolution
	if resolvedObjects != nil {
		resolvedObjects[h.String()] = packObj.Data
	}

	return nil
}

// checkoutBranch checks out a branch to the working directory
func checkoutBranch(repo *Repository, branchName string) error {
	// Get the branch hash
	h, err := repo.GetBranch(branchName)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	// Get commit object
	obj, err := repo.ObjectDB.Get(h)
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	commit, ok := obj.(*object.Commit)
	if !ok {
		return fmt.Errorf("expected commit, got %T", obj)
	}

	// Get tree object
	treeObj, err := repo.ObjectDB.Get(commit.Tree)
	if err != nil {
		return fmt.Errorf("failed to get tree: %w", err)
	}

	tree, ok := treeObj.(*object.Tree)
	if !ok {
		return fmt.Errorf("expected tree, got %T", treeObj)
	}

	// Checkout tree to working directory
	if err := checkoutTree(repo, tree, repo.Path); err != nil {
		return fmt.Errorf("failed to checkout tree: %w", err)
	}

	return nil
}

// checkoutTree recursively checks out a tree to the working directory
func checkoutTree(repo *Repository, tree *object.Tree, basePath string) error {
	for _, entry := range tree.Entries() {
		path := filepath.Join(basePath, entry.Name)

		// Get object
		obj, err := repo.ObjectDB.Get(entry.Hash)
		if err != nil {
			return fmt.Errorf("failed to get object %s: %w", entry.Name, err)
		}

		switch entry.Mode {
		case object.ModeDir:
			// Directory - recurse
			subtree, ok := obj.(*object.Tree)
			if !ok {
				return fmt.Errorf("expected tree for directory %s", entry.Name)
			}

			// Create directory
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}

			// Recurse into subtree
			if err := checkoutTree(repo, subtree, path); err != nil {
				return err
			}

		case object.ModeRegular, object.ModeExecutable:
			// Regular file
			blob, ok := obj.(*object.Blob)
			if !ok {
				return fmt.Errorf("expected blob for file %s", entry.Name)
			}

			// Write file
			perm := os.FileMode(0644)
			if entry.Mode == object.ModeExecutable {
				perm = 0755
			}

			if err := os.WriteFile(path, blob.Content(), perm); err != nil {
				return fmt.Errorf("failed to write file %s: %w", path, err)
			}

		case object.ModeSymlink:
			// Symlink
			blob, ok := obj.(*object.Blob)
			if !ok {
				return fmt.Errorf("expected blob for symlink %s", entry.Name)
			}

			// Create symlink
			if err := os.Symlink(string(blob.Content()), path); err != nil {
				return fmt.Errorf("failed to create symlink %s: %w", path, err)
			}

		default:
			return fmt.Errorf("unsupported file mode %o for %s", entry.Mode, entry.Name)
		}
	}

	return nil
}

// createObjectStorage creates an object storage for the repository
func createObjectStorage(repo *Repository) (object.Storage, error) {
	// For now, use file-based storage
	// TODO: Support different storage backends
	objectsPath := filepath.Join(repo.GitDir, "objects")
	return newFileStorage(objectsPath, repo.Hasher), nil
}

// stringSliceContains checks if a string slice contains a value
func stringSliceContains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
