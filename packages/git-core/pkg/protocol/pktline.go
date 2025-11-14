package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Pkt-line format constants
const (
	// PktLineMaxLength is the maximum length of a pkt-line (65536 bytes)
	PktLineMaxLength = 65536

	// PktLineHeaderLength is the length of the pkt-line header (4 bytes)
	PktLineHeaderLength = 4

	// FlushPkt is the special flush packet marker
	FlushPkt = "0000"

	// DelimiterPkt is the delimiter packet marker (protocol v2)
	DelimiterPkt = "0001"

	// ResponseEndPkt is the response end packet marker (protocol v2)
	ResponseEndPkt = "0002"
)

// PktLineReader reads pkt-line formatted data
type PktLineReader struct {
	reader *bufio.Reader
}

// NewPktLineReader creates a new pkt-line reader
func NewPktLineReader(r io.Reader) *PktLineReader {
	return &PktLineReader{
		reader: bufio.NewReader(r),
	}
}

// ReadLine reads a single pkt-line
// Returns nil for flush packet, io.EOF at end of stream
func (r *PktLineReader) ReadLine() ([]byte, error) {
	// Read the 4-byte length header
	header := make([]byte, PktLineHeaderLength)
	_, err := io.ReadFull(r.reader, header)
	if err != nil {
		return nil, err
	}

	// Parse the hex length
	lengthStr := string(header)
	length, err := strconv.ParseInt(lengthStr, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid pkt-line header: %s", lengthStr)
	}

	// Special packets
	if length == 0 {
		return nil, nil // Flush packet
	}
	if length == 1 {
		return []byte{0x01}, nil // Delimiter packet
	}
	if length == 2 {
		return []byte{0x02}, nil // Response end packet
	}

	// Validate length
	if length < PktLineHeaderLength || length > PktLineMaxLength {
		return nil, fmt.Errorf("invalid pkt-line length: %d", length)
	}

	// Read the payload (length includes the 4-byte header)
	payloadLength := int(length) - PktLineHeaderLength
	payload := make([]byte, payloadLength)
	_, err = io.ReadFull(r.reader, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// ReadAll reads all pkt-lines until a flush packet
func (r *PktLineReader) ReadAll() ([][]byte, error) {
	var lines [][]byte

	for {
		line, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Flush packet signals end
		if line == nil {
			break
		}

		lines = append(lines, line)
	}

	return lines, nil
}

// PktLineWriter writes pkt-line formatted data
type PktLineWriter struct {
	writer io.Writer
}

// NewPktLineWriter creates a new pkt-line writer
func NewPktLineWriter(w io.Writer) *PktLineWriter {
	return &PktLineWriter{writer: w}
}

// WriteLine writes a single pkt-line
func (w *PktLineWriter) WriteLine(data []byte) error {
	if len(data) == 0 {
		return w.WriteFlush()
	}

	// Calculate total length (header + payload)
	length := PktLineHeaderLength + len(data)
	if length > PktLineMaxLength {
		return fmt.Errorf("pkt-line too long: %d bytes", length)
	}

	// Format the length as 4-character hex
	header := fmt.Sprintf("%04x", length)

	// Write header and payload
	if _, err := w.writer.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	return nil
}

// WriteString writes a string as a pkt-line
func (w *PktLineWriter) WriteString(s string) error {
	return w.WriteLine([]byte(s))
}

// WriteFlush writes a flush packet (0000)
func (w *PktLineWriter) WriteFlush() error {
	_, err := w.writer.Write([]byte(FlushPkt))
	return err
}

// WriteDelimiter writes a delimiter packet (0001)
func (w *PktLineWriter) WriteDelimiter() error {
	_, err := w.writer.Write([]byte(DelimiterPkt))
	return err
}

// WriteResponseEnd writes a response end packet (0002)
func (w *PktLineWriter) WriteResponseEnd() error {
	_, err := w.writer.Write([]byte(ResponseEndPkt))
	return err
}

// EncodePktLine encodes a single pkt-line
func EncodePktLine(data []byte) []byte {
	var buf bytes.Buffer
	w := NewPktLineWriter(&buf)
	_ = w.WriteLine(data) // Error handling in production code
	return buf.Bytes()
}

// EncodePktLineString encodes a string as a pkt-line
func EncodePktLineString(s string) []byte {
	return EncodePktLine([]byte(s))
}

// EncodeFlushPkt returns a flush packet
func EncodeFlushPkt() []byte {
	return []byte(FlushPkt)
}

// DecodePktLine decodes a single pkt-line from a byte slice
func DecodePktLine(data []byte) (payload []byte, rest []byte, err error) {
	if len(data) < PktLineHeaderLength {
		return nil, data, fmt.Errorf("insufficient data for pkt-line header")
	}

	// Parse length
	lengthStr := string(data[:PktLineHeaderLength])
	length, err := strconv.ParseInt(lengthStr, 16, 32)
	if err != nil {
		return nil, data, fmt.Errorf("invalid pkt-line header: %s", lengthStr)
	}

	// Special packets
	if length == 0 {
		return nil, data[PktLineHeaderLength:], nil // Flush packet
	}

	// Validate length
	if int(length) > len(data) {
		return nil, data, fmt.Errorf("incomplete pkt-line: need %d bytes, have %d", length, len(data))
	}

	// Extract payload
	payload = data[PktLineHeaderLength:length]
	rest = data[length:]

	return payload, rest, nil
}

// DecodePktLines decodes multiple pkt-lines until a flush packet
func DecodePktLines(data []byte) ([][]byte, error) {
	var lines [][]byte
	remaining := data

	for len(remaining) > 0 {
		payload, rest, err := DecodePktLine(remaining)
		if err != nil {
			return nil, err
		}

		// Flush packet signals end
		if payload == nil {
			break
		}

		lines = append(lines, payload)
		remaining = rest
	}

	return lines, nil
}

// IsFlushPkt checks if a line is a flush packet
func IsFlushPkt(line []byte) bool {
	return line == nil
}

// IsDelimiterPkt checks if a line is a delimiter packet
func IsDelimiterPkt(line []byte) bool {
	return len(line) == 1 && line[0] == 0x01
}

// IsResponseEndPkt checks if a line is a response end packet
func IsResponseEndPkt(line []byte) bool {
	return len(line) == 1 && line[0] == 0x02
}
