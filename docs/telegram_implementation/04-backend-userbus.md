# Backend Implementation - UserBus Telegram Methods

**Task Category**: Backend / Business Logic
**Estimated Time**: 3-4 hours
**Prerequisites**:
- [01-backend-database.md](./01-backend-database.md) - Database schema
- [02-backend-telegramdb.md](./02-backend-telegramdb.md) - Database access layer
- [03-backend-foundation.md](./03-backend-foundation.md) - Foundation package
**Dependencies**: None

---

## Overview

Add Telegram account linking methods to the existing `userbus` package. These methods enable users to:
- Generate link codes
- Link their Telegram account
- Check link status
- Unlink their Telegram account

**Pattern**: Follow existing userbus patterns (QueryByID, Create, Update, etc.)

---

## Files to Modify

```
business/domain/userbus/
├── userbus.go         # UPDATED - Add Telegram methods
├── telegram.go        # NEW - Telegram-specific methods
└── model.go           # UPDATED - Add TelegramLinkStatus model
```

---

## Task 1: Add Models

**File**: `business/domain/userbus/model.go`

Add to the end of the file:

```go
// TelegramLinkStatus represents the Telegram link status for a user.
type TelegramLinkStatus struct {
    Linked           bool
    TelegramUserID   *int64
    TelegramUsername *string
    LinkedAt         *time.Time
}

// TelegramLinkCode represents a temporary link code.
type TelegramLinkCode struct {
    Code      string
    ExpiresAt time.Time
}
```

---

## Task 2: Create Telegram Methods File

**File**: `business/domain/userbus/telegram.go`

```go
package userbus

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "strings"
    "time"

    "github.com/francowini/rafiki/business/sdk/sqldb/telegramdb"
    "github.com/francowini/rafiki/foundation/telegram"
    "github.com/google/uuid"
)

const (
    linkCodeLength    = 16
    linkCodeTTL       = 5 * time.Minute
    maxLinkCodesPerHour = 5
)

// GenerateLinkCode generates a new link code for connecting a Telegram account.
func (b *Business) GenerateLinkCode(ctx context.Context, userID uuid.UUID) (TelegramLinkCode, error) {
    // Check if user is already linked
    status, err := b.QueryTelegramLinkStatus(ctx, userID)
    if err != nil {
        return TelegramLinkCode{}, fmt.Errorf("query link status: %w", err)
    }

    if status.Linked {
        return TelegramLinkCode{}, telegram.ErrAlreadyLinked
    }

    // Rate limiting: Check how many codes generated in last hour
    count, err := b.telegramStore.CountRecentLinkCodes(ctx, userID)
    if err != nil {
        return TelegramLinkCode{}, fmt.Errorf("count recent link codes: %w", err)
    }

    if count >= maxLinkCodesPerHour {
        return TelegramLinkCode{}, fmt.Errorf("rate limit exceeded: max %d link codes per hour", maxLinkCodesPerHour)
    }

    // Generate random code
    code, err := generateLinkCode()
    if err != nil {
        return TelegramLinkCode{}, fmt.Errorf("generate link code: %w", err)
    }

    // Store in database
    lc := telegramdb.TelegramLinkCode{
        Code:        code,
        UserID:      userID,
        ExpiresAt:   time.Now().Add(linkCodeTTL),
        DateCreated: time.Now(),
    }

    if err := b.telegramStore.CreateLinkCode(ctx, lc); err != nil {
        return TelegramLinkCode{}, fmt.Errorf("create link code: %w", err)
    }

    b.log.Info(ctx, "link code generated",
        "user_id", userID,
        "code", code,
        "expires_at", lc.ExpiresAt)

    return TelegramLinkCode{
        Code:      code,
        ExpiresAt: lc.ExpiresAt,
    }, nil
}

// LinkTelegramAccount links a Telegram user to a Rafiki account using a link code.
func (b *Business) LinkTelegramAccount(ctx context.Context, code string, telegramUserID int64, telegramUsername string) error {
    // Validate and consume link code
    lc, err := b.telegramStore.QueryLinkCodeByCode(ctx, code)
    if err != nil {
        return telegram.ErrInvalidLinkCode
    }

    // Check if code is expired
    if time.Now().After(lc.ExpiresAt) {
        return telegram.ErrInvalidLinkCode
    }

    // Check if code already consumed
    if lc.ConsumedAt != nil {
        return telegram.ErrInvalidLinkCode
    }

    // Check if Telegram user is already linked to another account
    existingLink, err := b.telegramStore.QueryUserTelegramLinkByTelegramUserID(ctx, telegramUserID)
    if err == nil {
        // Already linked to another account
        if existingLink.UserID != lc.UserID {
            return fmt.Errorf("telegram account already linked to different user")
        }
        // Already linked to same account - idempotent
        return nil
    }

    // Create or update Telegram user record
    tu := telegramdb.TelegramUser{
        TelegramUserID: telegramUserID,
        Username:       stringPtr(telegramUsername),
        DateCreated:    time.Now(),
        DateUpdated:    time.Now(),
    }

    if err := b.telegramStore.CreateTelegramUser(ctx, tu); err != nil {
        return fmt.Errorf("create telegram user: %w", err)
    }

    // Create link
    link := telegramdb.UserTelegramLink{
        UserID:         lc.UserID,
        TelegramUserID: telegramUserID,
        LinkedAt:       time.Now(),
        DateCreated:    time.Now(),
        DateUpdated:    time.Now(),
    }

    if err := b.telegramStore.CreateUserTelegramLink(ctx, link); err != nil {
        return fmt.Errorf("create user telegram link: %w", err)
    }

    // Consume link code
    if err := b.telegramStore.ConsumeLinkCode(ctx, code, telegramUserID); err != nil {
        return fmt.Errorf("consume link code: %w", err)
    }

    b.log.Info(ctx, "telegram account linked",
        "user_id", lc.UserID,
        "telegram_user_id", telegramUserID,
        "telegram_username", telegramUsername)

    return nil
}

// QueryTelegramLinkStatus returns the Telegram link status for a user.
func (b *Business) QueryTelegramLinkStatus(ctx context.Context, userID uuid.UUID) (TelegramLinkStatus, error) {
    link, err := b.telegramStore.QueryUserTelegramLinkByUserID(ctx, userID)
    if err != nil {
        // Not linked
        return TelegramLinkStatus{Linked: false}, nil
    }

    // Get Telegram user details
    tu, err := b.telegramStore.QueryTelegramUserByID(ctx, link.TelegramUserID)
    if err != nil {
        return TelegramLinkStatus{Linked: false}, nil
    }

    return TelegramLinkStatus{
        Linked:           true,
        TelegramUserID:   &link.TelegramUserID,
        TelegramUsername: tu.Username,
        LinkedAt:         &link.LinkedAt,
    }, nil
}

// QueryUserIDByTelegramUserID returns the Rafiki user ID for a Telegram user.
func (b *Business) QueryUserIDByTelegramUserID(ctx context.Context, telegramUserID int64) (uuid.UUID, error) {
    link, err := b.telegramStore.QueryUserTelegramLinkByTelegramUserID(ctx, telegramUserID)
    if err != nil {
        return uuid.UUID{}, telegram.ErrNotLinked
    }

    return link.UserID, nil
}

// UnlinkTelegramAccount unlinks a Telegram account from a Rafiki user.
func (b *Business) UnlinkTelegramAccount(ctx context.Context, userID uuid.UUID) error {
    // Check if linked
    status, err := b.QueryTelegramLinkStatus(ctx, userID)
    if err != nil {
        return fmt.Errorf("query link status: %w", err)
    }

    if !status.Linked {
        return telegram.ErrNotLinked
    }

    // Delete link
    if err := b.telegramStore.DeleteUserTelegramLink(ctx, userID); err != nil {
        return fmt.Errorf("delete user telegram link: %w", err)
    }

    b.log.Info(ctx, "telegram account unlinked",
        "user_id", userID,
        "telegram_user_id", *status.TelegramUserID)

    return nil
}

// generateLinkCode generates a cryptographically random link code.
func generateLinkCode() (string, error) {
    bytes := make([]byte, linkCodeLength)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("generating random bytes: %w", err)
    }

    // Encode to base64 and make URL-safe
    code := base64.URLEncoding.EncodeToString(bytes)
    code = strings.TrimRight(code, "=")  // Remove padding
    code = "RAFIKI-" + code[:12]         // Prefix and trim to 12 chars

    return code, nil
}

func stringPtr(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
```

---

## Task 3: Update Business Struct

**File**: `business/domain/userbus/userbus.go`

Add to the `Business` struct:

```go
type Business struct {
    log           *logger.Logger
    userBus       *userbus.Core
    delegate      *delegate
    usrStore      userdb.Store
    telegramStore telegramdb.Store  // NEW
}
```

Update `NewBusiness` constructor:

```go
func NewBusiness(log *logger.Logger, delegate *delegate, usrStore userdb.Store, telegramStore telegramdb.Store) *Business {
    return &Business{
        log:           log,
        delegate:      delegate,
        usrStore:      usrStore,
        telegramStore: telegramStore,  // NEW
    }
}
```

**Note**: You'll need to update all places where `NewBusiness` is called to pass `telegramStore`.

---

## Task 4: Update Main Service Initialization

**File**: `api/services/partners/main.go`

Update userbus initialization to include telegram store:

```go
// Initialize stores
userStore := userdb.NewStore(log, db)
telegramStore := telegramdb.NewStore(log, db)  // NEW

// Initialize user business
userBus := userbus.NewBusiness(log, delegate, userStore, telegramStore)  // UPDATED
```

---

## Task 5: Write Tests

**File**: `business/domain/userbus/telegram_test.go`

```go
package userbus_test

import (
    "context"
    "testing"
    "time"

    "github.com/francowini/rafiki/business/domain/userbus"
    "github.com/google/uuid"
)

func TestUserBus_GenerateLinkCode(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    // Create test user
    userID := createTestUser(t, bus)

    // Generate link code
    linkCode, err := bus.GenerateLinkCode(ctx, userID)
    if err != nil {
        t.Fatalf("generating link code: %v", err)
    }

    if linkCode.Code == "" {
        t.Error("expected non-empty link code")
    }

    if linkCode.ExpiresAt.Before(time.Now()) {
        t.Error("expected future expiry time")
    }

    // Verify code starts with RAFIKI-
    if !strings.HasPrefix(linkCode.Code, "RAFIKI-") {
        t.Errorf("expected code to start with 'RAFIKI-', got %s", linkCode.Code)
    }
}

func TestUserBus_GenerateLinkCode_AlreadyLinked(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    userID := createTestUser(t, bus)

    // Generate and consume link code
    linkCode, _ := bus.GenerateLinkCode(ctx, userID)
    _ = bus.LinkTelegramAccount(ctx, linkCode.Code, 123456789, "test_user")

    // Try to generate another link code
    _, err := bus.GenerateLinkCode(ctx, userID)
    if err == nil {
        t.Error("expected error when user already linked")
    }
}

func TestUserBus_LinkTelegramAccount(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    userID := createTestUser(t, bus)

    // Generate link code
    linkCode, _ := bus.GenerateLinkCode(ctx, userID)

    // Link account
    err := bus.LinkTelegramAccount(ctx, linkCode.Code, 123456789, "test_user")
    if err != nil {
        t.Fatalf("linking telegram account: %v", err)
    }

    // Verify link status
    status, err := bus.QueryTelegramLinkStatus(ctx, userID)
    if err != nil {
        t.Fatalf("querying link status: %v", err)
    }

    if !status.Linked {
        t.Error("expected user to be linked")
    }

    if status.TelegramUserID == nil || *status.TelegramUserID != 123456789 {
        t.Error("expected telegram_user_id to match")
    }
}

func TestUserBus_LinkTelegramAccount_InvalidCode(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    // Try to link with invalid code
    err := bus.LinkTelegramAccount(ctx, "INVALID-CODE", 123456789, "test_user")
    if err == nil {
        t.Error("expected error with invalid link code")
    }
}

func TestUserBus_LinkTelegramAccount_ExpiredCode(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    userID := createTestUser(t, bus)

    // Generate link code
    linkCode, _ := bus.GenerateLinkCode(ctx, userID)

    // Wait for expiry (in real test, mock time.Now)
    time.Sleep(6 * time.Minute)

    // Try to link with expired code
    err := bus.LinkTelegramAccount(ctx, linkCode.Code, 123456789, "test_user")
    if err == nil {
        t.Error("expected error with expired link code")
    }
}

func TestUserBus_UnlinkTelegramAccount(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    userID := createTestUser(t, bus)

    // Link account
    linkCode, _ := bus.GenerateLinkCode(ctx, userID)
    _ = bus.LinkTelegramAccount(ctx, linkCode.Code, 123456789, "test_user")

    // Unlink
    err := bus.UnlinkTelegramAccount(ctx, userID)
    if err != nil {
        t.Fatalf("unlinking telegram account: %v", err)
    }

    // Verify unlinked
    status, _ := bus.QueryTelegramLinkStatus(ctx, userID)
    if status.Linked {
        t.Error("expected user to be unlinked")
    }
}

func TestUserBus_QueryUserIDByTelegramUserID(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    userID := createTestUser(t, bus)

    // Link account
    linkCode, _ := bus.GenerateLinkCode(ctx, userID)
    _ = bus.LinkTelegramAccount(ctx, linkCode.Code, 123456789, "test_user")

    // Query by Telegram user ID
    gotUserID, err := bus.QueryUserIDByTelegramUserID(ctx, 123456789)
    if err != nil {
        t.Fatalf("querying user id: %v", err)
    }

    if gotUserID != userID {
        t.Errorf("expected user_id %s, got %s", userID, gotUserID)
    }
}

func TestUserBus_RateLimit_LinkCodes(t *testing.T) {
    bus := setupTestBusiness(t)
    ctx := context.Background()

    userID := createTestUser(t, bus)

    // Generate maximum allowed codes
    for i := 0; i < 5; i++ {
        _, err := bus.GenerateLinkCode(ctx, userID)
        if err != nil {
            t.Fatalf("generating link code %d: %v", i+1, err)
        }
    }

    // Try to generate one more (should hit rate limit)
    _, err := bus.GenerateLinkCode(ctx, userID)
    if err == nil {
        t.Error("expected rate limit error")
    }
}
```

---

## Checklist

### Models
- [ ] Add `TelegramLinkStatus` to `model.go`
- [ ] Add `TelegramLinkCode` to `model.go`

### Telegram Methods
- [ ] Create `telegram.go` file
- [ ] Implement `GenerateLinkCode`
- [ ] Implement `LinkTelegramAccount`
- [ ] Implement `QueryTelegramLinkStatus`
- [ ] Implement `QueryUserIDByTelegramUserID`
- [ ] Implement `UnlinkTelegramAccount`
- [ ] Implement `generateLinkCode` helper

### Business Struct
- [ ] Add `telegramStore` field to `Business` struct
- [ ] Update `NewBusiness` constructor signature
- [ ] Update all call sites of `NewBusiness` in codebase

### Main Service
- [ ] Update `main.go` to create `telegramStore`
- [ ] Pass `telegramStore` to `userbus.NewBusiness`

### Tests
- [ ] Create `telegram_test.go`
- [ ] Test `GenerateLinkCode` (happy path)
- [ ] Test `GenerateLinkCode` (already linked error)
- [ ] Test `GenerateLinkCode` (rate limit)
- [ ] Test `LinkTelegramAccount` (happy path)
- [ ] Test `LinkTelegramAccount` (invalid code)
- [ ] Test `LinkTelegramAccount` (expired code)
- [ ] Test `LinkTelegramAccount` (already consumed code)
- [ ] Test `UnlinkTelegramAccount`
- [ ] Test `QueryUserIDByTelegramUserID`
- [ ] Test `QueryTelegramLinkStatus`

### Integration
- [ ] Run all userbus tests: `go test ./business/domain/userbus/...`
- [ ] Verify no regressions in existing userbus functionality
- [ ] Run golangci-lint on modified files

---

## Security Considerations

### Link Code Generation
- ✅ Cryptographically random (crypto/rand)
- ✅ 16-byte entropy (128 bits)
- ✅ URL-safe encoding
- ✅ Unique constraint in database (prevents collisions)

### Rate Limiting
- ✅ Max 5 codes per user per hour
- ✅ Prevents abuse/spam
- ✅ Database-based (persists across service restarts)

### Link Code Validation
- ✅ Expiry check (5 minutes TTL)
- ✅ Single-use check (consumed_at != NULL)
- ✅ Ownership check (code belongs to correct user)

### Privacy
- ✅ Only store telegram_user_id, username (minimal data)
- ✅ No phone number, profile photo, or other PII
- ✅ Easy deletion (CASCADE on user deletion)

---

**Status**: ⏭️ Ready for Implementation
**Next Task**: [05-backend-moment-plugin.md](./05-backend-moment-plugin.md) - Moment conversation plugin
