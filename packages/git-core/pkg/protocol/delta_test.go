package protocol

import (
	"bytes"
	"testing"
)

func TestReadDeltaSize(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		expectedSize uint64
	}{
		{
			name:         "small size",
			data:         []byte{0x05},
			expectedSize: 5,
		},
		{
			name:         "medium size",
			data:         []byte{0x80, 0x01}, // 128 + 0 = 128
			expectedSize: 128,
		},
		{
			name:         "large size",
			data:         []byte{0xFF, 0xFF, 0x03}, // 127 + (127 << 7) + (3 << 14)
			expectedSize: 127 + (127 << 7) + (3 << 14),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			size, err := readDeltaSize(reader)

			if err != nil {
				t.Errorf("readDeltaSize() unexpected error: %v", err)
				return
			}

			if size != tt.expectedSize {
				t.Errorf("readDeltaSize() = %d, want %d", size, tt.expectedSize)
			}
		})
	}
}

func TestParseCopyInstruction(t *testing.T) {
	tests := []struct {
		name           string
		opcode         byte
		data           []byte
		expectedOffset uint64
		expectedSize   uint64
	}{
		{
			name:           "copy with all offset and size bytes",
			opcode:         0x91, // 1001 0001 - offset byte 0, size byte 0
			data:           []byte{0x0A, 0x05},
			expectedOffset: 0x0A,
			expectedSize:   0x05,
		},
		{
			name:           "copy with multiple offset bytes",
			opcode:         0x93, // 1001 0011 - offset bytes 0 and 1, size byte 0
			data:           []byte{0x00, 0x01, 0x0A},
			expectedOffset: 0x0100,
			expectedSize:   0x0A,
		},
		{
			name:           "copy with size 0 (means 0x10000)",
			opcode:         0x81, // 1000 0001 - offset byte 0, no size bytes
			data:           []byte{0x0A},
			expectedOffset: 0x0A,
			expectedSize:   0x10000,
		},
		{
			name:           "copy with multiple size bytes",
			opcode:         0xB1, // 1011 0001 - offset byte 0, size bytes 0 and 1
			data:           []byte{0x10, 0x20, 0x30},
			expectedOffset: 0x10,
			expectedSize:   0x3020,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			instruction, err := parseCopyInstruction(reader, tt.opcode)

			if err != nil {
				t.Errorf("parseCopyInstruction() unexpected error: %v", err)
				return
			}

			if instruction.Offset != tt.expectedOffset {
				t.Errorf("parseCopyInstruction() offset = %d, want %d", instruction.Offset, tt.expectedOffset)
			}

			if instruction.Size != tt.expectedSize {
				t.Errorf("parseCopyInstruction() size = %d, want %d", instruction.Size, tt.expectedSize)
			}
		})
	}
}

func TestParseInsertInstruction(t *testing.T) {
	tests := []struct {
		name         string
		opcode       byte
		data         []byte
		expectedData []byte
	}{
		{
			name:         "insert 5 bytes",
			opcode:       0x05, // Insert 5 bytes
			data:         []byte{'h', 'e', 'l', 'l', 'o'},
			expectedData: []byte("hello"),
		},
		{
			name:         "insert 1 byte",
			opcode:       0x01,
			data:         []byte{'x'},
			expectedData: []byte("x"),
		},
		{
			name:         "insert 10 bytes",
			opcode:       0x0A,
			data:         []byte("0123456789"),
			expectedData: []byte("0123456789"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			instruction, err := parseInsertInstruction(reader, tt.opcode)

			if err != nil {
				t.Errorf("parseInsertInstruction() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(instruction.Data, tt.expectedData) {
				t.Errorf("parseInsertInstruction() data = %v, want %v", instruction.Data, tt.expectedData)
			}
		})
	}
}

func TestParseInsertInstructionErrors(t *testing.T) {
	tests := []struct {
		name   string
		opcode byte
		data   []byte
	}{
		{
			name:   "size is zero",
			opcode: 0x00,
			data:   []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			_, err := parseInsertInstruction(reader, tt.opcode)

			if err == nil {
				t.Error("parseInsertInstruction() expected error, got nil")
			}
		})
	}
}

func TestParseDelta(t *testing.T) {
	tests := []struct {
		name               string
		data               []byte
		expectedSourceSize uint64
		expectedTargetSize uint64
		expectedInstCount  int
	}{
		{
			name: "simple delta with insert",
			data: buildDeltaData(
				10, // source size
				15, // target size
				[]byte{0x05, 'h', 'e', 'l', 'l', 'o'}, // insert "hello"
			),
			expectedSourceSize: 10,
			expectedTargetSize: 15,
			expectedInstCount:  1,
		},
		{
			name: "delta with copy and insert",
			data: buildDeltaData(
				20,
				25,
				[]byte{
					0x91, 0x00, 0x05, // copy 5 bytes from offset 0
					0x03, 'x', 'y', 'z', // insert "xyz"
				},
			),
			expectedSourceSize: 20,
			expectedTargetSize: 25,
			expectedInstCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta, err := ParseDelta(tt.data)

			if err != nil {
				t.Errorf("ParseDelta() unexpected error: %v", err)
				return
			}

			if delta.SourceSize != tt.expectedSourceSize {
				t.Errorf("ParseDelta() source size = %d, want %d", delta.SourceSize, tt.expectedSourceSize)
			}

			if delta.TargetSize != tt.expectedTargetSize {
				t.Errorf("ParseDelta() target size = %d, want %d", delta.TargetSize, tt.expectedTargetSize)
			}

			if len(delta.Instructions) != tt.expectedInstCount {
				t.Errorf("ParseDelta() instruction count = %d, want %d", len(delta.Instructions), tt.expectedInstCount)
			}
		})
	}
}

func TestApplyDelta(t *testing.T) {
	tests := []struct {
		name         string
		base         []byte
		deltaData    []byte
		expectedData []byte
	}{
		{
			name: "simple insert",
			base: []byte("base"),
			deltaData: buildDeltaData(
				4,  // source size (len("base"))
				9,  // target size (len("basehello"))
				[]byte{
					0x91, 0x00, 0x04, // copy 4 bytes from offset 0 ("base")
					0x05, 'h', 'e', 'l', 'l', 'o', // insert "hello"
				},
			),
			expectedData: []byte("basehello"),
		},
		{
			name: "copy and rearrange",
			base: []byte("abcdef"),
			deltaData: buildDeltaData(
				6, // source size
				6, // target size
				[]byte{
					0x91, 0x03, 0x03, // copy 3 bytes from offset 3 ("def") - opcode 0x91 = offset byte 0, size byte 0
					0x91, 0x00, 0x03, // copy 3 bytes from offset 0 ("abc")
				},
			),
			expectedData: []byte("defabc"),
		},
		{
			name: "insert only",
			base: []byte(""),
			deltaData: buildDeltaData(
				0,
				5,
				[]byte{
					0x05, 'h', 'e', 'l', 'l', 'o',
				},
			),
			expectedData: []byte("hello"),
		},
		{
			name: "copy only",
			base: []byte("hello world"),
			deltaData: buildDeltaData(
				11,
				5,
				[]byte{
					0x91, 0x00, 0x05, // copy "hello"
				},
			),
			expectedData: []byte("hello"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta, err := ParseDelta(tt.deltaData)
			if err != nil {
				t.Errorf("ParseDelta() unexpected error: %v", err)
				return
			}

			result, err := ApplyDelta(tt.base, delta)
			if err != nil {
				t.Errorf("ApplyDelta() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(result, tt.expectedData) {
				t.Errorf("ApplyDelta() = %q, want %q", string(result), string(tt.expectedData))
			}
		})
	}
}

func TestApplyDeltaErrors(t *testing.T) {
	tests := []struct {
		name      string
		base      []byte
		deltaData []byte
	}{
		{
			name: "base size mismatch",
			base: []byte("abc"),
			deltaData: buildDeltaData(
				10, // Wrong source size
				5,
				[]byte{0x05, 'h', 'e', 'l', 'l', 'o'},
			),
		},
		{
			name: "copy out of bounds",
			base: []byte("abc"),
			deltaData: buildDeltaData(
				3,
				5,
				[]byte{
					0x92, 0x0A, 0x05, // Copy from offset 10 (out of bounds)
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta, err := ParseDelta(tt.deltaData)
			if err != nil {
				// Parsing error is also acceptable
				return
			}

			_, err = ApplyDelta(tt.base, delta)
			if err == nil {
				t.Error("ApplyDelta() expected error, got nil")
			}
		})
	}
}

func TestCopyInstructionApply(t *testing.T) {
	source := []byte("0123456789")
	var target bytes.Buffer

	instruction := &CopyInstruction{
		Offset: 2,
		Size:   5,
	}

	err := instruction.Apply(source, &target)
	if err != nil {
		t.Errorf("Apply() unexpected error: %v", err)
	}

	expected := "23456"
	if target.String() != expected {
		t.Errorf("Apply() = %q, want %q", target.String(), expected)
	}
}

func TestInsertInstructionApply(t *testing.T) {
	source := []byte("ignored")
	var target bytes.Buffer

	instruction := &InsertInstruction{
		Data: []byte("hello"),
	}

	err := instruction.Apply(source, &target)
	if err != nil {
		t.Errorf("Apply() unexpected error: %v", err)
	}

	expected := "hello"
	if target.String() != expected {
		t.Errorf("Apply() = %q, want %q", target.String(), expected)
	}
}

// Helper function to build delta data
func buildDeltaData(sourceSize, targetSize uint64, instructions []byte) []byte {
	var buf bytes.Buffer

	// Write source size
	writeDeltaSize(&buf, sourceSize)

	// Write target size
	writeDeltaSize(&buf, targetSize)

	// Write instructions
	buf.Write(instructions)

	return buf.Bytes()
}

func TestEncodeCopyInstruction(t *testing.T) {
	tests := []struct {
		name   string
		inst   *CopyInstruction
	}{
		{
			name: "small copy",
			inst: &CopyInstruction{
				Offset: 10,
				Size:   5,
			},
		},
		{
			name: "large offset",
			inst: &CopyInstruction{
				Offset: 0x1234,
				Size:   100,
			},
		},
		{
			name: "large size",
			inst: &CopyInstruction{
				Offset: 0,
				Size:   1000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := encodeCopyInstruction(&buf, tt.inst); err != nil {
				t.Errorf("encodeCopyInstruction() unexpected error: %v", err)
				return
			}

			// Read back and verify
			reader := bytes.NewReader(buf.Bytes())
			opcode, _ := reader.ReadByte()
			decoded, err := parseCopyInstruction(reader, opcode)
			if err != nil {
				t.Errorf("parseCopyInstruction() unexpected error: %v", err)
				return
			}

			if decoded.Offset != tt.inst.Offset {
				t.Errorf("offset = %d, want %d", decoded.Offset, tt.inst.Offset)
			}
			if decoded.Size != tt.inst.Size {
				t.Errorf("size = %d, want %d", decoded.Size, tt.inst.Size)
			}
		})
	}
}

func TestEncodeInsertInstruction(t *testing.T) {
	tests := []struct {
		name string
		inst *InsertInstruction
	}{
		{
			name: "small insert",
			inst: &InsertInstruction{
				Data: []byte("hello"),
			},
		},
		{
			name: "single byte",
			inst: &InsertInstruction{
				Data: []byte("x"),
			},
		},
		{
			name: "max size insert (127 bytes)",
			inst: &InsertInstruction{
				Data: make([]byte, 127),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := encodeInsertInstruction(&buf, tt.inst); err != nil {
				t.Errorf("encodeInsertInstruction() unexpected error: %v", err)
				return
			}

			// Read back and verify
			reader := bytes.NewReader(buf.Bytes())
			opcode, _ := reader.ReadByte()
			decoded, err := parseInsertInstruction(reader, opcode)
			if err != nil {
				t.Errorf("parseInsertInstruction() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(decoded.Data, tt.inst.Data) {
				t.Errorf("data mismatch")
			}
		})
	}
}

func TestEncodeInsertInstructionErrors(t *testing.T) {
	tests := []struct {
		name string
		inst *InsertInstruction
	}{
		{
			name: "empty data",
			inst: &InsertInstruction{
				Data: []byte{},
			},
		},
		{
			name: "data too large",
			inst: &InsertInstruction{
				Data: make([]byte, 128),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := encodeInsertInstruction(&buf, tt.inst)
			if err == nil {
				t.Error("encodeInsertInstruction() expected error, got nil")
			}
		})
	}
}

func TestEncodeDelta(t *testing.T) {
	tests := []struct {
		name  string
		delta *Delta
	}{
		{
			name: "simple delta",
			delta: &Delta{
				SourceSize: 10,
				TargetSize: 15,
				Instructions: []DeltaInstruction{
					&InsertInstruction{Data: []byte("hello")},
				},
			},
		},
		{
			name: "delta with copy and insert",
			delta: &Delta{
				SourceSize: 20,
				TargetSize: 25,
				Instructions: []DeltaInstruction{
					&CopyInstruction{Offset: 0, Size: 10},
					&InsertInstruction{Data: []byte("world")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeDelta(tt.delta)
			if err != nil {
				t.Errorf("EncodeDelta() unexpected error: %v", err)
				return
			}

			// Parse back
			decoded, err := ParseDelta(encoded)
			if err != nil {
				t.Errorf("ParseDelta() unexpected error: %v", err)
				return
			}

			if decoded.SourceSize != tt.delta.SourceSize {
				t.Errorf("source size = %d, want %d", decoded.SourceSize, tt.delta.SourceSize)
			}
			if decoded.TargetSize != tt.delta.TargetSize {
				t.Errorf("target size = %d, want %d", decoded.TargetSize, tt.delta.TargetSize)
			}
			if len(decoded.Instructions) != len(tt.delta.Instructions) {
				t.Errorf("instruction count = %d, want %d", len(decoded.Instructions), len(tt.delta.Instructions))
			}
		})
	}
}

func TestCreateDelta(t *testing.T) {
	tests := []struct {
		name   string
		source []byte
		target []byte
	}{
		{
			name:   "identical",
			source: []byte("hello world"),
			target: []byte("hello world"),
		},
		{
			name:   "append",
			source: []byte("hello"),
			target: []byte("hello world"),
		},
		{
			name:   "prepend",
			source: []byte("world"),
			target: []byte("hello world"),
		},
		{
			name:   "completely different",
			source: []byte("abc"),
			target: []byte("xyz"),
		},
		{
			name:   "empty source",
			source: []byte(""),
			target: []byte("hello"),
		},
		{
			name:   "empty target",
			source: []byte("hello"),
			target: []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := CreateDelta(tt.source, tt.target)

			if delta.SourceSize != uint64(len(tt.source)) {
				t.Errorf("source size = %d, want %d", delta.SourceSize, len(tt.source))
			}
			if delta.TargetSize != uint64(len(tt.target)) {
				t.Errorf("target size = %d, want %d", delta.TargetSize, len(tt.target))
			}

			// Apply delta and verify result
			result, err := ApplyDelta(tt.source, delta)
			if err != nil {
				t.Errorf("ApplyDelta() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(result, tt.target) {
				t.Errorf("ApplyDelta() = %q, want %q", string(result), string(tt.target))
			}
		})
	}
}

func TestCreateAndEncodeDelta(t *testing.T) {
	tests := []struct {
		name   string
		source []byte
		target []byte
	}{
		{
			name:   "simple modification",
			source: []byte("The quick brown fox"),
			target: []byte("The quick red fox"),
		},
		{
			name:   "large source reuse",
			source: []byte("abcdefghijklmnopqrstuvwxyz"),
			target: []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and encode delta
			deltaData, err := CreateAndEncodeDelta(tt.source, tt.target)
			if err != nil {
				t.Errorf("CreateAndEncodeDelta() unexpected error: %v", err)
				return
			}

			// Parse and apply
			delta, err := ParseDelta(deltaData)
			if err != nil {
				t.Errorf("ParseDelta() unexpected error: %v", err)
				return
			}

			result, err := ApplyDelta(tt.source, delta)
			if err != nil {
				t.Errorf("ApplyDelta() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(result, tt.target) {
				t.Errorf("result = %q, want %q", string(result), string(tt.target))
			}
		})
	}
}

func TestDeltaRoundTrip(t *testing.T) {
	// Test encoding and decoding a delta
	source := []byte("This is the original content with some text.")
	target := []byte("This is the modified content with some different text.")

	// Create delta
	delta := CreateDelta(source, target)

	// Encode delta
	encoded, err := EncodeDelta(delta)
	if err != nil {
		t.Fatalf("EncodeDelta() error: %v", err)
	}

	// Decode delta
	decoded, err := ParseDelta(encoded)
	if err != nil {
		t.Fatalf("ParseDelta() error: %v", err)
	}

	// Apply decoded delta
	result, err := ApplyDelta(source, decoded)
	if err != nil {
		t.Fatalf("ApplyDelta() error: %v", err)
	}

	// Verify result matches target
	if !bytes.Equal(result, target) {
		t.Errorf("Round trip failed:\ngot:  %q\nwant: %q", string(result), string(target))
	}
}
