package auth

import (
	"net/http"
)

// CustomAuthProvider implements custom authentication with user-provided logic
type CustomAuthProvider struct {
	headers map[string]string
	handler func(req *http.Request) error
}

// NewCustomAuthProvider creates a new custom authentication provider
func NewCustomAuthProvider(headers map[string]string, handler func(req *http.Request) error) *CustomAuthProvider {
	if headers == nil {
		headers = make(map[string]string)
	}
	return &CustomAuthProvider{
		headers: headers,
		handler: handler,
	}
}

// GetMethod returns the authentication method
func (p *CustomAuthProvider) GetMethod() AuthMethod {
	return AuthMethodCustom
}

// ApplyAuth applies custom authentication to the request
func (p *CustomAuthProvider) ApplyAuth(req *http.Request) error {
	// Apply custom headers
	for key, value := range p.headers {
		req.Header.Set(key, value)
	}

	// Call custom handler if provided
	if p.handler != nil {
		return p.handler(req)
	}

	return nil
}

// ValidateCredentials validates the custom authentication
func (p *CustomAuthProvider) ValidateCredentials() error {
	// Custom authentication is always considered valid if either headers or handler is set
	// The actual validation should be done by the custom handler
	return nil
}

// Clone creates a copy of the provider
func (p *CustomAuthProvider) Clone() AuthProvider {
	// Deep copy headers
	headersCopy := make(map[string]string, len(p.headers))
	for k, v := range p.headers {
		headersCopy[k] = v
	}

	return &CustomAuthProvider{
		headers: headersCopy,
		handler: p.handler,
	}
}

// SetHeader sets a custom header
func (p *CustomAuthProvider) SetHeader(key, value string) {
	p.headers[key] = value
}

// GetHeaders returns all custom headers
func (p *CustomAuthProvider) GetHeaders() map[string]string {
	return p.headers
}

// SetHandler sets the custom handler
func (p *CustomAuthProvider) SetHandler(handler func(req *http.Request) error) {
	p.handler = handler
}
