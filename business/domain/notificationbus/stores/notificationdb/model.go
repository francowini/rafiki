package notificationdb

import (
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/notificationbus"
)

type message struct {
	ID            uuid.UUID  `db:"message_id"`
	UserID        uuid.UUID  `db:"user_id"`
	MessageType   string     `db:"message_type"`
	Content       string     `db:"content"`
	TelegramMsgID *int64     `db:"telegram_msg_id"`
	Status        string     `db:"status"`
	ErrorMessage  *string    `db:"error_message"`
	RetryCount    int        `db:"retry_count"`
	ScheduledAt   time.Time  `db:"scheduled_at"`
	SentAt        *time.Time `db:"sent_at"`
	DateCreated   time.Time  `db:"date_created"`
}

func toDBMessage(bus notificationbus.Message) message {
	var telegramMsgID *int64
	if bus.TelegramMsgID != nil {
		v := bus.TelegramMsgID.Value()
		telegramMsgID = &v
	}

	var errorMessage *string
	if bus.ErrorMessage != nil {
		v := bus.ErrorMessage.String()
		errorMessage = &v
	}

	return message{
		ID:            bus.ID,
		UserID:        bus.UserID,
		MessageType:   bus.MessageType.String(),
		Content:       bus.Content,
		TelegramMsgID: telegramMsgID,
		Status:        bus.Status.String(),
		ErrorMessage:  errorMessage,
		RetryCount:    bus.RetryCount,
		ScheduledAt:   bus.ScheduledAt.UTC(),
		SentAt:        toUTCPtr(bus.SentAt),
		DateCreated:   bus.DateCreated.UTC(),
	}
}

func toBusMessage(db message) notificationbus.Message {
	// Parse and validate MessageType (default to test if invalid)
	messageType := notificationbus.MessageType(db.MessageType)
	if !messageType.Valid() {
		messageType = notificationbus.MessageTypeTest
	}

	// Parse and validate MessageStatus (default to pending if invalid)
	status := notificationbus.MessageStatus(db.Status)
	if !status.Valid() {
		status = notificationbus.StatusPending
	}

	return notificationbus.Message{
		ID:            db.ID,
		UserID:        db.UserID,
		MessageType:   messageType,
		Content:       db.Content,
		TelegramMsgID: toTelegramMsgIDPtr(db.TelegramMsgID),
		Status:        status,
		ErrorMessage:  toFailureReasonPtr(db.ErrorMessage),
		RetryCount:    db.RetryCount,
		ScheduledAt:   db.ScheduledAt.UTC(),
		SentAt:        toUTCPtrRead(db.SentAt),
		DateCreated:   db.DateCreated.UTC(),
	}
}

func toBusMessages(dbs []message) []notificationbus.Message {
	msgs := make([]notificationbus.Message, len(dbs))
	for i, db := range dbs {
		msgs[i] = toBusMessage(db)
	}
	return msgs
}

type telegramUser struct {
	UserID         uuid.UUID `db:"user_id"`
	TelegramChatID int64     `db:"telegram_chat_id"`
}

func toBusTelegramUsers(dbs []telegramUser) []notificationbus.TelegramUser {
	users := make([]notificationbus.TelegramUser, len(dbs))
	for i, db := range dbs {
		users[i] = notificationbus.TelegramUser{
			UserID:         db.UserID,
			TelegramChatID: notificationbus.TelegramChatID(db.TelegramChatID),
		}
	}
	return users
}

func toTelegramMsgIDPtr(v *int64) *notificationbus.TelegramMessageID {
	if v == nil {
		return nil
	}
	t := notificationbus.TelegramMessageID(*v)
	return &t
}

func toFailureReasonPtr(v *string) *notificationbus.FailureReason {
	if v == nil {
		return nil
	}
	f := notificationbus.FailureReason(*v)
	return &f
}

func toUTCPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}

func toUTCPtrRead(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}
