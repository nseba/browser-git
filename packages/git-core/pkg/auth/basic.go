package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// BasicAuthProvider implements HTTP Basic Authentication
type BasicAuthProvider struct {
	username string
	password string
}

// NewBasicAuthProvider creates a new basic authentication provider
func NewBasicAuthProvider(username, password string) *BasicAuthProvider {
	return &BasicAuthProvider{
		username: username,
		password: password,
	}
}

// GetMethod returns the authentication method
func (p *BasicAuthProvider) GetMethod() AuthMethod {
	return AuthMethodBasic
}

// ApplyAuth applies basic authentication to the request
func (p *BasicAuthProvider) ApplyAuth(req *http.Request) error {
	if err := p.ValidateCredentials(); err != nil {
		return err
	}

	// Create Basic Auth header
	auth := p.username + ":" + p.password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)

	return nil
}

// ValidateCredentials validates that username and password are set
func (p *BasicAuthProvider) ValidateCredentials() error {
	if strings.TrimSpace(p.username) == "" {
		return fmt.Errorf("basic auth username is required")
	}
	if strings.TrimSpace(p.password) == "" {
		return fmt.Errorf("basic auth password is required")
	}
	return nil
}

// Clone creates a copy of the provider
func (p *BasicAuthProvider) Clone() AuthProvider {
	return &BasicAuthProvider{
		username: p.username,
		password: p.password,
	}
}

// GetUsername returns the username
func (p *BasicAuthProvider) GetUsername() string {
	return p.username
}

// SetUsername sets the username
func (p *BasicAuthProvider) SetUsername(username string) {
	p.username = username
}

// SetPassword sets the password
func (p *BasicAuthProvider) SetPassword(password string) {
	p.password = password
}
