// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/nseba/browser-git/git-core/pkg/hash"
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
