package telegramsessionbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/retrycount"
	"github.com/francowini/rafiki/business/types/sessionstep"
	"github.com/francowini/rafiki/business/types/sessiontype"
	"github.com/francowini/rafiki/business/types/telegramchatid"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, session Session) error
	Update(ctx context.Context, session Session) error
	Delete(ctx context.Context, session Session) error
	QueryByID(ctx context.Context, sessionID uuid.UUID) (Session, error)
	QueryByUserAndType(ctx context.Context, userID uuid.UUID, sessionType sessiontype.SessionType) (Session, error)
	QueryByChatID(ctx context.Context, chatID telegramchatid.TelegramChatID) (Session, error)
	DeleteExpired(ctx context.Context, ttl time.Duration) (int, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// Business manages the set of APIs for telegram session access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a telegram session business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	b := &Business{
		log:    log,
		storer: storer,
	}
	return b
}

// NewBusinessWithDelegate constructs a telegram session business API with delegate support.
func NewBusinessWithDelegate(log *logger.Logger, dlg *delegate.Delegate, storer Storer) *Business {
	b := &Business{
		log:    log,
		storer: storer,
	}
	b.registerDelegateFunctions(dlg)
	return b
}

// Create adds a new session to the system.
// Returns ErrSessionExists if user already has an active session of this type.
func (b *Business) Create(ctx context.Context, ns NewSession) (Session, error) {
	// Check for existing active session of this type
	existing, err := b.storer.QueryByUserAndType(ctx, ns.UserID, ns.SessionType)
	if err != nil && !errors.Is(err, sqldb.ErrDBNotFound) {
		return Session{}, fmt.Errorf("query existing: %w", err)
	}
	if err == nil {
		b.log.Info(ctx, "telegramsessionbus.create.duplicate",
			"user_id", ns.UserID,
			"session_type", ns.SessionType,
			"existing_session_id", existing.ID,
		)
		return Session{}, ErrSessionExists
	}

	now := time.Now().UTC()
	// MustParse is safe here: values are compile-time constants with guaranteed valid ranges
	step1 := sessionstep.MustParse(1)
	retryZero := retrycount.MustParse(0)

	session := Session{
		ID:           uuid.New(),
		UserID:       ns.UserID,
		ChatID:       ns.ChatID,
		SessionType:  ns.SessionType,
		CurrentStep:  step1,
		TotalSteps:   ns.SessionType.TotalSteps(),
		RetryCount:   retryZero,
		ContextData:  NewContextData(),
		LastActivity: now,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := b.storer.Create(ctx, session); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return Session{}, ErrSessionExists
		}
		return Session{}, fmt.Errorf("create: %w", err)
	}

	b.log.Info(ctx, "telegramsessionbus.create.success",
		"session_id", session.ID,
		"user_id", session.UserID,
		"session_type", session.SessionType,
	)

	return session, nil
}

// AdvanceStepWithData advances the session to the next step and stores the step data.
// This is an atomic operation - both the step advancement and data storage happen together.
func (b *Business) AdvanceStepWithData(ctx context.Context, session Session, stepKey string, data StepData) (Session, error) {
	if stepKey != session.StepKey() {
		return Session{}, fmt.Errorf("%w: expected %s, got %s", ErrInvalidStepKey, session.StepKey(), stepKey)
	}

	if session.IsFinalStep() {
		return Session{}, ErrAlreadyAtFinalStep
	}

	session.ContextData.SetStep(stepKey, data)

	nextStep, err := session.CurrentStep.Next(session.TotalSteps)
	if err != nil {
		return Session{}, fmt.Errorf("next step: %w", err)
	}

	session.CurrentStep = nextStep
	session.RetryCount = session.RetryCount.Reset()
	session.LastActivity = time.Now().UTC()
	session.DateUpdated = session.LastActivity

	if err := b.storer.Update(ctx, session); err != nil {
		return Session{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "telegramsessionbus.advance.success",
		"session_id", session.ID,
		"from_step", stepKey,
		"to_step", session.CurrentStep.String(),
	)

	return session, nil
}

// StoreDataForFinalStep stores the step data for the final step.
func (b *Business) StoreDataForFinalStep(ctx context.Context, session Session, stepKey string, data StepData) (Session, error) {
	if stepKey != session.StepKey() {
		return Session{}, fmt.Errorf("%w: expected %s, got %s", ErrInvalidStepKey, session.StepKey(), stepKey)
	}

	if !session.IsFinalStep() {
		return Session{}, fmt.Errorf("not at final step, use AdvanceStepWithData instead")
	}

	session.ContextData.SetStep(stepKey, data)
	session.LastActivity = time.Now().UTC()
	session.DateUpdated = session.LastActivity

	if err := b.storer.Update(ctx, session); err != nil {
		return Session{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "telegramsessionbus.final_step.success",
		"session_id", session.ID,
		"step", stepKey,
	)

	return session, nil
}

// IncrementRetry increments the retry count for the current step.
func (b *Business) IncrementRetry(ctx context.Context, session Session) (Session, error) {
	if session.RetryCount.IsMaxed() {
		return Session{}, ErrMaxRetriesExceeded
	}

	newCount, err := session.RetryCount.Increment()
	if err != nil {
		return Session{}, fmt.Errorf("increment retry: %w", err)
	}

	session.RetryCount = newCount
	session.LastActivity = time.Now().UTC()
	session.DateUpdated = session.LastActivity

	if err := b.storer.Update(ctx, session); err != nil {
		return Session{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "telegramsessionbus.retry.incremented",
		"session_id", session.ID,
		"step", session.CurrentStep.String(),
		"retry_count", newCount.Value(),
	)

	return session, nil
}

// Delete removes a session from the system.
func (b *Business) Delete(ctx context.Context, session Session) error {
	if err := b.storer.Delete(ctx, session); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	b.log.Info(ctx, "telegramsessionbus.delete.success",
		"session_id", session.ID,
		"user_id", session.UserID,
	)

	return nil
}

// QueryByID finds a session by its unique ID.
func (b *Business) QueryByID(ctx context.Context, sessionID uuid.UUID) (Session, error) {
	session, err := b.storer.QueryByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("query by id: %w", err)
	}
	return session, nil
}

// QueryByUserAndType finds an active session for a user and session type.
func (b *Business) QueryByUserAndType(ctx context.Context, userID uuid.UUID, sessionType sessiontype.SessionType) (Session, error) {
	session, err := b.storer.QueryByUserAndType(ctx, userID, sessionType)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("query by user and type: %w", err)
	}
	return session, nil
}

// QueryByChatID finds an active session by Telegram chat ID.
func (b *Business) QueryByChatID(ctx context.Context, chatID telegramchatid.TelegramChatID) (Session, error) {
	session, err := b.storer.QueryByChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("query by chat id: %w", err)
	}
	return session, nil
}

// CleanupExpired deletes all sessions that have exceeded the TTL.
func (b *Business) CleanupExpired(ctx context.Context, ttl time.Duration) (int, error) {
	count, err := b.storer.DeleteExpired(ctx, ttl)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired: %w", err)
	}

	if count > 0 {
		b.log.Info(ctx, "telegramsessionbus.cleanup.success",
			"deleted_count", count,
			"ttl_minutes", ttl.Minutes(),
		)
	}

	return count, nil
}

// HasActiveSession checks if a user has an active session of the given type.
func (b *Business) HasActiveSession(ctx context.Context, userID uuid.UUID, sessionType sessiontype.SessionType) (bool, error) {
	_, err := b.storer.QueryByUserAndType(ctx, userID, sessionType)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("check active session: %w", err)
	}
	return true, nil
}
