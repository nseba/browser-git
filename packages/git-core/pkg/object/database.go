package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// Database is the interface for storing and retrieving Git objects
type Database interface {
	// Get retrieves an object by its hash
	Get(h hash.Hash) (Object, error)

	// Put stores an object and returns its hash
	Put(obj Object) (hash.Hash, error)

	// Has checks if an object exists
	Has(h hash.Hash) bool

	// Delete removes an object
	Delete(h hash.Hash) error

	// List returns all object hashes in the database
	List() ([]hash.Hash, error)

	// Close closes the database
	Close() error
}

// Reader is the interface for reading objects from storage
type Reader interface {
	// Read reads compressed object data for the given hash
	Read(h hash.Hash) ([]byte, error)

	// Has checks if an object exists
	Has(h hash.Hash) bool
}

// Writer is the interface for writing objects to storage
type Writer interface {
	// Write writes compressed object data with the given hash
	Write(h hash.Hash, data []byte) error

	// Delete removes an object
	Delete(h hash.Hash) error
}

// Storage is the interface for object storage backends
type Storage interface {
	Reader
	Writer

	// List returns all object hashes
	List() ([]hash.Hash, error)

	// Close closes the storage
	Close() error
}

// Compress compresses object data using zlib
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, fmt.Errorf("failed to compress object: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compressor: %w", err)
	}

	return buf.Bytes(), nil
}

// Decompress decompresses object data using zlib
func Decompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create decompressor: %w", err)
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("failed to decompress object: %w", err)
	}

	return buf.Bytes(), nil
}

// ObjectDatabase implements the Database interface
type ObjectDatabase struct {
	storage Storage
	hasher  hash.Hasher
}

// NewObjectDatabase creates a new object database
func NewObjectDatabase(storage Storage, hasher hash.Hasher) *ObjectDatabase {
	return &ObjectDatabase{
		storage: storage,
		hasher:  hasher,
	}
}

// Get retrieves an object by its hash
func (db *ObjectDatabase) Get(h hash.Hash) (Object, error) {
	// Read compressed data from storage
	compressed, err := db.storage.Read(h)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	// Decompress
	data, err := Decompress(compressed)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress object: %w", err)
	}

	// Parse object
	obj, err := ParseObjectWithHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object: %w", err)
	}

	// Set hash
	obj.SetHash(h)

	return obj, nil
}

// Put stores an object and returns its hash
func (db *ObjectDatabase) Put(obj Object) (hash.Hash, error) {
	// Serialize object
	data, err := serializeObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object: %w", err)
	}

	// Compute hash
	h := db.hasher.Hash(data)
	obj.SetHash(h)

	// Compress
	compressed, err := Compress(data)
	if err != nil {
		return nil, fmt.Errorf("failed to compress object: %w", err)
	}

	// Write to storage
	if err := db.storage.Write(h, compressed); err != nil {
		return nil, fmt.Errorf("failed to write object: %w", err)
	}

	return h, nil
}

// Has checks if an object exists
func (db *ObjectDatabase) Has(h hash.Hash) bool {
	return db.storage.Has(h)
}

// Delete removes an object
func (db *ObjectDatabase) Delete(h hash.Hash) error {
	return db.storage.Delete(h)
}

// List returns all object hashes in the database
func (db *ObjectDatabase) List() ([]hash.Hash, error) {
	return db.storage.List()
}

// Close closes the database
func (db *ObjectDatabase) Close() error {
	return db.storage.Close()
}

// serializeObject serializes an object with its header
func serializeObject(obj Object) ([]byte, error) {
	var buf bytes.Buffer
	if err := obj.SerializeWithHeader(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetType retrieves only the type of an object without fully parsing it
func GetType(data []byte) (Type, error) {
	// Decompress if needed
	decompressed := data
	if len(data) > 2 && data[0] == 0x78 && (data[1] == 0x9c || data[1] == 0x01 || data[1] == 0xda) {
		var err error
		decompressed, err = Decompress(data)
		if err != nil {
			return "", err
		}
	}

	// Parse header to get type
	headerEnd := bytes.IndexByte(decompressed, 0)
	if headerEnd == -1 {
		return "", fmt.Errorf("invalid object: missing null byte in header")
	}

	header := string(decompressed[:headerEnd])
	var objType string
	var size int64
	_, err := fmt.Sscanf(header, "%s %d", &objType, &size)
	if err != nil {
		return "", fmt.Errorf("invalid object header: %w", err)
	}

	return ParseType(objType)
}
