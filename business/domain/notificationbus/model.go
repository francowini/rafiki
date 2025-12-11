package notificationbus

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("message not found")
)

// MessageType represents the type of notification message.
type MessageType string

// Message type constants.
const (
	MessageTypeMorning MessageType = "morning"
	MessageTypeEvening MessageType = "evening"
	MessageTypeTest    MessageType = "test"
	MessageTypeWelcome MessageType = "welcome"
)

// String returns the string representation of the message type.
func (mt MessageType) String() string {
	return string(mt)
}

// MessageStatus represents delivery status.
type MessageStatus string

// Message status constants.
const (
	StatusPending MessageStatus = "pending"
	StatusSent    MessageStatus = "sent"
	StatusFailed  MessageStatus = "failed"
)

// String returns the string representation of the message status.
func (ms MessageStatus) String() string {
	return string(ms)
}

// Message represents a single notification message.
type Message struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	MessageType   MessageType
	Content       string
	TelegramMsgID *int64
	Status        MessageStatus
	ErrorMessage  *string
	RetryCount    int
	ScheduledAt   time.Time
	SentAt        *time.Time
	DateCreated   time.Time
}

// NewMessage for creating a message.
type NewMessage struct {
	UserID      uuid.UUID
	MessageType MessageType
	Content     string
	ScheduledAt time.Time
}

// TelegramUser represents a user with Telegram enabled (for queries).
type TelegramUser struct {
	UserID         uuid.UUID
	TelegramChatID int64
}
