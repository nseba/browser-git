package protocol

import (
	"bytes"
	"testing"
)

func TestBuildInfoRefsURL(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		service  ServiceType
		expected string
	}{
		{
			name:     "basic URL with .git",
			repoURL:  "https://github.com/user/repo.git",
			service:  UploadPackService,
			expected: "https://github.com/user/repo.git/info/refs?service=git-upload-pack",
		},
		{
			name:     "URL without .git",
			repoURL:  "https://github.com/user/repo",
			service:  UploadPackService,
			expected: "https://github.com/user/repo.git/info/refs?service=git-upload-pack",
		},
		{
			name:     "receive-pack service",
			repoURL:  "https://gitlab.com/user/project.git",
			service:  ReceivePackService,
			expected: "https://gitlab.com/user/project.git/info/refs?service=git-receive-pack",
		},
		{
			name:     "URL with trailing slash",
			repoURL:  "https://example.com/repo/",
			service:  UploadPackService,
			expected: "https://example.com/repo/info/refs?service=git-upload-pack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildInfoRefsURL(tt.repoURL, tt.service)
			if err != nil {
				t.Errorf("buildInfoRefsURL() error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("buildInfoRefsURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseRefLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectHash  string
		expectRef   string
		expectError bool
	}{
		{
			name:        "valid ref line",
			line:        "abc1234567890123456789012345678901234567 refs/heads/main",
			expectHash:  "abc1234567890123456789012345678901234567",
			expectRef:   "refs/heads/main",
			expectError: false,
		},
		{
			name:        "tag reference",
			line:        "def1234567890123456789012345678901234567 refs/tags/v1.0.0",
			expectHash:  "def1234567890123456789012345678901234567",
			expectRef:   "refs/tags/v1.0.0",
			expectError: false,
		},
		{
			name:        "HEAD reference",
			line:        "123abc4567890123456789012345678901234567 HEAD",
			expectHash:  "123abc4567890123456789012345678901234567",
			expectRef:   "HEAD",
			expectError: false,
		},
		{
			name:        "invalid format - no space",
			line:        "abcdefghij",
			expectError: true,
		},
		{
			name:        "invalid hash length",
			line:        "abc123 refs/heads/main",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, ref, err := parseRefLine(tt.line)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseRefLine(%q) expected error, got nil", tt.line)
				}
				return
			}

			if err != nil {
				t.Errorf("parseRefLine(%q) unexpected error: %v", tt.line, err)
				return
			}

			if hash != tt.expectHash {
				t.Errorf("parseRefLine(%q) hash = %q, want %q", tt.line, hash, tt.expectHash)
			}

			if ref != tt.expectRef {
				t.Errorf("parseRefLine(%q) ref = %q, want %q", tt.line, ref, tt.expectRef)
			}
		})
	}
}

func TestParseCapabilitiesLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectHash  string
		expectRef   string
		expectCaps  []string
		expectError bool
	}{
		{
			name:        "line with capabilities",
			line:        "abc1234567890123456789012345678901234567 HEAD\x00multi_ack thin-pack side-band",
			expectHash:  "abc1234567890123456789012345678901234567",
			expectRef:   "HEAD",
			expectCaps:  []string{"multi_ack", "thin-pack", "side-band"},
			expectError: false,
		},
		{
			name:        "line without capabilities",
			line:        "abc1234567890123456789012345678901234567 refs/heads/main",
			expectHash:  "abc1234567890123456789012345678901234567",
			expectRef:   "refs/heads/main",
			expectCaps:  nil,
			expectError: false,
		},
		{
			name:        "line with symref capability",
			line:        "abc1234567890123456789012345678901234567 HEAD\x00symref=HEAD:refs/heads/main multi_ack",
			expectHash:  "abc1234567890123456789012345678901234567",
			expectRef:   "HEAD",
			expectCaps:  []string{"symref=HEAD:refs/heads/main", "multi_ack"},
			expectError: false,
		},
		{
			name:        "line with agent capability",
			line:        "abc1234567890123456789012345678901234567 HEAD\x00agent=git/2.30.0 side-band-64k",
			expectHash:  "abc1234567890123456789012345678901234567",
			expectRef:   "HEAD",
			expectCaps:  []string{"agent=git/2.30.0", "side-band-64k"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, ref, caps, err := parseCapabilitiesLine(tt.line)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseCapabilitiesLine(%q) expected error, got nil", tt.line)
				}
				return
			}

			if err != nil {
				t.Errorf("parseCapabilitiesLine(%q) unexpected error: %v", tt.line, err)
				return
			}

			if hash != tt.expectHash {
				t.Errorf("parseCapabilitiesLine() hash = %q, want %q", hash, tt.expectHash)
			}

			if ref != tt.expectRef {
				t.Errorf("parseCapabilitiesLine() ref = %q, want %q", ref, tt.expectRef)
			}

			if len(caps) != len(tt.expectCaps) {
				t.Errorf("parseCapabilitiesLine() caps length = %d, want %d", len(caps), len(tt.expectCaps))
				return
			}

			for i, cap := range caps {
				if cap != tt.expectCaps[i] {
					t.Errorf("parseCapabilitiesLine() caps[%d] = %q, want %q", i, cap, tt.expectCaps[i])
				}
			}
		})
	}
}

func TestParseDiscoveryResponse(t *testing.T) {
	// Simulate a real Git server response
	response := buildMockDiscoveryResponse(
		"# service=git-upload-pack\n",
		"abc1234567890123456789012345678901234567 HEAD\x00multi_ack thin-pack side-band side-band-64k ofs-delta symref=HEAD:refs/heads/main agent=git/2.30.0\n",
		"abc1234567890123456789012345678901234567 refs/heads/main\n",
		"def4567890123456789012345678901234567890 refs/heads/feature\n",
		"1234567890abcdef1234567890abcdef12345678 refs/tags/v1.0.0\n",
	)

	discovery, err := parseDiscoveryResponse(bytes.NewReader(response), UploadPackService)
	if err != nil {
		t.Fatalf("parseDiscoveryResponse() error: %v", err)
	}

	// Check service
	if discovery.Service != UploadPackService {
		t.Errorf("Service = %v, want %v", discovery.Service, UploadPackService)
	}

	// Check capabilities
	expectedCaps := []string{"multi_ack", "thin-pack", "side-band", "side-band-64k", "ofs-delta", "symref=HEAD:refs/heads/main", "agent=git/2.30.0"}
	if len(discovery.Capabilities) != len(expectedCaps) {
		t.Errorf("Capabilities length = %d, want %d", len(discovery.Capabilities), len(expectedCaps))
	}

	for i, cap := range discovery.Capabilities {
		if cap != expectedCaps[i] {
			t.Errorf("Capabilities[%d] = %q, want %q", i, cap, expectedCaps[i])
		}
	}

	// Check references
	if len(discovery.References) != 4 {
		t.Errorf("References length = %d, want 4", len(discovery.References))
	}

	// Check symrefs
	if target, ok := discovery.SymRefs["HEAD"]; !ok || target != "refs/heads/main" {
		t.Errorf("SymRefs[HEAD] = %q, want refs/heads/main", target)
	}

	// Check default branch
	defaultBranch, err := discovery.GetDefaultBranch()
	if err != nil {
		t.Errorf("GetDefaultBranch() error: %v", err)
	}
	if defaultBranch != "refs/heads/main" {
		t.Errorf("GetDefaultBranch() = %q, want refs/heads/main", defaultBranch)
	}

	// Check HasCapability
	if !discovery.HasCapability("multi_ack") {
		t.Error("HasCapability(multi_ack) = false, want true")
	}
	if discovery.HasCapability("nonexistent") {
		t.Error("HasCapability(nonexistent) = true, want false")
	}

	// Check GetReference
	if ref, ok := discovery.GetReference("refs/heads/feature"); !ok {
		t.Error("GetReference(refs/heads/feature) not found")
	} else if ref.Hash != "def4567890123456789012345678901234567890" {
		t.Errorf("GetReference(refs/heads/feature).Hash = %q, want def4567890123456789012345678901234567890", ref.Hash)
	}
}

func TestParseDiscoveryResponseMinimal(t *testing.T) {
	// Test with minimal response (no capabilities, single ref)
	response := buildMockDiscoveryResponse(
		"# service=git-upload-pack\n",
		"abc1234567890123456789012345678901234567 refs/heads/main\n",
	)

	discovery, err := parseDiscoveryResponse(bytes.NewReader(response), UploadPackService)
	if err != nil {
		t.Fatalf("parseDiscoveryResponse() error: %v", err)
	}

	if len(discovery.References) != 1 {
		t.Errorf("References length = %d, want 1", len(discovery.References))
	}

	if len(discovery.Capabilities) != 0 {
		t.Errorf("Capabilities length = %d, want 0", len(discovery.Capabilities))
	}
}

func TestParseDiscoveryResponseErrors(t *testing.T) {
	tests := []struct {
		name     string
		response []byte
	}{
		{
			name:     "wrong service",
			response: buildMockDiscoveryResponse("# service=wrong-service\n"),
		},
		{
			name:     "missing flush after service",
			response: encodePktLines("# service=git-upload-pack\n", "abc1234567890123456789012345678901234567 refs/heads/main\n"),
		},
		{
			name:     "invalid ref line",
			response: buildMockDiscoveryResponse("# service=git-upload-pack\n", "invalid line\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDiscoveryResponse(bytes.NewReader(tt.response), UploadPackService)
			if err == nil {
				t.Error("parseDiscoveryResponse() expected error, got nil")
			}
		})
	}
}

// Helper function to build a mock discovery response
func buildMockDiscoveryResponse(lines ...string) []byte {
	var buf bytes.Buffer
	writer := NewPktLineWriter(&buf)

	// Write service line
	writer.WriteString(lines[0])
	writer.WriteFlush()

	// Write ref lines
	for i := 1; i < len(lines); i++ {
		writer.WriteString(lines[i])
	}
	writer.WriteFlush()

	return buf.Bytes()
}

// Helper function to encode lines without flush
func encodePktLines(lines ...string) []byte {
	var buf bytes.Buffer
	writer := NewPktLineWriter(&buf)

	for _, line := range lines {
		writer.WriteString(line)
	}

	return buf.Bytes()
}
