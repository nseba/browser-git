package protocol

import (
	"errors"
	"fmt"
	"strings"
)

// Common protocol errors
var (
	// ErrCORS indicates a CORS-related error
	ErrCORS = errors.New("CORS error")

	// ErrAuthentication indicates an authentication failure
	ErrAuthentication = errors.New("authentication failed")

	// ErrNotFound indicates the repository was not found
	ErrNotFound = errors.New("repository not found")

	// ErrForbidden indicates access was denied
	ErrForbidden = errors.New("access forbidden")

	// ErrServerError indicates a server-side error
	ErrServerError = errors.New("server error")

	// ErrInvalidResponse indicates an invalid response from the server
	ErrInvalidResponse = errors.New("invalid response")

	// ErrNetworkError indicates a network connectivity issue
	ErrNetworkError = errors.New("network error")
)

// ProtocolError represents a Git protocol error with additional context
type ProtocolError struct {
	Type       error  // The base error type
	Message    string // Human-readable error message
	StatusCode int    // HTTP status code (if applicable)
	URL        string // The URL that caused the error
	Hint       string // Helpful hint for resolution
}

// Error implements the error interface
func (e *ProtocolError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s\nHint: %s", e.Type, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error type
func (e *ProtocolError) Unwrap() error {
	return e.Type
}

// DetectCORSError detects if an error is likely due to CORS restrictions
func DetectCORSError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Common CORS error indicators
	corsIndicators := []string{
		"CORS",
		"Cross-Origin",
		"Access-Control-Allow-Origin",
		"blocked by CORS policy",
		"No 'Access-Control-Allow-Origin'",
		"CORS header 'Access-Control-Allow-Origin' missing",
	}

	for _, indicator := range corsIndicators {
		if strings.Contains(errMsg, indicator) {
			return true
		}
	}

	// Status code 0 often indicates a CORS error in browsers
	if statusCode == 0 {
		return true
	}

	return false
}

// WrapProtocolError wraps an error with protocol context
func WrapProtocolError(err error, statusCode int, url string) error {
	if err == nil {
		return nil
	}

	// Detect CORS errors
	if DetectCORSError(err, statusCode) {
		return &ProtocolError{
			Type:       ErrCORS,
			Message:    "CORS policy prevented the request",
			StatusCode: statusCode,
			URL:        url,
			Hint: `The Git server does not allow cross-origin requests from your browser.

Possible solutions:
1. Use a CORS proxy (e.g., https://cors-anywhere.herokuapp.com/)
2. Configure the Git server to send CORS headers
3. Use a browser extension to bypass CORS (for development only)
4. Run your application on the same origin as the Git server

Example using a CORS proxy:
  const proxyURL = 'https://cors-anywhere.herokuapp.com/';
  const repoURL = proxyURL + 'https://github.com/user/repo.git';`,
		}
	}

	// Handle HTTP status code errors
	switch statusCode {
	case 401:
		return &ProtocolError{
			Type:       ErrAuthentication,
			Message:    "Authentication required",
			StatusCode: statusCode,
			URL:        url,
			Hint: `The repository requires authentication.

Solutions:
1. Provide a personal access token or password
2. Check that your credentials are correct
3. Ensure your token has the necessary permissions

Example:
  client.SetAuth(username, token);`,
		}

	case 403:
		return &ProtocolError{
			Type:       ErrForbidden,
			Message:    "Access denied",
			StatusCode: statusCode,
			URL:        url,
			Hint: `You don't have permission to access this repository.

Solutions:
1. Check that you're authenticated with the correct account
2. Verify that you have read access to the repository
3. Check if the repository is private and you have been granted access`,
		}

	case 404:
		return &ProtocolError{
			Type:       ErrNotFound,
			Message:    "Repository not found",
			StatusCode: statusCode,
			URL:        url,
			Hint: `The repository could not be found.

Solutions:
1. Check that the repository URL is correct
2. Verify that the repository exists
3. Check if the repository has been renamed or deleted
4. Ensure the repository is public or you have access`,
		}

	case 500, 502, 503, 504:
		return &ProtocolError{
			Type:       ErrServerError,
			Message:    fmt.Sprintf("Server error (HTTP %d)", statusCode),
			StatusCode: statusCode,
			URL:        url,
			Hint: `The Git server encountered an error.

Solutions:
1. Wait a few moments and try again
2. Check the server status page
3. Contact the repository administrator`,
		}
	}

	// Network errors
	if strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "dial") {
		return &ProtocolError{
			Type:       ErrNetworkError,
			Message:    err.Error(),
			StatusCode: statusCode,
			URL:        url,
			Hint: `Network connectivity issue.

Solutions:
1. Check your internet connection
2. Verify that the server URL is accessible
3. Check if there's a firewall blocking the connection
4. Try accessing the repository through a browser`,
		}
	}

	// Generic protocol error
	return &ProtocolError{
		Type:       ErrInvalidResponse,
		Message:    err.Error(),
		StatusCode: statusCode,
		URL:        url,
	}
}

// IsCORSError checks if an error is a CORS error
func IsCORSError(err error) bool {
	var protocolErr *ProtocolError
	if errors.As(err, &protocolErr) {
		return errors.Is(protocolErr.Type, ErrCORS)
	}
	return false
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	var protocolErr *ProtocolError
	if errors.As(err, &protocolErr) {
		return errors.Is(protocolErr.Type, ErrAuthentication)
	}
	return false
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	var protocolErr *ProtocolError
	if errors.As(err, &protocolErr) {
		return errors.Is(protocolErr.Type, ErrNotFound)
	}
	return false
}

// GetErrorHint extracts the hint from a protocol error
func GetErrorHint(err error) string {
	var protocolErr *ProtocolError
	if errors.As(err, &protocolErr) {
		return protocolErr.Hint
	}
	return ""
}
