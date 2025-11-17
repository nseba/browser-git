package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// OAuthProvider implements OAuth 2.0 authentication
type OAuthProvider struct {
	accessToken  string
	refreshToken string
}

// NewOAuthProvider creates a new OAuth authentication provider
func NewOAuthProvider(accessToken, refreshToken string) *OAuthProvider {
	return &OAuthProvider{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}
}

// GetMethod returns the authentication method
func (p *OAuthProvider) GetMethod() AuthMethod {
	return AuthMethodOAuth
}

// ApplyAuth applies OAuth authentication to the request
func (p *OAuthProvider) ApplyAuth(req *http.Request) error {
	if err := p.ValidateCredentials(); err != nil {
		return err
	}

	// OAuth uses Bearer token authentication
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	return nil
}

// ValidateCredentials validates that the access token is set
func (p *OAuthProvider) ValidateCredentials() error {
	if strings.TrimSpace(p.accessToken) == "" {
		return fmt.Errorf("OAuth access token is required")
	}
	return nil
}

// Clone creates a copy of the provider
func (p *OAuthProvider) Clone() AuthProvider {
	return &OAuthProvider{
		accessToken:  p.accessToken,
		refreshToken: p.refreshToken,
	}
}

// GetAccessToken returns the access token
func (p *OAuthProvider) GetAccessToken() string {
	return p.accessToken
}

// SetAccessToken sets the access token
func (p *OAuthProvider) SetAccessToken(token string) {
	p.accessToken = token
}

// GetRefreshToken returns the refresh token
func (p *OAuthProvider) GetRefreshToken() string {
	return p.refreshToken
}

// SetRefreshToken sets the refresh token
func (p *OAuthProvider) SetRefreshToken(token string) {
	p.refreshToken = token
}

// NeedsRefresh returns true if the access token should be refreshed
// Note: This is a simple check - in a real implementation, you would
// track token expiry and refresh proactively
func (p *OAuthProvider) NeedsRefresh() bool {
	return p.refreshToken != "" && p.accessToken == ""
}
