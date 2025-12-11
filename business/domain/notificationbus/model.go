package notificationbus

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound           = errors.New("message not found")
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrContentEmpty       = errors.New("content cannot be empty")
	ErrDuplicateSchedule  = errors.New("message already scheduled for this user/type/date")
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

// Valid returns true if the message type is a known valid type.
func (mt MessageType) Valid() bool {
	switch mt {
	case MessageTypeMorning, MessageTypeEvening, MessageTypeTest, MessageTypeWelcome:
		return true
	}
	return false
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

// Valid returns true if the message status is a known valid status.
func (ms MessageStatus) Valid() bool {
	switch ms {
	case StatusPending, StatusSent, StatusFailed:
		return true
	}
	return false
}

// TelegramMessageID represents a Telegram message ID.
type TelegramMessageID int64

// Value returns the int64 value of the Telegram message ID.
func (t TelegramMessageID) Value() int64 {
	return int64(t)
}

// TelegramChatID represents a Telegram chat ID.
type TelegramChatID int64

// Value returns the int64 value of the Telegram chat ID.
func (t TelegramChatID) Value() int64 {
	return int64(t)
}

// FailureReason represents the reason for a message failure.
type FailureReason string

// String returns the string representation of the failure reason.
func (f FailureReason) String() string {
	return string(f)
}

// Message represents a single notification message.
type Message struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	MessageType   MessageType
	Content       string
	TelegramMsgID *TelegramMessageID
	Status        MessageStatus
	ErrorMessage  *FailureReason
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
	TelegramChatID TelegramChatID
}
