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

// MarshalJSON implements json.Marshaler interface.
func (t TelegramChatID) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
// Accepts both number and string formats: 123456789 or "123456789"
func (t *TelegramChatID) UnmarshalJSON(data []byte) error {
	// Try parsing as number first
	value, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		// Try parsing as quoted string
		s := string(data)
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			value, err = strconv.ParseInt(s[1:len(s)-1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid telegram chat ID: %w", err)
			}
		} else {
			return fmt.Errorf("invalid telegram chat ID format: %s", s)
		}
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

// =============================================================================
// Null represents a nullable TelegramChatID for optional fields.
// =============================================================================

// Null represents a nullable TelegramChatID.
type Null struct {
	ChatID TelegramChatID
	Valid  bool // Valid is true if ChatID is set
}

// NewNull creates a valid Null with the given TelegramChatID.
func NewNull(chatID TelegramChatID) Null {
	return Null{
		ChatID: chatID,
		Valid:  true,
	}
}

// NullFromInt64 parses an int64 and creates a Null.
// If value is 0, returns an invalid (null) Null.
func NullFromInt64(value int64) (Null, error) {
	if value == 0 {
		return Null{}, nil
	}

	chatID, err := Parse(value)
	if err != nil {
		return Null{}, err
	}

	return NewNull(chatID), nil
}

// Value returns the int64 value, or 0 if null.
func (n Null) Value() int64 {
	if !n.Valid {
		return 0
	}
	return n.ChatID.Value()
}

// String returns the string representation, or empty string if null.
func (n Null) String() string {
	if !n.Valid {
		return ""
	}
	return n.ChatID.String()
}

// IsNull returns true if this Null represents a null value.
func (n Null) IsNull() bool {
	return !n.Valid
}

// Equal provides support for the go-cmp package and testing.
func (n Null) Equal(n2 Null) bool {
	if n.Valid != n2.Valid {
		return false
	}
	if !n.Valid {
		return true // Both are null
	}
	return n.ChatID.Equal(n2.ChatID)
}

// MarshalText provides support for logging and any marshal needs.
func (n Null) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.ChatID.MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// Treats "null" or empty string as a null value.
func (n *Null) UnmarshalText(data []byte) error {
	s := string(data)

	// Treat "null" or empty string as null
	if s == "" || s == "null" {
		*n = Null{}
		return nil
	}

	// Delegate to TelegramChatID parsing
	var chatID TelegramChatID
	if err := chatID.UnmarshalText(data); err != nil {
		return err
	}

	*n = NewNull(chatID)
	return nil
}
