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

// CreateDelta creates a delta from source to target
// This uses a simple greedy algorithm to find matching blocks
func CreateDelta(source, target []byte) *Delta {
	instructions := []DeltaInstruction{}

	// Simple greedy algorithm: for each position in target,
	// try to find the longest match in source
	targetPos := 0
	minMatchLen := 4 // Minimum match length to be worth a copy instruction

	for targetPos < len(target) {
		bestMatchOffset := -1
		bestMatchLen := 0

		// Search for matches in source
		for sourcePos := 0; sourcePos < len(source); sourcePos++ {
			matchLen := 0
			for matchLen < len(target)-targetPos &&
				matchLen < len(source)-sourcePos &&
				source[sourcePos+matchLen] == target[targetPos+matchLen] {
				matchLen++
			}

			if matchLen > bestMatchLen {
				bestMatchLen = matchLen
				bestMatchOffset = sourcePos
			}
		}

		// If we found a good match, create a copy instruction
		if bestMatchLen >= minMatchLen {
			instructions = append(instructions, &CopyInstruction{
				Offset: uint64(bestMatchOffset),
				Size:   uint64(bestMatchLen),
			})
			targetPos += bestMatchLen
		} else {
			// No good match found, insert the byte(s)
			// Collect consecutive bytes that don't match
			insertStart := targetPos
			for targetPos < len(target) {
				// Check if there's a match at this position
				hasMatch := false
				for sourcePos := 0; sourcePos < len(source); sourcePos++ {
					matchLen := 0
					for matchLen < len(target)-targetPos &&
						matchLen < len(source)-sourcePos &&
						source[sourcePos+matchLen] == target[targetPos+matchLen] {
						matchLen++
					}
					if matchLen >= minMatchLen {
						hasMatch = true
						break
					}
				}

				if hasMatch {
					break
				}

				targetPos++

				// Git's delta format limits insert size to 127 bytes
				if targetPos-insertStart >= 127 {
					break
				}
			}

			if targetPos > insertStart {
				instructions = append(instructions, &InsertInstruction{
					Data: target[insertStart:targetPos],
				})
			}
		}
	}

	return &Delta{
		SourceSize:   uint64(len(source)),
		TargetSize:   uint64(len(target)),
		Instructions: instructions,
	}
}

// EncodeDelta encodes a delta into the Git delta format
func EncodeDelta(delta *Delta) ([]byte, error) {
	var buf bytes.Buffer

	// Write source size
	if err := writeDeltaSize(&buf, delta.SourceSize); err != nil {
		return nil, fmt.Errorf("failed to write source size: %w", err)
	}

	// Write target size
	if err := writeDeltaSize(&buf, delta.TargetSize); err != nil {
		return nil, fmt.Errorf("failed to write target size: %w", err)
	}

	// Write instructions
	for _, instruction := range delta.Instructions {
		switch inst := instruction.(type) {
		case *CopyInstruction:
			if err := encodeCopyInstruction(&buf, inst); err != nil {
				return nil, fmt.Errorf("failed to encode copy instruction: %w", err)
			}
		case *InsertInstruction:
			if err := encodeInsertInstruction(&buf, inst); err != nil {
				return nil, fmt.Errorf("failed to encode insert instruction: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown instruction type: %T", inst)
		}
	}

	return buf.Bytes(), nil
}

// writeDeltaSize writes a variable-length size encoding
func writeDeltaSize(buf *bytes.Buffer, size uint64) error {
	for {
		b := byte(size & 0x7F)
		size >>= 7

		if size > 0 {
			b |= 0x80 // Set MSB if more bytes follow
		}

		if err := buf.WriteByte(b); err != nil {
			return err
		}

		if size == 0 {
			break
		}
	}
	return nil
}

// encodeCopyInstruction encodes a copy instruction
// Format: 1xxxxxxx [offset1] [offset2] [offset3] [offset4] [size1] [size2] [size3]
func encodeCopyInstruction(buf *bytes.Buffer, inst *CopyInstruction) error {
	opcode := byte(0x80) // MSB set indicates copy

	// Encode offset (up to 4 bytes)
	offsetBytes := make([]byte, 4)
	offset := inst.Offset
	for i := 0; i < 4; i++ {
		if offset > 0 {
			offsetBytes[i] = byte(offset & 0xFF)
			opcode |= 1 << uint(i) // Set bit to indicate this byte is present
			offset >>= 8
		}
	}

	// Encode size (up to 3 bytes)
	// Special case: size of 0x10000 (65536) is encoded as 0
	sizeBytes := make([]byte, 3)
	size := inst.Size
	if size == 0x10000 {
		size = 0
	}
	for i := 0; i < 3; i++ {
		if size > 0 || (i == 0 && inst.Size > 0) {
			sizeBytes[i] = byte(size & 0xFF)
			if size > 0 {
				opcode |= 1 << uint(i+4) // Set bit to indicate this byte is present
			}
			size >>= 8
		}
	}

	// Write opcode
	if err := buf.WriteByte(opcode); err != nil {
		return err
	}

	// Write offset bytes (only those that are present)
	for i := 0; i < 4; i++ {
		if opcode&(1<<uint(i)) != 0 {
			if err := buf.WriteByte(offsetBytes[i]); err != nil {
				return err
			}
		}
	}

	// Write size bytes (only those that are present)
	for i := 0; i < 3; i++ {
		if opcode&(1<<uint(i+4)) != 0 {
			if err := buf.WriteByte(sizeBytes[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

// encodeInsertInstruction encodes an insert instruction
// Format: 0xxxxxxx [data...]
func encodeInsertInstruction(buf *bytes.Buffer, inst *InsertInstruction) error {
	size := len(inst.Data)

	if size == 0 {
		return fmt.Errorf("insert instruction cannot have zero size")
	}

	if size > 127 {
		return fmt.Errorf("insert instruction size %d exceeds maximum 127", size)
	}

	// Write opcode (size with MSB clear)
	opcode := byte(size & 0x7F)
	if err := buf.WriteByte(opcode); err != nil {
		return err
	}

	// Write data
	if _, err := buf.Write(inst.Data); err != nil {
		return err
	}

	return nil
}

// CreateAndEncodeDelta creates and encodes a delta in one step
func CreateAndEncodeDelta(source, target []byte) ([]byte, error) {
	delta := CreateDelta(source, target)
	return EncodeDelta(delta)
}
