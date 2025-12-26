// Package telegramapp provides HTTP handlers for Telegram webhook integration.
package telegramapp

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/telegramchatid"
)

// TelegramUpdate represents an incoming update from Telegram Bot API.
// See: https://core.telegram.org/bots/api#update
type TelegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// Decode implements the web.Decoder interface.
func (tu *TelegramUpdate) Decode(data []byte) error {
	return json.Unmarshal(data, tu)
}

// TelegramMessage represents a message from Telegram.
type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      TelegramChat  `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text,omitempty"`
}

// TelegramUser represents a Telegram user.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat.
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// TelegramMessageArgs contains data for async job processing.
type TelegramMessageArgs struct {
	SessionID uuid.UUID                       `json:"session_id"`
	UserID    uuid.UUID                       `json:"user_id"`
	ChatID    telegramchatid.TelegramChatID   `json:"chat_id"`
	Text      string                          `json:"text"`
}
