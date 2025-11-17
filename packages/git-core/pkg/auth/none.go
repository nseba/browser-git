package auth

import (
	"net/http"
)

// NoneAuthProvider represents no authentication
type NoneAuthProvider struct{}

// GetMethod returns the authentication method
func (p *NoneAuthProvider) GetMethod() AuthMethod {
	return AuthMethodNone
}

// ApplyAuth does nothing for no authentication
func (p *NoneAuthProvider) ApplyAuth(req *http.Request) error {
	// No authentication to apply
	return nil
}

// ValidateCredentials always succeeds for no authentication
func (p *NoneAuthProvider) ValidateCredentials() error {
	return nil
}

// Clone creates a copy of the provider
func (p *NoneAuthProvider) Clone() AuthProvider {
	return &NoneAuthProvider{}
}
