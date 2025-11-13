package index

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Gitignore represents gitignore patterns
type Gitignore struct {
	patterns []pattern
}

// pattern represents a single gitignore pattern
type pattern struct {
	pattern    string
	negation   bool   // true if pattern starts with !
	dirOnly    bool   // true if pattern ends with /
	isAbsolute bool   // true if pattern starts with /
	isRegex    bool   // true if pattern contains special chars
}

// LoadGitignore loads .gitignore files from the repository
func LoadGitignore(workTreePath string) (*Gitignore, error) {
	gi := &Gitignore{
		patterns: make([]pattern, 0),
	}

	// Load .gitignore from work tree root
	gitignorePath := filepath.Join(workTreePath, ".gitignore")
	if err := gi.loadFile(gitignorePath); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Add default ignores
	gi.addDefaultIgnores()

	return gi, nil
}

// loadFile loads patterns from a .gitignore file
func (gi *Gitignore) loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		gi.addPattern(line)
	}

	return scanner.Err()
}

// addPattern adds a pattern to the gitignore
func (gi *Gitignore) addPattern(line string) {
	p := pattern{
		pattern: line,
	}

	// Check for negation
	if strings.HasPrefix(line, "!") {
		p.negation = true
		line = line[1:]
		p.pattern = line
	}

	// Check for directory-only
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
		p.pattern = line
	}

	// Check for absolute path
	if strings.HasPrefix(line, "/") {
		p.isAbsolute = true
		line = strings.TrimPrefix(line, "/")
		p.pattern = line
	}

	// Check for special characters
	if strings.ContainsAny(line, "*?[") {
		p.isRegex = true
	}

	gi.patterns = append(gi.patterns, p)
}

// addDefaultIgnores adds default ignore patterns
func (gi *Gitignore) addDefaultIgnores() {
	// Always ignore .git directory
	gi.addPattern(".git/")
}

// Match checks if a path matches any ignore pattern
func (gi *Gitignore) Match(path string) bool {
	// Convert backslashes to forward slashes
	path = filepath.ToSlash(path)

	matched := false

	for _, p := range gi.patterns {
		if gi.matchPattern(path, p) {
			if p.negation {
				matched = false // Negation un-ignores the file
			} else {
				matched = true
			}
		}
	}

	return matched
}

// matchPattern checks if a path matches a single pattern
func (gi *Gitignore) matchPattern(path string, p pattern) bool {
	pattern := p.pattern

	// Handle directory-only patterns
	if p.dirOnly {
		// For now, we don't distinguish directories from files in this context
		// This would require additional file system checks
	}

	// Handle absolute patterns
	if p.isAbsolute {
		// Pattern matches from repository root
		return gi.matchGlob(path, pattern)
	}

	// Pattern can match anywhere in the path
	// Try matching against the full path
	if gi.matchGlob(path, pattern) {
		return true
	}

	// Try matching against each component
	parts := strings.Split(path, "/")
	for i := range parts {
		subPath := strings.Join(parts[i:], "/")
		if gi.matchGlob(subPath, pattern) {
			return true
		}
	}

	return false
}

// matchGlob performs glob pattern matching
func (gi *Gitignore) matchGlob(path, pattern string) bool {
	// Handle ** (match any number of directories)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			// Check prefix
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			// Check suffix
			if suffix != "" && !strings.HasSuffix(path, suffix) {
				// Try matching with suffix as pattern
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				return matched
			}

			return true
		}
	}

	// Handle * and ? patterns
	if strings.ContainsAny(pattern, "*?[]") {
		// Try exact match first
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Try matching just the filename
		matched, err = filepath.Match(pattern, filepath.Base(path))
		return err == nil && matched
	}

	// Exact match
	return path == pattern || strings.HasSuffix(path, "/"+pattern) || filepath.Base(path) == pattern
}

// ShouldIgnore checks if a file should be ignored
func ShouldIgnore(workTreePath string, path string) bool {
	gi, err := LoadGitignore(workTreePath)
	if err != nil {
		return false
	}
	return gi.Match(path)
}
