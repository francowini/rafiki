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
	baseURL     string // Allow override for testing
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
		baseURL: apiURL,
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
		baseURL: apiURL,
	}, nil
}

// MessageRequest represents a request to send to the Messages API.
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
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	//nolint:canonicalheader // Anthropic API requires lowercase header names
	httpReq.Header.Set("x-api-key", c.apiKey)
	//nolint:canonicalheader // Anthropic API requires lowercase header names
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort cleanup

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
	// Safe: length check guarantees Content[0] exists
	if len(apiResp.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	return &MessageResponse{
		Content: apiResp.Content[0].Text,
		Usage: Usage{
			InputTokens:  apiResp.Usage.InputTokens,
			OutputTokens: apiResp.Usage.OutputTokens,
		},
	}, nil
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

// SetBaseURL overrides the API URL. Used for testing with mock servers.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}
