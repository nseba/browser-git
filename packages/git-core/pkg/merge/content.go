package merge

import (
	"bytes"
	"fmt"
	"strings"
)

// MergeContent performs a three-way content merge on text files
// Returns (mergedContent, hasConflict, error)
func MergeContent(base, ours, theirs []byte) ([]byte, bool, error) {
	// Check if any content is binary
	if isBinaryContent(base) || isBinaryContent(ours) || isBinaryContent(theirs) {
		// Cannot merge binary files
		return nil, true, nil
	}

	// Split into lines
	baseLines := splitLines(base)
	ourLines := splitLines(ours)
	theirLines := splitLines(theirs)

	// Perform three-way merge
	merged, hasConflict := mergeLines(baseLines, ourLines, theirLines)

	// Join lines back together
	result := joinLines(merged)

	return result, hasConflict, nil
}

// Line represents a line in a file with conflict information
type Line struct {
	Content    string
	InConflict bool
	ConflictSide string // "ours", "theirs", or ""
}

// splitLines splits content into lines
func splitLines(content []byte) []string {
	if len(content) == 0 {
		return []string{}
	}

	str := string(content)
	// Handle both \n and \r\n line endings
	str = strings.ReplaceAll(str, "\r\n", "\n")
	return strings.Split(str, "\n")
}

// joinLines joins lines back into content
func joinLines(lines []Line) []byte {
	var buf bytes.Buffer

	for _, line := range lines {
		buf.WriteString(line.Content)
		if !strings.HasSuffix(line.Content, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes()
}

// mergeLines performs a simple three-way merge on lines
// This is a simplified implementation - a production version would use
// a more sophisticated algorithm like diff3
func mergeLines(base, ours, theirs []string) ([]Line, bool) {
	result := make([]Line, 0)
	hasConflict := false

	// Merge the diffs
	baseIdx := 0
	ourIdx := 0
	theirIdx := 0

	for baseIdx < len(base) || ourIdx < len(ours) || theirIdx < len(theirs) {
		// If we're at the end of base, check if both sides added the same thing
		if baseIdx >= len(base) {
			if ourIdx < len(ours) && theirIdx < len(theirs) {
				if ours[ourIdx] == theirs[theirIdx] {
					// Both sides added the same line
					result = append(result, Line{Content: ours[ourIdx]})
					ourIdx++
					theirIdx++
				} else {
					// Conflict: both sides added different lines
					hasConflict = true
					result = append(result, Line{Content: "<<<<<<< HEAD", InConflict: true})
					for ourIdx < len(ours) {
						result = append(result, Line{Content: ours[ourIdx], InConflict: true, ConflictSide: "ours"})
						ourIdx++
					}
					result = append(result, Line{Content: "=======", InConflict: true})
					for theirIdx < len(theirs) {
						result = append(result, Line{Content: theirs[theirIdx], InConflict: true, ConflictSide: "theirs"})
						theirIdx++
					}
					result = append(result, Line{Content: ">>>>>>> MERGE", InConflict: true})
				}
			} else if ourIdx < len(ours) {
				// Only ours added lines
				result = append(result, Line{Content: ours[ourIdx]})
				ourIdx++
			} else if theirIdx < len(theirs) {
				// Only theirs added lines
				result = append(result, Line{Content: theirs[theirIdx]})
				theirIdx++
			}
			continue
		}

		baseLine := base[baseIdx]
		ourLine := ""
		theirLine := ""

		if ourIdx < len(ours) {
			ourLine = ours[ourIdx]
		}
		if theirIdx < len(theirs) {
			theirLine = theirs[theirIdx]
		}

		// Case 1: All three match - no change
		if baseLine == ourLine && baseLine == theirLine {
			result = append(result, Line{Content: baseLine})
			baseIdx++
			ourIdx++
			theirIdx++
			continue
		}

		// Case 2: Only ours changed
		if baseLine == theirLine && baseLine != ourLine {
			result = append(result, Line{Content: ourLine})
			baseIdx++
			ourIdx++
			theirIdx++
			continue
		}

		// Case 3: Only theirs changed
		if baseLine == ourLine && baseLine != theirLine {
			result = append(result, Line{Content: theirLine})
			baseIdx++
			ourIdx++
			theirIdx++
			continue
		}

		// Case 4: Both changed to the same thing
		if ourLine == theirLine && baseLine != ourLine {
			result = append(result, Line{Content: ourLine})
			baseIdx++
			ourIdx++
			theirIdx++
			continue
		}

		// Case 5: Conflict - both changed differently
		hasConflict = true
		result = append(result, Line{Content: "<<<<<<< HEAD", InConflict: true})
		result = append(result, Line{Content: ourLine, InConflict: true, ConflictSide: "ours"})
		result = append(result, Line{Content: "=======", InConflict: true})
		result = append(result, Line{Content: theirLine, InConflict: true, ConflictSide: "theirs"})
		result = append(result, Line{Content: ">>>>>>> MERGE", InConflict: true})
		baseIdx++
		ourIdx++
		theirIdx++
	}

	return result, hasConflict
}

// DiffOp represents a diff operation
type DiffOp struct {
	Type  string // "keep", "add", "delete"
	Line  string
	Index int
}

// computeSimpleDiff computes a simple diff between two line arrays
// Returns the operations needed to transform 'from' into 'to'
func computeSimpleDiff(from, to []string) []DiffOp {
	ops := make([]DiffOp, 0)

	// This is a very simple diff implementation
	// A production version would use Myers diff or similar
	fromIdx := 0
	toIdx := 0

	for fromIdx < len(from) || toIdx < len(to) {
		if fromIdx >= len(from) {
			// Additions at the end
			ops = append(ops, DiffOp{Type: "add", Line: to[toIdx], Index: toIdx})
			toIdx++
		} else if toIdx >= len(to) {
			// Deletions at the end
			ops = append(ops, DiffOp{Type: "delete", Line: from[fromIdx], Index: fromIdx})
			fromIdx++
		} else if from[fromIdx] == to[toIdx] {
			// Lines match
			ops = append(ops, DiffOp{Type: "keep", Line: from[fromIdx], Index: fromIdx})
			fromIdx++
			toIdx++
		} else {
			// Lines don't match - determine if it's an add, delete, or change
			// Look ahead to see if the line appears later
			foundInTo := findLineIndex(to[toIdx:], from[fromIdx])
			foundInFrom := findLineIndex(from[fromIdx:], to[toIdx])

			if foundInTo >= 0 && (foundInFrom < 0 || foundInTo < foundInFrom) {
				// The current 'to' line is new (insertion)
				ops = append(ops, DiffOp{Type: "add", Line: to[toIdx], Index: toIdx})
				toIdx++
			} else if foundInFrom >= 0 {
				// The current 'from' line was deleted
				ops = append(ops, DiffOp{Type: "delete", Line: from[fromIdx], Index: fromIdx})
				fromIdx++
			} else {
				// Change (delete + add)
				ops = append(ops, DiffOp{Type: "delete", Line: from[fromIdx], Index: fromIdx})
				ops = append(ops, DiffOp{Type: "add", Line: to[toIdx], Index: toIdx})
				fromIdx++
				toIdx++
			}
		}
	}

	return ops
}

// findLineIndex finds the index of a line in a slice
func findLineIndex(lines []string, target string) int {
	for i, line := range lines {
		if line == target {
			return i
		}
	}
	return -1
}

// MergeContentWithConflicts merges content and returns conflict markers if there are conflicts
func MergeContentWithConflicts(base, ours, theirs []byte) ([]byte, bool, error) {
	merged, hasConflict, err := MergeContent(base, ours, theirs)
	if err != nil {
		return nil, false, err
	}

	return merged, hasConflict, nil
}

// FormatConflictMarkers formats conflict markers for display
func FormatConflictMarkers(ours, theirs []byte, path string) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("<<<<<<< HEAD (%s)\n", path))
	buf.Write(ours)
	if len(ours) > 0 && ours[len(ours)-1] != '\n' {
		buf.WriteString("\n")
	}
	buf.WriteString("=======\n")
	buf.Write(theirs)
	if len(theirs) > 0 && theirs[len(theirs)-1] != '\n' {
		buf.WriteString("\n")
	}
	buf.WriteString(">>>>>>> MERGE\n")

	return buf.String()
}
