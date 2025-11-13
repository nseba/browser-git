package object

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// Tag represents a Git tag object (annotated tag)
type Tag struct {
	Target     hash.Hash // Hash of the tagged object
	TargetType Type      // Type of the tagged object (commit, tree, blob, tag)
	Name       string    // Tag name
	Tagger     Signature // Person who created the tag
	Message    string    // Tag message
	hash       hash.Hash
}

// NewTag creates a new tag object
func NewTag() *Tag {
	return &Tag{}
}

// Type returns the object type (tag)
func (t *Tag) Type() Type {
	return TagType
}

// Hash returns the hash of the tag
func (t *Tag) Hash() hash.Hash {
	return t.hash
}

// SetHash sets the hash of the tag
func (t *Tag) SetHash(h hash.Hash) {
	t.hash = h
}

// Size returns the size of the serialized tag content
func (t *Tag) Size() int64 {
	var buf bytes.Buffer
	_ = t.Serialize(&buf)
	return int64(buf.Len())
}

// Serialize writes the tag content to a writer (without header)
func (t *Tag) Serialize(w io.Writer) error {
	// Write object line
	if _, err := fmt.Fprintf(w, "object %s\n", t.Target.String()); err != nil {
		return err
	}

	// Write type line
	if _, err := fmt.Fprintf(w, "type %s\n", t.TargetType); err != nil {
		return err
	}

	// Write tag name line
	if _, err := fmt.Fprintf(w, "tag %s\n", t.Name); err != nil {
		return err
	}

	// Write tagger line
	if _, err := fmt.Fprintf(w, "tagger %s\n", t.Tagger.Format()); err != nil {
		return err
	}

	// Write empty line before message
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	// Write message
	if _, err := w.Write([]byte(t.Message)); err != nil {
		return err
	}

	return nil
}

// SerializeWithHeader writes the complete tag (header + content) to a writer
func (t *Tag) SerializeWithHeader(w io.Writer) error {
	// Write header: "tag <size>\0"
	header := fmt.Sprintf("%s %d\x00", TagType, t.Size())
	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}

	// Write content
	return t.Serialize(w)
}

// Bytes returns the complete serialized tag (header + content)
func (t *Tag) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := t.SerializeWithHeader(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseTag parses a tag from raw content data (without header)
func ParseTag(data []byte) (*Tag, error) {
	tag := NewTag()
	lines := strings.Split(string(data), "\n")

	// Parse header lines
	i := 0
	for i < len(lines) {
		line := lines[i]
		if line == "" {
			// Empty line separates headers from message
			i++
			break
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid tag line: %s", line)
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "object":
			h, err := hash.ParseHash(value)
			if err != nil {
				return nil, fmt.Errorf("invalid object hash: %w", err)
			}
			tag.Target = h

		case "type":
			objType, err := ParseType(value)
			if err != nil {
				return nil, fmt.Errorf("invalid object type: %w", err)
			}
			tag.TargetType = objType

		case "tag":
			tag.Name = value

		case "tagger":
			sig, err := ParseSignature(value)
			if err != nil {
				return nil, fmt.Errorf("invalid tagger signature: %w", err)
			}
			tag.Tagger = sig

		default:
			// Ignore unknown headers for forward compatibility
		}

		i++
	}

	// Parse message (remaining lines)
	if i < len(lines) {
		tag.Message = strings.Join(lines[i:], "\n")
	}

	return tag, nil
}

// ComputeHash computes and sets the hash of the tag using the given hasher
func (t *Tag) ComputeHash(hasher hash.Hasher) error {
	data, err := t.Bytes()
	if err != nil {
		return err
	}
	t.hash = hasher.Hash(data)
	return nil
}

// IsLightweight returns true if this is a lightweight tag (should not be used for Tag objects)
// Lightweight tags are just refs pointing to commits, not tag objects
func (t *Tag) IsLightweight() bool {
	return false // Tag objects are always annotated tags
}
