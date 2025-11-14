package protocol

import (
	"errors"
	"strings"
	"testing"
)

func TestDetectCORSError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		expected   bool
	}{
		{
			name:       "explicit CORS error message",
			err:        errors.New("blocked by CORS policy"),
			statusCode: 0,
			expected:   true,
		},
		{
			name:       "Access-Control-Allow-Origin missing",
			err:        errors.New("No 'Access-Control-Allow-Origin' header"),
			statusCode: 200,
			expected:   true,
		},
		{
			name:       "status code 0 (typical CORS)",
			err:        errors.New("network error"),
			statusCode: 0,
			expected:   true,
		},
		{
			name:       "Cross-Origin error",
			err:        errors.New("Cross-Origin Request Blocked"),
			statusCode: 0,
			expected:   true,
		},
		{
			name:       "regular 404 error",
			err:        errors.New("not found"),
			statusCode: 404,
			expected:   false,
		},
		{
			name:       "regular 500 error",
			err:        errors.New("internal server error"),
			statusCode: 500,
			expected:   false,
		},
		{
			name:       "nil error",
			err:        nil,
			statusCode: 200,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectCORSError(tt.err, tt.statusCode)
			if result != tt.expected {
				t.Errorf("DetectCORSError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWrapProtocolError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		statusCode     int
		url            string
		expectedType   error
		expectHint     bool
	}{
		{
			name:         "CORS error",
			err:          errors.New("blocked by CORS policy"),
			statusCode:   0,
			url:          "https://github.com/user/repo.git",
			expectedType: ErrCORS,
			expectHint:   true,
		},
		{
			name:         "401 authentication error",
			err:          errors.New("unauthorized"),
			statusCode:   401,
			url:          "https://github.com/user/repo.git",
			expectedType: ErrAuthentication,
			expectHint:   true,
		},
		{
			name:         "403 forbidden error",
			err:          errors.New("forbidden"),
			statusCode:   403,
			url:          "https://github.com/user/repo.git",
			expectedType: ErrForbidden,
			expectHint:   true,
		},
		{
			name:         "404 not found error",
			err:          errors.New("not found"),
			statusCode:   404,
			url:          "https://github.com/user/repo.git",
			expectedType: ErrNotFound,
			expectHint:   true,
		},
		{
			name:         "500 server error",
			err:          errors.New("internal server error"),
			statusCode:   500,
			url:          "https://github.com/user/repo.git",
			expectedType: ErrServerError,
			expectHint:   true,
		},
		{
			name:         "network connection error",
			err:          errors.New("connection refused"),
			statusCode:   0,
			url:          "https://github.com/user/repo.git",
			expectedType: ErrCORS, // status 0 triggers CORS detection first
			expectHint:   true,
		},
		{
			name:         "nil error",
			err:          nil,
			statusCode:   200,
			url:          "https://github.com/user/repo.git",
			expectedType: nil,
			expectHint:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapProtocolError(tt.err, tt.statusCode, tt.url)

			if tt.err == nil {
				if result != nil {
					t.Errorf("WrapProtocolError(nil) should return nil, got %v", result)
				}
				return
			}

			var protocolErr *ProtocolError
			if !errors.As(result, &protocolErr) {
				t.Fatalf("WrapProtocolError() did not return *ProtocolError")
			}

			if !errors.Is(protocolErr.Type, tt.expectedType) {
				t.Errorf("WrapProtocolError() type = %v, want %v", protocolErr.Type, tt.expectedType)
			}

			if protocolErr.StatusCode != tt.statusCode {
				t.Errorf("WrapProtocolError() status code = %d, want %d", protocolErr.StatusCode, tt.statusCode)
			}

			if protocolErr.URL != tt.url {
				t.Errorf("WrapProtocolError() URL = %s, want %s", protocolErr.URL, tt.url)
			}

			if tt.expectHint && protocolErr.Hint == "" {
				t.Error("WrapProtocolError() expected hint, but got empty string")
			}

			if !tt.expectHint && protocolErr.Hint != "" {
				t.Errorf("WrapProtocolError() unexpected hint: %s", protocolErr.Hint)
			}
		})
	}
}

func TestProtocolErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProtocolError
		contains []string
	}{
		{
			name: "error with hint",
			err: &ProtocolError{
				Type:    ErrCORS,
				Message: "CORS policy prevented the request",
				Hint:    "Use a CORS proxy",
			},
			contains: []string{"CORS", "Hint:", "CORS proxy"},
		},
		{
			name: "error without hint",
			err: &ProtocolError{
				Type:    ErrNotFound,
				Message: "Repository not found",
			},
			contains: []string{"not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()

			for _, substr := range tt.contains {
				if !strings.Contains(errStr, substr) {
					t.Errorf("Error() = %q, should contain %q", errStr, substr)
				}
			}
		})
	}
}

func TestProtocolErrorUnwrap(t *testing.T) {
	protocolErr := &ProtocolError{
		Type:    ErrCORS,
		Message: "test",
	}

	unwrapped := protocolErr.Unwrap()
	if unwrapped != ErrCORS {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, ErrCORS)
	}
}

func TestIsCORSError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "CORS protocol error",
			err: &ProtocolError{
				Type:    ErrCORS,
				Message: "CORS error",
			},
			expected: true,
		},
		{
			name: "authentication protocol error",
			err: &ProtocolError{
				Type:    ErrAuthentication,
				Message: "auth error",
			},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCORSError(tt.err)
			if result != tt.expected {
				t.Errorf("IsCORSError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "authentication protocol error",
			err: &ProtocolError{
				Type:    ErrAuthentication,
				Message: "auth error",
			},
			expected: true,
		},
		{
			name: "CORS protocol error",
			err: &ProtocolError{
				Type:    ErrCORS,
				Message: "CORS error",
			},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthenticationError(tt.err)
			if result != tt.expected {
				t.Errorf("IsAuthenticationError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "not found protocol error",
			err: &ProtocolError{
				Type:    ErrNotFound,
				Message: "not found",
			},
			expected: true,
		},
		{
			name: "CORS protocol error",
			err: &ProtocolError{
				Type:    ErrCORS,
				Message: "CORS error",
			},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetErrorHint(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedHint string
	}{
		{
			name: "protocol error with hint",
			err: &ProtocolError{
				Type:    ErrCORS,
				Message: "CORS error",
				Hint:    "Use a CORS proxy",
			},
			expectedHint: "Use a CORS proxy",
		},
		{
			name: "protocol error without hint",
			err: &ProtocolError{
				Type:    ErrNotFound,
				Message: "not found",
			},
			expectedHint: "",
		},
		{
			name:         "regular error",
			err:          errors.New("some error"),
			expectedHint: "",
		},
		{
			name:         "nil error",
			err:          nil,
			expectedHint: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorHint(tt.err)
			if result != tt.expectedHint {
				t.Errorf("GetErrorHint() = %q, want %q", result, tt.expectedHint)
			}
		})
	}
}

func TestErrorsIntegration(t *testing.T) {
	// Simulate a CORS error scenario
	originalErr := errors.New("blocked by CORS policy: No 'Access-Control-Allow-Origin' header")
	wrappedErr := WrapProtocolError(originalErr, 0, "https://github.com/user/repo.git")

	// Should be detected as CORS error
	if !IsCORSError(wrappedErr) {
		t.Error("Expected CORS error to be detected")
	}

	// Should have a helpful hint
	hint := GetErrorHint(wrappedErr)
	if hint == "" {
		t.Error("Expected error hint for CORS error")
	}

	if !strings.Contains(hint, "CORS proxy") {
		t.Errorf("Expected hint to mention CORS proxy, got: %s", hint)
	}

	// Error message should be descriptive
	errMsg := wrappedErr.Error()
	if !strings.Contains(errMsg, "Hint") {
		t.Errorf("Expected error message to contain hint, got: %s", errMsg)
	}
}
