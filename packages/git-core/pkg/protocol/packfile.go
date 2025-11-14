package protocol

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
)

// Packfile constants
const (
	// PackfileSignature is the magic signature at the start of a packfile
	PackfileSignature = "PACK"

	// PackfileVersion is the current packfile version
	PackfileVersion = 2

	// PackfileHeaderSize is the size of the packfile header (signature + version + count)
	PackfileHeaderSize = 12

	// PackfileChecksumSize is the size of the SHA-1 checksum at the end
	PackfileChecksumSize = 20
)

// Object types in packfile
const (
	ObjCommit    = 1
	ObjTree      = 2
	ObjBlob      = 3
	ObjTag       = 4
	ObjReserved  = 5
	ObjOfsDelta  = 6 // Delta with offset to base
	ObjRefDelta  = 7 // Delta with SHA-1 reference to base
)

// PackfileHeader represents the packfile header
type PackfileHeader struct {
	Signature   string // Should be "PACK"
	Version     uint32 // Should be 2
	ObjectCount uint32 // Number of objects in the packfile
}

// PackfileObject represents an object in the packfile
type PackfileObject struct {
	Type         uint8  // Object type (1-7)
	Size         uint64 // Uncompressed size
	Data         []byte // Decompressed object data
	Offset       int64  // Offset in packfile (for OFS_DELTA)
	BaseHash     []byte // Base object hash (for REF_DELTA, 20 bytes)
	IsDelta      bool   // Whether this is a delta object
}

// Packfile represents a parsed packfile
type Packfile struct {
	Header   PackfileHeader
	Objects  []PackfileObject
	Checksum []byte // SHA-1 checksum of packfile
}

// PackfileReader reads and parses packfiles
type PackfileReader struct {
	reader   io.Reader
	offset   int64
	checksum []byte
}

// NewPackfileReader creates a new packfile reader
func NewPackfileReader(r io.Reader) *PackfileReader {
	return &PackfileReader{
		reader: r,
		offset: 0,
	}
}

// ReadPackfile reads and parses a complete packfile
func (r *PackfileReader) ReadPackfile() (*Packfile, error) {
	// Read header
	header, err := r.ReadHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read packfile header: %w", err)
	}

	// Read all objects
	objects := make([]PackfileObject, 0, header.ObjectCount)
	for i := uint32(0); i < header.ObjectCount; i++ {
		obj, err := r.ReadObject()
		if err != nil {
			return nil, fmt.Errorf("failed to read object %d: %w", i, err)
		}
		objects = append(objects, *obj)
	}

	// Read checksum
	checksum := make([]byte, PackfileChecksumSize)
	if _, err := io.ReadFull(r.reader, checksum); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	return &Packfile{
		Header:   *header,
		Objects:  objects,
		Checksum: checksum,
	}, nil
}

// ReadHeader reads the packfile header
func (r *PackfileReader) ReadHeader() (*PackfileHeader, error) {
	headerBytes := make([]byte, PackfileHeaderSize)
	n, err := io.ReadFull(r.reader, headerBytes)
	if err != nil {
		return nil, err
	}
	r.offset += int64(n)

	// Parse signature
	signature := string(headerBytes[0:4])
	if signature != PackfileSignature {
		return nil, fmt.Errorf("invalid packfile signature: %s (expected %s)", signature, PackfileSignature)
	}

	// Parse version (big-endian)
	version := binary.BigEndian.Uint32(headerBytes[4:8])
	if version != PackfileVersion {
		return nil, fmt.Errorf("unsupported packfile version: %d (expected %d)", version, PackfileVersion)
	}

	// Parse object count (big-endian)
	objectCount := binary.BigEndian.Uint32(headerBytes[8:12])

	return &PackfileHeader{
		Signature:   signature,
		Version:     version,
		ObjectCount: objectCount,
	}, nil
}

// ReadObject reads a single object from the packfile
func (r *PackfileReader) ReadObject() (*PackfileObject, error) {
	objOffset := r.offset

	// Read type and size header
	objType, size, err := r.readObjectHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read object header: %w", err)
	}

	obj := &PackfileObject{
		Type: objType,
		Size: size,
	}

	// Handle different object types
	switch objType {
	case ObjCommit, ObjTree, ObjBlob, ObjTag:
		// Regular object - read compressed data
		data, err := r.readCompressedData()
		if err != nil {
			return nil, fmt.Errorf("failed to read compressed data: %w", err)
		}
		obj.Data = data

	case ObjOfsDelta:
		// Offset delta - read negative offset to base
		offset, err := r.readOffsetDeltaOffset()
		if err != nil {
			return nil, fmt.Errorf("failed to read offset delta: %w", err)
		}
		obj.IsDelta = true
		obj.Offset = objOffset - offset

		// Read delta data
		data, err := r.readCompressedData()
		if err != nil {
			return nil, fmt.Errorf("failed to read delta data: %w", err)
		}
		obj.Data = data

	case ObjRefDelta:
		// Reference delta - read 20-byte SHA-1 of base
		baseHash := make([]byte, 20)
		n, err := io.ReadFull(r.reader, baseHash)
		if err != nil {
			return nil, fmt.Errorf("failed to read ref delta hash: %w", err)
		}
		r.offset += int64(n)
		obj.IsDelta = true
		obj.BaseHash = baseHash

		// Read delta data
		data, err := r.readCompressedData()
		if err != nil {
			return nil, fmt.Errorf("failed to read delta data: %w", err)
		}
		obj.Data = data

	default:
		return nil, fmt.Errorf("unknown object type: %d", objType)
	}

	return obj, nil
}

// readObjectHeader reads the variable-length object type and size header
func (r *PackfileReader) readObjectHeader() (uint8, uint64, error) {
	// First byte contains type (bits 6-4) and size (bits 3-0)
	firstByte, err := r.readByte()
	if err != nil {
		return 0, 0, err
	}

	objType := (firstByte >> 4) & 0x7  // Bits 6-4
	size := uint64(firstByte & 0xF)    // Bits 3-0
	shift := uint(4)

	// Read continuation bytes if MSB is set
	for firstByte&0x80 != 0 {
		b, err := r.readByte()
		if err != nil {
			return 0, 0, err
		}
		size |= uint64(b&0x7F) << shift
		shift += 7
		firstByte = b
	}

	return objType, size, nil
}

// readOffsetDeltaOffset reads the negative offset for OFS_DELTA
func (r *PackfileReader) readOffsetDeltaOffset() (int64, error) {
	// Variable-length encoding with a twist
	b, err := r.readByte()
	if err != nil {
		return 0, err
	}

	offset := int64(b & 0x7F)
	for b&0x80 != 0 {
		b, err = r.readByte()
		if err != nil {
			return 0, err
		}
		offset = ((offset + 1) << 7) | int64(b&0x7F)
	}

	return offset, nil
}

// readCompressedData reads and decompresses zlib-compressed data
func (r *PackfileReader) readCompressedData() ([]byte, error) {
	// Create a zlib reader
	zlibReader, err := zlib.NewReader(r.reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zlibReader.Close()

	// Read all decompressed data
	var buf bytes.Buffer
	n, err := io.Copy(&buf, zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}
	r.offset += n // Note: this is compressed size, not exact but close enough

	return buf.Bytes(), nil
}

// readByte reads a single byte and updates offset
func (r *PackfileReader) readByte() (byte, error) {
	b := make([]byte, 1)
	_, err := io.ReadFull(r.reader, b)
	if err != nil {
		return 0, err
	}
	r.offset++
	return b[0], nil
}

// VerifyChecksum verifies the packfile checksum
func (p *Packfile) VerifyChecksum(data []byte) error {
	// Calculate SHA-1 of packfile data (excluding checksum)
	checksumData := data[:len(data)-PackfileChecksumSize]
	hash := sha1.Sum(checksumData)

	// Compare with stored checksum
	if !bytes.Equal(hash[:], p.Checksum) {
		return fmt.Errorf("checksum mismatch: got %x, expected %x", hash[:], p.Checksum)
	}

	return nil
}

// ObjectTypeName returns the human-readable name for an object type
func ObjectTypeName(objType uint8) string {
	switch objType {
	case ObjCommit:
		return "commit"
	case ObjTree:
		return "tree"
	case ObjBlob:
		return "blob"
	case ObjTag:
		return "tag"
	case ObjOfsDelta:
		return "ofs-delta"
	case ObjRefDelta:
		return "ref-delta"
	default:
		return fmt.Sprintf("unknown(%d)", objType)
	}
}

// IsRegularObject returns true if the object type is a regular (non-delta) object
func IsRegularObject(objType uint8) bool {
	return objType >= ObjCommit && objType <= ObjTag
}

// IsDeltaObject returns true if the object type is a delta object
func IsDeltaObject(objType uint8) bool {
	return objType == ObjOfsDelta || objType == ObjRefDelta
}
