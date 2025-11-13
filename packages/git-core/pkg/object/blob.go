package object

import (
	"bytes"
	"fmt"
	"io"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// Blob represents a Git blob object (file content)
type Blob struct {
	content []byte
	hash    hash.Hash
}

// NewBlob creates a new blob object with the given content
func NewBlob(content []byte) *Blob {
	return &Blob{
		content: content,
	}
}

// NewBlobFromString creates a new blob object from a string
func NewBlobFromString(content string) *Blob {
	return NewBlob([]byte(content))
}

// Type returns the object type (blob)
func (b *Blob) Type() Type {
	return BlobType
}

// Hash returns the hash of the blob
func (b *Blob) Hash() hash.Hash {
	return b.hash
}

// SetHash sets the hash of the blob
func (b *Blob) SetHash(h hash.Hash) {
	b.hash = h
}

// Size returns the size of the blob content
func (b *Blob) Size() int64 {
	return int64(len(b.content))
}

// Content returns the blob content
func (b *Blob) Content() []byte {
	return b.content
}

// ContentString returns the blob content as a string
func (b *Blob) ContentString() string {
	return string(b.content)
}

// Serialize writes the blob content to a writer (without header)
func (b *Blob) Serialize(w io.Writer) error {
	_, err := w.Write(b.content)
	return err
}

// SerializeWithHeader writes the complete blob (header + content) to a writer
func (b *Blob) SerializeWithHeader(w io.Writer) error {
	// Write header: "blob <size>\0"
	header := fmt.Sprintf("%s %d\x00", BlobType, b.Size())
	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}

	// Write content
	return b.Serialize(w)
}

// Bytes returns the complete serialized blob (header + content)
func (b *Blob) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := b.SerializeWithHeader(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseBlob parses a blob from raw content data (without header)
func ParseBlob(data []byte) (*Blob, error) {
	return NewBlob(data), nil
}

// ComputeHash computes and sets the hash of the blob using the given hasher
func (b *Blob) ComputeHash(hasher hash.Hasher) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	b.hash = hasher.Hash(data)
	return nil
}

// Equals compares two blobs for equality
func (b *Blob) Equals(other *Blob) bool {
	if b == nil || other == nil {
		return b == other
	}
	return bytes.Equal(b.content, other.content)
}
