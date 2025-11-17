package protocol

import (
	"bytes"
	"testing"
)

func TestBuildReceivePackURL(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "URL with .git extension",
			repoURL:  "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git/git-receive-pack",
		},
		{
			name:     "URL without .git extension",
			repoURL:  "https://github.com/user/repo",
			expected: "https://github.com/user/repo.git/git-receive-pack",
		},
		{
			name:     "URL with trailing slash",
			repoURL:  "https://github.com/user/repo.git/",
			expected: "https://github.com/user/repo.git/git-receive-pack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildReceivePackURL(tt.repoURL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEncodePushRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *PushRequest
		validate func(t *testing.T, encoded []byte)
	}{
		{
			name: "Single ref update with capabilities",
			request: &PushRequest{
				Updates: []RefUpdate{
					{
						Name:    "refs/heads/main",
						OldHash: "0000000000000000000000000000000000000000",
						NewHash: "abcd1234abcd1234abcd1234abcd1234abcd1234",
					},
				},
				Capabilities: []string{"report-status", "side-band-64k"},
			},
			validate: func(t *testing.T, encoded []byte) {
				// Should contain the ref update with capabilities
				if !bytes.Contains(encoded, []byte("refs/heads/main")) {
					t.Error("encoded request should contain ref name")
				}
				if !bytes.Contains(encoded, []byte("report-status")) {
					t.Error("encoded request should contain capabilities")
				}
			},
		},
		{
			name: "Multiple ref updates",
			request: &PushRequest{
				Updates: []RefUpdate{
					{
						Name:    "refs/heads/main",
						OldHash: "1111111111111111111111111111111111111111",
						NewHash: "2222222222222222222222222222222222222222",
					},
					{
						Name:    "refs/heads/feature",
						OldHash: "3333333333333333333333333333333333333333",
						NewHash: "4444444444444444444444444444444444444444",
					},
				},
				Capabilities: []string{"report-status"},
			},
			validate: func(t *testing.T, encoded []byte) {
				if !bytes.Contains(encoded, []byte("refs/heads/main")) {
					t.Error("encoded request should contain main branch")
				}
				if !bytes.Contains(encoded, []byte("refs/heads/feature")) {
					t.Error("encoded request should contain feature branch")
				}
			},
		},
		{
			name: "Delete ref",
			request: &PushRequest{
				Updates: []RefUpdate{
					NewRefUpdateForDelete("refs/heads/old-branch", "abcd1234abcd1234abcd1234abcd1234abcd1234"),
				},
				Capabilities: []string{"report-status"},
			},
			validate: func(t *testing.T, encoded []byte) {
				if !bytes.Contains(encoded, []byte("refs/heads/old-branch")) {
					t.Error("encoded request should contain branch name")
				}
				// Should have 40 zeros for new hash (delete)
				if !bytes.Contains(encoded, []byte("0000000000000000000000000000000000000000")) {
					t.Error("encoded request should contain zero hash for delete")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encodePushRequest(tt.request)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, encoded)
		})
	}
}

func TestParsePushResponse(t *testing.T) {
	tests := []struct {
		name           string
		responseData   string
		reportStatus   bool
		sideBand       bool
		expectedUnpack string
		expectedRefs   int
		expectError    bool
	}{
		{
			name: "Successful push with status",
			responseData: "0030unpack ok\n" +
				"0019ok refs/heads/main\n" +
				"0000",
			reportStatus:   true,
			sideBand:       false,
			expectedUnpack: "ok",
			expectedRefs:   1,
			expectError:    false,
		},
		{
			name: "Push with failed ref update",
			responseData: "0030unpack ok\n" +
				"0033ng refs/heads/main non-fast-forward\n" +
				"0000",
			reportStatus:   true,
			sideBand:       false,
			expectedUnpack: "ok",
			expectedRefs:   1,
			expectError:    false,
		},
		{
			name: "Push with unpack failure",
			responseData: "003Dunpack error: disk full\n" +
				"0000",
			reportStatus:   true,
			sideBand:       false,
			expectedUnpack: "error: disk full",
			expectedRefs:   0,
			expectError:    false,
		},
		{
			name: "Multiple ref updates",
			responseData: "0030unpack ok\n" +
				"0019ok refs/heads/main\n" +
				"001Cok refs/heads/feature\n" +
				"0000",
			reportStatus:   true,
			sideBand:       false,
			expectedUnpack: "ok",
			expectedRefs:   2,
			expectError:    false,
		},
		{
			name:           "No status report requested",
			responseData:   "",
			reportStatus:   false,
			sideBand:       false,
			expectedUnpack: "ok",
			expectedRefs:   0,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.responseData))
			response, err := parsePushResponse(reader, tt.reportStatus, tt.sideBand)

			if tt.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if response.UnpackStatus != tt.expectedUnpack {
				t.Errorf("expected unpack status %s, got %s", tt.expectedUnpack, response.UnpackStatus)
			}

			if len(response.RefStatuses) != tt.expectedRefs {
				t.Errorf("expected %d ref statuses, got %d", tt.expectedRefs, len(response.RefStatuses))
			}
		})
	}
}

func TestNewRefUpdate(t *testing.T) {
	update := NewRefUpdate("refs/heads/main", "oldHash", "newHash")
	if update.Name != "refs/heads/main" {
		t.Errorf("expected name refs/heads/main, got %s", update.Name)
	}
	if update.OldHash != "oldHash" {
		t.Errorf("expected old hash oldHash, got %s", update.OldHash)
	}
	if update.NewHash != "newHash" {
		t.Errorf("expected new hash newHash, got %s", update.NewHash)
	}
}

func TestNewRefUpdateForNew(t *testing.T) {
	update := NewRefUpdateForNew("refs/heads/new-branch", "newHash")
	if update.Name != "refs/heads/new-branch" {
		t.Errorf("expected name refs/heads/new-branch, got %s", update.Name)
	}
	if update.OldHash != "0000000000000000000000000000000000000000" {
		t.Errorf("expected zero hash for old hash, got %s", update.OldHash)
	}
	if update.NewHash != "newHash" {
		t.Errorf("expected new hash newHash, got %s", update.NewHash)
	}
}

func TestNewRefUpdateForDelete(t *testing.T) {
	update := NewRefUpdateForDelete("refs/heads/old-branch", "oldHash")
	if update.Name != "refs/heads/old-branch" {
		t.Errorf("expected name refs/heads/old-branch, got %s", update.Name)
	}
	if update.OldHash != "oldHash" {
		t.Errorf("expected old hash oldHash, got %s", update.OldHash)
	}
	if update.NewHash != "0000000000000000000000000000000000000000" {
		t.Errorf("expected zero hash for new hash, got %s", update.NewHash)
	}
}

func TestBuildPushCapabilities(t *testing.T) {
	caps := BuildPushCapabilities()

	// Check that we have some capabilities
	if len(caps) == 0 {
		t.Error("expected capabilities, got none")
	}

	// Check for essential capabilities
	hasReportStatus := false
	for _, cap := range caps {
		if cap == "report-status" {
			hasReportStatus = true
			break
		}
	}
	if !hasReportStatus {
		t.Error("expected report-status capability")
	}
}
