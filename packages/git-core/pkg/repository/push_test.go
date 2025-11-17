package repository

import (
	"testing"

	"github.com/nseba/browser-git/git-core/pkg/hash"
	"github.com/nseba/browser-git/git-core/pkg/object"
)

func TestDefaultPushOptions(t *testing.T) {
	opts := DefaultPushOptions()

	if opts.Remote != "origin" {
		t.Errorf("expected default remote 'origin', got %s", opts.Remote)
	}

	if opts.Force {
		t.Error("expected force to be false by default")
	}

	if len(opts.RefSpecs) != 0 {
		t.Errorf("expected empty refspecs by default, got %d", len(opts.RefSpecs))
	}
}

func TestPushIsAncestor(t *testing.T) {
	// This test would require a full repository setup
	// For now, we'll test the basic logic

	// Create a temporary test repository
	tmpDir := t.TempDir()

	// Initialize repository
	initOpts := InitOptions{
		Bare:          false,
		InitialBranch: "main",
		HashAlgorithm: "sha1",
	}

	if err := Init(tmpDir, initOpts); err != nil {
		t.Fatalf("failed to init repository: %v", err)
	}

	// Open repository
	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Note: We can't fully test isAncestor without creating commits
	// This would require a more complete test setup
	// For now, we just verify the function exists and handles basic cases

	// Test with invalid hashes
	_, err = repo.isAncestor("invalid", "invalid")
	if err == nil {
		t.Error("expected error for invalid hash")
	}
}

func TestParseRefSpec(t *testing.T) {
	tests := []struct {
		name     string
		refspec  string
		expected struct {
			local  string
			remote string
		}
		expectError bool
	}{
		{
			name:    "simple branch name",
			refspec: "main",
			expected: struct {
				local  string
				remote string
			}{
				local:  "refs/heads/main",
				remote: "refs/heads/main",
			},
			expectError: false,
		},
		{
			name:    "full refspec",
			refspec: "refs/heads/main:refs/heads/main",
			expected: struct {
				local  string
				remote string
			}{
				local:  "refs/heads/main",
				remote: "refs/heads/main",
			},
			expectError: false,
		},
		{
			name:    "delete refspec",
			refspec: ":refs/heads/old-branch",
			expected: struct {
				local  string
				remote string
			}{
				local:  "",
				remote: "refs/heads/old-branch",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: parseAndBuildRefsToPush requires a repository instance
			// For a proper test, we'd need to create a test repository
			// This is a placeholder test structure
		})
	}
}

func TestHasNewOrDeletedRefs(t *testing.T) {
	zeroHash := "0000000000000000000000000000000000000000"
	normalHash := "abcd1234abcd1234abcd1234abcd1234abcd1234"

	tests := []struct {
		name     string
		refs     []refToPush
		expected bool
	}{
		{
			name: "new ref",
			refs: []refToPush{
				{
					oldHash: zeroHash,
					newHash: normalHash,
				},
			},
			expected: true,
		},
		{
			name: "deleted ref",
			refs: []refToPush{
				{
					oldHash: normalHash,
					newHash: zeroHash,
				},
			},
			expected: true,
		},
		{
			name: "updated ref",
			refs: []refToPush{
				{
					oldHash: "1111111111111111111111111111111111111111",
					newHash: normalHash,
				},
			},
			expected: false,
		},
		{
			name:     "no refs",
			refs:     []refToPush{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasNewOrDeletedRefs(tt.refs)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetRemoteURL(t *testing.T) {
	// Create a temporary test repository
	tmpDir := t.TempDir()

	// Initialize repository
	initOpts := InitOptions{
		Bare:          false,
		InitialBranch: "main",
		HashAlgorithm: "sha1",
	}

	if err := Init(tmpDir, initOpts); err != nil {
		t.Fatalf("failed to init repository: %v", err)
	}

	// Open repository
	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Set up a remote
	testURL := "https://github.com/user/repo.git"
	if err := setupRemote(repo, "origin", testURL); err != nil {
		t.Fatalf("failed to setup remote: %v", err)
	}

	// Test getting remote URL
	url, err := repo.GetRemoteURL("origin")
	if err != nil {
		t.Fatalf("failed to get remote URL: %v", err)
	}

	if url != testURL {
		t.Errorf("expected URL %s, got %s", testURL, url)
	}

	// Test getting non-existent remote
	_, err = repo.GetRemoteURL("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent remote")
	}
}

func TestCollectObjectsForCommits(t *testing.T) {
	// This test would require creating a full repository with commits
	// For now, we'll just verify the function exists and handles empty input

	// Create a temporary test repository
	tmpDir := t.TempDir()

	// Initialize repository
	initOpts := InitOptions{
		Bare:          false,
		InitialBranch: "main",
		HashAlgorithm: "sha1",
	}

	if err := Init(tmpDir, initOpts); err != nil {
		t.Fatalf("failed to init repository: %v", err)
	}

	// Open repository
	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Test with empty commits list
	objects, err := repo.collectObjectsForCommits([]hash.Hash{})
	if err != nil {
		t.Fatalf("failed to collect objects: %v", err)
	}

	if len(objects) != 0 {
		t.Errorf("expected 0 objects for empty commits list, got %d", len(objects))
	}
}

func TestCreatePackfileForPush(t *testing.T) {
	// Create a temporary test repository
	tmpDir := t.TempDir()

	// Initialize repository
	initOpts := InitOptions{
		Bare:          false,
		InitialBranch: "main",
		HashAlgorithm: "sha1",
	}

	if err := Init(tmpDir, initOpts); err != nil {
		t.Fatalf("failed to init repository: %v", err)
	}

	// Open repository
	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	// Test with empty objects list
	packfile, err := repo.createPackfileForPush([]object.Object{})
	if err != nil {
		t.Fatalf("failed to create packfile: %v", err)
	}

	// Packfile should have at least the header
	if len(packfile) == 0 {
		t.Error("expected non-empty packfile")
	}
}
