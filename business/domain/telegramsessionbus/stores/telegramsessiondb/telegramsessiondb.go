package telegramsessiondb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/sessiontype"
	"github.com/francowini/rafiki/business/types/telegramchatid"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages the set of APIs for telegram session database access.
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

// Create adds a new session to the database.
func (s *Store) Create(ctx context.Context, sess telegramsessionbus.Session) error {
	const q = `
	INSERT INTO telegram_sessions (
		session_id, user_id, chat_id, session_type, current_step,
		total_steps, retry_count, context_data, last_activity,
		date_created, date_updated
	) VALUES (
		:session_id, :user_id, :chat_id, :session_type, :current_step,
		:total_steps, :retry_count, :context_data, :last_activity,
		:date_created, :date_updated
	)`

	dbSession, err := toDBSession(sess)
	if err != nil {
		return fmt.Errorf("to db session: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbSession); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return telegramsessionbus.ErrSessionExists
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about a session in the database.
func (s *Store) Update(ctx context.Context, sess telegramsessionbus.Session) error {
	const q = `
	UPDATE telegram_sessions SET
		current_step = :current_step,
		retry_count = :retry_count,
		context_data = :context_data,
		last_activity = :last_activity,
		date_updated = :date_updated
	WHERE
		session_id = :session_id`

	dbSession, err := toDBSession(sess)
	if err != nil {
		return fmt.Errorf("to db session: %w", err)
	}

	rowsAffected, err := sqldb.NamedExecContextWithResult(ctx, s.log, s.db, q, dbSession)
	if err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	if rowsAffected == 0 {
		return sqldb.ErrDBNotFound
	}

	return nil
}

// Delete removes a session from the database.
func (s *Store) Delete(ctx context.Context, sess telegramsessionbus.Session) error {
	const q = `
	DELETE FROM telegram_sessions
	WHERE session_id = :session_id`

	data := struct {
		ID string `db:"session_id"`
	}{
		ID: sess.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID retrieves a session by its unique ID.
func (s *Store) QueryByID(ctx context.Context, sessionID uuid.UUID) (telegramsessionbus.Session, error) {
	data := map[string]any{
		"session_id": sessionID,
	}

	const q = `
	SELECT
		session_id, user_id, chat_id, session_type, current_step,
		total_steps, retry_count, context_data, last_activity,
		date_created, date_updated
	FROM
		telegram_sessions
	WHERE
		session_id = :session_id`

	var dbSess session
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbSess); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return telegramsessionbus.Session{}, sqldb.ErrDBNotFound
		}
		return telegramsessionbus.Session{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSession(dbSess)
}

// QueryByUserAndType retrieves an active session for a user and session type.
func (s *Store) QueryByUserAndType(ctx context.Context, userID uuid.UUID, sessionType sessiontype.SessionType) (telegramsessionbus.Session, error) {
	data := map[string]any{
		"user_id":      userID,
		"session_type": sessionType.Value(),
	}

	const q = `
	SELECT
		session_id, user_id, chat_id, session_type, current_step,
		total_steps, retry_count, context_data, last_activity,
		date_created, date_updated
	FROM
		telegram_sessions
	WHERE
		user_id = :user_id AND session_type = :session_type`

	var dbSess session
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbSess); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return telegramsessionbus.Session{}, sqldb.ErrDBNotFound
		}
		return telegramsessionbus.Session{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSession(dbSess)
}

// QueryByChatID retrieves an active session by Telegram chat ID.
func (s *Store) QueryByChatID(ctx context.Context, chatID telegramchatid.TelegramChatID) (telegramsessionbus.Session, error) {
	data := map[string]any{
		"chat_id": chatID.Value(),
	}

	const q = `
	SELECT
		session_id, user_id, chat_id, session_type, current_step,
		total_steps, retry_count, context_data, last_activity,
		date_created, date_updated
	FROM
		telegram_sessions
	WHERE
		chat_id = :chat_id
	LIMIT 1`

	var dbSess session
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbSess); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return telegramsessionbus.Session{}, sqldb.ErrDBNotFound
		}
		return telegramsessionbus.Session{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSession(dbSess)
}

// DeleteExpired removes all sessions that have exceeded the TTL.
func (s *Store) DeleteExpired(ctx context.Context, ttl time.Duration) (int, error) {
	cutoff := time.Now().UTC().Add(-ttl)

	data := map[string]any{
		"cutoff": cutoff,
	}

	const q = `
	DELETE FROM telegram_sessions
	WHERE last_activity < :cutoff`

	rowsAffected, err := sqldb.NamedExecContextWithResult(ctx, s.log, s.db, q, data)
	if err != nil {
		return 0, fmt.Errorf("namedexeccontext: %w", err)
	}

	return int(rowsAffected), nil
}

// DeleteByUserID removes all sessions for a specific user.
func (s *Store) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM telegram_sessions WHERE user_id = :user_id`

	data := map[string]any{"user_id": userID}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}
