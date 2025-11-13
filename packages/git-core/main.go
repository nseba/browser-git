// +build js,wasm

package main

import (
	"syscall/js"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
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
		"init":    js.FuncOf(initRepository),
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
	}))

	println("BrowserGit WASM module loaded - version", Version)

	<-c
}

// getVersion returns the version of the git-core module
func getVersion(this js.Value, args []js.Value) interface{} {
	return js.ValueOf(Version)
}

// initRepository initializes a new Git repository
// Placeholder implementation
func initRepository(this js.Value, args []js.Value) interface{} {
	// TODO: Implement repository initialization
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"message": "Repository initialized (placeholder)",
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
