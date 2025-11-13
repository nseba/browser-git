package repository

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config represents a Git configuration
type Config struct {
	sections map[string]map[string]string
}

// NewConfig creates a new empty config
func NewConfig() *Config {
	return &Config{
		sections: make(map[string]map[string]string),
	}
}

// LoadConfig loads a Git config file from the specified path
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return NewConfig(), nil
		}
		return nil, err
	}
	defer file.Close()

	return ParseConfig(file)
}

// LoadConfigFromRepo loads the config from a repository's .git directory
func LoadConfigFromRepo(gitDir string) (*Config, error) {
	configPath := filepath.Join(gitDir, "config")
	return LoadConfig(configPath)
}

// ParseConfig parses a Git config from a reader
func ParseConfig(r *os.File) (*Config, error) {
	config := NewConfig()
	scanner := bufio.NewScanner(r)

	var currentSection string
	var currentSubsection string

	sectionRegex := regexp.MustCompile(`^\[([^\]]+)\]$`)
	keyValueRegex := regexp.MustCompile(`^\s*([^=\s]+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section header
		if matches := sectionRegex.FindStringSubmatch(line); matches != nil {
			sectionParts := strings.SplitN(matches[1], " ", 2)
			currentSection = strings.ToLower(sectionParts[0])

			if len(sectionParts) > 1 {
				// Section with subsection: [section "subsection"]
				currentSubsection = strings.Trim(sectionParts[1], "\" ")
			} else {
				currentSubsection = ""
			}

			sectionKey := currentSection
			if currentSubsection != "" {
				sectionKey = fmt.Sprintf("%s.%s", currentSection, currentSubsection)
			}

			if config.sections[sectionKey] == nil {
				config.sections[sectionKey] = make(map[string]string)
			}
			continue
		}

		// Check for key-value pair
		if matches := keyValueRegex.FindStringSubmatch(line); matches != nil {
			key := strings.ToLower(strings.TrimSpace(matches[1]))
			value := strings.TrimSpace(matches[2])

			// Remove quotes from value if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}

			sectionKey := currentSection
			if currentSubsection != "" {
				sectionKey = fmt.Sprintf("%s.%s", currentSection, currentSubsection)
			}

			if config.sections[sectionKey] == nil {
				config.sections[sectionKey] = make(map[string]string)
			}

			config.sections[sectionKey][key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

// Get retrieves a configuration value
// section can be "core" or "core.user" for subsections
// key is the configuration key
func (c *Config) Get(section, key string) (string, bool) {
	section = strings.ToLower(section)
	key = strings.ToLower(key)

	if sec, ok := c.sections[section]; ok {
		if val, ok := sec[key]; ok {
			return val, true
		}
	}
	return "", false
}

// GetBool retrieves a boolean configuration value
func (c *Config) GetBool(section, key string) (bool, bool) {
	val, ok := c.Get(section, key)
	if !ok {
		return false, false
	}

	val = strings.ToLower(val)
	return val == "true" || val == "yes" || val == "on" || val == "1", true
}

// Set sets a configuration value
func (c *Config) Set(section, key, value string) {
	section = strings.ToLower(section)
	key = strings.ToLower(key)

	if c.sections[section] == nil {
		c.sections[section] = make(map[string]string)
	}

	c.sections[section][key] = value
}

// SetBool sets a boolean configuration value
func (c *Config) SetBool(section, key string, value bool) {
	if value {
		c.Set(section, key, "true")
	} else {
		c.Set(section, key, "false")
	}
}

// Unset removes a configuration value
func (c *Config) Unset(section, key string) {
	section = strings.ToLower(section)
	key = strings.ToLower(key)

	if sec, ok := c.sections[section]; ok {
		delete(sec, key)
	}
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return c.Write(file)
}

// Write writes the configuration to a writer
func (c *Config) Write(file *os.File) error {
	// Group sections
	for section, keys := range c.sections {
		// Write section header
		if strings.Contains(section, ".") {
			// Section with subsection
			parts := strings.SplitN(section, ".", 2)
			if _, err := fmt.Fprintf(file, "[%s \"%s\"]\n", parts[0], parts[1]); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(file, "[%s]\n", section); err != nil {
				return err
			}
		}

		// Write key-value pairs
		for key, value := range keys {
			// Quote value if it contains spaces
			if strings.Contains(value, " ") {
				value = fmt.Sprintf("\"%s\"", value)
			}
			if _, err := fmt.Fprintf(file, "\t%s = %s\n", key, value); err != nil {
				return err
			}
		}

		// Empty line between sections
		if _, err := fmt.Fprintln(file); err != nil {
			return err
		}
	}

	return nil
}

// GetHashAlgorithm returns the configured hash algorithm (default: sha1)
func (c *Config) GetHashAlgorithm() string {
	// Check for extensions.objectformat (modern Git)
	if algo, ok := c.Get("extensions", "objectformat"); ok {
		return strings.ToLower(algo)
	}

	// Default to SHA-1
	return "sha1"
}

// SetHashAlgorithm sets the hash algorithm
func (c *Config) SetHashAlgorithm(algo string) {
	algo = strings.ToLower(algo)
	if algo == "sha256" {
		c.Set("extensions", "objectformat", "sha256")
	} else {
		c.Unset("extensions", "objectformat")
	}
}

// GetUser returns the configured user name and email
func (c *Config) GetUser() (name string, email string) {
	name, _ = c.Get("user", "name")
	email, _ = c.Get("user", "email")
	return
}

// SetUser sets the user name and email
func (c *Config) SetUser(name, email string) {
	if name != "" {
		c.Set("user", "name", name)
	}
	if email != "" {
		c.Set("user", "email", email)
	}
}

// IsBare returns whether this is a bare repository
func (c *Config) IsBare() bool {
	bare, ok := c.GetBool("core", "bare")
	return ok && bare
}

// GetInitialBranch returns the configured initial branch name
func (c *Config) GetInitialBranch() string {
	if branch, ok := c.Get("init", "defaultbranch"); ok {
		return branch
	}
	return "main"
}

// SetInitialBranch sets the initial branch name
func (c *Config) SetInitialBranch(branch string) {
	c.Set("init", "defaultbranch", branch)
}

// GetRepositoryFormatVersion returns the repository format version
func (c *Config) GetRepositoryFormatVersion() int {
	if version, ok := c.Get("core", "repositoryformatversion"); ok {
		if version == "0" {
			return 0
		}
		if version == "1" {
			return 1
		}
	}
	return 0
}

// ListSections returns all section names
func (c *Config) ListSections() []string {
	sections := make([]string, 0, len(c.sections))
	for section := range c.sections {
		sections = append(sections, section)
	}
	return sections
}

// ListKeys returns all keys in a section
func (c *Config) ListKeys(section string) []string {
	section = strings.ToLower(section)
	if sec, ok := c.sections[section]; ok {
		keys := make([]string, 0, len(sec))
		for key := range sec {
			keys = append(keys, key)
		}
		return keys
	}
	return nil
}
