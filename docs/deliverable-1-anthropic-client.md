# Deliverable 1: Anthropic Client - Backend Implementation

## Overview

HTTP client for the Anthropic Messages API, enabling AI-guided validation of user responses in the Telegram moment tracking flow.

## Architecture Compliance

- **Layer**: Foundation (infrastructure)
- **Domain Type**: N/A (not a business domain)
- **Imports**: Standard library only (context, net/http, encoding/json, errors, fmt, io, time)
- **Status**: ALIGNED with business-model-dependencies.md

Foundation packages have ZERO business domain dependencies. This follows the same pattern as `foundation/telegram/`.

---

## Package Structure

```
foundation/anthropic/
├── anthropic.go    # Client struct, constructor, main methods
└── errors.go       # Error types and classification
```

---

## Public API

### Types

```go
// Client communicates with the Anthropic Messages API.
type Client struct {
    apiKey      string
    model       string
    maxTokens   int
    temperature float64
    httpClient  *http.Client
}

// Config holds configuration for the Anthropic client.
type Config struct {
    APIKey      string
    Model       string
    MaxTokens   int
    Temperature float64
}

// MessageRequest represents a request to the Messages API.
type MessageRequest struct {
    SystemPrompt string
    UserMessage  string
}

// MessageResponse represents the raw response from the Anthropic API.
type MessageResponse struct {
    // Content is the raw text response from Claude.
    // The caller is responsible for parsing this into their domain-specific format.
    Content string

    // Usage contains token usage information.
    Usage Usage
}

// Usage contains token usage information from the API response.
type Usage struct {
    InputTokens  int
    OutputTokens int
}
```

### Constructor

```go
// NewClient creates a new Anthropic client with defaults.
func NewClient(apiKey, model string) (*Client, error)

// NewClientWithConfig creates a client with custom configuration.
func NewClientWithConfig(cfg Config) (*Client, error)
```

### Methods

```go
// SendMessage sends a message to Claude and returns the parsed response.
func (c *Client) SendMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
```

---

## Implementation

### anthropic.go

```go
// Package anthropic provides an HTTP client for the Anthropic Messages API.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiURL             = "https://api.anthropic.com/v1/messages"
	apiVersion         = "2023-06-01"
	defaultTimeout     = 30 * time.Second
	defaultMaxTokens   = 500
	defaultTemperature = 0.7
	maxTokensLimit     = 4096 // Anthropic API limit
)

// Client communicates with the Anthropic Messages API.
type Client struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

// Config holds configuration for the Anthropic client.
type Config struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
}

// NewClient creates a new Anthropic client with the specified API key and model.
// Uses default maxTokens (500) and temperature (0.7).
func NewClient(apiKey, model string) (*Client, error) {
	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}
	if model == "" {
		return nil, ErrEmptyModel
	}

	return &Client{
		apiKey:      apiKey,
		model:       model,
		maxTokens:   defaultMaxTokens,
		temperature: defaultTemperature,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}, nil
}

// NewClientWithConfig creates a client with custom configuration.
func NewClientWithConfig(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, ErrEmptyAPIKey
	}
	if cfg.Model == "" {
		return nil, ErrEmptyModel
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}
	if maxTokens > maxTokensLimit {
		return nil, fmt.Errorf("max tokens must be <= %d, got %d", maxTokensLimit, maxTokens)
	}

	temperature := cfg.Temperature
	if temperature == 0 {
		temperature = defaultTemperature
	}

	return &Client{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   maxTokens,
		temperature: temperature,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}, nil
}

// MessageRequest represents a request to send to the Messages API.
type MessageRequest struct {
	SystemPrompt string
	UserMessage  string
}

// MessageResponse represents the parsed response from Claude.
// Status is "approved" or "needs_refinement" as defined in moment-tracker.md prompts.
type MessageResponse struct {
	Status     string         `json:"status"`      // "approved" | "needs_refinement"
	Feedback   *string        `json:"feedback"`    // nil if approved
	ParsedData map[string]any `json:"parsed_data"` // Step-specific extracted data
}

// anthropicRequest is the API request format.
type anthropicRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	System      string    `json:"system"`
	Messages    []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the API response format.
type anthropicResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []contentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      usage          `json:"usage"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// SendMessage sends a message to Claude and returns the parsed response.
func (c *Client) SendMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Build API request
	apiReq := anthropicRequest{
		Model:       c.model,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		System:      req.SystemPrompt,
		Messages: []message{
			{Role: "user", Content: req.UserMessage},
		},
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode, respBody)
	}

	// Parse API response
	var apiResp anthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Extract text from content
	if len(apiResp.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	// Safe: len check above guarantees Content[0] exists
	responseText := apiResp.Content[0].Text

	// Parse Claude's JSON response
	var msgResp MessageResponse
	if err := json.Unmarshal([]byte(responseText), &msgResp); err != nil {
		return nil, fmt.Errorf("parse claude response: %w", err)
	}

	return &msgResp, nil
}

// handleErrorResponse converts HTTP error responses to typed errors.
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	// Truncate body for error message (max 200 chars)
	bodyStr := string(body)
	if len(bodyStr) > 200 {
		bodyStr = bodyStr[:200] + "..."
	}

	switch statusCode {
	case http.StatusTooManyRequests:
		return &RateLimitError{
			StatusCode: statusCode,
			Message:    bodyStr,
		}
	case http.StatusUnauthorized:
		return &APIError{
			StatusCode: statusCode,
			Retryable:  false,
			Message:    "invalid API key",
		}
	case http.StatusBadRequest:
		return &APIError{
			StatusCode: statusCode,
			Retryable:  false,
			Message:    bodyStr,
		}
	default:
		// 5xx errors are retryable
		retryable := statusCode >= 500
		return &APIError{
			StatusCode: statusCode,
			Retryable:  retryable,
			Message:    bodyStr,
		}
	}
}
```

### errors.go

```go
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
```

---

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PARTNER_ANTHROPIC_APIKEY` | Anthropic API key | - | Yes |
| `PARTNER_ANTHROPIC_MODEL` | Claude model ID | `claude-haiku-4-5` | No |
| `PARTNER_ANTHROPIC_MAXTOKENS` | Max response tokens | `500` | No |
| `PARTNER_ANTHROPIC_TEMPERATURE` | Response temperature | `0.7` | No |

### main.go Config Struct Addition

```go
// Add to cfg struct in api/services/partners/main.go

Anthropic struct {
    APIKey      string  `conf:"required,mask,env:ANTHROPIC_APIKEY"`
    Model       string  `conf:"default:claude-haiku-4-5,env:ANTHROPIC_MODEL"`
    MaxTokens   int     `conf:"default:500,env:ANTHROPIC_MAXTOKENS"`
    Temperature float64 `conf:"default:0.7,env:ANTHROPIC_TEMPERATURE"`
} `conf:"namespace:anthropic"`
```

### Initialization in main.go

```go
// After config parsing

anthropicClient, err := anthropic.NewClientWithConfig(anthropic.Config{
    APIKey:      cfg.Anthropic.APIKey,
    Model:       cfg.Anthropic.Model,
    MaxTokens:   cfg.Anthropic.MaxTokens,
    Temperature: cfg.Anthropic.Temperature,
})
if err != nil {
    return fmt.Errorf("create anthropic client: %w", err)
}

// Pass to job worker initialization
```

---

## Usage Example

```go
// In app/jobs/telegrammessage/

func (w *Worker) processMessage(ctx context.Context, session *Session, userText string) error {
    // Build prompt from moment-tracker.md templates
    systemPrompt := buildSystemPrompt()
    stepPrompt := buildStepPrompt(session.CurrentStep, session.ParsedData, userText)

    // Call Anthropic
    resp, err := w.anthropicClient.SendMessage(ctx, anthropic.MessageRequest{
        SystemPrompt: systemPrompt,
        UserMessage:  stepPrompt,
    })
    if err != nil {
        if anthropic.IsRetryableError(err) {
            // Let River queue retry
            return fmt.Errorf("anthropic call (retryable): %w", err)
        }
        // Non-retryable, fail the job
        return fmt.Errorf("anthropic call: %w", err)
    }

    // Handle response
    if resp.Status == "approved" {
        // Store parsed data, advance step
        session.ParsedData[currentStepKey] = resp.ParsedData
        session.CurrentStep++
    } else {
        // Send feedback, increment retry count
        session.RetryCount++
        sendTelegramMessage(resp.Feedback)
    }

    return nil
}
```

---

## Error Handling Strategy

| Error Type | HTTP Status | Retryable | Action |
|------------|-------------|-----------|--------|
| `RateLimitError` | 429 | Yes | River retries with backoff |
| `APIError` (5xx) | 500-599 | Yes | River retries with backoff |
| `APIError` (4xx) | 400-499 | No | Fail job, log error |
| `ErrEmptyAPIKey` | N/A | No | Startup failure |
| `ErrEmptyModel` | N/A | No | Startup failure |
| `ErrEmptyResponse` | 200 | No | Fail job, investigate |

---

## API Reference

### Anthropic Messages API

- **Endpoint**: `POST https://api.anthropic.com/v1/messages`
- **Version**: `2023-06-01`
- **Documentation**: https://docs.anthropic.com/en/api/messages

### Request Headers

```
Content-Type: application/json
x-api-key: <API_KEY>
anthropic-version: 2023-06-01
```

### Request Body

```json
{
  "model": "claude-3-5-sonnet-20241022",
  "max_tokens": 500,
  "temperature": 0.7,
  "system": "<system prompt from moment-tracker.md>",
  "messages": [
    {"role": "user", "content": "<user message with context>"}
  ]
}
```

### Response Body

```json
{
  "id": "msg_...",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "{\"status\":\"approved\",\"feedback\":null,\"parsed_data\":{...}}"
    }
  ],
  "stop_reason": "end_turn",
  "usage": {"input_tokens": 150, "output_tokens": 80}
}
```

---

## Files to Create

| File | Purpose |
|------|---------|
| `foundation/anthropic/anthropic.go` | Client implementation |
| `foundation/anthropic/errors.go` | Error types |

## Files to Modify

| File | Change |
|------|--------|
| `api/services/partners/main.go` | Add Anthropic config struct |
| `.env.example` | Add PARTNER_ANTHROPIC_* variables |

---

## Deployment Notes

1. **Environment Setup**:
   - Add `PARTNER_ANTHROPIC_APIKEY` to server `.env`
   - API key from https://console.anthropic.com/

2. **No Database Changes**: This deliverable has no schema changes

3. **No Nginx Changes**: No new routes needed

4. **Verification**:
   - Service starts without errors
   - Config outputs show `ANTHROPIC_APIKEY: ********` (masked)

---

## Errors-to-Avoid Compliance

| Error # | Rule | Compliance |
|---------|------|------------|
| #4 | Strong types with Value() | N/A - API request/response types use primitives (acceptable per #15) |
| #6 | Structured logging | N/A - Foundation packages don't include logging (matches `foundation/telegram/` pattern). Logging handled by app layer (job worker). |
| #14 | Max limits on parameters | ✓ MaxTokens validated against `maxTokensLimit` (4096) |
| #32 | Slice safety | ✓ Length check before accessing `Content[0]` with safety comment |

**Note on Logging**: Foundation packages are low-level infrastructure and do NOT include logging dependencies. The calling code (app layer) is responsible for logging. This follows the established pattern in `foundation/telegram/client.go`.
