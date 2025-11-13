package object

import (
	"fmt"
	"io"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// Type represents the type of a Git object
type Type string

const (
	// BlobType represents a blob object (file content)
	BlobType Type = "blob"
	// TreeType represents a tree object (directory)
	TreeType Type = "tree"
	// CommitType represents a commit object
	CommitType Type = "commit"
	// TagType represents a tag object
	TagType Type = "tag"
)

// Object is the interface that all Git objects must implement
type Object interface {
	// Type returns the type of the object
	Type() Type

	// Hash returns the hash of the object
	Hash() hash.Hash

	// Size returns the size of the serialized object content
	Size() int64

	// Serialize writes the object content to a writer (without the header)
	Serialize(w io.Writer) error

	// SerializeWithHeader writes the complete object (header + content) to a writer
	SerializeWithHeader(w io.Writer) error

	// SetHash sets the hash of the object (computed after serialization)
	SetHash(h hash.Hash)
}

// ParseObject parses a Git object from raw data
func ParseObject(objType Type, data []byte) (Object, error) {
	switch objType {
	case BlobType:
		return ParseBlob(data)
	case TreeType:
		return ParseTree(data)
	case CommitType:
		return ParseCommit(data)
	case TagType:
		return ParseTag(data)
	default:
		return nil, fmt.Errorf("unknown object type: %s", objType)
	}
}

// ParseObjectWithHeader parses a Git object from data that includes the header
func ParseObjectWithHeader(data []byte) (Object, error) {
	// Parse header: "<type> <size>\0"
	headerEnd := -1
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			headerEnd = i
			break
		}
	}

	if headerEnd == -1 {
		return nil, fmt.Errorf("invalid object: missing null byte in header")
	}

	// Parse type and size from header
	header := string(data[:headerEnd])
	var objType string
	var size int64
	_, err := fmt.Sscanf(header, "%s %d", &objType, &size)
	if err != nil {
		return nil, fmt.Errorf("invalid object header: %w", err)
	}

	// Extract content
	content := data[headerEnd+1:]
	if int64(len(content)) != size {
		return nil, fmt.Errorf("object size mismatch: expected %d, got %d", size, len(content))
	}

	return ParseObject(Type(objType), content)
}

// IsValidType checks if a type string is a valid Git object type
func IsValidType(t Type) bool {
	return t == BlobType || t == TreeType || t == CommitType || t == TagType
}

// ParseType parses an object type string
func ParseType(s string) (Type, error) {
	t := Type(s)
	if !IsValidType(t) {
		return "", fmt.Errorf("invalid object type: %s", s)
	}
	return t, nil
}
