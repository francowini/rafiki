package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// =============================================================================
// Unit Tests (mock HTTP server)
// =============================================================================

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		model   string
		wantErr error
	}{
		{
			name:    "valid config",
			apiKey:  "test-key",
			model:   "claude-haiku-4-5",
			wantErr: nil,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			model:   "claude-haiku-4-5",
			wantErr: ErrEmptyAPIKey,
		},
		{
			name:    "empty model",
			apiKey:  "test-key",
			model:   "",
			wantErr: ErrEmptyModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, tt.model)
			if err != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil && client == nil {
				t.Error("NewClient() returned nil client with no error")
			}
		})
	}
}

func TestNewClientWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with defaults",
			cfg: Config{
				APIKey: "test-key",
				Model:  "claude-haiku-4-5",
			},
			wantErr: false,
		},
		{
			name: "valid config with custom values",
			cfg: Config{
				APIKey:      "test-key",
				Model:       "claude-haiku-4-5",
				MaxTokens:   1000,
				Temperature: 0.5,
			},
			wantErr: false,
		},
		{
			name: "max tokens exceeds limit",
			cfg: Config{
				APIKey:    "test-key",
				Model:     "claude-haiku-4-5",
				MaxTokens: 5000, // exceeds 4096 limit
			},
			wantErr: true,
			errMsg:  "max tokens must be <= 4096",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithConfig(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("NewClientWithConfig() expected error, got nil")
					return
				}
				return
			}
			if err != nil {
				t.Errorf("NewClientWithConfig() unexpected error = %v", err)
				return
			}
			if client == nil {
				t.Error("NewClientWithConfig() returned nil client with no error")
			}
		})
	}
}

func TestSendMessage_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers (use canonical form for Get)
		//nolint:canonicalheader // Anthropic API uses lowercase headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing or incorrect x-api-key header")
		}
		//nolint:canonicalheader // Anthropic API uses lowercase headers
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("missing or incorrect anthropic-version header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing or incorrect Content-Type header")
		}

		// Return mock response
		resp := anthropicResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []contentBlock{
				{
					Type: "text",
					Text: "Hello! How can I help you today?",
				},
			},
			StopReason: "end_turn",
			Usage:      usage{InputTokens: 100, OutputTokens: 50},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test helper, error irrelevant
	}))
	defer server.Close()

	// Create client and override URL
	client, err := NewClient("test-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	// Send message
	ctx := context.Background()
	resp, err := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "You are a helpful assistant",
		UserMessage:  "Hello!",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	// Verify response
	if resp.Content != "Hello! How can I help you today?" {
		t.Errorf("Content = %v, want 'Hello! How can I help you today?'", resp.Content)
	}
	if resp.Usage.InputTokens != 100 {
		t.Errorf("InputTokens = %v, want 100", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 50 {
		t.Errorf("OutputTokens = %v, want 50", resp.Usage.OutputTokens)
	}
}

func TestSendMessage_JSONContent(t *testing.T) {
	// Test that JSON content is returned as raw string for caller to parse
	jsonContent := `{"status":"approved","data":{"value":"test"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []contentBlock{
				{
					Type: "text",
					Text: jsonContent,
				},
			},
			StopReason: "end_turn",
			Usage:      usage{InputTokens: 50, OutputTokens: 25},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test helper, error irrelevant
	}))
	defer server.Close()

	client, err := NewClient("test-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	resp, err := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "Return JSON",
		UserMessage:  "test",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	// Verify raw content is returned - caller can parse as needed
	if resp.Content != jsonContent {
		t.Errorf("Content = %v, want %v", resp.Content, jsonContent)
	}
}

func TestSendMessage_RateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate_limited"}`)) //nolint:errcheck // Test helper, error irrelevant
	}))
	defer server.Close()

	client, err := NewClient("test-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	_, sendErr := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "test",
		UserMessage:  "test",
	})

	if !IsRateLimitError(sendErr) {
		t.Errorf("Expected RateLimitError, got %T: %v", sendErr, sendErr)
	}
	if !IsRetryableError(sendErr) {
		t.Error("RateLimitError should be retryable")
	}
}

func TestSendMessage_UnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_api_key"}`)) //nolint:errcheck // Test helper, error irrelevant
	}))
	defer server.Close()

	client, err := NewClient("bad-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	_, sendErr := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "test",
		UserMessage:  "test",
	})

	if !IsAPIError(sendErr) {
		t.Errorf("Expected APIError, got %T: %v", sendErr, sendErr)
	}
	if IsRetryableError(sendErr) {
		t.Error("Unauthorized error should not be retryable")
	}
}

func TestSendMessage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error"}`)) //nolint:errcheck // Test helper, error irrelevant
	}))
	defer server.Close()

	client, err := NewClient("test-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	_, sendErr := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "test",
		UserMessage:  "test",
	})

	if !IsAPIError(sendErr) {
		t.Errorf("Expected APIError, got %T: %v", sendErr, sendErr)
	}
	if !IsRetryableError(sendErr) {
		t.Error("5xx error should be retryable")
	}
}

func TestSendMessage_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			ID:         "msg_123",
			Type:       "message",
			Role:       "assistant",
			Content:    []contentBlock{}, // Empty content
			StopReason: "end_turn",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // Test helper, error irrelevant
	}))
	defer server.Close()

	client, err := NewClient("test-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	_, sendErr := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "test",
		UserMessage:  "test",
	})

	if sendErr != ErrEmptyResponse {
		t.Errorf("Expected ErrEmptyResponse, got %v", sendErr)
	}
}

func TestSendMessage_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient("test-key", "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetBaseURL(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, sendErr := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "test",
		UserMessage:  "test",
	})

	if sendErr == nil {
		t.Error("Expected error for canceled context")
	}
}

// =============================================================================
// Integration Test (requires real API key)
// =============================================================================

// TestIntegration_SendMessage is an integration test that calls the real Anthropic API.
// Run with: ANTHROPIC_APIKEY=your-key go test -v -run TestIntegration
func TestIntegration_SendMessage(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_APIKEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: ANTHROPIC_APIKEY not set")
	}

	client, err := NewClient(apiKey, "claude-haiku-4-5")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.SendMessage(ctx, MessageRequest{
		SystemPrompt: "You are a helpful assistant. Be concise.",
		UserMessage:  "Say hello in exactly 3 words.",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	t.Logf("Content: %s", resp.Content)
	t.Logf("Usage: input=%d, output=%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)

	if resp.Content == "" {
		t.Error("Expected non-empty content")
	}
	if resp.Usage.InputTokens == 0 {
		t.Error("Expected non-zero input tokens")
	}
	if resp.Usage.OutputTokens == 0 {
		t.Error("Expected non-zero output tokens")
	}
}
