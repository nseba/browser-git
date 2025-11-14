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

// Helper function to write variable-length size
func writeDeltaSize(buf *bytes.Buffer, size uint64) {
	for {
		b := byte(size & 0x7F)
		size >>= 7
		if size != 0 {
			b |= 0x80 // Set MSB to indicate more bytes
		}
		buf.WriteByte(b)
		if size == 0 {
			break
		}
	}
}
