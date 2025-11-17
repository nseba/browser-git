package auth

import (
	"net/http"
	"testing"
)

// TestNoneAuthProvider tests the NoneAuthProvider
func TestNoneAuthProvider(t *testing.T) {
	provider := &NoneAuthProvider{}

	if provider.GetMethod() != AuthMethodNone {
		t.Errorf("GetMethod() = %v, want %v", provider.GetMethod(), AuthMethodNone)
	}

	if err := provider.ValidateCredentials(); err != nil {
		t.Errorf("ValidateCredentials() error = %v, want nil", err)
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	if err := provider.ApplyAuth(req); err != nil {
		t.Errorf("ApplyAuth() error = %v, want nil", err)
	}

	// Verify no Authorization header was added
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		t.Errorf("Authorization header = %v, want empty", authHeader)
	}

	// Test cloning
	clone := provider.Clone()
	if clone.GetMethod() != AuthMethodNone {
		t.Errorf("Clone().GetMethod() = %v, want %v", clone.GetMethod(), AuthMethodNone)
	}
}

// TestBasicAuthProvider tests the BasicAuthProvider
func TestBasicAuthProvider(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		provider := NewBasicAuthProvider("testuser", "testpass")

		if provider.GetMethod() != AuthMethodBasic {
			t.Errorf("GetMethod() = %v, want %v", provider.GetMethod(), AuthMethodBasic)
		}

		if err := provider.ValidateCredentials(); err != nil {
			t.Errorf("ValidateCredentials() error = %v, want nil", err)
		}

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err != nil {
			t.Errorf("ApplyAuth() error = %v, want nil", err)
		}

		// Verify Authorization header was added
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header is empty, want Basic auth header")
		}
		if authHeader[:6] != "Basic " {
			t.Errorf("Authorization header = %v, want Basic prefix", authHeader)
		}
	})

	t.Run("empty username", func(t *testing.T) {
		provider := NewBasicAuthProvider("", "testpass")

		if err := provider.ValidateCredentials(); err == nil {
			t.Error("ValidateCredentials() error = nil, want error for empty username")
		}

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err == nil {
			t.Error("ApplyAuth() error = nil, want error for empty username")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		provider := NewBasicAuthProvider("testuser", "")

		if err := provider.ValidateCredentials(); err == nil {
			t.Error("ValidateCredentials() error = nil, want error for empty password")
		}
	})

	t.Run("clone", func(t *testing.T) {
		original := NewBasicAuthProvider("testuser", "testpass")
		clone := original.Clone().(*BasicAuthProvider)

		if clone.GetUsername() != "testuser" {
			t.Errorf("Clone username = %v, want testuser", clone.GetUsername())
		}

		// Modify clone shouldn't affect original
		clone.SetUsername("newuser")
		if original.GetUsername() == "newuser" {
			t.Error("Modifying clone affected original")
		}
	})
}

// TestTokenAuthProvider tests the TokenAuthProvider
func TestTokenAuthProvider(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		provider := NewTokenAuthProvider("ghp_testtoken123")

		if provider.GetMethod() != AuthMethodToken {
			t.Errorf("GetMethod() = %v, want %v", provider.GetMethod(), AuthMethodToken)
		}

		if err := provider.ValidateCredentials(); err != nil {
			t.Errorf("ValidateCredentials() error = %v, want nil", err)
		}

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err != nil {
			t.Errorf("ApplyAuth() error = %v, want nil", err)
		}

		// Verify Authorization header was added
		authHeader := req.Header.Get("Authorization")
		if authHeader != "Bearer ghp_testtoken123" {
			t.Errorf("Authorization header = %v, want Bearer ghp_testtoken123", authHeader)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		provider := NewTokenAuthProvider("")

		if err := provider.ValidateCredentials(); err == nil {
			t.Error("ValidateCredentials() error = nil, want error for empty token")
		}

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err == nil {
			t.Error("ApplyAuth() error = nil, want error for empty token")
		}
	})

	t.Run("clone", func(t *testing.T) {
		original := NewTokenAuthProvider("token123")
		clone := original.Clone().(*TokenAuthProvider)

		if clone.GetToken() != "token123" {
			t.Errorf("Clone token = %v, want token123", clone.GetToken())
		}

		// Modify clone shouldn't affect original
		clone.SetToken("newtoken")
		if original.GetToken() == "newtoken" {
			t.Error("Modifying clone affected original")
		}
	})
}

// TestOAuthProvider tests the OAuthProvider
func TestOAuthProvider(t *testing.T) {
	t.Run("valid access token", func(t *testing.T) {
		provider := NewOAuthProvider("access_token_123", "refresh_token_456")

		if provider.GetMethod() != AuthMethodOAuth {
			t.Errorf("GetMethod() = %v, want %v", provider.GetMethod(), AuthMethodOAuth)
		}

		if err := provider.ValidateCredentials(); err != nil {
			t.Errorf("ValidateCredentials() error = %v, want nil", err)
		}

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err != nil {
			t.Errorf("ApplyAuth() error = %v, want nil", err)
		}

		// Verify Authorization header was added
		authHeader := req.Header.Get("Authorization")
		if authHeader != "Bearer access_token_123" {
			t.Errorf("Authorization header = %v, want Bearer access_token_123", authHeader)
		}
	})

	t.Run("empty access token", func(t *testing.T) {
		provider := NewOAuthProvider("", "refresh_token_456")

		if err := provider.ValidateCredentials(); err == nil {
			t.Error("ValidateCredentials() error = nil, want error for empty access token")
		}
	})

	t.Run("needs refresh", func(t *testing.T) {
		provider := NewOAuthProvider("", "refresh_token_456")

		if !provider.NeedsRefresh() {
			t.Error("NeedsRefresh() = false, want true when access token is empty and refresh token exists")
		}

		provider.SetAccessToken("new_access_token")
		if provider.NeedsRefresh() {
			t.Error("NeedsRefresh() = true, want false when access token is set")
		}
	})

	t.Run("clone", func(t *testing.T) {
		original := NewOAuthProvider("access_123", "refresh_456")
		clone := original.Clone().(*OAuthProvider)

		if clone.GetAccessToken() != "access_123" {
			t.Errorf("Clone access token = %v, want access_123", clone.GetAccessToken())
		}
		if clone.GetRefreshToken() != "refresh_456" {
			t.Errorf("Clone refresh token = %v, want refresh_456", clone.GetRefreshToken())
		}

		// Modify clone shouldn't affect original
		clone.SetAccessToken("new_access")
		if original.GetAccessToken() == "new_access" {
			t.Error("Modifying clone affected original")
		}
	})
}

// TestCustomAuthProvider tests the CustomAuthProvider
func TestCustomAuthProvider(t *testing.T) {
	t.Run("custom headers", func(t *testing.T) {
		headers := map[string]string{
			"X-Custom-Auth": "custom_value",
			"X-API-Key":     "api_key_123",
		}
		provider := NewCustomAuthProvider(headers, nil)

		if provider.GetMethod() != AuthMethodCustom {
			t.Errorf("GetMethod() = %v, want %v", provider.GetMethod(), AuthMethodCustom)
		}

		if err := provider.ValidateCredentials(); err != nil {
			t.Errorf("ValidateCredentials() error = %v, want nil", err)
		}

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err != nil {
			t.Errorf("ApplyAuth() error = %v, want nil", err)
		}

		// Verify custom headers were added
		if req.Header.Get("X-Custom-Auth") != "custom_value" {
			t.Errorf("X-Custom-Auth header = %v, want custom_value", req.Header.Get("X-Custom-Auth"))
		}
		if req.Header.Get("X-API-Key") != "api_key_123" {
			t.Errorf("X-API-Key header = %v, want api_key_123", req.Header.Get("X-API-Key"))
		}
	})

	t.Run("custom handler", func(t *testing.T) {
		handlerCalled := false
		handler := func(req *http.Request) error {
			handlerCalled = true
			req.Header.Set("X-Handler-Called", "true")
			return nil
		}

		provider := NewCustomAuthProvider(nil, handler)

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err != nil {
			t.Errorf("ApplyAuth() error = %v, want nil", err)
		}

		if !handlerCalled {
			t.Error("Custom handler was not called")
		}
		if req.Header.Get("X-Handler-Called") != "true" {
			t.Error("Custom handler did not modify request")
		}
	})

	t.Run("headers and handler", func(t *testing.T) {
		headers := map[string]string{"X-Custom": "value"}
		handler := func(req *http.Request) error {
			req.Header.Set("X-Handler", "called")
			return nil
		}

		provider := NewCustomAuthProvider(headers, handler)

		req, _ := http.NewRequest("GET", "https://example.com", nil)
		if err := provider.ApplyAuth(req); err != nil {
			t.Errorf("ApplyAuth() error = %v, want nil", err)
		}

		// Verify both headers and handler were applied
		if req.Header.Get("X-Custom") != "value" {
			t.Error("Custom header was not added")
		}
		if req.Header.Get("X-Handler") != "called" {
			t.Error("Handler was not called")
		}
	})

	t.Run("clone", func(t *testing.T) {
		headers := map[string]string{"X-Test": "value"}
		original := NewCustomAuthProvider(headers, nil)
		clone := original.Clone().(*CustomAuthProvider)

		// Verify headers were cloned
		cloneHeaders := clone.GetHeaders()
		if cloneHeaders["X-Test"] != "value" {
			t.Error("Headers were not cloned properly")
		}

		// Modify clone shouldn't affect original
		clone.SetHeader("X-New", "new_value")
		if original.GetHeaders()["X-New"] == "new_value" {
			t.Error("Modifying clone affected original")
		}
	})
}

// TestNewAuthProvider tests the factory function
func TestNewAuthProvider(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		provider, err := NewAuthProvider(nil)
		if err != nil {
			t.Errorf("NewAuthProvider(nil) error = %v, want nil", err)
		}
		if provider.GetMethod() != AuthMethodNone {
			t.Errorf("NewAuthProvider(nil) method = %v, want %v", provider.GetMethod(), AuthMethodNone)
		}
	})

	t.Run("none method", func(t *testing.T) {
		config := &AuthConfig{Method: AuthMethodNone}
		provider, err := NewAuthProvider(config)
		if err != nil {
			t.Errorf("NewAuthProvider(none) error = %v, want nil", err)
		}
		if provider.GetMethod() != AuthMethodNone {
			t.Errorf("NewAuthProvider(none) method = %v, want %v", provider.GetMethod(), AuthMethodNone)
		}
	})

	t.Run("basic method", func(t *testing.T) {
		config := &AuthConfig{
			Method:   AuthMethodBasic,
			Username: "user",
			Password: "pass",
		}
		provider, err := NewAuthProvider(config)
		if err != nil {
			t.Errorf("NewAuthProvider(basic) error = %v, want nil", err)
		}
		if provider.GetMethod() != AuthMethodBasic {
			t.Errorf("NewAuthProvider(basic) method = %v, want %v", provider.GetMethod(), AuthMethodBasic)
		}
	})

	t.Run("token method", func(t *testing.T) {
		config := &AuthConfig{
			Method: AuthMethodToken,
			Token:  "token123",
		}
		provider, err := NewAuthProvider(config)
		if err != nil {
			t.Errorf("NewAuthProvider(token) error = %v, want nil", err)
		}
		if provider.GetMethod() != AuthMethodToken {
			t.Errorf("NewAuthProvider(token) method = %v, want %v", provider.GetMethod(), AuthMethodToken)
		}
	})

	t.Run("oauth method", func(t *testing.T) {
		config := &AuthConfig{
			Method:      AuthMethodOAuth,
			AccessToken: "access123",
		}
		provider, err := NewAuthProvider(config)
		if err != nil {
			t.Errorf("NewAuthProvider(oauth) error = %v, want nil", err)
		}
		if provider.GetMethod() != AuthMethodOAuth {
			t.Errorf("NewAuthProvider(oauth) method = %v, want %v", provider.GetMethod(), AuthMethodOAuth)
		}
	})

	t.Run("custom method", func(t *testing.T) {
		config := &AuthConfig{
			Method:        AuthMethodCustom,
			CustomHeaders: map[string]string{"X-API-Key": "key123"},
		}
		provider, err := NewAuthProvider(config)
		if err != nil {
			t.Errorf("NewAuthProvider(custom) error = %v, want nil", err)
		}
		if provider.GetMethod() != AuthMethodCustom {
			t.Errorf("NewAuthProvider(custom) method = %v, want %v", provider.GetMethod(), AuthMethodCustom)
		}
	})
}
