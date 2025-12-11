// Package telegram provides a client for the Telegram Bot API.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://api.telegram.org"

// ErrEmptyBotToken is returned when bot token is empty.
var ErrEmptyBotToken = errors.New("telegram: bot token cannot be empty")

// Client for Telegram Bot API.
type Client struct {
	botToken   string
	httpClient *http.Client
}

// NewClient creates a Telegram client. Returns error if botToken is empty.
func NewClient(botToken string) (*Client, error) {
	if strings.TrimSpace(botToken) == "" {
		return nil, ErrEmptyBotToken
	}

	return &Client{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// SendMessageRequest for Telegram API.
type SendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendMessageResponse from Telegram API.
type SendMessageResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
	Description string `json:"description,omitempty"`
}

// SendMessage sends a text message to a Telegram chat.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) (*SendMessageResponse, error) {
	req := SendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", baseURL, c.botToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort cleanup

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check for non-2xx responses (infrastructure errors like 502, 503, etc.)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Truncate body for error message (max 200 chars)
		bodySnippet := string(respBody)
		if len(bodySnippet) > 200 {
			bodySnippet = bodySnippet[:200] + "..."
		}
		return nil, fmt.Errorf("telegram http error: status=%d %s, body=%s",
			resp.StatusCode, http.StatusText(resp.StatusCode), bodySnippet)
	}

	var apiResp SendMessageResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("telegram api error: %s", apiResp.Description)
	}

	return &apiResp, nil
}
