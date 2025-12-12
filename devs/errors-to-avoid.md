# Critical Errors to Avoid

This document catalogs critical errors discovered during code review that should be avoided in future development. Use this as a checklist when implementing new features.

---

## Table of Contents

### Backend (Go)

1. [Security: Child Entity Ownership Validation](#1-security-child-entity-ownership-validation)
2. [Error Handling: Sentinel vs Generic Errors](#2-error-handling-sentinel-vs-generic-errors)
3. [String Length: UTF-8 Rune Count vs Byte Count](#3-string-length-utf-8-rune-count-vs-byte-count)
4. [Strong Types: Missing Value() Method](#4-strong-types-missing-value-method)
5. [SQL: WHERE Clause Building](#5-sql-where-clause-building)
6. [Logging: Missing Structured Logging in Business Layer](#6-logging-missing-structured-logging-in-business-layer)
7. [Code Duplication: Repeated Decrypt-Parse Patterns](#7-code-duplication-repeated-decrypt-parse-patterns)
8. [Thread Safety: Package-Level Random Sources](#8-thread-safety-package-level-random-sources)
9. [Timezone: Using time.Local in Database Conversions](#9-timezone-using-timelocal-in-database-conversions)
10. [Idempotency: Duplicate Scheduled Messages](#10-idempotency-duplicate-scheduled-messages)
11. [Input Validation: Missing Business Layer Validation](#11-input-validation-missing-business-layer-validation)
12. [SQL Views: ORDER BY in View Definition](#12-sql-views-order-by-in-view-definition)
13. [Shell Scripts: Unvalidated Environment Variables](#13-shell-scripts-unvalidated-environment-variables)
14. [API Parameters: Missing Max Limits](#14-api-parameters-missing-max-limits)
15. [Strong Types: When NOT to Use Them](#15-strong-types-when-not-to-use-them)

### Frontend (TypeScript/React)

F1. [Security: Markdown Injection in User Content](#f1-security-markdown-injection-in-user-content)
F2. [Async: Stale Responses in useEffect](#f2-async-stale-responses-in-useeffect)
F3. [Data: Silent Data Truncation](#f3-data-silent-data-truncation)
F4. [State: useSyncExternalStore + useState Duplication](#f4-state-usesyncexternalstore--usestate-duplication)
F5. [Accessibility: Clickable Div Without Keyboard Support](#f5-accessibility-clickable-div-without-keyboard-support)

---

## 1. Security: Child Entity Ownership Validation

### Severity: 🔴 CRITICAL (Security Vulnerability)

### Problem

When creating a child entity that references a parent entity by ID (e.g., creating a `LifeVision` linked to a `Value`), failing to validate that the authenticated user owns the parent entity allows **cross-user data linking attacks**.

### Bad Example

```go
// ❌ BAD: No ownership validation
func (b *Business) Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error) {
    value, err := b.valueBus.QueryByID(ctx, nlv.ValueID)
    if err != nil {
        return LifeVision{}, err
    }

    // Attacker can link their life vision to ANY user's value if they know the UUID
    lifeVision := LifeVision{
        UserID:  value.UserID,  // Inherits from value - no validation!
        ValueID: nlv.ValueID,
    }
    return b.storer.Create(ctx, lifeVision)
}
```

### Good Example

```go
// ✅ GOOD: Validates authenticated user owns the parent entity
type NewLifeVision struct {
    UserID  uuid.UUID  // Authenticated user (for ownership validation)
    ValueID uuid.UUID
    Content lifevisioncontent.LifeVisionContent
}

func (b *Business) Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error) {
    value, err := b.valueBus.QueryByID(ctx, nlv.ValueID)
    if err != nil {
        return LifeVision{}, err
    }

    // Security: Verify authenticated user owns the parent value
    if value.UserID != nlv.UserID {
        return LifeVision{}, ErrNotValueOwner
    }

    lifeVision := LifeVision{
        UserID:  value.UserID,
        ValueID: nlv.ValueID,
    }
    return b.storer.Create(ctx, lifeVision)
}
```

### App Layer Pattern

```go
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
    // Get authenticated user
    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return errs.New(errs.Unauthenticated, err)
    }

    // Pass userID to business layer for ownership validation
    nlv, err := toBusNewLifeVision(appLifeVision, userID)

    lifeVision, err := a.lifeVisionBus.Create(ctx, nlv)
    if errors.Is(err, lifevisionbus.ErrNotValueOwner) {
        return errs.New(errs.PermissionDenied, err)
    }
}
```

### Checklist

- [ ] Child entity creation includes authenticated userID in the request struct
- [ ] Business layer validates `parent.UserID == authenticatedUserID`
- [ ] Return `PermissionDenied` error for ownership violations
- [ ] App layer handles the ownership error appropriately

---

## 2. Error Handling: Sentinel vs Generic Errors

### Severity: 🟠 Major (Error Handling Bug)

### Problem

When the business layer encounters a permission error (e.g., user doesn't own the parent entity), returning a generic `fmt.Errorf()` message instead of a sentinel error prevents the app layer from properly detecting and handling the error with the correct HTTP status code.

### Bad Example

```go
// ❌ BAD: Generic error message - app layer can't detect this
func (b *Business) Update(ctx context.Context, entity Entity, upd UpdateEntity) (Entity, error) {
    if upd.ParentID != nil {
        parent, err := b.parentBus.QueryByID(ctx, *upd.ParentID)
        if err != nil {
            return Entity{}, err
        }

        // Generic error - app layer will return 500 Internal Error
        if parent.UserID != entity.UserID {
            return Entity{}, fmt.Errorf("cannot move entity to parent owned by different user")
        }
    }
    // ...
}
```

### Good Example

```go
// ✅ GOOD: Sentinel error - app layer can detect with errors.Is()
var ErrNotParentOwner = errors.New("user does not own the specified parent")

func (b *Business) Update(ctx context.Context, entity Entity, upd UpdateEntity) (Entity, error) {
    if upd.ParentID != nil {
        parent, err := b.parentBus.QueryByID(ctx, *upd.ParentID)
        if err != nil {
            return Entity{}, err
        }

        // Sentinel error - app layer can return 403 PermissionDenied
        if parent.UserID != entity.UserID {
            return Entity{}, ErrNotParentOwner
        }
    }
    // ...
}
```

### App Layer Pattern

```go
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
    // ...
    entity, err = a.entityBus.Update(ctx, entity, upd)
    if err != nil {
        if errors.Is(err, parentbus.ErrNotFound) {
            return errs.New(errs.NotFound, errors.New("parent not found"))
        }
        // Can now properly detect permission error
        if errors.Is(err, entitybus.ErrNotParentOwner) {
            return errs.New(errs.PermissionDenied, errors.New("user does not own the specified parent"))
        }
        return errs.Newf(errs.Internal, "update: %s", err)
    }
    // ...
}
```

### Checklist

- [ ] All permission violations return sentinel errors (defined in model.go)
- [ ] Sentinel errors are defined with `errors.New()` for `errors.Is()` compatibility
- [ ] App layer handles ALL sentinel errors with appropriate HTTP status codes
- [ ] Both Create AND Update operations validate ownership when changing parent references

---

## 3. String Length: UTF-8 Rune Count vs Byte Count

### Severity: 🟠 Major (Data Validation Bug)

### Problem

Using `len(string)` for character count validation fails for multi-byte UTF-8 characters. A 10-character string in Chinese/Japanese/emoji can have 30+ bytes.

### Bad Example

```go
// ❌ BAD: len() counts bytes, not characters
func Parse(value string) (Content, error) {
    if len(value) < 10 {  // "日本語です" is 5 chars but 15 bytes
        return Content{}, ErrTooShort
    }
    if len(value) > 500 {  // Will reject valid 200-char strings with multibyte chars
        return Content{}, ErrTooLong
    }
    return Content{value: value}, nil
}
```

### Good Example

```go
import "unicode/utf8"

// ✅ GOOD: utf8.RuneCountInString() counts actual characters
func Parse(value string) (Content, error) {
    runeCount := utf8.RuneCountInString(value)

    if runeCount < 10 {
        return Content{}, ErrTooShort
    }
    if runeCount > 500 {
        return Content{}, ErrTooLong
    }
    return Content{value: value}, nil
}
```

### Test Case

```go
func TestUTF8Validation(t *testing.T) {
    // 10 Japanese characters = 30 bytes
    content := "日本語で話す勉強中"  // 9 chars

    // With len(): would pass (27 bytes > 10)
    // With RuneCount(): correctly fails (9 chars < 10)
    _, err := Parse(content)
    if err != ErrTooShort {
        t.Error("should reject strings under 10 characters")
    }
}
```

### Affected Files (to fix)

- `business/types/valuecontent/valuecontent.go` - uses `len()`
- Any other content validation types

---

## 4. Strong Types: Missing Value() Method

### Severity: 🟡 Minor (API Inconsistency)

### Problem

Strong types should expose their underlying value through a `Value()` method for consistency with the codebase pattern. Having only `String()` makes it unclear if it's a formatted representation or the raw value.

### Bad Example

```go
// ❌ BAD: Missing Value() method
type Content struct {
    value string
}

func (c Content) String() string {
    return c.value
}

// Is String() the raw value or a formatted representation? Unclear.
```

### Good Example

```go
// ✅ GOOD: Value() for raw access, String() for representation
type Content struct {
    value string
}

// Value returns the raw underlying value
func (c Content) Value() string {
    return c.value
}

// String returns the string representation (may be same as Value)
func (c Content) String() string {
    return c.value
}
```

### Standard Strong Type Template

```go
type MyType struct {
    value string  // or int, etc.
}

// Value returns the underlying value
func (t MyType) Value() string { return t.value }

// String returns the string representation
func (t MyType) String() string { return t.value }

// Equal supports go-cmp and testing
func (t MyType) Equal(t2 MyType) bool { return t.value == t2.value }

// MarshalText supports logging and JSON
func (t MyType) MarshalText() ([]byte, error) { return []byte(t.value), nil }

// Parse creates a validated instance
func Parse(value string) (MyType, error) { ... }

// MustParse panics on error (tests only)
func MustParse(value string) MyType { ... }
```

---

## 5. SQL: WHERE Clause Building

### Severity: 🟢 Low (Code Quality)

### Problem

Building WHERE clauses with manual index checking is verbose and error-prone.

### Bad Example

```go
// ❌ BAD: Verbose loop with index checking
func buildWhereClause(filter QueryFilter, data map[string]any) string {
    var conditions []string

    if filter.ID != nil {
        conditions = append(conditions, "id = :id")
    }
    if filter.UserID != nil {
        conditions = append(conditions, "user_id = :user_id")
    }

    if len(conditions) == 0 {
        return ""
    }

    whereClause := " WHERE "
    for i, condition := range conditions {
        if i > 0 {
            whereClause += " AND "
        }
        whereClause += condition
    }
    return whereClause
}
```

### Good Example

```go
import "strings"

// ✅ GOOD: Use strings.Join for cleaner code
func buildWhereClause(filter QueryFilter, data map[string]any) string {
    var conditions []string

    if filter.ID != nil {
        data["id"] = *filter.ID
        conditions = append(conditions, "id = :id")
    }
    if filter.UserID != nil {
        data["user_id"] = *filter.UserID
        conditions = append(conditions, "user_id = :user_id")
    }

    if len(conditions) == 0 {
        return ""
    }

    return " WHERE " + strings.Join(conditions, " AND ")
}
```

---

## 6. Logging: Missing Structured Logging in Business Layer

### Severity: 🟠 Major (Observability Gap)

### Problem

Business layer methods that don't include structured logging make debugging production issues difficult. Without entry/exit logging, you can't trace request flow or identify where failures occur.

### Bad Example

```go
// ❌ BAD: No logging - silent failures, no observability
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Item, error) {
    if filter.UserID == uuid.Nil {
        return nil, ErrInvalidUserID
    }

    items, err := b.storer.Query(ctx, filter, orderBy, page)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }

    return items, nil
}
```

### Good Example

```go
// ✅ GOOD: Structured logging at entry, errors, and success
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Item, error) {
    // Log entry with relevant parameters
    b.log.Info(ctx, "domain.query", "userID", filter.UserID, "page", pg.Number(), "rows", pg.RowsPerPage())

    // Log validation failures
    if filter.UserID == uuid.Nil {
        b.log.Error(ctx, "domain.query", "err", ErrInvalidUserID)
        return nil, ErrInvalidUserID
    }

    items, err := b.storer.Query(ctx, filter, orderBy, pg)
    if err != nil {
        b.log.Error(ctx, "domain.query", "err", err, "userID", filter.UserID)
        return nil, fmt.Errorf("storer.query: %w", err)
    }

    // Log success with result count
    b.log.Info(ctx, "domain.query.success", "userID", filter.UserID, "count", len(items))

    return items, nil
}
```

### Business Struct Pattern

```go
// Business struct must include logger
type Business struct {
    log    *logger.Logger  // Required for structured logging
    storer Storer
}

// NewBusiness must accept logger as first parameter
func NewBusiness(log *logger.Logger, storer Storer) ExtBusiness {
    return &Business{
        log:    log,
        storer: storer,
    }
}
```

### Logging Format Guidelines

```go
// Entry log: action name + input parameters
b.log.Info(ctx, "domain.method", "param1", value1, "param2", value2)

// Error log: action name + error + relevant context
b.log.Error(ctx, "domain.method", "err", err, "userID", userID)

// Success log: action name + result summary
b.log.Info(ctx, "domain.method.success", "userID", userID, "count", len(items))
```

### Checklist

- [ ] Business struct includes `*logger.Logger` field
- [ ] NewBusiness accepts logger as first parameter
- [ ] All public methods log on entry with parameters
- [ ] All errors are logged before returning
- [ ] Success cases log result summary (count, IDs, etc.)
- [ ] Log messages use `domain.method` naming convention

---

## 7. Code Duplication: Repeated Decrypt-Parse Patterns

### Severity: 🟡 Minor (Maintainability)

### Problem

When multiple fields require the same decrypt→parse→assign pattern, copying the same code block for each field creates maintenance burden and inconsistency risk.

### Bad Example

```go
// ❌ BAD: Repeated pattern for each field
func toBusItem(db dbItem, enc encrypt.Encryptor) (Item, error) {
    item := Item{ID: db.ID}

    if db.Field1.Valid {
        decrypted, err := enc.Decrypt(db.Field1.String)
        if err != nil {
            return Item{}, fmt.Errorf("decrypt field1: %w", err)
        }
        parsed, err := content.Parse(decrypted)
        if err != nil {
            return Item{}, fmt.Errorf("parse field1: %w", err)
        }
        item.Field1 = &parsed
    }

    if db.Field2.Valid {
        decrypted, err := enc.Decrypt(db.Field2.String)
        if err != nil {
            return Item{}, fmt.Errorf("decrypt field2: %w", err)
        }
        parsed, err := content.Parse(decrypted)
        if err != nil {
            return Item{}, fmt.Errorf("parse field2: %w", err)
        }
        item.Field2 = &parsed
    }
    // ... repeated 5 more times
}
```

### Good Example

```go
// ✅ GOOD: Extract helper function for common pattern
func decryptContent(enc encrypt.Encryptor, field sql.NullString, fieldName string) (*content.Content, error) {
    if !field.Valid {
        return nil, nil
    }

    decrypted, err := enc.Decrypt(field.String)
    if err != nil {
        return nil, fmt.Errorf("decrypt %s: %w", fieldName, err)
    }

    parsed, err := content.Parse(decrypted)
    if err != nil {
        return nil, fmt.Errorf("parse %s: %w", fieldName, err)
    }

    return &parsed, nil
}

func toBusItem(db dbItem, enc encrypt.Encryptor) (Item, error) {
    item := Item{ID: db.ID}
    var err error

    if item.Field1, err = decryptContent(enc, db.Field1, "field1"); err != nil {
        return Item{}, err
    }
    if item.Field2, err = decryptContent(enc, db.Field2, "field2"); err != nil {
        return Item{}, err
    }
    // ... clean and concise

    return item, nil
}
```

### When to Extract Helpers

- Pattern repeated 3+ times
- Same error wrapping logic needed
- NULL handling is identical
- Makes the calling code significantly cleaner

### Checklist

- [ ] Identify repeated patterns in conversion functions
- [ ] Extract helper with clear name describing the operation
- [ ] Include field name parameter for error context
- [ ] Handle NULL/empty cases consistently
- [ ] Keep specialized parsing (intensity, category) inline if unique

---

## 8. Thread Safety: Package-Level Random Sources

### Severity: 🟠 Major (Race Condition)

### Problem

Creating a package-level `rand.Rand` instance with `rand.New(rand.NewSource(...))` is not safe for concurrent use. Multiple goroutines calling `rng.Intn()` simultaneously can cause data races.

### Bad Example

```go
// ❌ BAD: Not thread-safe
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func GetRandomTemplate() string {
    templates := []string{"a", "b", "c"}
    return templates[rng.Intn(len(templates))] // DATA RACE!
}
```

### Good Example (Go 1.22+)

```go
import "math/rand/v2"

// ✅ GOOD: Use thread-safe rand/v2 package
func GetRandomTemplate() string {
    templates := []string{"a", "b", "c"}
    return templates[rand.IntN(len(templates))] // Thread-safe
}
```

### Good Example (Pre Go 1.22)

```go
import (
    "math/rand"
    "sync"
)

// ✅ GOOD: Protect with mutex
var (
    rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
    rngMu sync.Mutex
)

func GetRandomTemplate() string {
    templates := []string{"a", "b", "c"}
    rngMu.Lock()
    idx := rng.Intn(len(templates))
    rngMu.Unlock()
    return templates[idx]
}
```

### Checklist

- [ ] Never use `rand.New(rand.NewSource(...))` at package level without mutex
- [ ] Prefer `math/rand/v2` on Go 1.22+ for thread-safe global generator
- [ ] Run tests with `-race` flag to detect data races

---

## 9. Timezone: Using time.Local in Database Conversions

### Severity: 🟠 Major (Non-Deterministic Behavior)

### Problem

Using `time.Local` when converting timestamps from the database ties behavior to the host's timezone, causing inconsistent results across different servers or containers.

### Bad Example

```go
// ❌ BAD: Tied to server timezone
func toBusMessage(db message) Message {
    return Message{
        ScheduledAt: db.ScheduledAt.In(time.Local), // Different on each server!
        DateCreated: db.DateCreated.In(time.Local),
    }
}

func toLocalPtr(t *time.Time) *time.Time {
    if t == nil {
        return nil
    }
    local := t.In(time.Local) // Non-deterministic
    return &local
}
```

### Good Example

```go
// ✅ GOOD: Always use UTC for storage layer
func toBusMessage(db message) Message {
    return Message{
        ScheduledAt: db.ScheduledAt.UTC(), // Consistent everywhere
        DateCreated: db.DateCreated.UTC(),
    }
}

func toUTCPtr(t *time.Time) *time.Time {
    if t == nil {
        return nil
    }
    utc := t.UTC()
    return &utc
}
```

### Checklist

- [ ] Store all timestamps in UTC in the database
- [ ] Convert to UTC when writing: `time.Now().UTC()` or `.UTC()`
- [ ] Return UTC from store layer, let presentation layer handle timezone display
- [ ] Never use `time.Local` in business or store layers

---

## 10. Idempotency: Duplicate Scheduled Messages

### Severity: 🟠 Major (Data Integrity)

### Problem

Scheduled operations (like sending daily notifications) can create duplicate records if the job runs multiple times within the same time window without idempotency controls.

### Bad Example

```go
// ❌ BAD: No duplicate prevention
func (w *Worker) scheduleMorningMessage(ctx context.Context, userID uuid.UUID) error {
    // If job runs at 08:03 and 08:07, two messages are created!
    _, err = w.notificationBus.Create(ctx, notificationbus.NewMessage{
        UserID:      userID,
        MessageType: notificationbus.MessageTypeMorning,
        Content:     content,
        ScheduledAt: time.Now(),
    })
    return err
}
```

### Good Example

**1. Add unique constraint in SQL:**

```sql
-- Prevent duplicate messages per user/type/date
CREATE UNIQUE INDEX IF NOT EXISTS notification_messages_schedule_unique_idx
  ON notification_messages(user_id, message_type, DATE(scheduled_at))
  WHERE status = 'pending';
```

**2. Handle constraint violation gracefully:**

```go
// ✅ GOOD: Idempotent - duplicate is expected behavior
func (w *Worker) scheduleMorningMessage(ctx context.Context, userID uuid.UUID) error {
    _, err = w.notificationBus.Create(ctx, notificationbus.NewMessage{
        UserID:      userID,
        MessageType: notificationbus.MessageTypeMorning,
        Content:     content,
        ScheduledAt: time.Now(),
    })
    if err != nil {
        if errors.Is(err, notificationbus.ErrDuplicateSchedule) {
            // Already scheduled for today - not an error
            return nil
        }
        return fmt.Errorf("create message: %w", err)
    }
    return nil
}
```

**3. Detect unique violation in store:**

```go
func (s *Store) Create(ctx context.Context, msg Message) error {
    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbMsg); err != nil {
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return ErrDuplicateSchedule
        }
        return fmt.Errorf("namedexeccontext: %w", err)
    }
    return nil
}
```

### Checklist

- [ ] Add unique constraint on scheduling key (user_id, type, date)
- [ ] Define sentinel error (e.g., `ErrDuplicateSchedule`)
- [ ] Store layer detects `ErrDBDuplicatedEntry` and returns sentinel
- [ ] Worker treats duplicate as success, not error

---

## 11. Input Validation: Missing Business Layer Validation

### Severity: 🟠 Major (Data Integrity)

### Problem

Persisting data without validation in the business layer allows invalid data to reach the database, relying solely on database constraints (if any) for validation.

### Bad Example

```go
// ❌ BAD: No validation - trusts all input
func (b *Business) Create(ctx context.Context, nm NewMessage) (Message, error) {
    msg := Message{
        ID:          uuid.New(),
        MessageType: nm.MessageType, // Could be invalid!
        Content:     nm.Content,     // Could be empty!
    }
    return b.storer.Create(ctx, msg)
}
```

### Good Example

```go
// ✅ GOOD: Validate before persisting
func (b *Business) Create(ctx context.Context, nm NewMessage) (Message, error) {
    // Validate MessageType
    if !nm.MessageType.Valid() {
        return Message{}, ErrInvalidMessageType
    }

    // Validate Content
    if strings.TrimSpace(nm.Content) == "" {
        return Message{}, ErrContentEmpty
    }

    msg := Message{
        ID:          uuid.New(),
        MessageType: nm.MessageType,
        Content:     nm.Content,
    }

    if err := b.storer.Create(ctx, msg); err != nil {
        return Message{}, fmt.Errorf("create: %w", err)
    }
    return msg, nil
}

// Add Valid() method to enum types
func (mt MessageType) Valid() bool {
    switch mt {
    case MessageTypeMorning, MessageTypeEvening, MessageTypeTest:
        return true
    }
    return false
}
```

### Checklist

- [ ] Enum types have `Valid()` method
- [ ] Business layer validates all input before persisting
- [ ] Define sentinel errors for each validation failure
- [ ] Empty/whitespace strings are rejected where required

---

## 12. SQL Views: ORDER BY in View Definition

### Severity: 🟢 Low (Performance/Maintainability)

### Problem

Including `ORDER BY` in a SQL view definition is often ignored by the database optimizer and can mislead developers about result ordering.

### Bad Example

```sql
-- ❌ BAD: ORDER BY in view is often ignored
CREATE OR REPLACE VIEW view_notification_content AS
SELECT user_id, value_id, content, display_order
FROM values
ORDER BY user_id, display_order;  -- May be ignored!
```

### Good Example

```sql
-- ✅ GOOD: View defines projection only
CREATE OR REPLACE VIEW view_notification_content AS
SELECT user_id, value_id, content, display_order
FROM values;
-- Note: ORDER BY intentionally omitted - consumers should order when querying
```

```go
// Consumer applies ordering
const q = `
SELECT * FROM view_notification_content
WHERE user_id = :user_id
ORDER BY display_order`  // Explicit ordering at query time
```

### Checklist

- [ ] Remove ORDER BY from view definitions
- [ ] Add comment explaining ordering is consumer's responsibility
- [ ] Apply ORDER BY in the query/store layer instead

---

## 13. Shell Scripts: Unvalidated Environment Variables

### Severity: 🟠 Major (Runtime Failure)

### Problem

Using environment variables in shell scripts without validation leads to cryptic errors when variables are missing or empty.

### Bad Example

```bash
# ❌ BAD: No validation - cryptic error if vars missing
if [ -n "$PARTNER_DB_HOST" ]; then
    DB_URL="postgresql://${PARTNER_DB_USER}:${PARTNER_DB_PASSWORD}@${PARTNER_DB_HOST}:${PARTNER_DB_PORT}/${PARTNER_DB_NAME}"
    psql "$DB_URL" -c "$SQL"  # Fails with confusing error
fi
```

### Good Example

```bash
# ✅ GOOD: Validate all required variables
if [ -n "$PARTNER_DB_HOST" ]; then
    echo "Using external database: $PARTNER_DB_HOST"

    # Validate required database variables
    for var in PARTNER_DB_USER PARTNER_DB_PASSWORD PARTNER_DB_NAME PARTNER_DB_PORT; do
        if [ -z "${!var}" ]; then
            echo -e "${RED}Error: $var is not set${NC}"
            exit 1
        fi
    done

    DB_URL="postgresql://${PARTNER_DB_USER}:${PARTNER_DB_PASSWORD}@${PARTNER_DB_HOST}:${PARTNER_DB_PORT}/${PARTNER_DB_NAME}"
    psql "$DB_URL" -c "$SQL"
fi
```

### Checklist

- [ ] Validate all required environment variables before use
- [ ] Provide clear error messages indicating which variable is missing
- [ ] Use `${VAR:-default}` for optional variables with defaults
- [ ] Exit early with non-zero code on validation failure

---

## 14. API Parameters: Missing Max Limits

### Severity: 🟠 Major (Resource Exhaustion / DoS Risk)

### Problem

API endpoints that accept numeric parameters (like `days`, `limit`, `count`) without maximum bounds can be abused to cause excessive resource consumption or denial of service.

### Bad Example

```go
// ❌ BAD: No max limit - attacker could request days=999999
func (a *app) queryStats(ctx context.Context, r *http.Request) web.Encoder {
    days := 30
    if daysStr := r.URL.Query().Get("days"); daysStr != "" {
        parsedDays, err := strconv.Atoi(daysStr)
        if err != nil || parsedDays < 1 {
            return errs.New(errs.InvalidArgument, fmt.Errorf("days must be a positive integer"))
        }
        days = parsedDays  // No upper bound!
    }
    // ...
}
```

### Good Example

```go
// ✅ GOOD: Define and enforce max limit with field-specific error
const maxStatsDays = 365

func (a *app) queryStats(ctx context.Context, r *http.Request) web.Encoder {
    days := 30
    if daysStr := r.URL.Query().Get("days"); daysStr != "" {
        parsedDays, err := strconv.Atoi(daysStr)
        if err != nil || parsedDays < 1 {
            return errs.NewFieldErrors("days", fmt.Errorf("must be a positive integer"))
        }
        if parsedDays > maxStatsDays {
            return errs.NewFieldErrors("days", fmt.Errorf("must be <= %d", maxStatsDays))
        }
        days = parsedDays
    }
    // ...
}
```

### Checklist

- [ ] All numeric API parameters have sensible max limits defined as constants
- [ ] Return field-specific validation errors (not generic InvalidArgument)
- [ ] Document limits in API documentation

---

## 15. Strong Types: When NOT to Use Them

### Severity: 🟢 Low (Over-Engineering)

### Problem

Creating strong types for simple values that have no validation rules adds unnecessary complexity. The pattern in CLAUDE.md is for types that have business validation (like intensity 0-10, content length limits). Simple counts, totals, or database-returned aggregates don't need strong types.

### Bad Example

```go
// ❌ BAD: Over-engineering - Count has no validation rules
package count

type Count struct {
    value int
}

func Parse(value int) (Count, error) {
    if value < 0 {
        return Count{}, errors.New("count cannot be negative")
    }
    return Count{value: value}, nil
}

// Stats using unnecessary strong type
type Stats struct {
    ThisWeek   count.Count  // Over-engineered
    ThisMonth  count.Count  // Over-engineered
    Last30Days count.Count  // Over-engineered
}
```

### Good Example

```go
// ✅ GOOD: Simple int is appropriate for counts
// Counts are database aggregates - they're naturally non-negative
// No business rules to enforce beyond what SQL provides
type Stats struct {
    ThisWeek   int  // Simple, appropriate
    ThisMonth  int  // Simple, appropriate
    Last30Days int  // Simple, appropriate
}
```

### When to Use Strong Types

Use strong types when:
- Value has validation rules (intensity 0-10, content min/max length)
- Value requires parsing/formatting (dates, UUIDs, enums)
- Type safety prevents mixing different concepts (UserID vs OrderID)

Don't use strong types when:
- Value is a simple count/total from database
- No validation beyond basic type (int, string)
- Adding type doesn't prevent any real bugs

---

## Quick Reference Checklist (Backend)

When implementing new features, verify:

- [ ] **Security**: Child entities validate parent ownership against authenticated user
- [ ] **UTF-8**: String length validation uses `utf8.RuneCountInString()`
- [ ] **Strong Types**: Include both `Value()` and `String()` methods
- [ ] **Strong Types**: Don't over-engineer - simple counts don't need custom types
- [ ] **SQL**: Use `strings.Join()` for WHERE clause construction
- [ ] **Errors**: Define domain-specific errors (e.g., `ErrNotValueOwner`)
- [ ] **App Layer**: Handle all business errors with appropriate HTTP status codes
- [ ] **Logging**: Business layer methods include structured entry/error/success logging
- [ ] **DRY**: Extract helpers for patterns repeated 3+ times
- [ ] **Thread Safety**: Use `math/rand/v2` (Go 1.22+) or sync.Mutex with custom rand sources
- [ ] **Timezones**: Use UTC for database storage, never `time.Local`
- [ ] **Idempotency**: Scheduled/repeated operations use unique constraints to prevent duplicates
- [ ] **Validation**: Business layer validates inputs before persisting (MessageType, Content, etc.)
- [ ] **View Models**: Read-only query models should have comments explaining primitive type usage
- [ ] **API Parameters**: Numeric parameters have sensible max limits (e.g., days <= 365)

---

# Frontend Errors to Avoid

This section catalogs critical errors specific to frontend (TypeScript/React) development.

---

## F1. Security: Markdown Injection in User Content

### Severity: 🔴 CRITICAL (Security Vulnerability)

### Problem

When generating Markdown from user-provided content, failing to escape special characters allows users to inject malicious Markdown that can break formatting, inject links, or cause rendering issues.

### Bad Example

```typescript
// ❌ BAD: User content inserted directly into Markdown
function formatMoment(moment: ExportItem): string[] {
  const lines: string[] = [];

  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(moment.situation); // User could inject: "# HACKED [click here](evil.com)"
    lines.push('');
  }

  return lines;
}
```

### Good Example

```typescript
// ✅ GOOD: Escape Markdown special characters in user content
function escapeMarkdown(text: string): string {
  return text.replace(/([\\*_\[\]()#+-.,!`>|{}])/g, '\\$1');
}

function formatMoment(moment: ExportItem): string[] {
  const lines: string[] = [];

  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(escapeMarkdown(moment.situation)); // Safe: special chars escaped
    lines.push('');
  }

  return lines;
}
```

### Checklist

- [ ] All user-provided content is escaped before insertion into Markdown
- [ ] Escape function covers all Markdown special characters
- [ ] Headings and programmatic strings remain unescaped

---

## F2. Async: Stale Responses in useEffect

### Severity: 🟠 Major (Race Condition Bug)

### Problem

When fetching data in `useEffect`, rapid state changes (e.g., user quickly changing date range) can cause stale responses to overwrite newer data. Without a freshness guard, the UI may show outdated results.

### Bad Example

```typescript
// ❌ BAD: No guard against stale responses
useEffect(() => {
  if (!open) return;

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const response = await api.getData(filter);
      setData(response); // Stale response may overwrite newer data!
    } catch (err) {
      setError(err);
    } finally {
      setIsLoading(false);
    }
  };

  fetchData();
}, [filter, open]);
```

### Good Example

```typescript
// ✅ GOOD: Request ID guards against stale responses
const requestIdRef = useRef(0);

useEffect(() => {
  if (!open) return;

  const currentRequestId = ++requestIdRef.current;

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const response = await api.getData(filter);

      // Only update if this is still the latest request
      if (currentRequestId === requestIdRef.current) {
        setData(response);
      }
    } catch (err) {
      if (currentRequestId === requestIdRef.current) {
        setError(err);
      }
    } finally {
      if (currentRequestId === requestIdRef.current) {
        setIsLoading(false);
      }
    }
  };

  fetchData();
}, [filter, open]);
```

### Alternative: AbortController

```typescript
useEffect(() => {
  const controller = new AbortController();

  const fetchData = async () => {
    try {
      const response = await api.getData(filter, { signal: controller.signal });
      setData(response);
    } catch (err) {
      if (!controller.signal.aborted) {
        setError(err);
      }
    }
  };

  fetchData();

  return () => controller.abort(); // Cleanup: cancel pending request
}, [filter]);
```

### Checklist

- [ ] All async useEffect hooks have a freshness guard (requestId or AbortController)
- [ ] State updates are conditional on the request still being current
- [ ] Loading and error states also check freshness before updating

---

## F3. Data: Silent Data Truncation

### Severity: 🟠 Major (Data Loss / UX Issue)

### Problem

When exporting paginated data, if only the first page is fetched but the UI suggests a full export, users unknowingly receive partial data. This is especially problematic for date-range exports where users expect all entries.

### Bad Example

```typescript
// ❌ BAD: Only fetches first page, no truncation warning
const handleExport = async () => {
  const response = await api.export.getItems({
    startDate,
    endDate,
    rows: 100, // Backend limit
  });

  // User gets only first 100 items even if total is 500!
  generateMarkdown(response.items);
};
```

### Good Example

```typescript
// ✅ GOOD: Block export when results are truncated
const momentsCount = exportData?.items.filter((i) => i.itemType === 'moment').length || 0;
const thinksCount = exportData?.items.filter((i) => i.itemType === 'think').length || 0;
const hasMoreItems = !!exportData && exportData.items.length < exportData.total;

// In preview component: warn user about truncation
{hasMoreItems && (
  <p className="text-sm text-amber-600">
    Showing {exportData.items.length} of {exportData.total} entries.
    Consider using a smaller date range.
  </p>
)}

// Disable export button when truncated
<Button
  onClick={handleExport}
  disabled={isLoading || isExporting || hasMoreItems}
>
  Export
</Button>
```

### Better Solution: Full Pagination

```typescript
// ✅ BEST: Fetch all pages before export
const fetchAllItems = async (params: ExportParams): Promise<ExportItem[]> => {
  const allItems: ExportItem[] = [];
  let page = 1;
  let hasMore = true;

  while (hasMore) {
    const response = await api.export.getItems({ ...params, page });
    allItems.push(...response.items);
    hasMore = allItems.length < response.total;
    page++;
  }

  return allItems;
};
```

### Checklist

- [ ] Compare `items.length` vs `total` to detect truncation
- [ ] Either block export or show prominent warning when truncated
- [ ] Consider implementing full pagination for complete exports

---

## F4. State: useSyncExternalStore + useState Duplication

### Severity: 🟠 Major (State Sync Bug)

### Problem

When using `useSyncExternalStore` to read external state (like localStorage), creating a separate `useState` that copies the initial value creates a state duplication bug. The local state never updates when the external store changes.

### Bad Example

```typescript
// ❌ BAD: useState duplicates external store state and never syncs
function getSidebarCollapsed(): boolean {
  const saved = localStorage.getItem('sidebar-collapsed');
  return saved !== null ? JSON.parse(saved) : false;
}

function subscribeSidebarState(callback: () => void) {
  window.addEventListener('storage', callback);
  return () => window.removeEventListener('storage', callback);
}

export function AppSidebar() {
  // BUG: initialCollapsed is read once, then copied to useState
  const initialCollapsed = useSyncExternalStore(subscribeSidebarState, getSidebarCollapsed, () => false);
  const [collapsed, setCollapsed] = useState(initialCollapsed); // Never updates!

  const toggleCollapsed = () => {
    setCollapsed(!collapsed); // Only updates local state
    localStorage.setItem('sidebar-collapsed', JSON.stringify(!collapsed));
  };
}
```

### Good Example

```typescript
// ✅ GOOD: Use useSyncExternalStore directly, no useState duplication
const STORAGE_KEY = 'sidebar-collapsed';

function getSidebarCollapsed(): boolean {
  if (typeof window === 'undefined') return false;
  const saved = localStorage.getItem(STORAGE_KEY);
  return saved !== null ? JSON.parse(saved) : false;
}

function subscribeSidebarState(callback: () => void) {
  const handler = () => callback();
  window.addEventListener('storage', handler);
  window.addEventListener('sidebar-state-change', handler); // Custom event for same-tab
  return () => {
    window.removeEventListener('storage', handler);
    window.removeEventListener('sidebar-state-change', handler);
  };
}

export function AppSidebar() {
  // Single source of truth - directly from external store
  const collapsed = useSyncExternalStore(subscribeSidebarState, getSidebarCollapsed, () => false);

  const toggleCollapsed = useCallback(() => {
    const newState = !getSidebarCollapsed();
    localStorage.setItem(STORAGE_KEY, JSON.stringify(newState));
    // Dispatch custom event for same-tab updates (storage event only fires cross-tab)
    window.dispatchEvent(new Event('sidebar-state-change'));
  }, []);
}
```

### Checklist

- [ ] Never copy `useSyncExternalStore` result into `useState`
- [ ] Use custom events to trigger same-tab updates (storage events are cross-tab only)
- [ ] Read fresh value from external store when updating (not stale closure)

---

## F5. Accessibility: Clickable Div Without Keyboard Support

### Severity: 🟠 Major (Accessibility Violation)

### Problem

Using a `<div>` with `onClick` for interactive elements without proper ARIA attributes and keyboard handlers makes the component inaccessible to keyboard users and screen readers.

### Bad Example

```typescript
// ❌ BAD: Clickable div without accessibility
return (
  <div className="cursor-pointer" onClick={() => setExpanded(!expanded)}>
    <h3>{title}</h3>
    <ChevronDown className={expanded && 'rotate-180'} />
  </div>
);
```

### Good Example

```typescript
// ✅ GOOD: Proper accessibility for clickable element
const handleKeyDown = (e: React.KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault(); // Prevent scroll on Space
    setExpanded(!expanded);
  }
};

return (
  <div
    role="button"
    tabIndex={0}
    aria-expanded={expanded}
    className="cursor-pointer"
    onClick={() => setExpanded(!expanded)}
    onKeyDown={handleKeyDown}
  >
    <h3>{title}</h3>
    <ChevronDown className={expanded && 'rotate-180'} />
  </div>
);
```

### Alternative: Use a Button

```typescript
// ✅ BETTER: Use semantic HTML when possible
return (
  <button
    type="button"
    aria-expanded={expanded}
    className="w-full text-left cursor-pointer"
    onClick={() => setExpanded(!expanded)}
  >
    <h3>{title}</h3>
    <ChevronDown className={expanded && 'rotate-180'} />
  </button>
);
```

### Checklist

- [ ] Clickable divs have `role="button"` and `tabIndex={0}`
- [ ] Include `onKeyDown` handler for Enter and Space keys
- [ ] Use `aria-expanded` for expandable elements
- [ ] Call `e.preventDefault()` for Space to prevent page scroll
- [ ] Prefer semantic `<button>` when styling allows

---

## Quick Reference Checklist (Frontend)

When implementing new features, verify:

- [ ] **Markdown**: User content is escaped before insertion into Markdown/HTML
- [ ] **Async**: useEffect with fetch uses requestId or AbortController for freshness
- [ ] **Pagination**: Exports/downloads handle truncation (warn or fetch all pages)
- [ ] **Types**: Use `import type` for type-only imports (not `React.ReactNode`)
- [ ] **State Sync**: Local state derived from props syncs via useEffect when props change
- [ ] **Cleanup**: File downloads use try-catch-finally for resource cleanup
- [ ] **External Store**: Never copy `useSyncExternalStore` result into `useState`
- [ ] **Accessibility**: Clickable divs have `role="button"`, `tabIndex={0}`, and keyboard handlers
- [ ] **Validation**: API params are clamped/validated before sending (e.g., rows limit)
