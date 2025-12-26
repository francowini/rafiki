package telegramsessiondb

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/business/types/retrycount"
	"github.com/francowini/rafiki/business/types/sessionstep"
	"github.com/francowini/rafiki/business/types/sessiontype"
	"github.com/francowini/rafiki/business/types/telegramchatid"
)

type session struct {
	ID           uuid.UUID `db:"session_id"`
	UserID       uuid.UUID `db:"user_id"`
	ChatID       int64     `db:"chat_id"`
	SessionType  string    `db:"session_type"`
	CurrentStep  int       `db:"current_step"`
	TotalSteps   int       `db:"total_steps"`
	RetryCount   int       `db:"retry_count"`
	ContextData  []byte    `db:"context_data"`
	LastActivity time.Time `db:"last_activity"`
	DateCreated  time.Time `db:"date_created"`
	DateUpdated  time.Time `db:"date_updated"`
}

func toDBSession(bus telegramsessionbus.Session) (session, error) {
	contextJSON, err := json.Marshal(bus.ContextData)
	if err != nil {
		return session{}, fmt.Errorf("marshal context_data: %w", err)
	}

	return session{
		ID:           bus.ID,
		UserID:       bus.UserID,
		ChatID:       bus.ChatID.Value(),
		SessionType:  bus.SessionType.Value(),
		CurrentStep:  bus.CurrentStep.Value(),
		TotalSteps:   bus.TotalSteps,
		RetryCount:   bus.RetryCount.Value(),
		ContextData:  contextJSON,
		LastActivity: bus.LastActivity.UTC(),
		DateCreated:  bus.DateCreated.UTC(),
		DateUpdated:  bus.DateUpdated.UTC(),
	}, nil
}

func toBusSession(db session) (telegramsessionbus.Session, error) {
	chatID, err := telegramchatid.Parse(db.ChatID)
	if err != nil {
		return telegramsessionbus.Session{}, fmt.Errorf("parse chat_id: %w", err)
	}

	sessionType, err := sessiontype.Parse(db.SessionType)
	if err != nil {
		return telegramsessionbus.Session{}, fmt.Errorf("parse session_type: %w", err)
	}

	currentStep, err := sessionstep.Parse(db.CurrentStep)
	if err != nil {
		return telegramsessionbus.Session{}, fmt.Errorf("parse current_step: %w", err)
	}

	retryCount, err := retrycount.Parse(db.RetryCount)
	if err != nil {
		return telegramsessionbus.Session{}, fmt.Errorf("parse retry_count: %w", err)
	}

	var contextData telegramsessionbus.ContextData
	if len(db.ContextData) > 0 {
		if err := json.Unmarshal(db.ContextData, &contextData); err != nil {
			return telegramsessionbus.Session{}, fmt.Errorf("unmarshal context_data: %w", err)
		}
	} else {
		contextData = telegramsessionbus.NewContextData()
	}

	return telegramsessionbus.Session{
		ID:           db.ID,
		UserID:       db.UserID,
		ChatID:       chatID,
		SessionType:  sessionType,
		CurrentStep:  currentStep,
		TotalSteps:   db.TotalSteps,
		RetryCount:   retryCount,
		ContextData:  contextData,
		LastActivity: db.LastActivity.UTC(),
		DateCreated:  db.DateCreated.UTC(),
		DateUpdated:  db.DateUpdated.UTC(),
	}, nil
}
