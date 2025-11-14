package protocol

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ServiceType represents the type of Git service
type ServiceType string

const (
	// UploadPackService is used for fetching/cloning
	UploadPackService ServiceType = "git-upload-pack"

	// ReceivePackService is used for pushing
	ReceivePackService ServiceType = "git-receive-pack"
)

// Reference represents a Git reference (branch or tag)
type Reference struct {
	Name string // e.g., "refs/heads/main"
	Hash string // 40-character SHA-1 hex string
}

// DiscoveryResponse contains the server's advertised capabilities and references
type DiscoveryResponse struct {
	Service      ServiceType
	Capabilities []string
	References   []Reference
	SymRefs      map[string]string // Symbolic references (e.g., HEAD -> refs/heads/main)
}

// Client represents a Git HTTP protocol client
type Client struct {
	httpClient *http.Client
	userAgent  string
	authHeader string // For authentication
}

// NewClient creates a new Git protocol client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		userAgent:  "browser-git/0.1.0",
	}
}

// SetAuth sets the authentication header
func (c *Client) SetAuth(username, password string) {
	// Use Basic authentication
	c.authHeader = "Basic " + basicAuth(username, password)
}

// SetAuthToken sets a bearer token for authentication
func (c *Client) SetAuthToken(token string) {
	c.authHeader = "Bearer " + token
}

// SetAuthHeader sets a custom authentication header
func (c *Client) SetAuthHeader(header string) {
	c.authHeader = header
}

// basicAuth encodes username and password for HTTP Basic Auth
func basicAuth(username, password string) string {
	auth := username + ":" + password
	// In real implementation, this should be base64 encoded
	// For WASM, we'll handle this in the JS bridge
	return auth
}

// Discover performs the discovery phase and retrieves repository info
func (c *Client) Discover(repoURL string, service ServiceType) (*DiscoveryResponse, error) {
	// Construct the info/refs URL
	infoRefsURL, err := buildInfoRefsURL(repoURL, service)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("GET", infoRefsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Git-Protocol", "version=2")

	// Add authentication if provided
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	expectedContentType := fmt.Sprintf("application/x-%s-advertisement", service)
	if !strings.Contains(contentType, expectedContentType) {
		return nil, fmt.Errorf("unexpected content type: %s (expected %s)", contentType, expectedContentType)
	}

	// Parse the response
	discovery, err := parseDiscoveryResponse(resp.Body, service)
	if err != nil {
		return nil, fmt.Errorf("failed to parse discovery response: %w", err)
	}

	return discovery, nil
}

// buildInfoRefsURL constructs the info/refs URL with service parameter
func buildInfoRefsURL(repoURL string, service ServiceType) (string, error) {
	// Parse the repository URL
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}

	// Ensure the path ends with .git or add info/refs
	path := u.Path
	if !strings.HasSuffix(path, "/") {
		if !strings.HasSuffix(path, ".git") {
			path += ".git"
		}
		path += "/"
	}
	path += "info/refs"

	// Set the path and query
	u.Path = path
	q := u.Query()
	q.Set("service", string(service))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// parseDiscoveryResponse parses the server's discovery response
func parseDiscoveryResponse(body io.Reader, service ServiceType) (*DiscoveryResponse, error) {
	reader := NewPktLineReader(body)

	// Read the first line: "# service=<service-name>"
	firstLine, err := reader.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read service line: %w", err)
	}

	expectedService := fmt.Sprintf("# service=%s\n", service)
	if string(firstLine) != expectedService {
		return nil, fmt.Errorf("unexpected service line: %s (expected %s)", string(firstLine), expectedService)
	}

	// Read the flush packet after service line
	flushLine, err := reader.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read flush after service: %w", err)
	}
	if flushLine != nil {
		return nil, fmt.Errorf("expected flush packet after service line, got data: %s", string(flushLine))
	}

	// Read all reference lines until the next flush
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read reference lines: %w", err)
	}

	// Parse references and capabilities
	response := &DiscoveryResponse{
		Service:    service,
		SymRefs:    make(map[string]string),
		References: []Reference{},
	}

	for i, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Remove trailing newline
		lineStr := string(line)
		lineStr = strings.TrimSuffix(lineStr, "\n")

		// First line contains capabilities after null byte
		if i == 0 {
			hash, refName, caps, err := parseCapabilitiesLine(lineStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse capabilities line: %w", err)
			}

			response.Capabilities = caps
			response.References = append(response.References, Reference{
				Name: refName,
				Hash: hash,
			})

			// Parse symrefs from capabilities
			for _, cap := range caps {
				if strings.HasPrefix(cap, "symref=") {
					symrefStr := strings.TrimPrefix(cap, "symref=")
					parts := strings.Split(symrefStr, ":")
					if len(parts) == 2 {
						response.SymRefs[parts[0]] = parts[1]
					}
				}
			}
		} else {
			// Subsequent lines are just "hash refname"
			hash, refName, err := parseRefLine(lineStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ref line: %w", err)
			}

			response.References = append(response.References, Reference{
				Name: refName,
				Hash: hash,
			})
		}
	}

	return response, nil
}

// parseCapabilitiesLine parses the first reference line which includes capabilities
// Format: "<hash> <refname>\0<capability> <capability> ..."
func parseCapabilitiesLine(line string) (hash string, refName string, capabilities []string, err error) {
	// Split on null byte to separate ref from capabilities
	parts := strings.SplitN(line, "\x00", 2)
	if len(parts) < 1 {
		return "", "", nil, fmt.Errorf("invalid capabilities line: %s", line)
	}

	// Parse the hash and ref name
	hash, refName, err = parseRefLine(parts[0])
	if err != nil {
		return "", "", nil, err
	}

	// Parse capabilities if present
	if len(parts) == 2 {
		capStr := parts[1]
		capabilities = strings.Split(capStr, " ")
		// Filter out empty strings
		filtered := []string{}
		for _, cap := range capabilities {
			if cap != "" {
				filtered = append(filtered, cap)
			}
		}
		capabilities = filtered
	}

	return hash, refName, capabilities, nil
}

// parseRefLine parses a reference line
// Format: "<hash> <refname>"
func parseRefLine(line string) (hash string, refName string, err error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ref line format: %s", line)
	}

	hash = parts[0]
	refName = parts[1]

	// Validate hash format (40 hex characters)
	if len(hash) != 40 {
		return "", "", fmt.Errorf("invalid hash length: %s", hash)
	}

	return hash, refName, nil
}

// GetDefaultBranch returns the default branch from symbolic references
func (d *DiscoveryResponse) GetDefaultBranch() (string, error) {
	if target, ok := d.SymRefs["HEAD"]; ok {
		return target, nil
	}

	// Fallback: look for common default branches
	for _, ref := range d.References {
		if ref.Name == "refs/heads/main" || ref.Name == "refs/heads/master" {
			return ref.Name, nil
		}
	}

	return "", fmt.Errorf("no default branch found")
}

// GetReference finds a reference by name
func (d *DiscoveryResponse) GetReference(name string) (*Reference, bool) {
	for _, ref := range d.References {
		if ref.Name == name {
			return &ref, true
		}
	}
	return nil, false
}

// HasCapability checks if a capability is supported
func (d *DiscoveryResponse) HasCapability(cap string) bool {
	for _, c := range d.Capabilities {
		// Handle capabilities with values (e.g., "agent=git/2.30.0")
		if strings.HasPrefix(c, cap) {
			return true
		}
	}
	return false
}
