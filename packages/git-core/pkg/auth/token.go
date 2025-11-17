package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// TokenAuthProvider implements token-based authentication
// This is commonly used for GitHub Personal Access Tokens and similar services
type TokenAuthProvider struct {
	token string
}

// NewTokenAuthProvider creates a new token authentication provider
func NewTokenAuthProvider(token string) *TokenAuthProvider {
	return &TokenAuthProvider{
		token: token,
	}
}

// GetMethod returns the authentication method
func (p *TokenAuthProvider) GetMethod() AuthMethod {
	return AuthMethodToken
}

// ApplyAuth applies token authentication to the request
func (p *TokenAuthProvider) ApplyAuth(req *http.Request) error {
	if err := p.ValidateCredentials(); err != nil {
		return err
	}

	// For GitHub and many other services, tokens are sent as Bearer tokens
	// Some services may use different schemes, but Bearer is most common
	req.Header.Set("Authorization", "Bearer "+p.token)

	return nil
}

// ValidateCredentials validates that the token is set
func (p *TokenAuthProvider) ValidateCredentials() error {
	if strings.TrimSpace(p.token) == "" {
		return fmt.Errorf("authentication token is required")
	}
	return nil
}

// Clone creates a copy of the provider
func (p *TokenAuthProvider) Clone() AuthProvider {
	return &TokenAuthProvider{
		token: p.token,
	}
}

// SetToken sets the token
func (p *TokenAuthProvider) SetToken(token string) {
	p.token = token
}

// GetToken returns the token
func (p *TokenAuthProvider) GetToken() string {
	return p.token
}
