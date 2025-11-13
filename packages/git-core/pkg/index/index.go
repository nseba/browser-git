package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// Index represents the Git staging area (index)
type Index struct {
	Version int
	Entries []*Entry
}

// Entry represents a single file entry in the index
type Entry struct {
	// Metadata
	CTime     time.Time // Creation time
	MTime     time.Time // Modification time
	Dev       uint32    // Device
	Ino       uint32    // Inode
	Mode      uint32    // File mode
	UID       uint32    // User ID
	GID       uint32    // Group ID
	Size      uint32    // File size
	Hash      hash.Hash // SHA-1 hash of content
	Flags     uint16    // Flags
	Path      string    // File path
	StageFlag uint8     // Stage (0 = normal, 1-3 = conflict stages)
}

// FileMode constants
const (
	FileModeRegular    uint32 = 0100644
	FileModeExecutable uint32 = 0100755
	FileModeSymlink    uint32 = 0120000
	FileModeGitlink    uint32 = 0160000
)

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		Version: 2,
		Entries: make([]*Entry, 0),
	}
}

// AddEntry adds an entry to the index
func (idx *Index) AddEntry(entry *Entry) {
	// Remove existing entry with same path
	idx.RemoveEntry(entry.Path)

	// Add new entry
	idx.Entries = append(idx.Entries, entry)

	// Sort entries by path
	idx.Sort()
}

// RemoveEntry removes an entry from the index by path
func (idx *Index) RemoveEntry(path string) bool {
	for i, entry := range idx.Entries {
		if entry.Path == path {
			idx.Entries = append(idx.Entries[:i], idx.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// GetEntry retrieves an entry by path
func (idx *Index) GetEntry(path string) (*Entry, bool) {
	for _, entry := range idx.Entries {
		if entry.Path == path {
			return entry, true
		}
	}
	return nil, false
}

// HasEntry checks if an entry exists
func (idx *Index) HasEntry(path string) bool {
	_, ok := idx.GetEntry(path)
	return ok
}

// Sort sorts entries by path (required by Git index format)
func (idx *Index) Sort() {
	sort.Slice(idx.Entries, func(i, j int) bool {
		return idx.Entries[i].Path < idx.Entries[j].Path
	})
}

// EntryCount returns the number of entries
func (idx *Index) EntryCount() int {
	return len(idx.Entries)
}

// Clear removes all entries from the index
func (idx *Index) Clear() {
	idx.Entries = make([]*Entry, 0)
}

// NewEntryFromFile creates an index entry from a file on disk
func NewEntryFromFile(path string, workTreePath string) (*Entry, error) {
	fullPath := filepath.Join(workTreePath, path)

	// Get file info
	info, err := os.Lstat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Compute hash
	hasher := hash.NewSHA1()
	blobHash := hash.HashBlob(hasher, content)

	// Determine mode
	mode := FileModeRegular
	if info.Mode()&0111 != 0 {
		mode = FileModeExecutable
	}
	if info.Mode()&os.ModeSymlink != 0 {
		mode = FileModeSymlink
	}

	entry := &Entry{
		CTime:     info.ModTime(), // Use ModTime for both for now
		MTime:     info.ModTime(),
		Mode:      mode,
		Size:      uint32(info.Size()),
		Hash:      blobHash,
		Path:      filepath.ToSlash(path), // Convert to forward slashes
		StageFlag: 0,
	}

	return entry, nil
}

// IsModified checks if a file has been modified compared to the index entry
func (e *Entry) IsModified(workTreePath string) (bool, error) {
	fullPath := filepath.Join(workTreePath, e.Path)

	// Get current file info
	info, err := os.Lstat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // File was deleted
		}
		return false, err
	}

	// Check size first (quick check)
	if uint32(info.Size()) != e.Size {
		return true, nil
	}

	// Check modification time
	if !info.ModTime().Equal(e.MTime) {
		// Size matches but mtime differs - need to check content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return false, err
		}

		hasher := hash.NewSHA1()
		currentHash := hash.HashBlob(hasher, content)

		return !currentHash.Equals(e.Hash), nil
	}

	return false, nil
}

// Serialize writes the index to a writer (Git index format version 2)
func (idx *Index) Serialize(w io.Writer) error {
	buf := new(bytes.Buffer)

	// Write header
	// Signature: "DIRC" (DirCache)
	buf.Write([]byte("DIRC"))

	// Version
	binary.Write(buf, binary.BigEndian, uint32(idx.Version))

	// Number of entries
	binary.Write(buf, binary.BigEndian, uint32(len(idx.Entries)))

	// Write entries
	for _, entry := range idx.Entries {
		if err := entry.serialize(buf); err != nil {
			return err
		}
	}

	// Compute checksum (SHA-1 of the index file)
	indexData := buf.Bytes()
	checksum := sha1.Sum(indexData)

	// Write index data and checksum
	if _, err := w.Write(indexData); err != nil {
		return err
	}
	if _, err := w.Write(checksum[:]); err != nil {
		return err
	}

	return nil
}

// serialize writes an entry to a buffer
func (e *Entry) serialize(buf *bytes.Buffer) error {
	// ctime (seconds and nanoseconds)
	binary.Write(buf, binary.BigEndian, uint32(e.CTime.Unix()))
	binary.Write(buf, binary.BigEndian, uint32(e.CTime.Nanosecond()))

	// mtime (seconds and nanoseconds)
	binary.Write(buf, binary.BigEndian, uint32(e.MTime.Unix()))
	binary.Write(buf, binary.BigEndian, uint32(e.MTime.Nanosecond()))

	// dev, ino, mode, uid, gid, size
	binary.Write(buf, binary.BigEndian, e.Dev)
	binary.Write(buf, binary.BigEndian, e.Ino)
	binary.Write(buf, binary.BigEndian, e.Mode)
	binary.Write(buf, binary.BigEndian, e.UID)
	binary.Write(buf, binary.BigEndian, e.GID)
	binary.Write(buf, binary.BigEndian, e.Size)

	// Hash (20 bytes for SHA-1)
	buf.Write(e.Hash.Bytes())

	// Flags (assume-valid, extended, stage, name length)
	pathLen := len(e.Path)
	if pathLen > 0xFFF {
		pathLen = 0xFFF
	}
	flags := uint16(pathLen) | (uint16(e.StageFlag) << 12)
	binary.Write(buf, binary.BigEndian, flags)

	// Path (null-terminated)
	buf.WriteString(e.Path)
	buf.WriteByte(0)

	// Padding (entries are padded to 8-byte alignment)
	entrySize := 62 + len(e.Path) + 1 // 62 bytes of fixed data + path + null
	padding := (8 - (entrySize % 8)) % 8
	for i := 0; i < padding; i++ {
		buf.WriteByte(0)
	}

	return nil
}

// Deserialize reads an index from a reader
func Deserialize(r io.Reader) (*Index, error) {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) < 12 {
		return nil, fmt.Errorf("index file too short")
	}

	// Verify checksum
	if len(data) < 20 {
		return nil, fmt.Errorf("index file missing checksum")
	}

	indexData := data[:len(data)-20]
	storedChecksum := data[len(data)-20:]
	computedChecksum := sha1.Sum(indexData)

	if !bytes.Equal(storedChecksum, computedChecksum[:]) {
		return nil, fmt.Errorf("index checksum mismatch")
	}

	buf := bytes.NewReader(indexData)

	// Read header
	sig := make([]byte, 4)
	if _, err := buf.Read(sig); err != nil {
		return nil, err
	}
	if string(sig) != "DIRC" {
		return nil, fmt.Errorf("invalid index signature")
	}

	var version uint32
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return nil, err
	}
	if version != 2 && version != 3 && version != 4 {
		return nil, fmt.Errorf("unsupported index version: %d", version)
	}

	var numEntries uint32
	if err := binary.Read(buf, binary.BigEndian, &numEntries); err != nil {
		return nil, err
	}

	idx := &Index{
		Version: int(version),
		Entries: make([]*Entry, 0, numEntries),
	}

	// Read entries
	for i := uint32(0); i < numEntries; i++ {
		entry, err := deserializeEntry(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read entry %d: %w", i, err)
		}
		idx.Entries = append(idx.Entries, entry)
	}

	return idx, nil
}

// deserializeEntry reads an entry from a buffer
func deserializeEntry(buf *bytes.Reader) (*Entry, error) {
	entry := &Entry{}

	// ctime
	var ctimeSec, ctimeNsec uint32
	binary.Read(buf, binary.BigEndian, &ctimeSec)
	binary.Read(buf, binary.BigEndian, &ctimeNsec)
	entry.CTime = time.Unix(int64(ctimeSec), int64(ctimeNsec))

	// mtime
	var mtimeSec, mtimeNsec uint32
	binary.Read(buf, binary.BigEndian, &mtimeSec)
	binary.Read(buf, binary.BigEndian, &mtimeNsec)
	entry.MTime = time.Unix(int64(mtimeSec), int64(mtimeNsec))

	// dev, ino, mode, uid, gid, size
	binary.Read(buf, binary.BigEndian, &entry.Dev)
	binary.Read(buf, binary.BigEndian, &entry.Ino)
	binary.Read(buf, binary.BigEndian, &entry.Mode)
	binary.Read(buf, binary.BigEndian, &entry.UID)
	binary.Read(buf, binary.BigEndian, &entry.GID)
	binary.Read(buf, binary.BigEndian, &entry.Size)

	// Hash (20 bytes)
	hashBytes := make([]byte, 20)
	if _, err := buf.Read(hashBytes); err != nil {
		return nil, err
	}
	entry.Hash = hash.NewHash(hashBytes)

	// Flags
	var flags uint16
	binary.Read(buf, binary.BigEndian, &flags)
	entry.StageFlag = uint8((flags >> 12) & 0x3)
	pathLen := int(flags & 0xFFF)

	// Path
	pathBytes := make([]byte, pathLen)
	if _, err := buf.Read(pathBytes); err != nil {
		return nil, err
	}
	entry.Path = string(pathBytes)

	// Read null terminator
	var nullByte byte
	binary.Read(buf, binary.BigEndian, &nullByte)

	// Skip padding
	entrySize := 62 + pathLen + 1
	padding := (8 - (entrySize % 8)) % 8
	if padding > 0 {
		buf.Seek(int64(padding), io.SeekCurrent)
	}

	return entry, nil
}

// Save saves the index to a file
func (idx *Index) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return idx.Serialize(file)
}

// Load loads an index from a file
func Load(path string) (*Index, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index if file doesn't exist
			return NewIndex(), nil
		}
		return nil, err
	}
	defer file.Close()

	return Deserialize(file)
}
