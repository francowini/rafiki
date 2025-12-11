package notificationbus

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrInvalidTimeFormat is returned when a time string is not in HH:MM format.
var ErrInvalidTimeFormat = errors.New("time must be in HH:MM format (00:00-23:59)")

// TimeOfDay represents a validated time of day in HH:MM format.
type TimeOfDay struct {
	value string
}

// ParseTimeOfDay validates and creates a TimeOfDay from a string.
// The input must be in HH:MM format (e.g., "08:00", "21:30").
// Returns an error if the format is invalid.
func ParseTimeOfDay(value string) (TimeOfDay, error) {
	if _, err := time.Parse("15:04", value); err != nil {
		return TimeOfDay{}, fmt.Errorf("%w: got %q, parse error: %v", ErrInvalidTimeFormat, value, err)
	}
	return TimeOfDay{value: value}, nil
}

// MustParseTimeOfDay is like ParseTimeOfDay but panics on error.
// Use in tests or for known-valid constants only.
func MustParseTimeOfDay(value string) TimeOfDay {
	tod, err := ParseTimeOfDay(value)
	if err != nil {
		panic(err)
	}
	return tod
}

// String returns the HH:MM string representation.
func (t TimeOfDay) String() string {
	return t.value
}

// Value returns the underlying string value.
func (t TimeOfDay) Value() string {
	return t.value
}

// IsZero returns true if the TimeOfDay is uninitialized.
func (t TimeOfDay) IsZero() bool {
	return t.value == ""
}

// Equal provides support for the go-cmp package and testing.
func (t TimeOfDay) Equal(t2 TimeOfDay) bool {
	return t.value == t2.value
}

// MarshalText provides support for logging and any marshal needs.
func (t TimeOfDay) MarshalText() ([]byte, error) {
	return []byte(t.value), nil
}

// ScheduleConfig holds scheduling configuration for notifications.
type ScheduleConfig struct {
	MorningTime TimeOfDay      // Validated HH:MM format
	EveningTime TimeOfDay      // Validated HH:MM format
	Location    *time.Location // Timezone for scheduling
}

// NewScheduleConfig creates a validated ScheduleConfig.
// Returns an error if morning or evening times are invalid.
// Location can be nil (will use UTC as default in business logic).
func NewScheduleConfig(morningTime, eveningTime string, location *time.Location) (ScheduleConfig, error) {
	morning, err := ParseTimeOfDay(morningTime)
	if err != nil {
		return ScheduleConfig{}, fmt.Errorf("morning time: %w", err)
	}

	evening, err := ParseTimeOfDay(eveningTime)
	if err != nil {
		return ScheduleConfig{}, fmt.Errorf("evening time: %w", err)
	}

	return ScheduleConfig{
		MorningTime: morning,
		EveningTime: evening,
		Location:    location,
	}, nil
}

// TelegramSender interface for dependency injection.
type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, content string) (TelegramSendResponse, error)
}

// TelegramSendResponse represents the response from sending a Telegram message.
type TelegramSendResponse struct {
	MessageID int64
}
