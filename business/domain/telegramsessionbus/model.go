// Package telegramsessionbus provides business layer access for managing
// Telegram multi-step conversation sessions.
package telegramsessionbus

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/retrycount"
	"github.com/francowini/rafiki/business/types/sessionstep"
	"github.com/francowini/rafiki/business/types/sessiontype"
	"github.com/francowini/rafiki/business/types/telegramchatid"
)

// Set of error variables for session operations.
var (
	ErrNotFound           = errors.New("session not found")
	ErrSessionExists      = errors.New("user already has an active session of this type")
	ErrSessionExpired     = errors.New("session has expired")
	ErrAlreadyAtFinalStep = errors.New("session already at final step")
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
	ErrInvalidStepKey     = errors.New("step key does not match current step")
)

// Session represents an active Telegram multi-step conversation.
type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ChatID       telegramchatid.TelegramChatID
	SessionType  sessiontype.SessionType
	CurrentStep  sessionstep.SessionStep
	TotalSteps   int
	RetryCount   retrycount.RetryCount
	ContextData  ContextData
	LastActivity time.Time
	DateCreated  time.Time
	DateUpdated  time.Time
}

// ContextData holds all step data collected during the conversation.
type ContextData struct {
	Steps map[string]StepData `json:"steps"`
}

// NewContextData creates an empty ContextData with initialized map.
func NewContextData() ContextData {
	return ContextData{
		Steps: make(map[string]StepData),
	}
}

// GetStep retrieves step data by step number string.
func (c ContextData) GetStep(stepKey string) (StepData, bool) {
	data, ok := c.Steps[stepKey]
	return data, ok
}

// SetStep sets step data for a given step number string.
func (c *ContextData) SetStep(stepKey string, data StepData) {
	if c.Steps == nil {
		c.Steps = make(map[string]StepData)
	}
	c.Steps[stepKey] = data
}

// MarshalJSON implements json.Marshaler.
func (c ContextData) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Steps)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ContextData) UnmarshalJSON(data []byte) error {
	if c.Steps == nil {
		c.Steps = make(map[string]StepData)
	}
	return json.Unmarshal(data, &c.Steps)
}

// StepData contains the user response and parsed values for a single step.
// All keys are in English following CLAUDE.md language conventions.
type StepData struct {
	RawResponse  string            `json:"raw_response"`
	ParsedValues map[string]string `json:"parsed_values,omitempty"`
	CompletedAt  time.Time         `json:"completed_at"`
}

// NewStepData creates a StepData with the raw response and current timestamp.
func NewStepData(rawResponse string) StepData {
	return StepData{
		RawResponse:  rawResponse,
		ParsedValues: make(map[string]string),
		CompletedAt:  time.Now().UTC(),
	}
}

// WithParsedValue adds a parsed value and returns the StepData (builder pattern).
func (s StepData) WithParsedValue(key, value string) StepData {
	if s.ParsedValues == nil {
		s.ParsedValues = make(map[string]string)
	}
	s.ParsedValues[key] = value
	return s
}

// NewSession contains information needed to create a new session.
type NewSession struct {
	UserID      uuid.UUID
	ChatID      telegramchatid.TelegramChatID
	SessionType sessiontype.SessionType
}

// IsExpired checks if the session has exceeded the given TTL.
func (s Session) IsExpired(ttl time.Duration) bool {
	return time.Since(s.LastActivity) > ttl
}

// CanAdvance returns true if the session can advance to the next step.
func (s Session) CanAdvance() bool {
	return s.CurrentStep.Value() < s.TotalSteps
}

// IsFinalStep returns true if the session is at the final step.
func (s Session) IsFinalStep() bool {
	return s.CurrentStep.IsFinal(s.TotalSteps)
}

// StepKey returns the current step as a string key for ContextData access.
func (s Session) StepKey() string {
	return s.CurrentStep.String()
}
