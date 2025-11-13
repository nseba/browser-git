package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigParse tests parsing Git config
func TestConfigParse(t *testing.T) {
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false

[user]
	name = Test User
	email = test@example.com

[remote "origin"]
	url = https://github.com/test/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test core.repositoryformatversion
	if val, ok := config.Get("core", "repositoryformatversion"); !ok || val != "0" {
		t.Errorf("core.repositoryformatversion = %s, want 0", val)
	}

	// Test core.bare
	if bare, ok := config.GetBool("core", "bare"); !ok || bare {
		t.Errorf("core.bare = %v, want false", bare)
	}

	// Test user.name
	if name, ok := config.Get("user", "name"); !ok || name != "Test User" {
		t.Errorf("user.name = %q, want %q", name, "Test User")
	}

	// Test user.email
	if email, ok := config.Get("user", "email"); !ok || email != "test@example.com" {
		t.Errorf("user.email = %q, want %q", email, "test@example.com")
	}

	// Test remote.origin.url
	if url, ok := config.Get("remote.origin", "url"); !ok || url != "https://github.com/test/repo.git" {
		t.Errorf("remote.origin.url = %q, want %q", url, "https://github.com/test/repo.git")
	}
}

// TestConfigSetGet tests setting and getting config values
func TestConfigSetGet(t *testing.T) {
	config := NewConfig()

	// Set values
	config.Set("user", "name", "John Doe")
	config.Set("user", "email", "john@example.com")
	config.SetBool("core", "bare", true)

	// Get values
	if name, ok := config.Get("user", "name"); !ok || name != "John Doe" {
		t.Errorf("Get returned %q, want %q", name, "John Doe")
	}

	if email, ok := config.Get("user", "email"); !ok || email != "john@example.com" {
		t.Errorf("Get returned %q, want %q", email, "john@example.com")
	}

	if bare, ok := config.GetBool("core", "bare"); !ok || !bare {
		t.Errorf("GetBool returned %v, want true", bare)
	}
}

// TestConfigUnset tests unsetting config values
func TestConfigUnset(t *testing.T) {
	config := NewConfig()

	config.Set("user", "name", "Test")
	config.Unset("user", "name")

	if _, ok := config.Get("user", "name"); ok {
		t.Error("Value should not exist after Unset")
	}
}

// TestConfigSave tests saving config to file
func TestConfigSave(t *testing.T) {
	config := NewConfig()
	config.Set("core", "repositoryformatversion", "0")
	config.Set("core", "filemode", "true")
	config.Set("user", "name", "Test User")
	config.Set("user", "email", "test@example.com")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	err := config.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if name, ok := loaded.Get("user", "name"); !ok || name != "Test User" {
		t.Errorf("Loaded name = %q, want %q", name, "Test User")
	}
}

// TestConfigHashAlgorithm tests hash algorithm configuration
func TestConfigHashAlgorithm(t *testing.T) {
	config := NewConfig()

	// Default should be SHA-1
	if algo := config.GetHashAlgorithm(); algo != "sha1" {
		t.Errorf("Default hash algorithm = %s, want sha1", algo)
	}

	// Set to SHA-256
	config.SetHashAlgorithm("sha256")
	if algo := config.GetHashAlgorithm(); algo != "sha256" {
		t.Errorf("Hash algorithm = %s, want sha256", algo)
	}

	// Set back to SHA-1
	config.SetHashAlgorithm("sha1")
	if algo := config.GetHashAlgorithm(); algo != "sha1" {
		t.Errorf("Hash algorithm = %s, want sha1", algo)
	}
}

// TestConfigUser tests user configuration
func TestConfigUser(t *testing.T) {
	config := NewConfig()

	// Initially empty
	name, email := config.GetUser()
	if name != "" || email != "" {
		t.Errorf("Initial user = (%q, %q), want empty", name, email)
	}

	// Set user
	config.SetUser("Jane Doe", "jane@example.com")

	name, email = config.GetUser()
	if name != "Jane Doe" {
		t.Errorf("User name = %q, want %q", name, "Jane Doe")
	}
	if email != "jane@example.com" {
		t.Errorf("User email = %q, want %q", email, "jane@example.com")
	}
}

// TestConfigIsBare tests bare repository detection
func TestConfigIsBare(t *testing.T) {
	config := NewConfig()

	// Initially not bare
	if config.IsBare() {
		t.Error("New config should not be bare")
	}

	// Set to bare
	config.SetBool("core", "bare", true)
	if !config.IsBare() {
		t.Error("Config should be bare after setting")
	}
}

// TestConfigInitialBranch tests initial branch configuration
func TestConfigInitialBranch(t *testing.T) {
	config := NewConfig()

	// Default should be "main"
	if branch := config.GetInitialBranch(); branch != "main" {
		t.Errorf("Default initial branch = %s, want main", branch)
	}

	// Set custom branch
	config.SetInitialBranch("develop")
	if branch := config.GetInitialBranch(); branch != "develop" {
		t.Errorf("Initial branch = %s, want develop", branch)
	}
}

// TestConfigListSections tests listing all sections
func TestConfigListSections(t *testing.T) {
	config := NewConfig()
	config.Set("core", "bare", "false")
	config.Set("user", "name", "Test")
	config.Set("remote.origin", "url", "https://example.com")

	sections := config.ListSections()
	if len(sections) != 3 {
		t.Errorf("Number of sections = %d, want 3", len(sections))
	}

	// Check that expected sections exist
	expectedSections := map[string]bool{
		"core":          false,
		"user":          false,
		"remote.origin": false,
	}

	for _, section := range sections {
		if _, ok := expectedSections[section]; ok {
			expectedSections[section] = true
		}
	}

	for section, found := range expectedSections {
		if !found {
			t.Errorf("Expected section %s not found", section)
		}
	}
}

// TestConfigListKeys tests listing keys in a section
func TestConfigListKeys(t *testing.T) {
	config := NewConfig()
	config.Set("user", "name", "Test")
	config.Set("user", "email", "test@example.com")

	keys := config.ListKeys("user")
	if len(keys) != 2 {
		t.Errorf("Number of keys = %d, want 2", len(keys))
	}

	// Check that expected keys exist
	expectedKeys := map[string]bool{
		"name":  false,
		"email": false,
	}

	for _, key := range keys {
		if _, ok := expectedKeys[key]; ok {
			expectedKeys[key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected key %s not found", key)
		}
	}
}

// TestConfigCaseInsensitive tests that config is case-insensitive
func TestConfigCaseInsensitive(t *testing.T) {
	config := NewConfig()
	config.Set("User", "Name", "Test")

	// Should be able to retrieve with different case
	if val, ok := config.Get("user", "name"); !ok || val != "Test" {
		t.Errorf("Case-insensitive get failed: got %q", val)
	}

	if val, ok := config.Get("USER", "NAME"); !ok || val != "Test" {
		t.Errorf("Case-insensitive get failed: got %q", val)
	}
}

// TestConfigComments tests that comments are ignored
func TestConfigComments(t *testing.T) {
	configContent := `# This is a comment
[core]
	# Another comment
	bare = false
	; Semicolon comment
	filemode = true
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if bare, ok := config.GetBool("core", "bare"); !ok || bare {
		t.Errorf("core.bare = %v, want false", bare)
	}

	if filemode, ok := config.GetBool("core", "filemode"); !ok || !filemode {
		t.Errorf("core.filemode = %v, want true", filemode)
	}
}

// TestConfigQuotedValues tests handling of quoted values
func TestConfigQuotedValues(t *testing.T) {
	configContent := `[user]
	name = "John Doe"
	description = "A test user with spaces"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if name, ok := config.Get("user", "name"); !ok || name != "John Doe" {
		t.Errorf("user.name = %q, want %q", name, "John Doe")
	}

	if desc, ok := config.Get("user", "description"); !ok || desc != "A test user with spaces" {
		t.Errorf("user.description = %q, want %q", desc, "A test user with spaces")
	}
}

// TestConfigNonExistentFile tests loading non-existent config file
func TestConfigNonExistentFile(t *testing.T) {
	config, err := LoadConfig("/nonexistent/config")
	if err != nil {
		t.Fatalf("Loading non-existent file should not error: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Should be empty
	if len(config.sections) != 0 {
		t.Errorf("Empty config should have 0 sections, got %d", len(config.sections))
	}
}

// TestConfigWrite tests writing config with subsections
func TestConfigWrite(t *testing.T) {
	config := NewConfig()
	config.Set("core", "bare", "false")
	config.Set("remote.origin", "url", "https://github.com/test/repo.git")
	config.Set("remote.origin", "fetch", "+refs/heads/*:refs/remotes/origin/*")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	err := config.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read the file
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)

	// Verify format
	if !strings.Contains(contentStr, "[core]") {
		t.Error("Config should contain [core] section")
	}

	if !strings.Contains(contentStr, `[remote "origin"]`) {
		t.Error("Config should contain [remote \"origin\"] section")
	}
}
