package notificationbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, msg Message) error
	Update(ctx context.Context, msg Message) error
	QueryByID(ctx context.Context, messageID uuid.UUID) (Message, error)
	QueryPending(ctx context.Context, before time.Time) ([]Message, error)
	QueryTelegramUsers(ctx context.Context) ([]TelegramUser, error)
}

// Business manages the set of APIs for notification access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a notification business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// Create adds a new notification message to the system.
func (b *Business) Create(ctx context.Context, nm NewMessage) (Message, error) {
	now := time.Now()

	msg := Message{
		ID:          uuid.New(),
		UserID:      nm.UserID,
		MessageType: nm.MessageType,
		Content:     nm.Content,
		Status:      StatusPending,
		RetryCount:  0,
		ScheduledAt: nm.ScheduledAt,
		DateCreated: now,
	}

	if err := b.storer.Create(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("create: %w", err)
	}

	return msg, nil
}

// MarkSent marks a message as successfully sent.
func (b *Business) MarkSent(ctx context.Context, msg Message, telegramMsgID int64) (Message, error) {
	now := time.Now()
	msg.Status = StatusSent
	msg.TelegramMsgID = &telegramMsgID
	msg.SentAt = &now

	if err := b.storer.Update(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("update: %w", err)
	}

	return msg, nil
}

// MarkFailed marks a message as failed with an error message.
func (b *Business) MarkFailed(ctx context.Context, msg Message, errMsg string) (Message, error) {
	msg.Status = StatusFailed
	msg.ErrorMessage = &errMsg
	msg.RetryCount++

	if err := b.storer.Update(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("update: %w", err)
	}

	return msg, nil
}

// QueryByID finds the message by the specified ID.
func (b *Business) QueryByID(ctx context.Context, messageID uuid.UUID) (Message, error) {
	msg, err := b.storer.QueryByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Message{}, ErrNotFound
		}
		return Message{}, fmt.Errorf("query: %w", err)
	}

	return msg, nil
}

// QueryPending returns all pending messages scheduled before the given time.
func (b *Business) QueryPending(ctx context.Context, before time.Time) ([]Message, error) {
	msgs, err := b.storer.QueryPending(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("query pending: %w", err)
	}

	return msgs, nil
}

// QueryTelegramUsers returns all users with Telegram enabled.
func (b *Business) QueryTelegramUsers(ctx context.Context) ([]TelegramUser, error) {
	users, err := b.storer.QueryTelegramUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("query telegram users: %w", err)
	}

	return users, nil
}
