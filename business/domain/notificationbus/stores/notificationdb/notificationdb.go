package notificationdb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/notificationbus"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages the set of APIs for notification database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create adds a new notification message to the database.
func (s *Store) Create(ctx context.Context, msg notificationbus.Message) error {
	const q = `
	INSERT INTO notification_messages (
		message_id, user_id, message_type, content, telegram_msg_id,
		status, error_message, retry_count, scheduled_at, sent_at, date_created
	) VALUES (
		:message_id, :user_id, :message_type, :content, :telegram_msg_id,
		:status, :error_message, :retry_count, :scheduled_at, :sent_at, :date_created
	)`

	dbMsg := toDBMessage(msg)

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbMsg); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a notification message in the database.
func (s *Store) Update(ctx context.Context, msg notificationbus.Message) error {
	const q = `
	UPDATE notification_messages SET
		telegram_msg_id = :telegram_msg_id,
		status = :status,
		error_message = :error_message,
		retry_count = :retry_count,
		sent_at = :sent_at
	WHERE
		message_id = :message_id`

	dbMsg := toDBMessage(msg)

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbMsg); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID retrieves a single message by its ID.
func (s *Store) QueryByID(ctx context.Context, messageID uuid.UUID) (notificationbus.Message, error) {
	data := map[string]any{
		"message_id": messageID,
	}

	const q = `
	SELECT
		message_id, user_id, message_type, content, telegram_msg_id,
		status, error_message, retry_count, scheduled_at, sent_at, date_created
	FROM
		notification_messages
	WHERE
		message_id = :message_id`

	var dbMsg message
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbMsg); err != nil {
		return notificationbus.Message{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusMessage(dbMsg), nil
}

// QueryPending returns all pending messages scheduled before the given time.
func (s *Store) QueryPending(ctx context.Context, before time.Time) ([]notificationbus.Message, error) {
	data := map[string]any{
		"before": before.UTC(),
	}

	const q = `
	SELECT
		message_id, user_id, message_type, content, telegram_msg_id,
		status, error_message, retry_count, scheduled_at, sent_at, date_created
	FROM
		notification_messages
	WHERE
		status = 'pending' AND scheduled_at <= :before
	ORDER BY
		scheduled_at ASC`

	var dbMsgs []message
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbMsgs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusMessages(dbMsgs), nil
}

// QueryTelegramUsers returns all users with Telegram enabled.
func (s *Store) QueryTelegramUsers(ctx context.Context) ([]notificationbus.TelegramUser, error) {
	const q = `
	SELECT
		user_id, telegram_chat_id
	FROM
		users
	WHERE
		telegram_enabled = true AND telegram_chat_id IS NOT NULL`

	var dbUsers []telegramUser
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, map[string]any{}, &dbUsers); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTelegramUsers(dbUsers), nil
}
