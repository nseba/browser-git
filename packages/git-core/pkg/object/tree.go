package object

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// FileMode represents the Unix file mode stored in tree entries
type FileMode uint32

const (
	// ModeDir is the mode for directories (040000)
	ModeDir FileMode = 0040000
	// ModeRegular is the mode for regular files (100644)
	ModeRegular FileMode = 0100644
	// ModeExecutable is the mode for executable files (100755)
	ModeExecutable FileMode = 0100755
	// ModeSymlink is the mode for symbolic links (120000)
	ModeSymlink FileMode = 0120000
	// ModeGitlink is the mode for Git submodules (160000)
	ModeGitlink FileMode = 0160000
)

// TreeEntry represents an entry in a Git tree object
type TreeEntry struct {
	Mode FileMode
	Name string
	Hash hash.Hash
}

// Tree represents a Git tree object (directory)
type Tree struct {
	entries []TreeEntry
	hash    hash.Hash
}

// NewTree creates a new tree object
func NewTree() *Tree {
	return &Tree{
		entries: make([]TreeEntry, 0),
	}
}

// Type returns the object type (tree)
func (t *Tree) Type() Type {
	return TreeType
}

// Hash returns the hash of the tree
func (t *Tree) Hash() hash.Hash {
	return t.hash
}

// SetHash sets the hash of the tree
func (t *Tree) SetHash(h hash.Hash) {
	t.hash = h
}

// Size returns the size of the serialized tree content
func (t *Tree) Size() int64 {
	size := int64(0)
	for _, entry := range t.entries {
		// Mode (as string) + space + name + null byte + hash bytes
		modeStr := fmt.Sprintf("%o", entry.Mode)
		size += int64(len(modeStr) + 1 + len(entry.Name) + 1 + len(entry.Hash))
	}
	return size
}

// Entries returns all tree entries
func (t *Tree) Entries() []TreeEntry {
	return t.entries
}

// AddEntry adds an entry to the tree
func (t *Tree) AddEntry(entry TreeEntry) {
	t.entries = append(t.entries, entry)
}

// AddEntryWithMode adds an entry to the tree with the given mode, name, and hash
func (t *Tree) AddEntryWithMode(mode FileMode, name string, h hash.Hash) {
	t.AddEntry(TreeEntry{
		Mode: mode,
		Name: name,
		Hash: h,
	})
}

// Sort sorts the tree entries (required for correct Git tree format)
// Git sorts entries by name, but directories have an implicit "/" suffix
func (t *Tree) Sort() {
	sort.Slice(t.entries, func(i, j int) bool {
		nameI := t.entries[i].Name
		nameJ := t.entries[j].Name

		// For directories, add implicit "/" for sorting
		if t.entries[i].Mode == ModeDir {
			nameI += "/"
		}
		if t.entries[j].Mode == ModeDir {
			nameJ += "/"
		}

		return nameI < nameJ
	})
}

// Serialize writes the tree content to a writer (without header)
func (t *Tree) Serialize(w io.Writer) error {
	// Sort entries before serializing
	t.Sort()

	for _, entry := range t.entries {
		// Write mode (octal) + space + name + null byte
		modeStr := fmt.Sprintf("%o", entry.Mode)
		if _, err := w.Write([]byte(modeStr)); err != nil {
			return err
		}
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if _, err := w.Write([]byte(entry.Name)); err != nil {
			return err
		}
		if _, err := w.Write([]byte{0}); err != nil {
			return err
		}

		// Write hash bytes (20 or 32 bytes depending on algorithm)
		if _, err := w.Write(entry.Hash.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// SerializeWithHeader writes the complete tree (header + content) to a writer
func (t *Tree) SerializeWithHeader(w io.Writer) error {
	// Write header: "tree <size>\0"
	header := fmt.Sprintf("%s %d\x00", TreeType, t.Size())
	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}

	// Write content
	return t.Serialize(w)
}

// Bytes returns the complete serialized tree (header + content)
func (t *Tree) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := t.SerializeWithHeader(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseTree parses a tree from raw content data (without header)
func ParseTree(data []byte) (*Tree, error) {
	tree := NewTree()
	offset := 0

	for offset < len(data) {
		// Parse mode (octal string until space)
		spaceIdx := bytes.IndexByte(data[offset:], ' ')
		if spaceIdx == -1 {
			return nil, fmt.Errorf("invalid tree entry: missing space after mode")
		}

		modeStr := string(data[offset : offset+spaceIdx])
		mode64, err := strconv.ParseUint(modeStr, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid tree entry mode: %w", err)
		}
		mode := FileMode(mode64)
		offset += spaceIdx + 1

		// Parse name (until null byte)
		nullIdx := bytes.IndexByte(data[offset:], 0)
		if nullIdx == -1 {
			return nil, fmt.Errorf("invalid tree entry: missing null byte after name")
		}

		name := string(data[offset : offset+nullIdx])
		offset += nullIdx + 1

		// Parse hash (20 bytes for SHA-1, 32 bytes for SHA-256)
		// Detect hash size: if remaining data is exactly 20 or 32 bytes, use that
		// Otherwise, check if there's more data after position 20/32
		remaining := len(data) - offset
		hashSize := 20 // Default to SHA-1

		if remaining >= 32 {
			// Look ahead to see if there's another entry after 32 bytes
			// Next entry would start with a mode (digits)
			if remaining == 32 {
				hashSize = 32
			} else if offset+32 < len(data) {
				// Check if byte at offset+20 could be start of next mode (digit 0-9)
				nextByte20 := data[offset+20]
				nextByte32 := data[offset+32]
				// If offset+32 looks like start of mode, use SHA-256
				// If offset+20 looks like start of mode, use SHA-1
				if nextByte32 >= '0' && nextByte32 <= '9' {
					hashSize = 32
				} else if nextByte20 >= '0' && nextByte20 <= '9' {
					hashSize = 20
				} else {
					// Default to 20 if ambiguous
					hashSize = 20
				}
			}
		} else if remaining >= 20 {
			hashSize = 20
		} else {
			return nil, fmt.Errorf("invalid tree entry: incomplete hash (only %d bytes remaining)", remaining)
		}

		if offset+hashSize > len(data) {
			return nil, fmt.Errorf("invalid tree entry: incomplete hash")
		}

		entryHash := hash.NewHash(data[offset : offset+hashSize])
		offset += hashSize

		tree.AddEntry(TreeEntry{
			Mode: mode,
			Name: name,
			Hash: entryHash,
		})
	}

	return tree, nil
}

// ComputeHash computes and sets the hash of the tree using the given hasher
func (t *Tree) ComputeHash(hasher hash.Hasher) error {
	data, err := t.Bytes()
	if err != nil {
		return err
	}
	t.hash = hasher.Hash(data)
	return nil
}

// FindEntry finds an entry by name
func (t *Tree) FindEntry(name string) (*TreeEntry, bool) {
	for i, entry := range t.entries {
		if entry.Name == name {
			return &t.entries[i], true
		}
	}
	return nil, false
}

// IsValidMode checks if a file mode is valid
func IsValidMode(mode FileMode) bool {
	return mode == ModeDir || mode == ModeRegular || mode == ModeExecutable ||
		mode == ModeSymlink || mode == ModeGitlink
}

// String returns a string representation of the file mode
func (m FileMode) String() string {
	return fmt.Sprintf("%06o", m)
}
