package auth

import (
	"net/http"
)

// AuthMethod represents the type of authentication
type AuthMethod string

const (
	// AuthMethodNone represents no authentication
	AuthMethodNone AuthMethod = "none"

	// AuthMethodBasic represents HTTP Basic Authentication
	AuthMethodBasic AuthMethod = "basic"

	// AuthMethodToken represents token-based authentication (e.g., GitHub PAT)
	AuthMethodToken AuthMethod = "token"

	// AuthMethodOAuth represents OAuth 2.0 authentication
	AuthMethodOAuth AuthMethod = "oauth"

	// AuthMethodSSH represents SSH key-based authentication
	AuthMethodSSH AuthMethod = "ssh"

	// AuthMethodCustom represents custom authentication handler
	AuthMethodCustom AuthMethod = "custom"
)

// AuthProvider is the interface that all authentication providers must implement
type AuthProvider interface {
	// GetMethod returns the authentication method type
	GetMethod() AuthMethod

	// ApplyAuth applies authentication to an HTTP request
	// It may modify request headers or other properties
	ApplyAuth(req *http.Request) error

	// ValidateCredentials validates that the credentials are properly configured
	ValidateCredentials() error

	// Clone creates a copy of the auth provider
	Clone() AuthProvider
}

// Credentials represents generic authentication credentials
type Credentials struct {
	Method   AuthMethod
	Username string
	Password string
	Token    string
	Data     map[string]interface{}
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	// Method is the authentication method to use
	Method AuthMethod

	// Username for basic authentication
	Username string

	// Password for basic authentication
	Password string

	// Token for token-based authentication
	Token string

	// AccessToken for OAuth authentication
	AccessToken string

	// RefreshToken for OAuth authentication (optional)
	RefreshToken string

	// PrivateKey for SSH authentication (PEM format)
	PrivateKey string

	// CustomHeaders for custom authentication
	CustomHeaders map[string]string

	// CustomHandler for custom authentication logic
	// This is called before the request is sent
	CustomHandler func(req *http.Request) error
}

// NewAuthProvider creates a new authentication provider based on the config
func NewAuthProvider(config *AuthConfig) (AuthProvider, error) {
	if config == nil {
		return &NoneAuthProvider{}, nil
	}

	switch config.Method {
	case AuthMethodNone:
		return &NoneAuthProvider{}, nil
	case AuthMethodBasic:
		return NewBasicAuthProvider(config.Username, config.Password), nil
	case AuthMethodToken:
		return NewTokenAuthProvider(config.Token), nil
	case AuthMethodOAuth:
		return NewOAuthProvider(config.AccessToken, config.RefreshToken), nil
	case AuthMethodCustom:
		return NewCustomAuthProvider(config.CustomHeaders, config.CustomHandler), nil
	default:
		return &NoneAuthProvider{}, nil
	}
}
