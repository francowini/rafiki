package anthropic

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure cases.
var (
	ErrEmptyAPIKey   = errors.New("anthropic: API key cannot be empty")
	ErrEmptyModel    = errors.New("anthropic: model cannot be empty")
	ErrEmptyResponse = errors.New("anthropic: empty response from API")
)

// RateLimitError indicates the API rate limit was exceeded.
type RateLimitError struct {
	StatusCode int
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("anthropic: rate limited (status %d): %s", e.StatusCode, e.Message)
}

// IsRateLimitError checks if an error is a rate limit error.
func IsRateLimitError(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// APIError represents an error response from the Anthropic API.
type APIError struct {
	StatusCode int
	Retryable  bool
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("anthropic: API error (status %d): %s", e.StatusCode, e.Message)
}

// IsRetryable returns true if the error is retryable.
func (e *APIError) IsRetryable() bool {
	return e.Retryable
}

// IsAPIError checks if an error is an API error.
func IsAPIError(err error) bool {
	var ae *APIError
	return errors.As(err, &ae)
}

// IsRetryableError checks if an error should be retried.
func IsRetryableError(err error) bool {
	// Rate limit errors are retryable
	if IsRateLimitError(err) {
		return true
	}

	// Check API error retryable flag
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.Retryable
	}

	// Network errors are generally retryable
	return false
}
