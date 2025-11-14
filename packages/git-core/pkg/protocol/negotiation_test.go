package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeNegotiationRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *NegotiationRequest
		expected string
	}{
		{
			name: "simple want with capabilities",
			req: &NegotiationRequest{
				Wants:        []string{"abc1234567890123456789012345678901234567"},
				Capabilities: []string{"multi_ack", "thin-pack"},
				Done:         true,
			},
			expected: "0046want abc1234567890123456789012345678901234567 multi_ack thin-pack\n" +
				"0000" +
				"0009done\n",
		},
		{
			name: "multiple wants",
			req: &NegotiationRequest{
				Wants: []string{
					"abc1234567890123456789012345678901234567",
					"def4567890123456789012345678901234567890",
				},
				Capabilities: []string{"side-band-64k"},
				Done:         true,
			},
			expected: "0040want abc1234567890123456789012345678901234567 side-band-64k\n" +
				"0032want def4567890123456789012345678901234567890\n" +
				"0000" +
				"0009done\n",
		},
		{
			name: "want with haves",
			req: &NegotiationRequest{
				Wants:        []string{"abc1234567890123456789012345678901234567"},
				Haves:        []string{"123abc4567890123456789012345678901234567"},
				Capabilities: []string{"multi_ack"},
				Done:         true,
			},
			expected: "003cwant abc1234567890123456789012345678901234567 multi_ack\n" +
				"0000" +
				"0032have 123abc4567890123456789012345678901234567\n" +
				"0009done\n",
		},
		{
			name: "shallow clone",
			req: &NegotiationRequest{
				Wants:        []string{"abc1234567890123456789012345678901234567"},
				Capabilities: []string{"shallow"},
				Deepen:       1,
				Done:         true,
			},
			expected: "003awant abc1234567890123456789012345678901234567 shallow\n" +
				"000ddeepen 1\n" +
				"0000" +
				"0009done\n",
		},
		{
			name: "incomplete negotiation",
			req: &NegotiationRequest{
				Wants:        []string{"abc1234567890123456789012345678901234567"},
				Haves:        []string{"123abc4567890123456789012345678901234567"},
				Capabilities: []string{"multi_ack"},
				Done:         false,
			},
			expected: "003cwant abc1234567890123456789012345678901234567 multi_ack\n" +
				"0000" +
				"0032have 123abc4567890123456789012345678901234567\n" +
				"0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encodeNegotiationRequest(tt.req)
			if err != nil {
				t.Errorf("encodeNegotiationRequest() error: %v", err)
				return
			}

			if string(result) != tt.expected {
				t.Errorf("encodeNegotiationRequest() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestParseACKLine(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		expectedHash string
		expectedStatus ACKStatus
		expectError  bool
	}{
		{
			name:         "simple ACK",
			line:         "ACK abc1234567890123456789012345678901234567",
			expectedHash: "abc1234567890123456789012345678901234567",
			expectedStatus: ACKSingle,
			expectError:  false,
		},
		{
			name:         "ACK with continue",
			line:         "ACK abc1234567890123456789012345678901234567 continue",
			expectedHash: "abc1234567890123456789012345678901234567",
			expectedStatus: ACKContinue,
			expectError:  false,
		},
		{
			name:         "ACK with common",
			line:         "ACK def4567890123456789012345678901234567890 common",
			expectedHash: "def4567890123456789012345678901234567890",
			expectedStatus: ACKCommon,
			expectError:  false,
		},
		{
			name:         "ACK with ready",
			line:         "ACK 123abc4567890123456789012345678901234567 ready",
			expectedHash: "123abc4567890123456789012345678901234567",
			expectedStatus: ACKReady,
			expectError:  false,
		},
		{
			name:        "invalid ACK line",
			line:        "ACK",
			expectError: true,
		},
		{
			name:        "unknown status",
			line:        "ACK abc1234567890123456789012345678901234567 unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ack, err := parseACKLine(tt.line)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseACKLine(%q) expected error, got nil", tt.line)
				}
				return
			}

			if err != nil {
				t.Errorf("parseACKLine(%q) unexpected error: %v", tt.line, err)
				return
			}

			if ack.Hash != tt.expectedHash {
				t.Errorf("parseACKLine(%q) hash = %q, want %q", tt.line, ack.Hash, tt.expectedHash)
			}

			if ack.Status != tt.expectedStatus {
				t.Errorf("parseACKLine(%q) status = %q, want %q", tt.line, ack.Status, tt.expectedStatus)
			}
		})
	}
}

func TestParseNegotiationResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      []byte
		done          bool
		sideBand      bool
		expectedACKs  int
		expectedNAK   bool
		expectedError string
	}{
		{
			name: "simple NAK response",
			response: buildNegotiationResponse(
				"NAK\n",
			),
			done:        false,
			sideBand:    false,
			expectedACKs: 0,
			expectedNAK: true,
		},
		{
			name: "single ACK",
			response: buildNegotiationResponse(
				"ACK abc1234567890123456789012345678901234567\n",
			),
			done:        false,
			sideBand:    false,
			expectedACKs: 1,
			expectedNAK: false,
		},
		{
			name: "multiple ACKs",
			response: buildNegotiationResponse(
				"ACK abc1234567890123456789012345678901234567 common\n",
				"ACK def4567890123456789012345678901234567890 continue\n",
				"NAK\n",
			),
			done:        false,
			sideBand:    false,
			expectedACKs: 2,
			expectedNAK: true,
		},
		{
			name: "error response",
			response: buildNegotiationResponse(
				"ERR repository not found\n",
			),
			done:          false,
			sideBand:      false,
			expectedACKs:  0,
			expectedError: "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseNegotiationResponse(bytes.NewReader(tt.response), tt.done, tt.sideBand)
			if err != nil {
				t.Errorf("parseNegotiationResponse() unexpected error: %v", err)
				return
			}

			if len(resp.ACKs) != tt.expectedACKs {
				t.Errorf("parseNegotiationResponse() ACKs count = %d, want %d", len(resp.ACKs), tt.expectedACKs)
			}

			if resp.NAK != tt.expectedNAK {
				t.Errorf("parseNegotiationResponse() NAK = %v, want %v", resp.NAK, tt.expectedNAK)
			}

			if tt.expectedError != "" && !strings.Contains(resp.ErrorMsg, tt.expectedError) {
				t.Errorf("parseNegotiationResponse() error = %q, want %q", resp.ErrorMsg, tt.expectedError)
			}
		})
	}
}

func TestParseSideBandResponse(t *testing.T) {
	tests := []struct {
		name             string
		response         []byte
		done             bool
		expectedPackfile string
		expectedError    string
	}{
		{
			name: "packfile data",
			response: buildSideBandResponse(
				[]sideBandLine{
					{channel: 1, data: []byte("PACK")},
					{channel: 1, data: []byte("file")},
					{channel: 2, data: []byte("Progress: 100%")},
				},
			),
			done:             true,
			expectedPackfile: "PACKfile",
		},
		{
			name: "error message",
			response: buildSideBandResponse(
				[]sideBandLine{
					{channel: 3, data: []byte("fatal: repository not found")},
				},
			),
			done:          false,
			expectedError: "repository not found",
		},
		{
			name: "mixed data and progress",
			response: buildSideBandResponse(
				[]sideBandLine{
					{channel: 2, data: []byte("Counting objects...")},
					{channel: 1, data: []byte("data1")},
					{channel: 2, data: []byte("Compressing...")},
					{channel: 1, data: []byte("data2")},
				},
			),
			done:             true,
			expectedPackfile: "data1data2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewPktLineReader(bytes.NewReader(tt.response))
			resp, err := parseSideBandResponse(reader, tt.done)

			if err != nil {
				t.Errorf("parseSideBandResponse() unexpected error: %v", err)
				return
			}

			if tt.expectedPackfile != "" && string(resp.Packfile) != tt.expectedPackfile {
				t.Errorf("parseSideBandResponse() packfile = %q, want %q", string(resp.Packfile), tt.expectedPackfile)
			}

			if tt.expectedError != "" && !strings.Contains(resp.ErrorMsg, tt.expectedError) {
				t.Errorf("parseSideBandResponse() error = %q, want to contain %q", resp.ErrorMsg, tt.expectedError)
			}
		})
	}
}

func TestHasSideBandCapability(t *testing.T) {
	tests := []struct {
		name         string
		capabilities []string
		expected     bool
	}{
		{
			name:         "has side-band",
			capabilities: []string{"multi_ack", "side-band", "thin-pack"},
			expected:     true,
		},
		{
			name:         "has side-band-64k",
			capabilities: []string{"multi_ack", "side-band-64k", "thin-pack"},
			expected:     true,
		},
		{
			name:         "no side-band",
			capabilities: []string{"multi_ack", "thin-pack"},
			expected:     false,
		},
		{
			name:         "empty capabilities",
			capabilities: []string{},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSideBandCapability(tt.capabilities)
			if result != tt.expected {
				t.Errorf("hasSideBandCapability() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildCapabilities(t *testing.T) {
	caps := BuildCapabilities()

	// Check that common capabilities are included
	requiredCaps := []string{"multi_ack_detailed", "side-band-64k", "thin-pack", "ofs-delta"}
	for _, req := range requiredCaps {
		found := false
		for _, cap := range caps {
			if strings.Contains(cap, req) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("BuildCapabilities() missing required capability: %s", req)
		}
	}
}

func TestBuildUploadPackURL(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "URL with .git",
			repoURL:  "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git/git-upload-pack",
		},
		{
			name:     "URL without .git",
			repoURL:  "https://github.com/user/repo",
			expected: "https://github.com/user/repo.git/git-upload-pack",
		},
		{
			name:     "URL with trailing slash",
			repoURL:  "https://example.com/repo/",
			expected: "https://example.com/repo/git-upload-pack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildUploadPackURL(tt.repoURL)
			if err != nil {
				t.Errorf("buildUploadPackURL() error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("buildUploadPackURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Helper function to build a mock negotiation response
func buildNegotiationResponse(lines ...string) []byte {
	var buf bytes.Buffer
	writer := NewPktLineWriter(&buf)

	for _, line := range lines {
		writer.WriteString(line)
	}
	writer.WriteFlush()

	return buf.Bytes()
}

// Helper type for side-band lines
type sideBandLine struct {
	channel byte
	data    []byte
}

// Helper function to build a side-band response
func buildSideBandResponse(lines []sideBandLine) []byte {
	var buf bytes.Buffer
	writer := NewPktLineWriter(&buf)

	for _, line := range lines {
		data := append([]byte{line.channel}, line.data...)
		writer.WriteLine(data)
	}
	writer.WriteFlush()

	return buf.Bytes()
}
