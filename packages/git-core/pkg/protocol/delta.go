package protocol

import (
	"bytes"
	"fmt"
	"io"
)

// Delta represents a delta object that needs to be applied to a base
type Delta struct {
	SourceSize uint64 // Size of the base object
	TargetSize uint64 // Size of the resulting object
	Instructions []DeltaInstruction
}

// DeltaInstruction represents a single delta instruction
type DeltaInstruction interface {
	Apply(source []byte, target *bytes.Buffer) error
}

// CopyInstruction represents a copy operation from the source
type CopyInstruction struct {
	Offset uint64 // Offset in source to copy from
	Size   uint64 // Number of bytes to copy
}

// InsertInstruction represents an insert operation of new data
type InsertInstruction struct {
	Data []byte // Data to insert
}

// Apply applies a copy instruction
func (c *CopyInstruction) Apply(source []byte, target *bytes.Buffer) error {
	if c.Offset+c.Size > uint64(len(source)) {
		return fmt.Errorf("copy instruction out of bounds: offset=%d size=%d source_len=%d",
			c.Offset, c.Size, len(source))
	}

	data := source[c.Offset : c.Offset+c.Size]
	target.Write(data)
	return nil
}

// Apply applies an insert instruction
func (i *InsertInstruction) Apply(source []byte, target *bytes.Buffer) error {
	target.Write(i.Data)
	return nil
}

// ParseDelta parses delta instructions from compressed delta data
func ParseDelta(data []byte) (*Delta, error) {
	reader := bytes.NewReader(data)

	// Read source size
	sourceSize, err := readDeltaSize(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source size: %w", err)
	}

	// Read target size
	targetSize, err := readDeltaSize(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read target size: %w", err)
	}

	// Parse instructions
	instructions := []DeltaInstruction{}
	for {
		instruction, err := readDeltaInstruction(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read instruction: %w", err)
		}
		instructions = append(instructions, instruction)
	}

	return &Delta{
		SourceSize:   sourceSize,
		TargetSize:   targetSize,
		Instructions: instructions,
	}, nil
}

// ApplyDelta applies delta instructions to a base object
func ApplyDelta(base []byte, delta *Delta) ([]byte, error) {
	// Verify base size matches expected source size
	if uint64(len(base)) != delta.SourceSize {
		return nil, fmt.Errorf("base size mismatch: expected %d, got %d",
			delta.SourceSize, len(base))
	}

	// Apply all instructions
	var result bytes.Buffer
	result.Grow(int(delta.TargetSize))

	for i, instruction := range delta.Instructions {
		if err := instruction.Apply(base, &result); err != nil {
			return nil, fmt.Errorf("failed to apply instruction %d: %w", i, err)
		}
	}

	// Verify result size matches expected target size
	if uint64(result.Len()) != delta.TargetSize {
		return nil, fmt.Errorf("result size mismatch: expected %d, got %d",
			delta.TargetSize, result.Len())
	}

	return result.Bytes(), nil
}

// readDeltaSize reads a variable-length size encoding
func readDeltaSize(reader io.ByteReader) (uint64, error) {
	var size uint64
	var shift uint

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		size |= uint64(b&0x7F) << shift
		shift += 7

		// If MSB is not set, we're done
		if b&0x80 == 0 {
			break
		}
	}

	return size, nil
}

// readDeltaInstruction reads a single delta instruction
func readDeltaInstruction(reader io.ByteReader) (DeltaInstruction, error) {
	opcode, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Check if it's a copy instruction (MSB is set)
	if opcode&0x80 != 0 {
		return parseCopyInstruction(reader, opcode)
	}

	// Otherwise it's an insert instruction
	return parseInsertInstruction(reader, opcode)
}

// parseCopyInstruction parses a copy instruction
// Format: 1xxxxxxx [offset1] [offset2] [offset3] [offset4] [size1] [size2] [size3]
// The lower 7 bits of the opcode indicate which offset/size bytes are present
func parseCopyInstruction(reader io.ByteReader, opcode byte) (*CopyInstruction, error) {
	var offset uint64
	var size uint64

	// Read offset bytes (bits 0-3 of opcode indicate which bytes are present)
	for i := uint(0); i < 4; i++ {
		if opcode&(1<<i) != 0 {
			b, err := reader.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to read offset byte %d: %w", i, err)
			}
			offset |= uint64(b) << (i * 8)
		}
	}

	// Read size bytes (bits 4-6 of opcode indicate which bytes are present)
	for i := uint(0); i < 3; i++ {
		if opcode&(1<<(i+4)) != 0 {
			b, err := reader.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to read size byte %d: %w", i, err)
			}
			size |= uint64(b) << (i * 8)
		}
	}

	// Size of 0 means 0x10000 (65536)
	if size == 0 {
		size = 0x10000
	}

	return &CopyInstruction{
		Offset: offset,
		Size:   size,
	}, nil
}

// parseInsertInstruction parses an insert instruction
// Format: 0xxxxxxx [data...]
// The lower 7 bits indicate the number of bytes to insert
func parseInsertInstruction(reader io.ByteReader, opcode byte) (*InsertInstruction, error) {
	// The opcode itself is the size (lower 7 bits, since MSB is 0)
	size := int(opcode & 0x7F)

	if size == 0 {
		return nil, fmt.Errorf("invalid insert instruction: size is 0")
	}

	// Read the data to insert
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read insert data byte %d: %w", i, err)
		}
		data[i] = b
	}

	return &InsertInstruction{
		Data: data,
	}, nil
}

// ResolveDelta resolves a delta object given a function to retrieve base objects
type BaseObjectResolver func(hash string) ([]byte, error)

// ResolveOfsDelta resolves an offset delta
func ResolveOfsDelta(objects []PackfileObject, deltaIndex int) ([]byte, error) {
	delta := objects[deltaIndex]
	if delta.Type != ObjOfsDelta {
		return nil, fmt.Errorf("object is not an offset delta")
	}

	// Find base object by offset
	baseIndex := -1
	for i := 0; i < deltaIndex; i++ {
		if int64(i) == delta.Offset {
			baseIndex = i
			break
		}
	}

	if baseIndex == -1 {
		return nil, fmt.Errorf("base object not found at offset %d", delta.Offset)
	}

	// Get base object data (may need to recursively resolve if base is also a delta)
	var baseData []byte
	baseObj := objects[baseIndex]

	if baseObj.IsDelta {
		// Recursively resolve base delta
		resolved, err := ResolveOfsDelta(objects, baseIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve base delta: %w", err)
		}
		baseData = resolved
	} else {
		baseData = baseObj.Data
	}

	// Parse and apply delta
	parsedDelta, err := ParseDelta(delta.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse delta: %w", err)
	}

	result, err := ApplyDelta(baseData, parsedDelta)
	if err != nil {
		return nil, fmt.Errorf("failed to apply delta: %w", err)
	}

	return result, nil
}

// ResolveRefDelta resolves a reference delta
func ResolveRefDelta(deltaData []byte, baseData []byte) ([]byte, error) {
	// Parse and apply delta
	parsedDelta, err := ParseDelta(deltaData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse delta: %w", err)
	}

	result, err := ApplyDelta(baseData, parsedDelta)
	if err != nil {
		return nil, fmt.Errorf("failed to apply delta: %w", err)
	}

	return result, nil
}
