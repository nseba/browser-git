package repository

import (
	"testing"
)

// TestFetchOptions tests default fetch options
func TestFetchOptions(t *testing.T) {
	opts := DefaultFetchOptions()

	if opts.Remote != "origin" {
		t.Errorf("expected remote 'origin', got '%s'", opts.Remote)
	}

	if opts.Prune {
		t.Error("expected prune to be false by default")
	}

	if opts.Force {
		t.Error("expected force to be false by default")
	}

	if opts.Depth != 0 {
		t.Errorf("expected depth 0, got %d", opts.Depth)
	}
}

// TestPullOptions tests default pull options
func TestPullOptions(t *testing.T) {
	opts := DefaultPullOptions()

	if opts.Remote != "origin" {
		t.Errorf("expected remote 'origin', got '%s'", opts.Remote)
	}

	if opts.Branch != "" {
		t.Errorf("expected empty branch, got '%s'", opts.Branch)
	}

	if opts.Rebase {
		t.Error("expected rebase to be false by default")
	}

	if opts.FastForwardOnly {
		t.Error("expected fastForwardOnly to be false by default")
	}
}

// TestParseRefSpec tests refspec parsing
func TestParseRefSpec(t *testing.T) {
	tests := []struct {
		refspec string
		src     string
		dst     string
		force   bool
	}{
		{
			refspec: "refs/heads/main:refs/remotes/origin/main",
			src:     "refs/heads/main",
			dst:     "refs/remotes/origin/main",
			force:   false,
		},
		{
			refspec: "+refs/heads/main:refs/remotes/origin/main",
			src:     "refs/heads/main",
			dst:     "refs/remotes/origin/main",
			force:   true,
		},
		{
			refspec: "refs/heads/*:refs/remotes/origin/*",
			src:     "refs/heads/*",
			dst:     "refs/remotes/origin/*",
			force:   false,
		},
		{
			refspec: "+refs/heads/*:refs/remotes/origin/*",
			src:     "refs/heads/*",
			dst:     "refs/remotes/origin/*",
			force:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.refspec, func(t *testing.T) {
			src, dst, force := parseRefSpec(tt.refspec)

			if src != tt.src {
				t.Errorf("expected src '%s', got '%s'", tt.src, src)
			}

			if dst != tt.dst {
				t.Errorf("expected dst '%s', got '%s'", tt.dst, dst)
			}

			if force != tt.force {
				t.Errorf("expected force %v, got %v", tt.force, force)
			}
		})
	}
}

// TestMatchesPattern tests pattern matching
func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		refName string
		pattern string
		matches bool
	}{
		{
			refName: "refs/heads/main",
			pattern: "refs/heads/*",
			matches: true,
		},
		{
			refName: "refs/heads/feature/test",
			pattern: "refs/heads/*",
			matches: true,
		},
		{
			refName: "refs/tags/v1.0",
			pattern: "refs/heads/*",
			matches: false,
		},
		{
			refName: "refs/heads/main",
			pattern: "refs/heads/main",
			matches: true,
		},
		{
			refName: "refs/heads/develop",
			pattern: "refs/heads/main",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.refName+" vs "+tt.pattern, func(t *testing.T) {
			matches := matchesPattern(tt.refName, tt.pattern)

			if matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

// TestCalculateDestRef tests destination ref calculation
func TestCalculateDestRef(t *testing.T) {
	tests := []struct {
		refName    string
		srcPattern string
		dstPattern string
		remote     string
		expected   string
	}{
		{
			refName:    "refs/heads/main",
			srcPattern: "refs/heads/*",
			dstPattern: "refs/remotes/origin/*",
			remote:     "origin",
			expected:   "refs/remotes/origin/main",
		},
		{
			refName:    "refs/heads/feature/test",
			srcPattern: "refs/heads/*",
			dstPattern: "refs/remotes/origin/*",
			remote:     "origin",
			expected:   "refs/remotes/origin/feature/test",
		},
		{
			refName:    "refs/heads/main",
			srcPattern: "refs/heads/main",
			dstPattern: "refs/remotes/origin/main",
			remote:     "origin",
			expected:   "refs/remotes/origin/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.refName, func(t *testing.T) {
			result := calculateDestRef(tt.refName, tt.srcPattern, tt.dstPattern, tt.remote)

			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestRefUpdate tests the RefUpdate structure
func TestRefUpdate(t *testing.T) {
	update := RefUpdate{
		RefName: "refs/remotes/origin/main",
		OldHash: "abc123",
		NewHash: "def456",
		Forced:  false,
	}

	if update.RefName != "refs/remotes/origin/main" {
		t.Errorf("unexpected ref name: %s", update.RefName)
	}

	if update.OldHash != "abc123" {
		t.Errorf("unexpected old hash: %s", update.OldHash)
	}

	if update.NewHash != "def456" {
		t.Errorf("unexpected new hash: %s", update.NewHash)
	}

	if update.Forced {
		t.Error("expected forced to be false")
	}
}

// Note: Integration tests that actually fetch from a remote would go in
// packages/git-core/pkg/protocol/protocol_integration_test.go
