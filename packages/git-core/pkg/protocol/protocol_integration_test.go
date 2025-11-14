package protocol

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// TestPackfileIntegration tests the complete packfile workflow
func TestPackfileIntegration(t *testing.T) {
	t.Run("write and read packfile with mixed object types", func(t *testing.T) {
		// Create realistic Git objects
		blob1 := []byte("console.log('Hello, World!');\n")
		blob2 := []byte("# README\n\nThis is a test project.\n")
		tree := []byte("100644 index.js\x00" + strings.Repeat("\x00", 20) +
			"100644 README.md\x00" + strings.Repeat("\x00", 20))
		commit := []byte("tree " + strings.Repeat("0", 40) + "\n" +
			"author Test User <test@example.com> 1234567890 +0000\n" +
			"committer Test User <test@example.com> 1234567890 +0000\n\n" +
			"Initial commit\n")

		// Create packfile with these objects
		objects := []PackfileObject{
			{Type: ObjBlob, Size: uint64(len(blob1)), Data: blob1},
			{Type: ObjBlob, Size: uint64(len(blob2)), Data: blob2},
			{Type: ObjTree, Size: uint64(len(tree)), Data: tree},
			{Type: ObjCommit, Size: uint64(len(commit)), Data: commit},
		}

		// Write packfile
		var buf bytes.Buffer
		writer := NewPackfileWriter(&buf)
		if err := writer.WritePackfile(objects); err != nil {
			t.Fatalf("WritePackfile() error: %v", err)
		}

		// Verify packfile was written
		if buf.Len() == 0 {
			t.Fatal("Packfile buffer is empty")
		}

		// Read packfile
		reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
		packfile, err := reader.ReadPackfile()
		if err != nil {
			t.Fatalf("ReadPackfile() error: %v", err)
		}

		// Verify header
		if packfile.Header.Signature != PackfileSignature {
			t.Errorf("Invalid signature: %s", packfile.Header.Signature)
		}
		if packfile.Header.Version != PackfileVersion {
			t.Errorf("Invalid version: %d", packfile.Header.Version)
		}
		if packfile.Header.ObjectCount != uint32(len(objects)) {
			t.Errorf("Object count = %d, want %d", packfile.Header.ObjectCount, len(objects))
		}

		// Verify all objects
		if len(packfile.Objects) != len(objects) {
			t.Fatalf("Object count = %d, want %d", len(packfile.Objects), len(objects))
		}

		for i, expected := range objects {
			actual := packfile.Objects[i]
			if actual.Type != expected.Type {
				t.Errorf("Object %d type = %d (%s), want %d (%s)",
					i, actual.Type, ObjectTypeName(actual.Type),
					expected.Type, ObjectTypeName(expected.Type))
			}
			if actual.Size != expected.Size {
				t.Errorf("Object %d size = %d, want %d", i, actual.Size, expected.Size)
			}
			if !bytes.Equal(actual.Data, expected.Data) {
				t.Errorf("Object %d data mismatch:\ngot:  %q\nwant: %q",
					i, string(actual.Data), string(expected.Data))
			}
		}

		// Verify checksum exists
		if len(packfile.Checksum) != PackfileChecksumSize {
			t.Errorf("Checksum size = %d, want %d", len(packfile.Checksum), PackfileChecksumSize)
		}
	})
}

// TestDeltaIntegration tests delta encoding and resolution workflow
func TestDeltaIntegration(t *testing.T) {
	t.Run("create and apply delta", func(t *testing.T) {
		// Simulate a file change
		originalFile := []byte("This is the original content\nwith multiple lines\nof text data\n")
		modifiedFile := []byte("This is the modified content\nwith multiple lines\nof different text data\nand extra lines\n")

		// Create delta from original to modified
		delta := CreateDelta(originalFile, modifiedFile)
		deltaBytes, err := EncodeDelta(delta)
		if err != nil {
			t.Fatalf("EncodeDelta() error: %v", err)
		}

		t.Logf("Delta size: %d bytes (original: %d, modified: %d, savings: %.1f%%)",
			len(deltaBytes), len(originalFile), len(modifiedFile),
			(1.0-float64(len(deltaBytes))/float64(len(modifiedFile)))*100)

		// Parse and apply delta
		parsedDelta, err := ParseDelta(deltaBytes)
		if err != nil {
			t.Fatalf("ParseDelta() error: %v", err)
		}

		resolved, err := ApplyDelta(originalFile, parsedDelta)
		if err != nil {
			t.Fatalf("ApplyDelta() error: %v", err)
		}

		// Verify resolved content matches modified file
		if !bytes.Equal(resolved, modifiedFile) {
			t.Errorf("Resolved delta doesn't match:\ngot:  %q\nwant: %q",
				string(resolved), string(modifiedFile))
		}
	})

	t.Run("delta chain", func(t *testing.T) {
		// Simulate multiple file versions
		v1 := []byte("Version 1 of the file\n")
		v2 := []byte("Version 2 of the file with updates\n")
		v3 := []byte("Version 3 of the file with more updates and additions\n")

		// Create delta chain: v1 -> v2 -> v3
		delta1, err := CreateAndEncodeDelta(v1, v2)
		if err != nil {
			t.Fatalf("CreateAndEncodeDelta(v1->v2) error: %v", err)
		}

		delta2, err := CreateAndEncodeDelta(v2, v3)
		if err != nil {
			t.Fatalf("CreateAndEncodeDelta(v2->v3) error: %v", err)
		}

		// Apply first delta
		resolved1, err := ResolveRefDelta(delta1, v1)
		if err != nil {
			t.Fatalf("ResolveRefDelta(delta1) error: %v", err)
		}
		if !bytes.Equal(resolved1, v2) {
			t.Errorf("First delta resolution failed")
		}

		// Apply second delta
		resolved2, err := ResolveRefDelta(delta2, v2)
		if err != nil {
			t.Fatalf("ResolveRefDelta(delta2) error: %v", err)
		}
		if !bytes.Equal(resolved2, v3) {
			t.Errorf("Second delta resolution failed")
		}
	})

	t.Run("resolve reference delta", func(t *testing.T) {
		baseData := []byte("package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n")
		targetData := []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n")

		// Create delta
		deltaBytes, err := CreateAndEncodeDelta(baseData, targetData)
		if err != nil {
			t.Fatalf("CreateAndEncodeDelta() error: %v", err)
		}

		// Create packfile with base and ref delta
		baseHash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
		objects := []PackfileObject{
			{
				Type: ObjBlob,
				Size: uint64(len(baseData)),
				Data: baseData,
			},
			{
				Type:     ObjRefDelta,
				Size:     uint64(len(deltaBytes)),
				Data:     deltaBytes,
				BaseHash: baseHash,
				IsDelta:  true,
			},
		}

		// Write and read packfile
		var buf bytes.Buffer
		writer := NewPackfileWriter(&buf)
		if err := writer.WritePackfile(objects); err != nil {
			t.Fatalf("WritePackfile() error: %v", err)
		}

		reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
		packfile, err := reader.ReadPackfile()
		if err != nil {
			t.Fatalf("ReadPackfile() error: %v", err)
		}

		// Resolve the ref delta (we'd normally look up base by hash, but for testing we use the first object)
		resolved, err := ResolveRefDelta(packfile.Objects[1].Data, packfile.Objects[0].Data)
		if err != nil {
			t.Fatalf("ResolveRefDelta() error: %v", err)
		}

		// Verify resolved content matches target
		if !bytes.Equal(resolved, targetData) {
			t.Errorf("Resolved ref delta doesn't match:\ngot:  %q\nwant: %q",
				string(resolved), string(targetData))
		}
	})
}

// TestProtocolErrorHandlingIntegration tests error detection and messages
func TestProtocolErrorHandlingIntegration(t *testing.T) {
	t.Run("CORS error detection", func(t *testing.T) {
		// Simulate CORS error with status code 0
		err := fmt.Errorf("network request failed")
		isCORS := DetectCORSError(err, 0)
		if !isCORS {
			t.Error("Expected CORS detection for status code 0")
		}
		t.Log("CORS error successfully detected for status code 0")

		// Test with CORS-specific error message
		corsErr := fmt.Errorf("blocked by CORS policy")
		isCORS2 := DetectCORSError(corsErr, 200)
		if !isCORS2 {
			t.Error("Expected CORS detection from error message")
		}
		t.Log("CORS error successfully detected from message")
	})

	t.Run("authentication error detection", func(t *testing.T) {
		// Create a protocol error for authentication failure
		err := &ProtocolError{
			StatusCode: 401,
			Message:    "Unauthorized",
			Type:       ErrAuthentication,
		}

		if !IsAuthenticationError(err) {
			t.Error("Expected authentication error detection")
		}
		t.Log("Authentication error successfully detected")
	})

	t.Run("not found error detection", func(t *testing.T) {
		// Create a protocol error for not found
		err := &ProtocolError{
			StatusCode: 404,
			Message:    "Not Found",
			Type:       ErrNotFound,
		}

		if !IsNotFoundError(err) {
			t.Error("Expected not found error detection")
		}
		t.Log("Not found error successfully detected")
	})
}

// TestPktLineIntegration tests pkt-line encoding/decoding
func TestPktLineIntegration(t *testing.T) {
	t.Run("encode and decode pkt-lines", func(t *testing.T) {
		lines := []string{
			"want 1234567890abcdef",
			"have abcdef1234567890",
			"done",
		}

		// Encode
		var buf bytes.Buffer
		writer := NewPktLineWriter(&buf)
		for _, line := range lines {
			if err := writer.WriteLine([]byte(line)); err != nil {
				t.Fatalf("WriteLine() error: %v", err)
			}
		}
		if err := writer.WriteFlush(); err != nil {
			t.Fatalf("WriteFlush() error: %v", err)
		}

		// Decode
		reader := NewPktLineReader(&buf)
		for i, expected := range lines {
			line, err := reader.ReadLine()
			if err != nil {
				t.Fatalf("ReadLine(%d) error: %v", i, err)
			}
			if string(line) != expected {
				t.Errorf("Line %d = %q, want %q", i, string(line), expected)
			}
		}

		// Verify flush packet
		line, err := reader.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine(flush) error: %v", err)
		}
		if line != nil {
			t.Errorf("Expected flush packet (nil), got: %q", string(line))
		}
	})
}

// TestCompletePackfileWorkflow tests the complete workflow of creating,
// writing, reading, and resolving a packfile with deltas
func TestCompletePackfileWorkflow(t *testing.T) {
	// Simulate a repository with multiple commits
	file1v1 := []byte("Initial content\n")
	file1v2 := []byte("Initial content\nwith modifications\n")
	file2 := []byte("Another file\n")
	treeData := []byte("100644 file1.txt\x00" + strings.Repeat("\x00", 20) +
		"100644 file2.txt\x00" + strings.Repeat("\x00", 20))
	commitData := []byte("tree abc123\nauthor Test\n\nCommit message\n")

	// Create deltas for modified files
	file1Delta, err := CreateAndEncodeDelta(file1v1, file1v2)
	if err != nil {
		t.Fatalf("CreateAndEncodeDelta() error: %v", err)
	}

	// Build packfile with base objects (no deltas for simplicity)
	objects := []PackfileObject{
		{Type: ObjBlob, Size: uint64(len(file1v1)), Data: file1v1},
		{Type: ObjBlob, Size: uint64(len(file1v2)), Data: file1v2},
		{Type: ObjBlob, Size: uint64(len(file2)), Data: file2},
		{Type: ObjTree, Size: uint64(len(treeData)), Data: treeData},
		{Type: ObjCommit, Size: uint64(len(commitData)), Data: commitData},
	}

	// Write packfile
	var buf bytes.Buffer
	writer := NewPackfileWriter(&buf)
	if err := writer.WritePackfile(objects); err != nil {
		t.Fatalf("WritePackfile() error: %v", err)
	}

	packfileSize := buf.Len()
	t.Logf("Packfile size: %d bytes for %d objects", packfileSize, len(objects))

	// Read packfile
	reader := NewPackfileReader(bytes.NewReader(buf.Bytes()))
	packfile, err := reader.ReadPackfile()
	if err != nil {
		t.Fatalf("ReadPackfile() error: %v", err)
	}

	// Verify packfile structure
	if len(packfile.Objects) != len(objects) {
		t.Fatalf("Object count = %d, want %d", len(packfile.Objects), len(objects))
	}

	// Verify we can read all objects correctly
	for i := range objects {
		if !bytes.Equal(packfile.Objects[i].Data, objects[i].Data) {
			t.Errorf("Object %d data mismatch", i)
		}
	}

	// Test delta separately
	resolved, err := ResolveRefDelta(file1Delta, file1v1)
	if err != nil {
		t.Fatalf("ResolveRefDelta() error: %v", err)
	}

	if !bytes.Equal(resolved, file1v2) {
		t.Errorf("Resolved delta doesn't match:\ngot:  %q\nwant: %q",
			string(resolved), string(file1v2))
	}

	t.Log("Complete packfile workflow successful")
}
