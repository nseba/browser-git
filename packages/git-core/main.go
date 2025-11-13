// +build js,wasm

package main

import (
	"path/filepath"
	"syscall/js"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/index"
	"github.com/nseba/browser-git/git-core/pkg/object"
	"github.com/nseba/browser-git/git-core/pkg/repository"
)

// Version information
const (
	Version = "0.1.0"
)

func main() {
	// Wait forever - WASM modules need to keep running
	c := make(chan struct{}, 0)

	// Export functions to JavaScript
	js.Global().Set("gitCore", js.ValueOf(map[string]interface{}{
		"version": js.FuncOf(getVersion),
		"hash": js.ValueOf(map[string]interface{}{
			"sha1":     js.FuncOf(hashSHA1),
			"sha256":   js.FuncOf(hashSHA256),
			"hashBlob": js.FuncOf(hashBlob),
		}),
		"object": js.ValueOf(map[string]interface{}{
			"createBlob":   js.FuncOf(createBlob),
			"createTree":   js.FuncOf(createTree),
			"createCommit": js.FuncOf(createCommit),
			"createTag":    js.FuncOf(createTag),
			"parseObject":  js.FuncOf(parseObject),
			"compress":     js.FuncOf(compressObject),
			"decompress":   js.FuncOf(decompressObject),
		}),
		"repository": js.ValueOf(map[string]interface{}{
			"init":          js.FuncOf(initRepository),
			"open":          js.FuncOf(openRepository),
			"isRepository":  js.FuncOf(isRepository),
			"find":          js.FuncOf(findRepository),
			"add":           js.FuncOf(addFiles),
			"commit":        js.FuncOf(createCommitFromIndex),
			"status":        js.FuncOf(getStatus),
			"listBranches":  js.FuncOf(listBranches),
			"createBranch":  js.FuncOf(createBranch),
			"deleteBranch":  js.FuncOf(deleteBranch),
			"renameBranch":  js.FuncOf(renameBranch),
			"currentBranch": js.FuncOf(currentBranch),
		}),
	}))

	println("BrowserGit WASM module loaded - version", Version)

	<-c
}

// getVersion returns the version of the git-core module
func getVersion(this js.Value, args []js.Value) interface{} {
	return js.ValueOf(Version)
}

// initRepository initializes a new Git repository
// Args: path (string), options (optional: { bare, initialBranch, hashAlgorithm })
// Returns: { success, path, gitDir } or { error }
func initRepository(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing path argument")
	}

	path := args[0].String()

	// Parse options
	opts := repository.DefaultInitOptions()
	if len(args) >= 2 && args[1].Type() == js.TypeObject {
		optsJS := args[1]

		if !optsJS.Get("bare").IsUndefined() {
			opts.Bare = optsJS.Get("bare").Bool()
		}
		if !optsJS.Get("initialBranch").IsUndefined() {
			opts.InitialBranch = optsJS.Get("initialBranch").String()
		}
		if !optsJS.Get("hashAlgorithm").IsUndefined() {
			opts.HashAlgorithm = optsJS.Get("hashAlgorithm").String()
		}
	}

	// Initialize repository
	if err := repository.Init(path, opts); err != nil {
		return jsError("failed to initialize repository: " + err.Error())
	}

	// Get git dir
	gitDir := path
	if !opts.Bare {
		gitDir = path + "/.git"
	}

	return js.ValueOf(map[string]interface{}{
		"success": true,
		"path":    path,
		"gitDir":  gitDir,
	})
}

// openRepository opens an existing repository
// Args: path (string)
// Returns: { success, path, gitDir, config } or { error }
func openRepository(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing path argument")
	}

	path := args[0].String()

	repo, err := repository.Open(path)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Get config values
	userName, userEmail := repo.Config.GetUser()

	return js.ValueOf(map[string]interface{}{
		"success": true,
		"path":    repo.Path,
		"gitDir":  repo.GitDir,
		"config": map[string]interface{}{
			"bare":          repo.Config.IsBare(),
			"hashAlgorithm": repo.Config.GetHashAlgorithm(),
			"userName":      userName,
			"userEmail":     userEmail,
		},
	})
}

// isRepository checks if a path contains a repository
// Args: path (string)
// Returns: boolean
func isRepository(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(false)
	}

	path := args[0].String()
	return js.ValueOf(repository.IsRepository(path))
}

// findRepository finds a repository starting from path
// Args: path (string)
// Returns: { found, path } or { found: false }
func findRepository(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing path argument")
	}

	path := args[0].String()
	repoPath, err := repository.FindRepository(path)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"found": false,
		})
	}

	return js.ValueOf(map[string]interface{}{
		"found": true,
		"path":  repoPath,
	})
}

// hashSHA1 computes SHA-1 hash of data
// Args: data (string or Uint8Array)
// Returns: hex-encoded hash string
func hashSHA1(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "missing data argument",
		})
	}

	// Get data from argument
	data := jsValueToBytes(args[0])

	// Hash the data
	hasher := hash.NewSHA1()
	h := hasher.Hash(data)

	return js.ValueOf(h.String())
}

// hashSHA256 computes SHA-256 hash of data
// Args: data (string or Uint8Array)
// Returns: hex-encoded hash string
func hashSHA256(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "missing data argument",
		})
	}

	// Get data from argument
	data := jsValueToBytes(args[0])

	// Hash the data
	hasher := hash.NewSHA256()
	h := hasher.Hash(data)

	return js.ValueOf(h.String())
}

// hashBlob computes Git blob hash of content
// Args: content (string or Uint8Array), algorithm (optional, default: "sha1")
// Returns: hex-encoded hash string
func hashBlob(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "missing content argument",
		})
	}

	// Get content from argument
	content := jsValueToBytes(args[0])

	// Get algorithm (default to SHA-1)
	algo := hash.SHA1
	if len(args) >= 2 && args[1].Type() == js.TypeString {
		algoStr := args[1].String()
		if algoStr == "sha256" {
			algo = hash.SHA256
		}
	}

	// Create hasher and hash blob
	hasher, err := hash.NewHasher(algo)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": err.Error(),
		})
	}

	h := hash.HashBlob(hasher, content)

	return js.ValueOf(h.String())
}

// jsValueToBytes converts a JS value to bytes
// Handles both strings and Uint8Array
func jsValueToBytes(val js.Value) []byte {
	switch val.Type() {
	case js.TypeString:
		return []byte(val.String())
	case js.TypeObject:
		// Assume it's a Uint8Array
		length := val.Get("length").Int()
		data := make([]byte, length)
		js.CopyBytesToGo(data, val)
		return data
	default:
		return []byte{}
	}
}

// createBlob creates a blob object
// Args: content (string or Uint8Array)
// Returns: { type, size, hash } or { error }
func createBlob(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing content argument")
	}

	content := jsValueToBytes(args[0])
	blob := object.NewBlob(content)

	hasher := hash.NewSHA1()
	if err := blob.ComputeHash(hasher); err != nil {
		return jsError(err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"type": string(blob.Type()),
		"size": blob.Size(),
		"hash": blob.Hash().String(),
	})
}

// createTree creates a tree object
// Args: entries (array of {mode, name, hash})
// Returns: { type, size, hash } or { error }
func createTree(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing entries argument")
	}

	entriesJS := args[0]
	if entriesJS.Type() != js.TypeObject || entriesJS.Get("length").IsUndefined() {
		return jsError("entries must be an array")
	}

	tree := object.NewTree()
	length := entriesJS.Get("length").Int()

	for i := 0; i < length; i++ {
		entry := entriesJS.Index(i)

		modeJS := entry.Get("mode")
		nameJS := entry.Get("name")
		hashJS := entry.Get("hash")

		if modeJS.IsUndefined() || nameJS.IsUndefined() || hashJS.IsUndefined() {
			return jsError("each entry must have mode, name, and hash")
		}

		mode := object.FileMode(modeJS.Int())
		name := nameJS.String()
		hashStr := hashJS.String()

		h, err := hash.ParseHash(hashStr)
		if err != nil {
			return jsError("invalid hash: " + err.Error())
		}

		tree.AddEntryWithMode(mode, name, h)
	}

	hasher := hash.NewSHA1()
	if err := tree.ComputeHash(hasher); err != nil {
		return jsError(err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"type": string(tree.Type()),
		"size": tree.Size(),
		"hash": tree.Hash().String(),
	})
}

// createCommit creates a commit object
// Args: { tree, parents[], author, committer, message }
// Returns: { type, size, hash } or { error }
func createCommit(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing commit data argument")
	}

	data := args[0]
	commit := object.NewCommit()

	// Parse tree
	treeStr := data.Get("tree").String()
	if treeStr == "" {
		return jsError("missing tree field")
	}
	treeHash, err := hash.ParseHash(treeStr)
	if err != nil {
		return jsError("invalid tree hash: " + err.Error())
	}
	commit.Tree = treeHash

	// Parse parents
	parentsJS := data.Get("parents")
	if !parentsJS.IsUndefined() && parentsJS.Type() == js.TypeObject {
		length := parentsJS.Get("length").Int()
		for i := 0; i < length; i++ {
			parentStr := parentsJS.Index(i).String()
			parentHash, err := hash.ParseHash(parentStr)
			if err != nil {
				return jsError("invalid parent hash: " + err.Error())
			}
			commit.AddParent(parentHash)
		}
	}

	// Parse author
	authorJS := data.Get("author")
	if authorJS.IsUndefined() {
		return jsError("missing author field")
	}
	commit.Author = parseSignature(authorJS)

	// Parse committer
	committerJS := data.Get("committer")
	if committerJS.IsUndefined() {
		return jsError("missing committer field")
	}
	commit.Committer = parseSignature(committerJS)

	// Parse message
	commit.Message = data.Get("message").String()

	hasher := hash.NewSHA1()
	if err := commit.ComputeHash(hasher); err != nil {
		return jsError(err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"type": string(commit.Type()),
		"size": commit.Size(),
		"hash": commit.Hash().String(),
	})
}

// createTag creates a tag object
// Args: { target, targetType, name, tagger, message }
// Returns: { type, size, hash } or { error }
func createTag(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing tag data argument")
	}

	data := args[0]
	tag := object.NewTag()

	// Parse target
	targetStr := data.Get("target").String()
	if targetStr == "" {
		return jsError("missing target field")
	}
	targetHash, err := hash.ParseHash(targetStr)
	if err != nil {
		return jsError("invalid target hash: " + err.Error())
	}
	tag.Target = targetHash

	// Parse target type
	targetTypeStr := data.Get("targetType").String()
	if targetTypeStr == "" {
		return jsError("missing targetType field")
	}
	targetType, err := object.ParseType(targetTypeStr)
	if err != nil {
		return jsError("invalid target type: " + err.Error())
	}
	tag.TargetType = targetType

	// Parse name
	tag.Name = data.Get("name").String()
	if tag.Name == "" {
		return jsError("missing name field")
	}

	// Parse tagger
	taggerJS := data.Get("tagger")
	if taggerJS.IsUndefined() {
		return jsError("missing tagger field")
	}
	tag.Tagger = parseSignature(taggerJS)

	// Parse message
	tag.Message = data.Get("message").String()

	hasher := hash.NewSHA1()
	if err := tag.ComputeHash(hasher); err != nil {
		return jsError(err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"type": string(tag.Type()),
		"size": tag.Size(),
		"hash": tag.Hash().String(),
	})
}

// parseObject parses an object from raw data
// Args: data (Uint8Array with header)
// Returns: object representation or { error }
func parseObject(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing data argument")
	}

	data := jsValueToBytes(args[0])
	obj, err := object.ParseObjectWithHeader(data)
	if err != nil {
		return jsError("failed to parse object: " + err.Error())
	}

	return serializeObjectToJS(obj)
}

// compressObject compresses data using zlib
// Args: data (Uint8Array)
// Returns: Uint8Array or { error }
func compressObject(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing data argument")
	}

	data := jsValueToBytes(args[0])
	compressed, err := object.Compress(data)
	if err != nil {
		return jsError("compression failed: " + err.Error())
	}

	// Convert to Uint8Array
	dst := js.Global().Get("Uint8Array").New(len(compressed))
	js.CopyBytesToJS(dst, compressed)
	return dst
}

// decompressObject decompresses zlib data
// Args: data (Uint8Array)
// Returns: Uint8Array or { error }
func decompressObject(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing data argument")
	}

	data := jsValueToBytes(args[0])
	decompressed, err := object.Decompress(data)
	if err != nil {
		return jsError("decompression failed: " + err.Error())
	}

	// Convert to Uint8Array
	dst := js.Global().Get("Uint8Array").New(len(decompressed))
	js.CopyBytesToJS(dst, decompressed)
	return dst
}

// Helper functions

func jsError(msg string) js.Value {
	return js.ValueOf(map[string]interface{}{
		"error": msg,
	})
}

func parseSignature(val js.Value) object.Signature {
	name := val.Get("name").String()
	email := val.Get("email").String()
	timestamp := val.Get("timestamp").Int()

	return object.Signature{
		Name:  name,
		Email: email,
		When:  time.Unix(int64(timestamp), 0).UTC(),
	}
}

func serializeObjectToJS(obj object.Object) js.Value {
	result := map[string]interface{}{
		"type": string(obj.Type()),
		"size": obj.Size(),
		"hash": obj.Hash().String(),
	}

	switch o := obj.(type) {
	case *object.Blob:
		result["content"] = string(o.Content())

	case *object.Tree:
		entries := o.Entries()
		jsEntries := make([]interface{}, len(entries))
		for i, e := range entries {
			jsEntries[i] = map[string]interface{}{
				"mode": int(e.Mode),
				"name": e.Name,
				"hash": e.Hash.String(),
			}
		}
		result["entries"] = jsEntries

	case *object.Commit:
		result["tree"] = o.Tree.String()

		parents := make([]interface{}, len(o.Parents))
		for i, p := range o.Parents {
			parents[i] = p.String()
		}
		result["parents"] = parents

		result["author"] = map[string]interface{}{
			"name":      o.Author.Name,
			"email":     o.Author.Email,
			"timestamp": o.Author.When.Unix(),
		}

		result["committer"] = map[string]interface{}{
			"name":      o.Committer.Name,
			"email":     o.Committer.Email,
			"timestamp": o.Committer.When.Unix(),
		}

		result["message"] = o.Message

	case *object.Tag:
		result["target"] = o.Target.String()
		result["targetType"] = string(o.TargetType)
		result["name"] = o.Name
		result["tagger"] = map[string]interface{}{
			"name":      o.Tagger.Name,
			"email":     o.Tagger.Email,
			"timestamp": o.Tagger.When.Unix(),
		}
		result["message"] = o.Message
	}

	return js.ValueOf(result)
}

// addFiles adds files to the index (staging area)
// Args: repoPath (string), paths (array of strings), options (optional: { force, updateOnly })
// Returns: { success, filesAdded } or { error }
func addFiles(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("missing repoPath or paths arguments")
	}

	repoPath := args[0].String()
	pathsJS := args[1]

	// Parse paths array
	if pathsJS.Type() != js.TypeObject || pathsJS.Get("length").IsUndefined() {
		return jsError("paths must be an array")
	}

	length := pathsJS.Get("length").Int()
	paths := make([]string, length)
	for i := 0; i < length; i++ {
		paths[i] = pathsJS.Index(i).String()
	}

	// Parse options
	opts := index.AddOptions{
		Force:      false,
		UpdateOnly: false,
	}
	if len(args) >= 3 && args[2].Type() == js.TypeObject {
		optsJS := args[2]
		if !optsJS.Get("force").IsUndefined() {
			opts.Force = optsJS.Get("force").Bool()
		}
		if !optsJS.Get("updateOnly").IsUndefined() {
			opts.UpdateOnly = optsJS.Get("updateOnly").Bool()
		}
	}

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		return jsError("failed to load index: " + err.Error())
	}

	// Add files to index
	workTreePath := repo.WorkTree()
	if err := idx.Add(workTreePath, paths, opts); err != nil {
		return jsError("failed to add files: " + err.Error())
	}

	// Save index
	if err := idx.Save(indexPath); err != nil {
		return jsError("failed to save index: " + err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"success":    true,
		"filesAdded": len(paths),
	})
}

// createCommitFromIndex creates a commit from the index
// Args: repoPath (string), message (string), options (optional: { author: {name, email}, committer: {name, email} })
// Returns: { success, commitHash } or { error }
func createCommitFromIndex(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("missing repoPath or message arguments")
	}

	repoPath := args[0].String()
	message := args[1].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		return jsError("failed to load index: " + err.Error())
	}

	// Parse options
	var author, committer object.Signature
	userName, userEmail := repo.Config.GetUser()

	if len(args) >= 3 && args[2].Type() == js.TypeObject {
		optsJS := args[2]

		// Parse author
		if !optsJS.Get("author").IsUndefined() {
			author = parseSignature(optsJS.Get("author"))
		} else {
			author = index.DefaultSignature(userName, userEmail)
		}

		// Parse committer
		if !optsJS.Get("committer").IsUndefined() {
			committer = parseSignature(optsJS.Get("committer"))
		} else {
			committer = index.DefaultSignature(userName, userEmail)
		}
	} else {
		author = index.DefaultSignature(userName, userEmail)
		committer = index.DefaultSignature(userName, userEmail)
	}

	// Get parent commit
	parents, err := index.GetParentCommit(repo)
	if err != nil {
		parents = nil // Initial commit has no parents
	}

	// Write blobs to object database
	workTreePath := repo.WorkTree()
	if err := idx.WriteBlobs(workTreePath, repo.ObjectDB); err != nil {
		return jsError("failed to write blobs: " + err.Error())
	}

	// Create commit
	commitOpts := index.CommitOptions{
		Message:   message,
		Author:    author,
		Committer: committer,
		Parents:   parents,
	}

	commitHash, err := idx.CreateCommit(repo.Hasher, repo.ObjectDB, commitOpts)
	if err != nil {
		return jsError("failed to create commit: " + err.Error())
	}

	// Update HEAD
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		// Detached HEAD - update HEAD directly
		if err := repo.SetHEAD(commitHash.String()); err != nil {
			return jsError("failed to update HEAD: " + err.Error())
		}
	} else {
		// Update branch reference
		if err := repo.UpdateRef("refs/heads/"+currentBranch, commitHash); err != nil {
			return jsError("failed to update branch: " + err.Error())
		}
	}

	return js.ValueOf(map[string]interface{}{
		"success":    true,
		"commitHash": commitHash.String(),
	})
}

// getStatus gets the status of the repository
// Args: repoPath (string), options (optional: { includeUntracked, includeIgnored })
// Returns: { untracked[], modified[], staged[], deleted[], added[], isClean } or { error }
func getStatus(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing repoPath argument")
	}

	repoPath := args[0].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Load index
	indexPath := filepath.Join(repo.GitDir, "index")
	idx, err := index.Load(indexPath)
	if err != nil {
		return jsError("failed to load index: " + err.Error())
	}

	// Parse options
	opts := index.DefaultStatusOptions()
	if len(args) >= 2 && args[1].Type() == js.TypeObject {
		optsJS := args[1]
		if !optsJS.Get("includeUntracked").IsUndefined() {
			opts.IncludeUntracked = optsJS.Get("includeUntracked").Bool()
		}
		if !optsJS.Get("includeIgnored").IsUndefined() {
			opts.IncludeIgnored = optsJS.Get("includeIgnored").Bool()
		}
	}

	// Get HEAD commit
	var headCommit *object.Commit
	headStr, err := repo.HEAD()
	if err == nil {
		// Try to resolve HEAD
		var commitHash hash.Hash
		if headStr[:5] == "ref: " {
			// Symbolic ref
			refName := headStr[5:]
			commitHash, err = repo.ResolveRef(refName)
		} else {
			// Direct hash
			commitHash, err = hash.ParseHash(headStr)
		}

		if err == nil {
			// Load commit
			obj, err := repo.ObjectDB.Get(commitHash)
			if err == nil {
				headCommit, _ = obj.(*object.Commit)
			}
		}
	}

	// Get status
	workTreePath := repo.WorkTree()
	status, err := index.GetStatus(workTreePath, idx, headCommit, repo.ObjectDB, opts)
	if err != nil {
		return jsError("failed to get status: " + err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"success":    true,
		"untracked":  status.Untracked,
		"modified":   status.Modified,
		"staged":     status.Staged,
		"deleted":    status.Deleted,
		"added":      status.Added,
		"isClean":    status.IsClean(),
		"hasChanges": status.HasChanges(),
	})
}

// listBranches lists all branches in the repository
// Args: repoPath (string)
// Returns: { success, branches[], currentBranch } or { error }
func listBranches(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing repoPath argument")
	}

	repoPath := args[0].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// List branches
	branches, err := repo.ListBranches()
	if err != nil {
		return jsError("failed to list branches: " + err.Error())
	}

	// Get current branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		currentBranch = "" // Detached HEAD
	}

	return js.ValueOf(map[string]interface{}{
		"success":       true,
		"branches":      branches,
		"currentBranch": currentBranch,
	})
}

// createBranch creates a new branch
// Args: repoPath (string), name (string), commitHash (string, optional - defaults to HEAD)
// Returns: { success, branchName } or { error }
func createBranch(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("missing repoPath or name arguments")
	}

	repoPath := args[0].String()
	branchName := args[1].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Get commit hash
	var commitHash hash.Hash
	if len(args) >= 3 && args[2].Type() == js.TypeString {
		// Use provided commit hash
		hashStr := args[2].String()
		commitHash, err = hash.ParseHash(hashStr)
		if err != nil {
			return jsError("invalid commit hash: " + err.Error())
		}
	} else {
		// Use HEAD
		headStr, err := repo.HEAD()
		if err != nil {
			return jsError("failed to get HEAD: " + err.Error())
		}

		// Resolve HEAD to commit hash
		if headStr[:5] == "ref: " {
			refName := headStr[5:]
			commitHash, err = repo.ResolveRef(refName)
			if err != nil {
				return jsError("failed to resolve HEAD: " + err.Error())
			}
		} else {
			commitHash, err = hash.ParseHash(headStr)
			if err != nil {
				return jsError("invalid HEAD hash: " + err.Error())
			}
		}
	}

	// Create branch
	if err := repo.CreateBranch(branchName, commitHash); err != nil {
		return jsError("failed to create branch: " + err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"success":    true,
		"branchName": branchName,
	})
}

// deleteBranch deletes a branch
// Args: repoPath (string), name (string)
// Returns: { success, branchName } or { error }
func deleteBranch(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("missing repoPath or name arguments")
	}

	repoPath := args[0].String()
	branchName := args[1].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Delete branch
	if err := repo.DeleteBranch(branchName); err != nil {
		return jsError("failed to delete branch: " + err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"success":    true,
		"branchName": branchName,
	})
}

// renameBranch renames a branch
// Args: repoPath (string), oldName (string), newName (string)
// Returns: { success, oldName, newName } or { error }
func renameBranch(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return jsError("missing repoPath, oldName, or newName arguments")
	}

	repoPath := args[0].String()
	oldName := args[1].String()
	newName := args[2].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Rename branch
	if err := repo.RenameBranch(oldName, newName); err != nil {
		return jsError("failed to rename branch: " + err.Error())
	}

	return js.ValueOf(map[string]interface{}{
		"success": true,
		"oldName": oldName,
		"newName": newName,
	})
}

// currentBranch returns the current branch
// Args: repoPath (string)
// Returns: { success, branchName } or { error, detached: true }
func currentBranch(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsError("missing repoPath argument")
	}

	repoPath := args[0].String()

	// Open repository
	repo, err := repository.Open(repoPath)
	if err != nil {
		return jsError("failed to open repository: " + err.Error())
	}

	// Get current branch
	branchName, err := repo.CurrentBranch()
	if err != nil {
		// Detached HEAD
		return js.ValueOf(map[string]interface{}{
			"success":  false,
			"detached": true,
			"error":    err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"success":    true,
		"branchName": branchName,
	})
}
