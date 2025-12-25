# Telegram Moment Tracking - Deliverable 2: Database Schema

## Overview

This deliverable implements the database schema for Telegram moment tracking sessions. It includes:
1. New `telegram_sessions` table for multi-step conversation tracking
2. Addition of `telegram_linked_at` column to `users` table
3. Unique index on `users.telegram_chat_id` for one-to-one mapping
4. Three new business types for validated domain values

## Architecture Compliance

- **Domain Type**: Child Domain (`telegramsessionbus`)
- **Parent Domain**: `userbus` (Root Domain)
- **Imports**: `telegramsessionbus` imports `userbus.ExtBusiness` interface only
- **Status**: ALIGNED with `devs/business-model-dependencies.md`

**Key Architecture Decisions:**
- FK with `ON DELETE CASCADE` ensures sessions are deleted when user is deleted
- No cross-domain imports (job worker orchestrates between domains)
- Strong types for all validated values (following CLAUDE.md Business Types pattern)

---

## Database Schema (Version 20)

Add the following to `/Users/francowini/Documents/rafiki/business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 20
-- Description: Add Telegram integration for ACT Moment tracking (Deliverable 2)
-- Adds telegram_linked_at to users, telegram_sessions table with wellness tracking support

-- =============================================================================
-- Add telegram_linked_at to users table
-- =============================================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_linked_at TIMESTAMP NULL;

COMMENT ON COLUMN users.telegram_linked_at IS 'Timestamp when user first linked their Telegram account (NULL if never linked)';

-- =============================================================================
-- Create UNIQUE index on users.telegram_chat_id for one-to-one mapping
-- =============================================================================
-- Enforces business rule: one Telegram account -> one Rafiki user
-- Allows fast lookups when Telegram webhook receives message
CREATE UNIQUE INDEX IF NOT EXISTS users_telegram_chat_id_unique_idx
    ON users(telegram_chat_id)
    WHERE telegram_chat_id IS NOT NULL;

-- =============================================================================
-- Create telegram_sessions table for multi-step conversation tracking
-- =============================================================================
CREATE TABLE IF NOT EXISTS telegram_sessions (
    session_id          UUID        NOT NULL,
    user_id             UUID        NOT NULL,
    chat_id             BIGINT      NOT NULL,
    current_step        INTEGER     NOT NULL CHECK (current_step BETWEEN 1 AND 6),
    retry_count         INTEGER     NOT NULL DEFAULT 0 CHECK (retry_count BETWEEN 0 AND 2),
    initial_intensity   INTEGER     NULL CHECK (initial_intensity IS NULL OR (initial_intensity >= 0 AND initial_intensity <= 10)),
    abandonment_step    INTEGER     NULL CHECK (abandonment_step IS NULL OR (abandonment_step BETWEEN 1 AND 6)),
    parsed_data         JSONB       NOT NULL DEFAULT '{}',
    last_activity       TIMESTAMP   NOT NULL,
    date_created        TIMESTAMP   NOT NULL,
    date_updated        TIMESTAMP   NOT NULL,

    PRIMARY KEY (session_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- =============================================================================
-- Indexes for telegram_sessions
-- =============================================================================

-- UNIQUE constraint: one active session per user
-- Business logic also enforces this, but DB constraint provides defense-in-depth
CREATE UNIQUE INDEX IF NOT EXISTS telegram_sessions_user_id_unique_idx
    ON telegram_sessions(user_id);

-- Fast lookup when Telegram webhook receives message (chat_id -> session)
CREATE INDEX IF NOT EXISTS telegram_sessions_chat_id_idx
    ON telegram_sessions(chat_id);

-- TTL cleanup: find sessions inactive for > 20 minutes
CREATE INDEX IF NOT EXISTS telegram_sessions_last_activity_idx
    ON telegram_sessions(last_activity);

-- User session history queries
CREATE INDEX IF NOT EXISTS telegram_sessions_user_date_idx
    ON telegram_sessions(user_id, date_created DESC);

-- =============================================================================
-- Comments documenting telegram_sessions schema
-- =============================================================================

COMMENT ON TABLE telegram_sessions IS 'Multi-step Telegram conversation sessions for ACT-based moment tracking (20 min TTL)';
COMMENT ON COLUMN telegram_sessions.session_id IS 'Unique session identifier';
COMMENT ON COLUMN telegram_sessions.user_id IS 'Rafiki user associated with session';
COMMENT ON COLUMN telegram_sessions.chat_id IS 'Telegram chat_id for message routing (non-zero int64)';
COMMENT ON COLUMN telegram_sessions.current_step IS 'Current step in 6-step ACT functional analysis (1=situacion, 2=sintomas, 3=conducta, 4=consecuencias, 5=valores, 6=intensidad)';
COMMENT ON COLUMN telegram_sessions.retry_count IS 'AI validation retry attempts for current step (0-2, auto-approves at 2)';
COMMENT ON COLUMN telegram_sessions.initial_intensity IS 'User-reported intensity (0-10) captured at session start for ACT validation';
COMMENT ON COLUMN telegram_sessions.abandonment_step IS 'Step where user abandoned session for analytics (NULL if active/completed)';
COMMENT ON COLUMN telegram_sessions.last_activity IS 'Timestamp of last user message (for 20-minute TTL cleanup)';

-- =============================================================================
-- parsed_data JSONB Schema Documentation
-- =============================================================================
-- The parsed_data column stores a flexible JSONB object with the following structure:
--
-- {
--   "step_1": {
--     "situacion": "Free-text description of what happened",
--     "pensamientos": "Free-text description of thoughts that appeared",
--     "retry_reason": "Optional: Why AI asked for refinement"
--   },
--   "step_2": {
--     "sintomas_fisicos": "Free-text physical symptoms (palpitations, sweating, etc.)",
--     "emociones": "Free-text emotional labels (anxiety, fear, sadness, etc.)",
--     "retry_reason": "Optional: Why AI asked for refinement"
--   },
--   "step_3": {
--     "conducta": "Free-text description of behavior/action taken",
--     "retry_reason": "Optional: Why AI asked for refinement"
--   },
--   "step_4": {
--     "consecuencias": "Free-text description of what happened after the behavior",
--     "retry_reason": "Optional: Why AI asked for refinement"
--   },
--   "step_5": {
--     "tipo": "alejamiento | acercamiento",
--     "descripcion": "Free-text reflection on whether behavior moved toward or away from values",
--     "retry_reason": "Optional: Why AI asked for refinement"
--   },
--   "step_6": {
--     "intensidad": 7,
--     "retry_reason": "Optional: If user provided non-numeric value"
--   }
-- }
--
-- Fields are added incrementally as user progresses through steps.
-- retry_reason tracks AI validation feedback for analytics.
```

---

## Business Types

Create the following strong types following the pattern in `business/types/`:

### 1. SessionStep (`business/types/sessionstep/sessionstep.go`)

```go
package sessionstep

import (
	"fmt"
)

// SessionStep represents a validated step in the Telegram ACT Moment flow (1-6).
type SessionStep struct {
	value int
}

// Pre-defined step constants for use in business logic
var (
	Step1Situacion     = SessionStep{1}
	Step2Sintomas      = SessionStep{2}
	Step3Conducta      = SessionStep{3}
	Step4Consecuencias = SessionStep{4}
	Step5Valores       = SessionStep{5}
	Step6Intensidad    = SessionStep{6}
)

// Value returns the int value of the session step.
func (s SessionStep) Value() int {
	return s.value
}

// String returns the string representation.
func (s SessionStep) String() string {
	return fmt.Sprintf("%d", s.value)
}

// Equal provides support for the go-cmp package and testing.
func (s SessionStep) Equal(s2 SessionStep) bool {
	return s.value == s2.value
}

// MarshalText provides support for logging and any marshal needs.
func (s SessionStep) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *SessionStep) UnmarshalText(data []byte) error {
	var value int
	if _, err := fmt.Sscanf(string(data), "%d", &value); err != nil {
		return fmt.Errorf("invalid session step value: %w", err)
	}
	parsed, err := Parse(value)
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// Parse validates and creates a SessionStep (1-6).
func Parse(value int) (SessionStep, error) {
	if value < 1 || value > 6 {
		return SessionStep{}, fmt.Errorf("session step must be between 1 and 6, got %d", value)
	}
	return SessionStep{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) SessionStep {
	step, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return step
}

// Next returns the next step, or error if already at step 6.
func (s SessionStep) Next() (SessionStep, error) {
	if s.value >= 6 {
		return SessionStep{}, fmt.Errorf("already at final step (6)")
	}
	return SessionStep{s.value + 1}, nil
}

// IsFirst returns true if this is step 1.
func (s SessionStep) IsFirst() bool {
	return s.value == 1
}

// IsFinal returns true if this is step 6 (intensidad).
func (s SessionStep) IsFinal() bool {
	return s.value == 6
}
```

### 2. RetryCount (`business/types/retrycount/retrycount.go`)

```go
package retrycount

import (
	"fmt"
)

// RetryCount represents a validated retry count (0-2) for AI validation failures.
type RetryCount struct {
	value int
}

// Value returns the int value of the retry count.
func (r RetryCount) Value() int {
	return r.value
}

// String returns the string representation.
func (r RetryCount) String() string {
	return fmt.Sprintf("%d", r.value)
}

// Equal provides support for the go-cmp package and testing.
func (r RetryCount) Equal(r2 RetryCount) bool {
	return r.value == r2.value
}

// MarshalText provides support for logging and any marshal needs.
func (r RetryCount) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (r *RetryCount) UnmarshalText(data []byte) error {
	var value int
	if _, err := fmt.Sscanf(string(data), "%d", &value); err != nil {
		return fmt.Errorf("invalid retry count value: %w", err)
	}
	parsed, err := Parse(value)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

// Parse validates and creates a RetryCount (0-2).
func Parse(value int) (RetryCount, error) {
	if value < 0 || value > 2 {
		return RetryCount{}, fmt.Errorf("retry count must be between 0 and 2, got %d", value)
	}
	return RetryCount{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) RetryCount {
	count, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return count
}

// Increment returns a new RetryCount incremented by 1.
// Returns error if already at max (2).
func (r RetryCount) Increment() (RetryCount, error) {
	if r.value >= 2 {
		return RetryCount{}, fmt.Errorf("retry count already at maximum (2)")
	}
	return RetryCount{r.value + 1}, nil
}

// IsMaxed returns true if retry count is at maximum (2).
// When maxed, session should auto-approve and advance to next step.
func (r RetryCount) IsMaxed() bool {
	return r.value >= 2
}

// Reset returns a new RetryCount set to 0.
// Use when advancing to a new step.
func (r RetryCount) Reset() RetryCount {
	return RetryCount{0}
}
```

### 3. TelegramChatID (`business/types/telegramchatid/telegramchatid.go`)

```go
package telegramchatid

import (
	"fmt"
	"strconv"
)

// TelegramChatID represents a validated Telegram chat_id (non-zero int64).
// Telegram chat IDs can be negative (groups) or positive (users), but never zero.
type TelegramChatID struct {
	value int64
}

// Value returns the int64 value of the chat ID.
func (t TelegramChatID) Value() int64 {
	return t.value
}

// String returns the string representation.
func (t TelegramChatID) String() string {
	return fmt.Sprintf("%d", t.value)
}

// Equal provides support for the go-cmp package and testing.
func (t TelegramChatID) Equal(t2 TelegramChatID) bool {
	return t.value == t2.value
}

// MarshalText provides support for logging and any marshal needs.
func (t TelegramChatID) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *TelegramChatID) UnmarshalText(data []byte) error {
	value, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID value: %w", err)
	}
	parsed, err := Parse(value)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// Parse validates and creates a TelegramChatID (non-zero int64).
func Parse(value int64) (TelegramChatID, error) {
	if value == 0 {
		return TelegramChatID{}, fmt.Errorf("telegram chat_id cannot be zero")
	}
	return TelegramChatID{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int64) TelegramChatID {
	chatID, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return chatID
}
```

---

## Domain Model Preview

The following model will be implemented in Deliverable 3 (`telegramsessionbus`):

```go
// business/domain/telegramsessionbus/model.go
package telegramsessionbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/francowini/rafiki/business/types/intensity"
	"github.com/francowini/rafiki/business/types/retrycount"
	"github.com/francowini/rafiki/business/types/sessionstep"
	"github.com/francowini/rafiki/business/types/telegramchatid"
)

// Session represents an active Telegram moment tracking conversation.
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ChatID           telegramchatid.TelegramChatID
	CurrentStep      sessionstep.SessionStep
	RetryCount       retrycount.RetryCount
	InitialIntensity *intensity.Intensity // Nullable: captured at session start
	AbandonmentStep  *sessionstep.SessionStep // Nullable: set on timeout/cancel
	ParsedData       json.RawMessage
	LastActivity     time.Time
	DateCreated      time.Time
	DateUpdated      time.Time
}

// Sentinel errors
var (
	ErrNotFound            = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("user already has an active session")
	ErrSessionExpired      = errors.New("session has expired (20 min TTL)")
	ErrAlreadyAtFinalStep  = errors.New("session already at final step")
	ErrMaxRetriesReached   = errors.New("maximum retry attempts reached")
)
```

---

## Therapeutic Value Summary

This schema design supports mental wellness through:

| Feature | Therapeutic Value |
|---------|-------------------|
| **JSONB parsed_data** | Preserves raw emotional expression without forced categorization |
| **retry_count (0-2)** | Compassionate re-asking without endless loops; auto-approves to reduce frustration |
| **20-minute TTL** | Extended reflection time for ACT functional analysis (increased from 15 min) |
| **initial_intensity** | Enables ACT validation: does naming emotions reduce intensity? |
| **abandonment_step** | Pattern detection for UX improvements (which steps cause drop-off?) |
| **retry_reason in JSONB** | Transparency + training data for AI prompt refinement |

---

## Errors-to-Avoid Compliance

This schema adheres to the following patterns from `devs/errors-to-avoid-backend.md`:

| Rule | Compliance |
|------|------------|
| **#4: Strong Types** | All validated values (step, retry_count, chat_id) use strong types with `Value()`, `String()`, `Equal()`, `MarshalText()`, `UnmarshalText()` methods |
| **#9: Timezone** | All timestamps stored in UTC (TIMESTAMP NOT NULL) |
| **#10: Idempotency** | UNIQUE index on `user_id` prevents duplicate active sessions |
| **#15: When NOT to Use Strong Types** | Simple JSONB `parsed_data` uses `json.RawMessage`, not over-engineered type |
| **#26: PostgreSQL Pagination** | Schema uses PostgreSQL-compatible syntax |
| **#31: Defense-in-Depth** | Both DB constraints AND business layer will enforce validation |

---

## TTL Cleanup Query

The cleanup job (to be implemented in Deliverable 5) should use this query:

```sql
-- Delete sessions inactive for more than 20 minutes
-- Set abandonment_step before deletion for analytics
UPDATE telegram_sessions
SET abandonment_step = current_step,
    date_updated = NOW()
WHERE last_activity < NOW() - INTERVAL '20 minutes';

DELETE FROM telegram_sessions
WHERE last_activity < NOW() - INTERVAL '20 minutes';
```

The `telegram_sessions_last_activity_idx` index supports efficient cleanup queries.

---

## Implementation Checklist

- [ ] Add Version 20 migration to `business/sdk/migrate/sql/migrate.sql`
- [ ] Create `business/types/sessionstep/sessionstep.go`
- [ ] Create `business/types/retrycount/retrycount.go`
- [ ] Create `business/types/telegramchatid/telegramchatid.go`
- [ ] Run migration: `make deploy` (migrations run automatically)
- [ ] Verify with: `PGPASSWORD=db psql -h localhost -U db -d rafiki -c "\d telegram_sessions"`

---

## Dependencies

- **Depends on**: Nothing (Deliverable 1 - Anthropic Client is independent)
- **Required by**: Deliverable 3 (Session Domain), Deliverable 4 (Webhook Handler)

---

## Next Steps

After this schema is deployed, proceed to **Deliverable 3: Session Domain** which implements `business/domain/telegramsessionbus/` with CRUD operations and state machine transitions.
