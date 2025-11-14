package protocol

import (
	"bytes"
	"io"
	"testing"
)

func TestEncodePktLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "0009hello",
		},
		{
			name:     "string with newline",
			input:    "test\n",
			expected: "0009test\n",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "0000", // Flush packet
		},
		{
			name:     "single character",
			input:    "a",
			expected: "0005a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodePktLineString(tt.input)
			if string(result) != tt.expected {
				t.Errorf("EncodePktLineString(%q) = %q, want %q", tt.input, string(result), tt.expected)
			}
		})
	}
}

func TestDecodePktLine(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedPayload string
		expectedRest    string
		expectError     bool
	}{
		{
			name:            "simple string",
			input:           "0009hello",
			expectedPayload: "hello",
			expectedRest:    "",
			expectError:     false,
		},
		{
			name:            "flush packet",
			input:           "0000remaining",
			expectedPayload: "",
			expectedRest:    "remaining",
			expectError:     false,
		},
		{
			name:            "multiple packets",
			input:           "0009hello0009world",
			expectedPayload: "hello",
			expectedRest:    "0009world",
			expectError:     false,
		},
		{
			name:         "invalid header",
			input:        "xxxxyyyy",
			expectError:  true,
		},
		{
			name:         "incomplete packet",
			input:        "000a",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, rest, err := DecodePktLine([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("DecodePktLine(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("DecodePktLine(%q) unexpected error: %v", tt.input, err)
				return
			}

			if string(payload) != tt.expectedPayload {
				t.Errorf("DecodePktLine(%q) payload = %q, want %q", tt.input, string(payload), tt.expectedPayload)
			}

			if string(rest) != tt.expectedRest {
				t.Errorf("DecodePktLine(%q) rest = %q, want %q", tt.input, string(rest), tt.expectedRest)
			}
		})
	}
}

func TestPktLineReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single line",
			input:    "0009hello0000",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "0006a\n0006b\n0006c\n0000",
			expected: []string{"a\n", "b\n", "c\n"},
		},
		{
			name:     "empty lines",
			input:    "00040000",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewPktLineReader(bytes.NewReader([]byte(tt.input)))
			lines, err := reader.ReadAll()

			if err != nil {
				t.Errorf("ReadAll() unexpected error: %v", err)
				return
			}

			if len(lines) != len(tt.expected) {
				t.Errorf("ReadAll() got %d lines, want %d", len(lines), len(tt.expected))
				return
			}

			for i, line := range lines {
				if string(line) != tt.expected[i] {
					t.Errorf("ReadAll() line %d = %q, want %q", i, string(line), tt.expected[i])
				}
			}
		})
	}
}

func TestPktLineWriter(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{
			name:     "single line",
			lines:    []string{"hello"},
			expected: "0009hello",
		},
		{
			name:     "multiple lines with flush",
			lines:    []string{"a", "b", ""},
			expected: "0005a0005b0000",
		},
		{
			name:     "line with newline",
			lines:    []string{"test\n"},
			expected: "0009test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewPktLineWriter(&buf)

			for _, line := range tt.lines {
				if err := writer.WriteString(line); err != nil {
					t.Errorf("WriteString(%q) unexpected error: %v", line, err)
					return
				}
			}

			if buf.String() != tt.expected {
				t.Errorf("PktLineWriter got %q, want %q", buf.String(), tt.expected)
			}
		})
	}
}

func TestPktLineRoundTrip(t *testing.T) {
	tests := [][]string{
		{"hello", "world"},
		{"line1\n", "line2\n", "line3\n"},
		{"a", "b", "c", "d", "e"},
		{"single"},
	}

	for _, tt := range tests {
		// Encode
		var buf bytes.Buffer
		writer := NewPktLineWriter(&buf)
		for _, line := range tt {
			if err := writer.WriteString(line); err != nil {
				t.Errorf("WriteString(%q) error: %v", line, err)
			}
		}
		writer.WriteFlush()

		// Decode
		reader := NewPktLineReader(&buf)
		lines, err := reader.ReadAll()
		if err != nil {
			t.Errorf("ReadAll() error: %v", err)
			continue
		}

		// Compare
		if len(lines) != len(tt) {
			t.Errorf("Round trip got %d lines, want %d", len(lines), len(tt))
			continue
		}

		for i, line := range lines {
			if string(line) != tt[i] {
				t.Errorf("Round trip line %d = %q, want %q", i, string(line), tt[i])
			}
		}
	}
}

func TestPktLineReaderEOF(t *testing.T) {
	// Test reading when there's no flush packet
	input := "0009hello"
	reader := NewPktLineReader(bytes.NewReader([]byte(input)))

	line, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("First ReadLine() error: %v", err)
	}
	if string(line) != "hello" {
		t.Errorf("ReadLine() = %q, want %q", string(line), "hello")
	}

	// Next read should return EOF
	_, err = reader.ReadLine()
	if err != io.EOF {
		t.Errorf("Second ReadLine() error = %v, want io.EOF", err)
	}
}

func TestFlushPacket(t *testing.T) {
	flush := EncodeFlushPkt()
	if string(flush) != "0000" {
		t.Errorf("EncodeFlushPkt() = %q, want %q", string(flush), "0000")
	}

	payload, rest, err := DecodePktLine(flush)
	if err != nil {
		t.Errorf("DecodePktLine(flush) error: %v", err)
	}
	if payload != nil {
		t.Errorf("DecodePktLine(flush) payload = %v, want nil", payload)
	}
	if len(rest) != 0 {
		t.Errorf("DecodePktLine(flush) rest = %v, want empty", rest)
	}
}

func TestIsFlushPkt(t *testing.T) {
	if !IsFlushPkt(nil) {
		t.Error("IsFlushPkt(nil) = false, want true")
	}
	if IsFlushPkt([]byte("hello")) {
		t.Error("IsFlushPkt(data) = true, want false")
	}
}
