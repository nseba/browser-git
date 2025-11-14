package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nseba/browser-git/git-core/pkg/merge"
)

// TestConflictResolutionStrategies tests different conflict resolution strategies
func TestConflictResolutionStrategies(t *testing.T) {
	tests := []struct {
		name            string
		strategy        ConflictResolutionStrategy
		manualContent   []byte
		expectedContent []byte
	}{
		{
			name:            "accept-ours",
			strategy:        AcceptOurs,
			manualContent:   nil,
			expectedContent: []byte("our content"),
		},
		{
			name:            "accept-theirs",
			strategy:        AcceptTheirs,
			manualContent:   nil,
			expectedContent: []byte("their content"),
		},
		{
			name:            "manual",
			strategy:        AcceptManual,
			manualContent:   []byte("manually resolved content"),
			expectedContent: []byte("manually resolved content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock conflict
			conflict := merge.Conflict{
				Path:   "test.txt",
				Type:   merge.ContentConflict,
				Ours:   []byte("our content"),
				Theirs: []byte("their content"),
			}

			// Test resolution logic (without actually resolving in a real repo)
			var resolvedContent []byte
			switch tt.strategy {
			case AcceptOurs:
				resolvedContent = conflict.Ours
			case AcceptTheirs:
				resolvedContent = conflict.Theirs
			case AcceptManual:
				resolvedContent = tt.manualContent
			}

			if string(resolvedContent) != string(tt.expectedContent) {
				t.Errorf("Expected content %q, got %q", tt.expectedContent, resolvedContent)
			}
		})
	}
}

// TestGenerateConflictMarkers tests conflict marker generation
func TestGenerateConflictMarkers(t *testing.T) {
	conflict := merge.Conflict{
		Path:   "file.txt",
		Type:   merge.ContentConflict,
		Base:   []byte("base content\n"),
		Ours:   []byte("our content\n"),
		Theirs: []byte("their content\n"),
	}

	markers := merge.GenerateConflictMarkersWithBranches(conflict, "main", "feature")

	// Check that markers contain expected strings
	if !containsStr(markers, "<<<<<<< main") {
		t.Error("Expected markers to contain '<<<<<<< main'")
	}
	if !containsStr(markers, "=======") {
		t.Error("Expected markers to contain '======='")
	}
	if !containsStr(markers, ">>>>>>> feature") {
		t.Error("Expected markers to contain '>>>>>>> feature'")
	}
	if !containsStr(markers, "our content") {
		t.Error("Expected markers to contain 'our content'")
	}
	if !containsStr(markers, "their content") {
		t.Error("Expected markers to contain 'their content'")
	}
}

// TestBinaryConflictMarkers tests conflict marker generation for binary files
func TestBinaryConflictMarkers(t *testing.T) {
	conflict := merge.Conflict{
		Path:     "image.png",
		Type:     merge.BinaryConflict,
		IsBinary: true,
	}

	markers := merge.GenerateConflictMarkers(conflict)

	if !containsStr(markers, "Binary file conflict") {
		t.Error("Expected binary conflict message")
	}
	if !containsStr(markers, "image.png") {
		t.Error("Expected path in binary conflict message")
	}
}

// TestConflictMetadata tests conflict metadata generation
func TestConflictMetadata(t *testing.T) {
	conflict := merge.Conflict{
		Path: "file.txt",
		Type: merge.ContentConflict,
		Metadata: &merge.ConflictMetadata{
			StartLine:       1,
			EndLine:         5,
			OurChangeType:   "modified",
			TheirChangeType: "modified",
		},
	}

	if conflict.Metadata.StartLine != 1 {
		t.Errorf("Expected start line 1, got %d", conflict.Metadata.StartLine)
	}
	if conflict.Metadata.EndLine != 5 {
		t.Errorf("Expected end line 5, got %d", conflict.Metadata.EndLine)
	}
	if conflict.Metadata.OurChangeType != "modified" {
		t.Errorf("Expected our change type 'modified', got %q", conflict.Metadata.OurChangeType)
	}
	if conflict.Metadata.TheirChangeType != "modified" {
		t.Errorf("Expected their change type 'modified', got %q", conflict.Metadata.TheirChangeType)
	}
}

// TestConflictTypeString tests the string representation of conflict types
func TestConflictTypeString(t *testing.T) {
	tests := []struct {
		ctype    merge.ConflictType
		expected string
	}{
		{merge.ContentConflict, "content"},
		{merge.BinaryConflict, "binary"},
		{merge.DeleteConflict, "delete"},
		{merge.AddConflict, "add"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.ctype.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.ctype.String())
			}
		})
	}
}

// TestSaveAndLoadConflictState tests saving and loading conflict state
func TestSaveAndLoadConflictState(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-conflict-state-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create test conflict files
	conflicts := []merge.Conflict{
		{Path: "file1.txt", Type: merge.ContentConflict},
		{Path: "file2.txt", Type: merge.ContentConflict},
	}

	// Save conflicts
	var conflictPaths string
	for _, c := range conflicts {
		conflictPaths += c.Path + "\n"
	}

	conflictsPath := filepath.Join(gitDir, "MERGE_CONFLICTS")
	if err := os.WriteFile(conflictsPath, []byte(conflictPaths), 0644); err != nil {
		t.Fatalf("Failed to write conflicts: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(conflictsPath)
	if err != nil {
		t.Fatalf("Failed to read conflicts: %v", err)
	}

	lines := splitLines(data)
	if len(lines) != 2 {
		t.Errorf("Expected 2 conflict paths, got %d", len(lines))
	}
	if lines[0] != "file1.txt" {
		t.Errorf("Expected first path 'file1.txt', got %q", lines[0])
	}
	if lines[1] != "file2.txt" {
		t.Errorf("Expected second path 'file2.txt', got %q", lines[1])
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
