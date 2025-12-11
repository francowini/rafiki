package notificationbus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vnotificationbus"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// VNotificationQuerier interface for querying values (dependency injection).
type VNotificationQuerier interface {
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]vnotificationbus.ValueWithVision, error)
}

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
	// Validate MessageType
	if !nm.MessageType.Valid() {
		b.log.Error(ctx, "notificationbus.create", "err", ErrInvalidMessageType, "message_type", nm.MessageType)
		return Message{}, ErrInvalidMessageType
	}

	// Validate Content
	if strings.TrimSpace(nm.Content) == "" {
		b.log.Error(ctx, "notificationbus.create", "err", ErrContentEmpty, "user_id", nm.UserID)
		return Message{}, ErrContentEmpty
	}

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
		// Don't log duplicate schedule as error - it's expected behavior
		if !errors.Is(err, ErrDuplicateSchedule) {
			b.log.Error(ctx, "notificationbus.create", "err", err, "message_id", msg.ID, "user_id", msg.UserID)
		}
		return Message{}, fmt.Errorf("create: %w", err)
	}

	b.log.Info(ctx, "notificationbus.create", "message_id", msg.ID, "user_id", msg.UserID, "type", msg.MessageType)
	return msg, nil
}

// MarkSent marks a message as successfully sent.
func (b *Business) MarkSent(ctx context.Context, msg Message, telegramMsgID TelegramMessageID) (Message, error) {
	now := time.Now()
	msg.Status = StatusSent
	msg.TelegramMsgID = &telegramMsgID
	msg.SentAt = &now

	if err := b.storer.Update(ctx, msg); err != nil {
		b.log.Error(ctx, "notificationbus.markSent", "err", err, "message_id", msg.ID)
		return Message{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "notificationbus.markSent", "message_id", msg.ID, "telegram_msg_id", telegramMsgID)
	return msg, nil
}

// MarkFailed marks a message as failed with an error message.
func (b *Business) MarkFailed(ctx context.Context, msg Message, errMsg FailureReason) (Message, error) {
	msg.Status = StatusFailed
	msg.ErrorMessage = &errMsg
	msg.RetryCount++

	if err := b.storer.Update(ctx, msg); err != nil {
		b.log.Error(ctx, "notificationbus.markFailed", "err", err, "message_id", msg.ID)
		return Message{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "notificationbus.markFailed", "message_id", msg.ID, "reason", errMsg, "retry_count", msg.RetryCount)
	return msg, nil
}

// QueryByID finds the message by the specified ID.
func (b *Business) QueryByID(ctx context.Context, messageID uuid.UUID) (Message, error) {
	msg, err := b.storer.QueryByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Message{}, ErrNotFound
		}
		b.log.Error(ctx, "notificationbus.queryByID", "err", err, "message_id", messageID)
		return Message{}, fmt.Errorf("query: %w", err)
	}

	return msg, nil
}

// QueryPending returns all pending messages scheduled before the given time.
func (b *Business) QueryPending(ctx context.Context, before time.Time) ([]Message, error) {
	msgs, err := b.storer.QueryPending(ctx, before)
	if err != nil {
		b.log.Error(ctx, "notificationbus.queryPending", "err", err, "before", before)
		return nil, fmt.Errorf("query pending: %w", err)
	}

	return msgs, nil
}

// QueryTelegramUsers returns all users with Telegram enabled.
func (b *Business) QueryTelegramUsers(ctx context.Context) ([]TelegramUser, error) {
	users, err := b.storer.QueryTelegramUsers(ctx)
	if err != nil {
		b.log.Error(ctx, "notificationbus.queryTelegramUsers", "err", err)
		return nil, fmt.Errorf("query telegram users: %w", err)
	}

	return users, nil
}

// ============================================================================
// Scheduling and Sending Logic (moved from job layer)
// ============================================================================

// ScheduleMorningMessageIfTime schedules a morning message if current time is in the morning window.
// Returns nil if already scheduled (idempotent behavior) or not in time window.
func (b *Business) ScheduleMorningMessageIfTime(
	ctx context.Context,
	userID uuid.UUID,
	cfg ScheduleConfig,
	now time.Time,
	vnotificationQuerier VNotificationQuerier,
) error {
	currentTime := now.Format("15:04")
	if !b.IsInTimeWindow(currentTime, cfg.MorningTime) {
		return nil
	}

	// Fetch user's values for message content
	values, err := vnotificationQuerier.QueryByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("query values: %w", err)
	}

	// Generate message content
	content := GenerateMorningMessage(values)

	// Create notification (idempotent - handles duplicates)
	_, err = b.Create(ctx, NewMessage{
		UserID:      userID,
		MessageType: MessageTypeMorning,
		Content:     content,
		ScheduledAt: now,
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateSchedule) {
			b.log.Info(ctx, "notificationbus.scheduleMorning", "msg", "already scheduled", "user_id", userID)
			return nil
		}
		return fmt.Errorf("create morning message: %w", err)
	}

	b.log.Info(ctx, "notificationbus.scheduleMorning", "msg", "scheduled", "user_id", userID)
	return nil
}

// ScheduleEveningMessageIfTime schedules an evening message if current time is in the evening window.
// Returns nil if already scheduled (idempotent behavior) or not in time window.
func (b *Business) ScheduleEveningMessageIfTime(
	ctx context.Context,
	userID uuid.UUID,
	cfg ScheduleConfig,
	now time.Time,
	vnotificationQuerier VNotificationQuerier,
) error {
	currentTime := now.Format("15:04")
	if !b.IsInTimeWindow(currentTime, cfg.EveningTime) {
		return nil
	}

	// Fetch user's values for message content
	values, err := vnotificationQuerier.QueryByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("query values: %w", err)
	}

	// Generate message content
	content := GenerateEveningMessage(values)

	// Create notification (idempotent - handles duplicates)
	_, err = b.Create(ctx, NewMessage{
		UserID:      userID,
		MessageType: MessageTypeEvening,
		Content:     content,
		ScheduledAt: now,
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateSchedule) {
			b.log.Info(ctx, "notificationbus.scheduleEvening", "msg", "already scheduled", "user_id", userID)
			return nil
		}
		return fmt.Errorf("create evening message: %w", err)
	}

	b.log.Info(ctx, "notificationbus.scheduleEvening", "msg", "scheduled", "user_id", userID)
	return nil
}

// IsInTimeWindow checks if current time is within 10 minutes of target time.
func (b *Business) IsInTimeWindow(currentTime, targetTime string) bool {
	current, err := time.Parse("15:04", currentTime)
	if err != nil {
		return false
	}

	target, err := time.Parse("15:04", targetTime)
	if err != nil {
		return false
	}

	// Calculate difference in minutes (0 to +9 minutes after target)
	diff := current.Sub(target).Minutes()
	return diff >= 0 && diff < 10
}

// SendPendingMessages sends all pending messages scheduled before now.
func (b *Business) SendPendingMessages(
	ctx context.Context,
	telegramSender TelegramSender,
	before time.Time,
) error {
	// Get pending messages
	pending, err := b.QueryPending(ctx, before)
	if err != nil {
		return fmt.Errorf("query pending: %w", err)
	}

	if len(pending) == 0 {
		return nil
	}

	b.log.Info(ctx, "notificationbus.sendPending", "msg", "processing", "count", len(pending))

	// Get telegram users for chat ID lookup
	users, err := b.QueryTelegramUsers(ctx)
	if err != nil {
		return fmt.Errorf("query telegram users: %w", err)
	}

	// Build chat ID map
	chatIDMap := make(map[uuid.UUID]TelegramChatID)
	for _, u := range users {
		chatIDMap[u.UserID] = u.TelegramChatID
	}

	// Send each message
	for _, msg := range pending {
		chatID, ok := chatIDMap[msg.UserID]
		if !ok {
			b.log.Error(ctx, "notificationbus.sendPending", "msg", "user not in telegram map", "user_id", msg.UserID)
			errMsg := FailureReason("user telegram not enabled")
			if _, markErr := b.MarkFailed(ctx, msg, errMsg); markErr != nil {
				b.log.Error(ctx, "notificationbus.sendPending", "msg", "mark failed error", "err", markErr)
			}
			continue
		}

		b.log.Info(ctx, "notificationbus.sendPending", "msg", "sending", "message_id", msg.ID, "chat_id", chatID.Value())

		resp, err := telegramSender.SendMessage(ctx, chatID.Value(), msg.Content)
		if err != nil {
			b.log.Error(ctx, "notificationbus.sendPending", "msg", "send failed", "message_id", msg.ID, "err", err)
			errMsg := FailureReason(err.Error())
			if _, markErr := b.MarkFailed(ctx, msg, errMsg); markErr != nil {
				b.log.Error(ctx, "notificationbus.sendPending", "msg", "mark failed error", "err", markErr)
			}
			continue
		}

		telegramMsgID := TelegramMessageID(resp.MessageID)
		if _, err := b.MarkSent(ctx, msg, telegramMsgID); err != nil {
			b.log.Error(ctx, "notificationbus.sendPending", "msg", "mark sent error", "err", err)
		}

		b.log.Info(ctx, "notificationbus.sendPending", "msg", "sent", "message_id", msg.ID, "telegram_msg_id", resp.MessageID)
	}

	return nil
}
