// Package telegramchatid represents a validated Telegram chat_id (non-zero int64).
package telegramchatid

import (
	"fmt"
	"strconv"
)

// TelegramChatID represents a validated Telegram chat_id (non-zero int64).
// Telegram chat IDs can be negative (groups) or positive (users), but never zero.
type TelegramChatID struct {
	value int64
}

// Value returns the int64 value of the chat ID.
func (t TelegramChatID) Value() int64 {
	return t.value
}

// String returns the string representation.
func (t TelegramChatID) String() string {
	return fmt.Sprintf("%d", t.value)
}

// Equal provides support for the go-cmp package and testing.
func (t TelegramChatID) Equal(t2 TelegramChatID) bool {
	return t.value == t2.value
}

// MarshalText provides support for logging and any marshal needs.
func (t TelegramChatID) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TelegramChatID) UnmarshalText(data []byte) error {
	value, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

// =============================================================================

// Parse validates and creates a TelegramChatID (non-zero int64).
func Parse(value int64) (TelegramChatID, error) {
	if value == 0 {
		return TelegramChatID{}, fmt.Errorf("telegram chat_id cannot be zero")
	}

	return TelegramChatID{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int64) TelegramChatID {
	chatID, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return chatID
}
