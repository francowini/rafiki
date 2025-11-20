# Backend Implementation - Database Access Layer (telegramdb)

**Task Category**: Backend / Database
**Estimated Time**: 6-8 hours
**Prerequisites**: [01-backend-database.md](./01-backend-database.md) - Database migration completed
**Dependencies**: Database schema version 1.04

---

## Overview

Implement the database access layer for Telegram tables. This layer provides CRUD operations for:
- `telegram_users` - Telegram user data
- `user_telegram_links` - User-Telegram linkage
- `telegram_link_codes` - Temporary link codes
- `conversation_states` - Conversation state machine

**Pattern**: Follow existing `business/sdk/sqldb/*` patterns (momentdb, userdb, etc.)

---

## Directory Structure

```
business/sdk/sqldb/telegramdb/
├── telegram_users.go         # TelegramUser CRUD
├── link_codes.go              # LinkCode CRUD
├── conversation_states.go     # ConversationState CRUD
├── filters.go                 # Query filters
├── models.go                  # Database models
└── telegramdb.go             # Store struct (main interface)
```

---

## Task 1: Create Store Interface

**File**: `business/sdk/sqldb/telegramdb/telegramdb.go`

```go
// Package telegramdb provides database access for Telegram integration.
package telegramdb

import (
    "context"
    "fmt"

    "github.com/francowini/rafiki/business/sdk/sqldb"
    "github.com/francowini/rafiki/foundation/logger"
    "github.com/jmoiron/sqlx"
)

// Store manages database operations for Telegram data.
type Store struct {
    log *logger.Logger
    db  sqlx.ExtContext
}

// NewStore constructs a new Store.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
    return &Store{
        log: log,
        db:  db,
    }
}

// WithinTran runs the provided function within a transaction.
func (s *Store) WithinTran(ctx context.Context, fn func(s *Store) error) error {
    return sqldb.WithinTran(ctx, s.log, s.db, func(tx sqlx.ExtContext) error {
        trS := &Store{
            log: s.log,
            db:  tx,
        }
        return fn(trS)
    })
}
```

**Why This Pattern**:
- Consistent with existing stores (momentdb, userdb)
- Supports transactions via `WithinTran`
- Dependency injection (log, db)

**Test**:
```go
// business/sdk/sqldb/telegramdb/telegramdb_test.go
package telegramdb_test

import (
    "testing"

    "github.com/francowini/rafiki/business/sdk/sqldb/telegramdb"
    "github.com/francowini/rafiki/foundation/logger"
)

func TestStore_Creation(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func() string { return "test-trace-id" })

    // Assuming test database setup
    db := setupTestDB(t)

    store := telegramdb.NewStore(log, db)

    if store == nil {
        t.Fatal("expected store to be created")
    }
}
```

---

## Task 2: Define Database Models

**File**: `business/sdk/sqldb/telegramdb/models.go`

```go
package telegramdb

import (
    "time"

    "github.com/google/uuid"
)

// TelegramUser represents a Telegram user in the database.
type TelegramUser struct {
    TelegramUserID int64      `db:"telegram_user_id"`
    Username       *string    `db:"username"`        // Nullable
    FirstName      *string    `db:"first_name"`      // Nullable
    LastName       *string    `db:"last_name"`       // Nullable
    LanguageCode   *string    `db:"language_code"`   // Nullable
    IsBot          bool       `db:"is_bot"`
    DateCreated    time.Time  `db:"date_created"`
    DateUpdated    time.Time  `db:"date_updated"`
}

// UserTelegramLink represents a link between a Rafiki user and Telegram user.
type UserTelegramLink struct {
    UserID         uuid.UUID `db:"user_id"`
    TelegramUserID int64     `db:"telegram_user_id"`
    LinkedAt       time.Time `db:"linked_at"`
    DateCreated    time.Time `db:"date_created"`
    DateUpdated    time.Time `db:"date_updated"`
}

// TelegramLinkCode represents a temporary link code for account linking.
type TelegramLinkCode struct {
    Code        string     `db:"code"`
    UserID      uuid.UUID  `db:"user_id"`
    ExpiresAt   time.Time  `db:"expires_at"`
    ConsumedAt  *time.Time `db:"consumed_at"`  // Nullable
    ConsumedBy  *int64     `db:"consumed_by"`  // Nullable
    DateCreated time.Time  `db:"date_created"`
}

// ConversationState represents an active Telegram conversation.
type ConversationState struct {
    ConversationID   uuid.UUID `db:"conversation_id"`
    TelegramUserID   int64     `db:"telegram_user_id"`
    UserID           uuid.UUID `db:"user_id"`
    ConversationType string    `db:"conversation_type"` // 'moment', 'habit', 'goal'
    CurrentStep      string    `db:"current_step"`      // 'moment:awaiting_situation'
    Data             []byte    `db:"data"`              // JSONB (raw bytes)
    LastMessageID    *int      `db:"last_message_id"`   // Nullable
    LastActivity     time.Time `db:"last_activity"`
    DateCreated      time.Time `db:"date_created"`
    DateUpdated      time.Time `db:"date_updated"`
}
```

**Why These Types**:
- Matches database schema exactly
- Uses `*string` for nullable fields (Go best practice)
- Uses `[]byte` for JSONB (will be marshaled/unmarshaled by application)
- Follows existing patterns (momentdb.Moment, userdb.User)

---

## Task 3: Implement TelegramUser CRUD

**File**: `business/sdk/sqldb/telegramdb/telegram_users.go`

```go
package telegramdb

import (
    "context"
    "errors"
    "fmt"

    "github.com/francowini/rafiki/business/sdk/sqldb"
    "github.com/jmoiron/sqlx"
)

// CreateTelegramUser inserts a new Telegram user into the database.
func (s *Store) CreateTelegramUser(ctx context.Context, tu TelegramUser) error {
    const q = `
    INSERT INTO telegram_users (
        telegram_user_id, username, first_name, last_name, language_code, is_bot,
        date_created, date_updated
    ) VALUES (
        :telegram_user_id, :username, :first_name, :last_name, :language_code, :is_bot,
        :date_created, :date_updated
    )
    ON CONFLICT (telegram_user_id) DO UPDATE SET
        username = EXCLUDED.username,
        first_name = EXCLUDED.first_name,
        last_name = EXCLUDED.last_name,
        language_code = EXCLUDED.language_code,
        date_updated = EXCLUDED.date_updated`

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, tu); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// QueryTelegramUserByID retrieves a Telegram user by telegram_user_id.
func (s *Store) QueryTelegramUserByID(ctx context.Context, telegramUserID int64) (TelegramUser, error) {
    const q = `
    SELECT
        telegram_user_id, username, first_name, last_name, language_code, is_bot,
        date_created, date_updated
    FROM telegram_users
    WHERE telegram_user_id = $1`

    var tu TelegramUser
    if err := sqldb.Get(ctx, s.log, s.db, q, &tu, telegramUserID); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return TelegramUser{}, fmt.Errorf("telegram user %d not found", telegramUserID)
        }
        return TelegramUser{}, fmt.Errorf("selecting telegram user: %w", err)
    }

    return tu, nil
}

// DeleteTelegramUser removes a Telegram user from the database.
func (s *Store) DeleteTelegramUser(ctx context.Context, telegramUserID int64) error {
    const q = `DELETE FROM telegram_users WHERE telegram_user_id = $1`

    if err := sqldb.ExecContext(ctx, s.log, s.db, q, telegramUserID); err != nil {
        return fmt.Errorf("deleting telegram user: %w", err)
    }

    return nil
}
```

**Test** (`telegram_users_test.go`):
```go
func TestTelegramUser_CRUD(t *testing.T) {
    store := setupTestStore(t)
    ctx := context.Background()

    // Create
    tu := TelegramUser{
        TelegramUserID: 123456789,
        Username:       stringPtr("test_user"),
        FirstName:      stringPtr("Test"),
        IsBot:          false,
        DateCreated:    time.Now(),
        DateUpdated:    time.Now(),
    }

    err := store.CreateTelegramUser(ctx, tu)
    if err != nil {
        t.Fatalf("creating telegram user: %v", err)
    }

    // Read
    got, err := store.QueryTelegramUserByID(ctx, 123456789)
    if err != nil {
        t.Fatalf("querying telegram user: %v", err)
    }

    if got.TelegramUserID != tu.TelegramUserID {
        t.Errorf("expected telegram_user_id %d, got %d", tu.TelegramUserID, got.TelegramUserID)
    }

    // Delete
    err = store.DeleteTelegramUser(ctx, 123456789)
    if err != nil {
        t.Fatalf("deleting telegram user: %v", err)
    }

    // Verify deletion
    _, err = store.QueryTelegramUserByID(ctx, 123456789)
    if err == nil {
        t.Error("expected error when querying deleted user")
    }
}
```

---

## Task 4: Implement Link Code CRUD

**File**: `business/sdk/sqldb/telegramdb/link_codes.go`

```go
package telegramdb

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/francowini/rafiki/business/sdk/sqldb"
    "github.com/google/uuid"
)

// CreateLinkCode inserts a new link code into the database.
func (s *Store) CreateLinkCode(ctx context.Context, lc TelegramLinkCode) error {
    const q = `
    INSERT INTO telegram_link_codes (
        code, user_id, expires_at, date_created
    ) VALUES (
        :code, :user_id, :expires_at, :date_created
    )`

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, lc); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// QueryLinkCodeByCode retrieves a link code by its code string.
func (s *Store) QueryLinkCodeByCode(ctx context.Context, code string) (TelegramLinkCode, error) {
    const q = `
    SELECT
        code, user_id, expires_at, consumed_at, consumed_by, date_created
    FROM telegram_link_codes
    WHERE code = $1`

    var lc TelegramLinkCode
    if err := sqldb.Get(ctx, s.log, s.db, q, &lc, code); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return TelegramLinkCode{}, fmt.Errorf("link code not found")
        }
        return TelegramLinkCode{}, fmt.Errorf("selecting link code: %w", err)
    }

    return lc, nil
}

// ConsumeLinkCode marks a link code as consumed.
func (s *Store) ConsumeLinkCode(ctx context.Context, code string, telegramUserID int64) error {
    const q = `
    UPDATE telegram_link_codes
    SET consumed_at = $1, consumed_by = $2
    WHERE code = $3 AND consumed_at IS NULL`

    if err := sqldb.ExecContext(ctx, s.log, s.db, q, time.Now(), telegramUserID, code); err != nil {
        return fmt.Errorf("consuming link code: %w", err)
    }

    return nil
}

// DeleteExpiredLinkCodes removes all expired link codes.
func (s *Store) DeleteExpiredLinkCodes(ctx context.Context) (int, error) {
    const q = `DELETE FROM telegram_link_codes WHERE expires_at < NOW()`

    result, err := s.db.ExecContext(ctx, q)
    if err != nil {
        return 0, fmt.Errorf("deleting expired link codes: %w", err)
    }

    count, _ := result.RowsAffected()
    return int(count), nil
}

// CountRecentLinkCodes counts link codes created by a user in the last hour (rate limiting).
func (s *Store) CountRecentLinkCodes(ctx context.Context, userID uuid.UUID) (int, error) {
    const q = `
    SELECT COUNT(*)
    FROM telegram_link_codes
    WHERE user_id = $1 AND date_created > NOW() - INTERVAL '1 hour'`

    var count int
    if err := sqldb.Get(ctx, s.log, s.db, q, &count, userID); err != nil {
        return 0, fmt.Errorf("counting recent link codes: %w", err)
    }

    return count, nil
}
```

**Test**:
```go
func TestLinkCode_CRUD(t *testing.T) {
    store := setupTestStore(t)
    ctx := context.Background()

    userID := uuid.New()

    // Create
    lc := TelegramLinkCode{
        Code:        "TEST-CODE-1234",
        UserID:      userID,
        ExpiresAt:   time.Now().Add(5 * time.Minute),
        DateCreated: time.Now(),
    }

    err := store.CreateLinkCode(ctx, lc)
    if err != nil {
        t.Fatalf("creating link code: %v", err)
    }

    // Read
    got, err := store.QueryLinkCodeByCode(ctx, "TEST-CODE-1234")
    if err != nil {
        t.Fatalf("querying link code: %v", err)
    }

    if got.Code != lc.Code {
        t.Errorf("expected code %s, got %s", lc.Code, got.Code)
    }

    // Consume
    err = store.ConsumeLinkCode(ctx, "TEST-CODE-1234", 123456789)
    if err != nil {
        t.Fatalf("consuming link code: %v", err)
    }

    // Verify consumed
    got, _ = store.QueryLinkCodeByCode(ctx, "TEST-CODE-1234")
    if got.ConsumedAt == nil {
        t.Error("expected consumed_at to be set")
    }
}
```

---

## Task 5: Implement ConversationState CRUD

**File**: `business/sdk/sqldb/telegramdb/conversation_states.go`

```go
package telegramdb

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/francowini/rafiki/business/sdk/sqldb"
    "github.com/google/uuid"
)

// CreateConversationState inserts a new conversation state.
func (s *Store) CreateConversationState(ctx context.Context, cs ConversationState) error {
    const q = `
    INSERT INTO conversation_states (
        conversation_id, telegram_user_id, user_id, conversation_type,
        current_step, data, last_message_id, last_activity,
        date_created, date_updated
    ) VALUES (
        :conversation_id, :telegram_user_id, :user_id, :conversation_type,
        :current_step, :data, :last_message_id, :last_activity,
        :date_created, :date_updated
    )`

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, cs); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// QueryConversationStateByTelegramUserID retrieves active conversation for a Telegram user.
func (s *Store) QueryConversationStateByTelegramUserID(ctx context.Context, telegramUserID int64) (ConversationState, error) {
    const q = `
    SELECT
        conversation_id, telegram_user_id, user_id, conversation_type,
        current_step, data, last_message_id, last_activity,
        date_created, date_updated
    FROM conversation_states
    WHERE telegram_user_id = $1
      AND current_step NOT IN ('completed', 'cancelled')
    ORDER BY last_activity DESC
    LIMIT 1`

    var cs ConversationState
    if err := sqldb.Get(ctx, s.log, s.db, q, &cs, telegramUserID); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return ConversationState{}, fmt.Errorf("no active conversation found")
        }
        return ConversationState{}, fmt.Errorf("selecting conversation state: %w", err)
    }

    return cs, nil
}

// UpdateConversationState updates an existing conversation state.
func (s *Store) UpdateConversationState(ctx context.Context, cs ConversationState) error {
    const q = `
    UPDATE conversation_states
    SET
        current_step = :current_step,
        data = :data,
        last_message_id = :last_message_id,
        last_activity = :last_activity,
        date_updated = :date_updated
    WHERE conversation_id = :conversation_id`

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, cs); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// DeleteConversationState removes a conversation state.
func (s *Store) DeleteConversationState(ctx context.Context, conversationID uuid.UUID) error {
    const q = `DELETE FROM conversation_states WHERE conversation_id = $1`

    if err := sqldb.ExecContext(ctx, s.log, s.db, q, conversationID); err != nil {
        return fmt.Errorf("deleting conversation state: %w", err)
    }

    return nil
}

// DeleteAbandonedConversations removes conversations with no activity in the last 5 minutes.
func (s *Store) DeleteAbandonedConversations(ctx context.Context) (int, error) {
    const q = `
    DELETE FROM conversation_states
    WHERE last_activity < NOW() - INTERVAL '5 minutes'
      AND current_step NOT IN ('completed', 'cancelled')`

    result, err := s.db.ExecContext(ctx, q)
    if err != nil {
        return 0, fmt.Errorf("deleting abandoned conversations: %w", err)
    }

    count, _ := result.RowsAffected()
    return int(count), nil
}
```

**Test**:
```go
func TestConversationState_CRUD(t *testing.T) {
    store := setupTestStore(t)
    ctx := context.Background()

    conversationID := uuid.New()
    userID := uuid.New()

    // Create
    cs := ConversationState{
        ConversationID:   conversationID,
        TelegramUserID:   123456789,
        UserID:           userID,
        ConversationType: "moment",
        CurrentStep:      "moment:awaiting_situation",
        Data:             []byte("{}"),
        LastActivity:     time.Now(),
        DateCreated:      time.Now(),
        DateUpdated:      time.Now(),
    }

    err := store.CreateConversationState(ctx, cs)
    if err != nil {
        t.Fatalf("creating conversation state: %v", err)
    }

    // Read
    got, err := store.QueryConversationStateByTelegramUserID(ctx, 123456789)
    if err != nil {
        t.Fatalf("querying conversation state: %v", err)
    }

    if got.ConversationID != conversationID {
        t.Errorf("expected conversation_id %s, got %s", conversationID, got.ConversationID)
    }

    // Update
    got.CurrentStep = "moment:awaiting_thoughts"
    got.DateUpdated = time.Now()
    err = store.UpdateConversationState(ctx, got)
    if err != nil {
        t.Fatalf("updating conversation state: %v", err)
    }

    // Verify update
    got, _ = store.QueryConversationStateByTelegramUserID(ctx, 123456789)
    if got.CurrentStep != "moment:awaiting_thoughts" {
        t.Errorf("expected step 'moment:awaiting_thoughts', got %s", got.CurrentStep)
    }

    // Delete
    err = store.DeleteConversationState(ctx, conversationID)
    if err != nil {
        t.Fatalf("deleting conversation state: %v", err)
    }
}
```

---

## Task 6: Implement User-Telegram Link CRUD

**File**: `business/sdk/sqldb/telegramdb/user_telegram_links.go`

```go
package telegramdb

import (
    "context"
    "errors"
    "fmt"

    "github.com/francowini/rafiki/business/sdk/sqldb"
    "github.com/google/uuid"
)

// CreateUserTelegramLink creates a link between a user and Telegram account.
func (s *Store) CreateUserTelegramLink(ctx context.Context, link UserTelegramLink) error {
    const q = `
    INSERT INTO user_telegram_links (
        user_id, telegram_user_id, linked_at, date_created, date_updated
    ) VALUES (
        :user_id, :telegram_user_id, :linked_at, :date_created, :date_updated
    )`

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, link); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}

// QueryUserTelegramLinkByUserID retrieves a link by user_id.
func (s *Store) QueryUserTelegramLinkByUserID(ctx context.Context, userID uuid.UUID) (UserTelegramLink, error) {
    const q = `
    SELECT
        user_id, telegram_user_id, linked_at, date_created, date_updated
    FROM user_telegram_links
    WHERE user_id = $1`

    var link UserTelegramLink
    if err := sqldb.Get(ctx, s.log, s.db, q, &link, userID); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return UserTelegramLink{}, fmt.Errorf("user not linked to telegram")
        }
        return UserTelegramLink{}, fmt.Errorf("selecting user telegram link: %w", err)
    }

    return link, nil
}

// QueryUserTelegramLinkByTelegramUserID retrieves a link by telegram_user_id.
func (s *Store) QueryUserTelegramLinkByTelegramUserID(ctx context.Context, telegramUserID int64) (UserTelegramLink, error) {
    const q = `
    SELECT
        user_id, telegram_user_id, linked_at, date_created, date_updated
    FROM user_telegram_links
    WHERE telegram_user_id = $1`

    var link UserTelegramLink
    if err := sqldb.Get(ctx, s.log, s.db, q, &link, telegramUserID); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return UserTelegramLink{}, fmt.Errorf("telegram user not linked to any account")
        }
        return UserTelegramLink{}, fmt.Errorf("selecting user telegram link: %w", err)
    }

    return link, nil
}

// DeleteUserTelegramLink removes the link between a user and Telegram account.
func (s *Store) DeleteUserTelegramLink(ctx context.Context, userID uuid.UUID) error {
    const q = `DELETE FROM user_telegram_links WHERE user_id = $1`

    if err := sqldb.ExecContext(ctx, s.log, s.db, q, userID); err != nil {
        return fmt.Errorf("deleting user telegram link: %w", err)
    }

    return nil
}
```

---

## Checklist

### telegramdb.go
- [ ] Create `Store` struct with log and db fields
- [ ] Implement `NewStore` constructor
- [ ] Implement `WithinTran` for transaction support
- [ ] Test store creation

### models.go
- [ ] Define `TelegramUser` struct
- [ ] Define `UserTelegramLink` struct
- [ ] Define `TelegramLinkCode` struct
- [ ] Define `ConversationState` struct
- [ ] Ensure all fields match database schema
- [ ] Use proper nullable types (`*string`, `*int`, `*time.Time`)

### telegram_users.go
- [ ] Implement `CreateTelegramUser` (with UPSERT)
- [ ] Implement `QueryTelegramUserByID`
- [ ] Implement `DeleteTelegramUser`
- [ ] Write tests for all methods

### link_codes.go
- [ ] Implement `CreateLinkCode`
- [ ] Implement `QueryLinkCodeByCode`
- [ ] Implement `ConsumeLinkCode`
- [ ] Implement `DeleteExpiredLinkCodes` (cleanup job)
- [ ] Implement `CountRecentLinkCodes` (rate limiting)
- [ ] Write tests for all methods

### conversation_states.go
- [ ] Implement `CreateConversationState`
- [ ] Implement `QueryConversationStateByTelegramUserID`
- [ ] Implement `UpdateConversationState`
- [ ] Implement `DeleteConversationState`
- [ ] Implement `DeleteAbandonedConversations` (cleanup job)
- [ ] Write tests for all methods

### user_telegram_links.go
- [ ] Implement `CreateUserTelegramLink`
- [ ] Implement `QueryUserTelegramLinkByUserID`
- [ ] Implement `QueryUserTelegramLinkByTelegramUserID`
- [ ] Implement `DeleteUserTelegramLink`
- [ ] Write tests for all methods

### Testing
- [ ] Set up test database (use testcontainers or local PostgreSQL)
- [ ] Run all tests with `go test ./business/sdk/sqldb/telegramdb/...`
- [ ] Verify 100% code coverage on CRUD operations
- [ ] Test error cases (not found, duplicate keys, etc.)
- [ ] Test transaction rollback scenarios

---

**Status**: ⏭️ Ready for Implementation
**Next Task**: [03-backend-foundation.md](./03-backend-foundation.md) - Foundation telegram package
