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
	return message{
		ID:            bus.ID,
		UserID:        bus.UserID,
		MessageType:   bus.MessageType.String(),
		Content:       bus.Content,
		TelegramMsgID: bus.TelegramMsgID,
		Status:        bus.Status.String(),
		ErrorMessage:  bus.ErrorMessage,
		RetryCount:    bus.RetryCount,
		ScheduledAt:   bus.ScheduledAt.UTC(),
		SentAt:        toUTCPtr(bus.SentAt),
		DateCreated:   bus.DateCreated.UTC(),
	}
}

func toBusMessage(db message) notificationbus.Message {
	return notificationbus.Message{
		ID:            db.ID,
		UserID:        db.UserID,
		MessageType:   notificationbus.MessageType(db.MessageType),
		Content:       db.Content,
		TelegramMsgID: db.TelegramMsgID,
		Status:        notificationbus.MessageStatus(db.Status),
		ErrorMessage:  db.ErrorMessage,
		RetryCount:    db.RetryCount,
		ScheduledAt:   db.ScheduledAt.In(time.Local),
		SentAt:        toLocalPtr(db.SentAt),
		DateCreated:   db.DateCreated.In(time.Local),
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
			TelegramChatID: db.TelegramChatID,
		}
	}
	return users
}

func toUTCPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}

func toLocalPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	local := t.In(time.Local)
	return &local
}
