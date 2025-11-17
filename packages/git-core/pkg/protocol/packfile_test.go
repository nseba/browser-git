package protocol

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

func TestReadPackfileHeader(t *testing.T) {
	tests := []struct {
		name             string
		data             []byte
		expectedVersion  uint32
		expectedCount    uint32
		expectError      bool
	}{
		{
			name:            "valid header",
			data:            buildPackfileHeader(2, 10),
			expectedVersion: 2,
			expectedCount:   10,
			expectError:     false,
		},
		{
			name:            "empty packfile",
			data:            buildPackfileHeader(2, 0),
			expectedVersion: 2,
			expectedCount:   0,
			expectError:     false,
		},
		{
			name:        "invalid signature",
			data:        []byte("BADX\x00\x00\x00\x02\x00\x00\x00\x0A"),
			expectError: true,
		},
		{
			name:        "unsupported version",
			data:        buildPackfileHeader(3, 5),
			expectError: true,
		},
		{
			name:        "truncated header",
			data:        []byte("PACK\x00\x00"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewPackfileReader(bytes.NewReader(tt.data))
			header, err := reader.ReadHeader()

			if tt.expectError {
				if err == nil {
					t.Errorf("ReadHeader() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ReadHeader() unexpected error: %v", err)
				return
			}

			if header.Signature != PackfileSignature {
				t.Errorf("ReadHeader() signature = %s, want %s", header.Signature, PackfileSignature)
			}

			if header.Version != tt.expectedVersion {
				t.Errorf("ReadHeader() version = %d, want %d", header.Version, tt.expectedVersion)
			}

			if header.ObjectCount != tt.expectedCount {
				t.Errorf("ReadHeader() object count = %d, want %d", header.ObjectCount, tt.expectedCount)
			}
		})
	}
}

func TestReadObjectHeader(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		expectedType uint8
		expectedSize uint64
	}{
		{
			name:         "small blob",
			data:         []byte{0x33}, // type=3 (blob), size=3
			expectedType: ObjBlob,
			expectedSize: 3,
		},
		{
			name:         "commit",
			data:         []byte{0x15}, // type=1 (commit), size=5
			expectedType: ObjCommit,
			expectedSize: 5,
		},
		{
			name:         "tree",
			data:         []byte{0x2F}, // type=2 (tree), size=15
			expectedType: ObjTree,
			expectedSize: 15,
		},
		{
			name: "large size (multi-byte)",
			data: []byte{
				0x9F,       // type=1, size=0x0F, MSB set (more bytes)
				0x01,       // size continuation: 0x01
			},
			expectedType: ObjCommit,
			expectedSize: 31, // 0x0F | (0x01 << 4) = 15 | 16 = 31
		},
		{
			name: "very large size",
			data: []byte{
				0x9F,       // type=1, size=0x0F, MSB set
				0x81,       // size continuation, MSB set
				0x02,       // size continuation
			},
			expectedType: ObjCommit,
			expectedSize: 4127, // 0x0F | (0x01 << 4) | (0x02 << 11) = 15 | 16 | 4096 = 4127
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewPackfileReader(bytes.NewReader(tt.data))
			objType, size, err := reader.readObjectHeader()

			if err != nil {
				t.Errorf("readObjectHeader() unexpected error: %v", err)
				return
			}

			if objType != tt.expectedType {
				t.Errorf("readObjectHeader() type = %d, want %d", objType, tt.expectedType)
			}

			if size != tt.expectedSize {
				t.Errorf("readObjectHeader() size = %d, want %d", size, tt.expectedSize)
			}
		})
	}
}

func TestReadCompressedData(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedData string
	}{
		{
			name:         "simple text",
			input:        "hello world",
			expectedData: "hello world",
		},
		{
			name:         "empty data",
			input:        "",
			expectedData: "",
		},
		{
			name:         "binary data",
			input:        "\x00\x01\x02\x03\x04",
			expectedData: "\x00\x01\x02\x03\x04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress the input data
			var compressedBuf bytes.Buffer
			zlibWriter := zlib.NewWriter(&compressedBuf)
			zlibWriter.Write([]byte(tt.input))
			zlibWriter.Close()

			// Read and decompress
			reader := NewPackfileReader(&compressedBuf)
			data, err := reader.readCompressedData()

			if err != nil {
				t.Errorf("readCompressedData() unexpected error: %v", err)
				return
			}

			if string(data) != tt.expectedData {
				t.Errorf("readCompressedData() = %q, want %q", string(data), tt.expectedData)
			}
		})
	}
}

func TestReadOffsetDeltaOffset(t *testing.T) {
	tests := []struct {
		name           string
		data           []byte
		expectedOffset int64
	}{
		{
			name:           "small offset",
			data:           []byte{0x05}, // offset = 5
			expectedOffset: 5,
		},
		{
			name:           "medium offset",
			data:           []byte{0x80, 0x01}, // offset = (0 + 1) << 7 | 1 = 129
			expectedOffset: 129,
		},
		{
			name: "large offset",
			data: []byte{
				0x81,       // MSB set, value = 1
				0x02,       // No MSB, value = 2
			},
			expectedOffset: 258, // ((1 + 1) << 7) | 2 = 256 + 2 = 258
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewPackfileReader(bytes.NewReader(tt.data))
			offset, err := reader.readOffsetDeltaOffset()

			if err != nil {
				t.Errorf("readOffsetDeltaOffset() unexpected error: %v", err)
				return
			}

			if offset != tt.expectedOffset {
				t.Errorf("readOffsetDeltaOffset() = %d, want %d", offset, tt.expectedOffset)
			}
		})
	}
}

func TestReadObject(t *testing.T) {
	tests := []struct {
		name         string
		buildData    func() []byte
		expectedType uint8
		expectedSize uint64
		isDelta      bool
	}{
		{
			name: "simple blob",
			buildData: func() []byte {
				// Type=3 (blob), size=5
				header := []byte{0x35} // 0011 0101
				data := compressData("hello")
				return append(header, data...)
			},
			expectedType: ObjBlob,
			expectedSize: 5,
			isDelta:      false,
		},
		{
			name: "commit object",
			buildData: func() []byte {
				// Type=1 (commit), size=10
				header := []byte{0x1A} // 0001 1010
				data := compressData("commit msg")
				return append(header, data...)
			},
			expectedType: ObjCommit,
			expectedSize: 10,
			isDelta:      false,
		},
		{
			name: "ref delta",
			buildData: func() []byte {
				// Type=7 (ref-delta), size=5
				header := []byte{0x75} // 0111 0101
				// 20-byte base hash
				baseHash := make([]byte, 20)
				for i := range baseHash {
					baseHash[i] = byte(i)
				}
				deltaData := compressData("delta")
				result := append(header, baseHash...)
				return append(result, deltaData...)
			},
			expectedType: ObjRefDelta,
			expectedSize: 5,
			isDelta:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.buildData()
			reader := NewPackfileReader(bytes.NewReader(data))
			obj, err := reader.ReadObject()

			if err != nil {
				t.Errorf("ReadObject() unexpected error: %v", err)
				return
			}

			if obj.Type != tt.expectedType {
				t.Errorf("ReadObject() type = %d, want %d", obj.Type, tt.expectedType)
			}

			if obj.Size != tt.expectedSize {
				t.Errorf("ReadObject() size = %d, want %d", obj.Size, tt.expectedSize)
			}

			if obj.IsDelta != tt.isDelta {
				t.Errorf("ReadObject() isDelta = %v, want %v", obj.IsDelta, tt.isDelta)
			}

			if tt.isDelta && tt.expectedType == ObjRefDelta {
				if len(obj.BaseHash) != 20 {
					t.Errorf("ReadObject() base hash length = %d, want 20", len(obj.BaseHash))
				}
			}
		})
	}
}

func TestObjectTypeName(t *testing.T) {
	tests := []struct {
		objType      uint8
		expectedName string
	}{
		{ObjCommit, "commit"},
		{ObjTree, "tree"},
		{ObjBlob, "blob"},
		{ObjTag, "tag"},
		{ObjOfsDelta, "ofs-delta"},
		{ObjRefDelta, "ref-delta"},
		{99, "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			name := ObjectTypeName(tt.objType)
			if name != tt.expectedName {
				t.Errorf("ObjectTypeName(%d) = %s, want %s", tt.objType, name, tt.expectedName)
			}
		})
	}
}

func TestIsRegularObject(t *testing.T) {
	tests := []struct {
		objType  uint8
		expected bool
	}{
		{ObjCommit, true},
		{ObjTree, true},
		{ObjBlob, true},
		{ObjTag, true},
		{ObjOfsDelta, false},
		{ObjRefDelta, false},
		{0, false},
		{99, false},
	}

	for _, tt := range tests {
		t.Run(ObjectTypeName(tt.objType), func(t *testing.T) {
			result := IsRegularObject(tt.objType)
			if result != tt.expected {
				t.Errorf("IsRegularObject(%d) = %v, want %v", tt.objType, result, tt.expected)
			}
		})
	}
}

func TestIsDeltaObject(t *testing.T) {
	tests := []struct {
		objType  uint8
		expected bool
	}{
		{ObjCommit, false},
		{ObjTree, false},
		{ObjBlob, false},
		{ObjTag, false},
		{ObjOfsDelta, true},
		{ObjRefDelta, true},
		{0, false},
		{99, false},
	}

	for _, tt := range tests {
		t.Run(ObjectTypeName(tt.objType), func(t *testing.T) {
			result := IsDeltaObject(tt.objType)
			if result != tt.expected {
				t.Errorf("IsDeltaObject(%d) = %v, want %v", tt.objType, result, tt.expected)
			}
		})
	}
}

// Helper function to build a packfile header
func buildPackfileHeader(version uint32, objectCount uint32) []byte {
	header := make([]byte, 12)
	copy(header[0:4], []byte("PACK"))
	binary.BigEndian.PutUint32(header[4:8], version)
	binary.BigEndian.PutUint32(header[8:12], objectCount)
	return header
}

// Helper function to compress data with zlib
func compressData(data string) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write([]byte(data))
	w.Close()
	return buf.Bytes()
}

func TestWritePackfileHeader(t *testing.T) {
	tests := []struct {
		name        string
		objectCount uint32
	}{
		{
			name:        "empty packfile",
			objectCount: 0,
		},
		{
			name:        "single object",
			objectCount: 1,
		},
		{
			name:        "multiple objects",
			objectCount: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewPackfileWriter(&buf)

			if err := writer.WriteHeader(tt.objectCount); err != nil {
				t.Errorf("WriteHeader() unexpected error: %v", err)
				return
			}

			// Flush internal buffer to writer
			writer.buf.WriteTo(&buf)

			// Read back the header
			reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
			header, err := reader.ReadHeader()
			if err != nil {
				t.Errorf("ReadHeader() unexpected error: %v", err)
				return
			}

			if header.ObjectCount != tt.objectCount {
				t.Errorf("object count = %d, want %d", header.ObjectCount, tt.objectCount)
			}
		})
	}
}

func TestWriteObjectHeader(t *testing.T) {
	tests := []struct {
		name     string
		objType  uint8
		size     uint64
	}{
		{
			name:    "small blob",
			objType: ObjBlob,
			size:    5,
		},
		{
			name:    "commit",
			objType: ObjCommit,
			size:    100,
		},
		{
			name:    "large tree",
			objType: ObjTree,
			size:    65536,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewPackfileWriter(&buf)

			if err := writer.writeObjectHeader(tt.objType, tt.size); err != nil {
				t.Errorf("writeObjectHeader() unexpected error: %v", err)
				return
			}

			// Flush internal buffer to writer
			writer.buf.WriteTo(&buf)

			// Read back the header
			reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
			objType, size, err := reader.readObjectHeader()
			if err != nil {
				t.Errorf("readObjectHeader() unexpected error: %v", err)
				return
			}

			if objType != tt.objType {
				t.Errorf("object type = %d, want %d", objType, tt.objType)
			}

			if size != tt.size {
				t.Errorf("size = %d, want %d", size, tt.size)
			}
		})
	}
}

func TestWriteCompressedData(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple text",
			data: "hello world",
		},
		{
			name: "empty data",
			data: "",
		},
		{
			name: "binary data",
			data: "\x00\x01\x02\x03\x04",
		},
		{
			name: "large data",
			data: string(make([]byte, 10000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewPackfileWriter(&buf)

			if err := writer.writeCompressedData([]byte(tt.data)); err != nil {
				t.Errorf("writeCompressedData() unexpected error: %v", err)
				return
			}

			// Flush internal buffer to writer
			writer.buf.WriteTo(&buf)

			// Read back the data
			reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
			data, err := reader.readCompressedData()
			if err != nil {
				t.Errorf("readCompressedData() unexpected error: %v", err)
				return
			}

			if string(data) != tt.data {
				t.Errorf("data mismatch: got %d bytes, want %d bytes", len(data), len(tt.data))
			}
		})
	}
}

func TestWriteObject(t *testing.T) {
	tests := []struct {
		name string
		obj  PackfileObject
	}{
		{
			name: "blob object",
			obj: PackfileObject{
				Type: ObjBlob,
				Size: 11,
				Data: []byte("hello world"),
			},
		},
		{
			name: "commit object",
			obj: PackfileObject{
				Type: ObjCommit,
				Size: 100,
				Data: []byte("tree abc123\nauthor John Doe\ncommitter John Doe\n\nCommit message"),
			},
		},
		{
			name: "ref delta",
			obj: PackfileObject{
				Type:     ObjRefDelta,
				Size:     10,
				Data:     []byte("delta data"),
				IsDelta:  true,
				BaseHash: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewPackfileWriter(&buf)

			if err := writer.WriteObject(&tt.obj); err != nil {
				t.Errorf("WriteObject() unexpected error: %v", err)
				return
			}

			// Flush internal buffer to writer
			writer.buf.WriteTo(&buf)

			// Read back the object
			reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
			obj, err := reader.ReadObject()
			if err != nil {
				t.Errorf("ReadObject() unexpected error: %v", err)
				return
			}

			if obj.Type != tt.obj.Type {
				t.Errorf("type = %d, want %d", obj.Type, tt.obj.Type)
			}

			if obj.Size != tt.obj.Size {
				t.Errorf("size = %d, want %d", obj.Size, tt.obj.Size)
			}

			if obj.IsDelta != tt.obj.IsDelta {
				t.Errorf("isDelta = %v, want %v", obj.IsDelta, tt.obj.IsDelta)
			}

			if tt.obj.IsDelta && tt.obj.Type == ObjRefDelta {
				if !bytes.Equal(obj.BaseHash, tt.obj.BaseHash) {
					t.Errorf("base hash mismatch")
				}
			}
		})
	}
}

func TestWritePackfile(t *testing.T) {
	tests := []struct {
		name    string
		objects []PackfileObject
	}{
		{
			name:    "empty packfile",
			objects: []PackfileObject{},
		},
		{
			name: "single blob",
			objects: []PackfileObject{
				{
					Type: ObjBlob,
					Size: 11,
					Data: []byte("hello world"),
				},
			},
		},
		{
			name: "multiple objects",
			objects: []PackfileObject{
				{
					Type: ObjBlob,
					Size: 5,
					Data: []byte("hello"),
				},
				{
					Type: ObjCommit,
					Size: 20,
					Data: []byte("tree abc\nauthor Bob"),
				},
				{
					Type: ObjTree,
					Size: 10,
					Data: []byte("tree data\n"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewPackfileWriter(&buf)

			if err := writer.WritePackfile(tt.objects); err != nil {
				t.Errorf("WritePackfile() unexpected error: %v", err)
				return
			}

			// Read back the packfile
			reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
			packfile, err := reader.ReadPackfile()
			if err != nil {
				t.Errorf("ReadPackfile() unexpected error: %v", err)
				return
			}

			if packfile.Header.ObjectCount != uint32(len(tt.objects)) {
				t.Errorf("object count = %d, want %d", packfile.Header.ObjectCount, len(tt.objects))
			}

			if len(packfile.Objects) != len(tt.objects) {
				t.Errorf("objects length = %d, want %d", len(packfile.Objects), len(tt.objects))
			}

			// Verify each object
			for i, expected := range tt.objects {
				actual := packfile.Objects[i]

				if actual.Type != expected.Type {
					t.Errorf("object[%d] type = %d, want %d", i, actual.Type, expected.Type)
				}

				if actual.Size != expected.Size {
					t.Errorf("object[%d] size = %d, want %d", i, actual.Size, expected.Size)
				}
			}
		})
	}
}

func TestPackfileRoundTrip(t *testing.T) {
	// Create a complex packfile with various object types
	objects := []PackfileObject{
		{
			Type: ObjBlob,
			Size: 26,
			Data: []byte("abcdefghijklmnopqrstuvwxyz"),
		},
		{
			Type: ObjTree,
			Size: 15,
			Data: []byte("100644 file.txt"),
		},
		{
			Type: ObjCommit,
			Size: 50,
			Data: []byte("tree abc123\nauthor Test <test@example.com>\n\nMsg"),
		},
		{
			Type:     ObjRefDelta,
			Size:     10,
			Data:     []byte("delta data"),
			IsDelta:  true,
			BaseHash: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
	}

	// Write packfile
	var buf bytes.Buffer
	writer := NewPackfileWriter(&buf)
	if err := writer.WritePackfile(objects); err != nil {
		t.Fatalf("WritePackfile() error: %v", err)
	}

	// Read packfile
	reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
	packfile, err := reader.ReadPackfile()
	if err != nil {
		t.Fatalf("ReadPackfile() error: %v", err)
	}

	// Verify
	if len(packfile.Objects) != len(objects) {
		t.Errorf("object count = %d, want %d", len(packfile.Objects), len(objects))
	}

	for i := range objects {
		if packfile.Objects[i].Type != objects[i].Type {
			t.Errorf("object[%d] type mismatch: got %d, want %d",
				i, packfile.Objects[i].Type, objects[i].Type)
		}
		if packfile.Objects[i].Size != objects[i].Size {
			t.Errorf("object[%d] size mismatch: got %d, want %d",
				i, packfile.Objects[i].Size, objects[i].Size)
		}
	}
}
