# Backend Errors to Avoid (Go)

This document catalogs critical errors specific to backend (Go) development.

---

## Table of Contents

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
16. [SQL Performance: Unbounded Aggregate Queries](#16-sql-performance-unbounded-aggregate-queries)
17. [Error Handling: Fragile String-Based Error Detection](#17-error-handling-fragile-string-based-error-detection)
18. [Business Logic: Missing Parent Entity Validation on Restore](#18-business-logic-missing-parent-entity-validation-on-restore)
19. [Error Handling: Conflating Query Errors with NotFound](#19-error-handling-conflating-query-errors-with-notfound)
20. [State Transitions: Allowing Same-State Transitions](#20-state-transitions-allowing-same-state-transitions)
21. [Delegate Handlers: Returning Errors Instead of Logging](#21-delegate-handlers-returning-errors-instead-of-logging)
22. [Update Structs: Double Pointers for Nullable Optional Fields](#22-update-structs-double-pointers-for-nullable-optional-fields)
23. [Bulk Operations: Missing Context in Decryption Errors](#23-bulk-operations-missing-context-in-decryption-errors)
24. [API Parameters: Missing Date Range Validation](#24-api-parameters-missing-date-range-validation)
25. [Code Duplication: Repeated Ownership Validation](#25-code-duplication-repeated-ownership-validation)
26. [SQL Syntax: Non-PostgreSQL Pagination](#26-sql-syntax-non-postgresql-pagination)
27. [Filter Logic: Conflicting WHERE Conditions](#27-filter-logic-conflicting-where-conditions)
28. [Boolean Parsing: Silent Failures](#28-boolean-parsing-silent-failures)
29. [Time Formatting: Hardcoded Format Strings](#29-time-formatting-hardcoded-format-strings)
30. [Transaction Safety: Non-Atomic Multi-Operation Business Logic](#30-transaction-safety-non-atomic-multi-operation-business-logic)
31. [Defense-in-Depth: Dropping Constraints Without Trigger Replacement](#31-defense-in-depth-dropping-constraints-without-trigger-replacement)

---

## 1. Security: Child Entity Ownership Validation

### Severity: CRITICAL (Security Vulnerability)

### Problem

When creating a child entity that references a parent entity by ID, failing to validate that the authenticated user owns the parent entity allows cross-user data linking attacks.

### Bad Example

```go
// BAD: No ownership validation
func (b *Business) Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error) {
    value, err := b.valueBus.QueryByID(ctx, nlv.ValueID)
    if err != nil {
        return LifeVision{}, err
    }
    // Attacker can link their life vision to ANY user's value
    lifeVision := LifeVision{
        UserID:  value.UserID,
        ValueID: nlv.ValueID,
    }
    return b.storer.Create(ctx, lifeVision)
}
```

### Good Example

```go
// GOOD: Validates authenticated user owns the parent entity
func (b *Business) Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error) {
    value, err := b.valueBus.QueryByID(ctx, nlv.ValueID)
    if err != nil {
        return LifeVision{}, err
    }
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

### Checklist

- [ ] Child entity creation includes authenticated userID in the request struct
- [ ] Business layer validates `parent.UserID == authenticatedUserID`
- [ ] Return `PermissionDenied` error for ownership violations

---

## 2. Error Handling: Sentinel vs Generic Errors

### Severity: Major (Error Handling Bug)

### Problem

Returning a generic `fmt.Errorf()` message instead of a sentinel error prevents the app layer from properly detecting and handling the error with the correct HTTP status code.

### Bad Example

```go
// BAD: Generic error - app layer will return 500
if parent.UserID != entity.UserID {
    return Entity{}, fmt.Errorf("cannot move entity to parent owned by different user")
}
```

### Good Example

```go
// GOOD: Sentinel error - app layer can return 403
var ErrNotParentOwner = errors.New("user does not own the specified parent")

if parent.UserID != entity.UserID {
    return Entity{}, ErrNotParentOwner
}
```

### Checklist

- [ ] All permission violations return sentinel errors (defined in model.go)
- [ ] Sentinel errors are defined with `errors.New()` for `errors.Is()` compatibility
- [ ] App layer handles ALL sentinel errors with appropriate HTTP status codes

---

## 3. String Length: UTF-8 Rune Count vs Byte Count

### Severity: Major (Data Validation Bug)

### Problem

Using `len(string)` for character count validation fails for multi-byte UTF-8 characters.

### Bad Example

```go
// BAD: len() counts bytes, not characters
if len(value) < 10 {  // "日本語です" is 5 chars but 15 bytes
    return Content{}, ErrTooShort
}
```

### Good Example

```go
import "unicode/utf8"

// GOOD: utf8.RuneCountInString() counts actual characters
runeCount := utf8.RuneCountInString(value)
if runeCount < 10 {
    return Content{}, ErrTooShort
}
```

### Checklist

- [ ] String length validation uses `utf8.RuneCountInString()`
- [ ] Never use `len()` for user-facing character limits

---

## 4. Strong Types: Missing Value() Method

### Severity: Minor (API Inconsistency)

### Problem

Strong types should expose their underlying value through a `Value()` method for consistency.

### Good Example

```go
type MyType struct {
    value string
}

func (t MyType) Value() string { return t.value }
func (t MyType) String() string { return t.value }
func (t MyType) Equal(t2 MyType) bool { return t.value == t2.value }
func (t MyType) MarshalText() ([]byte, error) { return []byte(t.value), nil }
```

### Checklist

- [ ] Strong types have both `Value()` and `String()` methods
- [ ] Include `Equal()` for testing and `MarshalText()` for logging

---

## 5. SQL: WHERE Clause Building

### Severity: Low (Code Quality)

### Problem

Building WHERE clauses with manual index checking is verbose and error-prone.

### Good Example

```go
import "strings"

// GOOD: Use strings.Join for cleaner code
func buildWhereClause(filter QueryFilter, data map[string]any) string {
    var conditions []string
    if filter.ID != nil {
        data["id"] = *filter.ID
        conditions = append(conditions, "id = :id")
    }
    if len(conditions) == 0 {
        return ""
    }
    return " WHERE " + strings.Join(conditions, " AND ")
}
```

---

## 6. Logging: Missing Structured Logging in Business Layer

### Severity: Major (Observability Gap)

### Problem

Business layer methods without structured logging make debugging production issues difficult.

### Good Example

```go
func (b *Business) Query(ctx context.Context, filter QueryFilter) ([]Item, error) {
    b.log.Info(ctx, "domain.query", "userID", filter.UserID)

    items, err := b.storer.Query(ctx, filter)
    if err != nil {
        b.log.Error(ctx, "domain.query", "err", err)
        return nil, fmt.Errorf("storer.query: %w", err)
    }

    b.log.Info(ctx, "domain.query.success", "count", len(items))
    return items, nil
}
```

### Checklist

- [ ] Business struct includes `*logger.Logger` field
- [ ] All public methods log on entry with parameters
- [ ] All errors are logged before returning

---

## 7. Code Duplication: Repeated Decrypt-Parse Patterns

### Severity: Minor (Maintainability)

### Problem

Copying the same decrypt→parse→assign pattern for each field creates maintenance burden.

### Good Example

```go
// GOOD: Extract helper function
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
```

### Checklist

- [ ] Extract helpers for patterns repeated 3+ times

---

## 8. Thread Safety: Package-Level Random Sources

### Severity: Major (Race Condition)

### Problem

Creating a package-level `rand.Rand` instance is not safe for concurrent use.

### Good Example (Go 1.22+)

```go
import "math/rand/v2"

// GOOD: Use thread-safe rand/v2 package
func GetRandomTemplate() string {
    templates := []string{"a", "b", "c"}
    return templates[rand.IntN(len(templates))]
}
```

### Checklist

- [ ] Use `math/rand/v2` on Go 1.22+ for thread-safe global generator
- [ ] Never use `rand.New(rand.NewSource(...))` at package level without mutex

---

## 9. Timezone: Using time.Local in Database Conversions

### Severity: Major (Non-Deterministic Behavior)

### Problem

Using `time.Local` ties behavior to the host's timezone, causing inconsistent results.

### Good Example

```go
// GOOD: Always use UTC for storage layer
func toBusMessage(db message) Message {
    return Message{
        ScheduledAt: db.ScheduledAt.UTC(),
        DateCreated: db.DateCreated.UTC(),
    }
}
```

### Checklist

- [ ] Store all timestamps in UTC in the database
- [ ] Never use `time.Local` in business or store layers

---

## 10. Idempotency: Duplicate Scheduled Messages

### Severity: Major (Data Integrity)

### Problem

Scheduled operations can create duplicate records if the job runs multiple times.

### Good Example

```sql
-- Prevent duplicate messages per user/type/date
CREATE UNIQUE INDEX IF NOT EXISTS notification_messages_schedule_unique_idx
  ON notification_messages(user_id, message_type, DATE(scheduled_at))
  WHERE status = 'pending';
```

```go
// Handle constraint violation gracefully
if errors.Is(err, notificationbus.ErrDuplicateSchedule) {
    return nil  // Already scheduled - not an error
}
```

### Checklist

- [ ] Add unique constraint on scheduling key (user_id, type, date)
- [ ] Worker treats duplicate as success, not error

---

## 11. Input Validation: Missing Business Layer Validation

### Severity: Major (Data Integrity)

### Problem

Persisting data without validation in the business layer allows invalid data to reach the database.

### Good Example

```go
func (b *Business) Create(ctx context.Context, nm NewMessage) (Message, error) {
    if !nm.MessageType.Valid() {
        return Message{}, ErrInvalidMessageType
    }
    if strings.TrimSpace(nm.Content) == "" {
        return Message{}, ErrContentEmpty
    }
    // ... create message
}
```

### Checklist

- [ ] Enum types have `Valid()` method
- [ ] Business layer validates all input before persisting

---

## 12. SQL Views: ORDER BY in View Definition

### Severity: Low (Performance/Maintainability)

### Problem

`ORDER BY` in a SQL view definition is often ignored by the database optimizer.

### Good Example

```sql
-- GOOD: View defines projection only, no ORDER BY
CREATE OR REPLACE VIEW view_notification_content AS
SELECT user_id, value_id, content, display_order
FROM values;
-- Consumer applies ORDER BY when querying
```

---

## 13. Shell Scripts: Unvalidated Environment Variables

### Severity: Major (Runtime Failure)

### Problem

Using environment variables without validation leads to cryptic errors.

### Good Example

```bash
# GOOD: Validate all required variables
for var in PARTNER_DB_USER PARTNER_DB_PASSWORD PARTNER_DB_NAME; do
    if [ -z "${!var}" ]; then
        echo "Error: $var is not set"
        exit 1
    fi
done
```

---

## 14. API Parameters: Missing Max Limits

### Severity: Major (Resource Exhaustion / DoS Risk)

### Problem

API endpoints that accept numeric parameters without maximum bounds can be abused.

### Good Example

```go
const maxStatsDays = 365

if parsedDays > maxStatsDays {
    return errs.NewFieldErrors("days", fmt.Errorf("must be <= %d", maxStatsDays))
}
```

### Checklist

- [ ] All numeric API parameters have sensible max limits

---

## 15. Strong Types: When NOT to Use Them

### Severity: Low (Over-Engineering)

### Problem

Creating strong types for simple values that have no validation rules adds unnecessary complexity.

### Good Example

```go
// GOOD: Simple int is appropriate for counts
type Stats struct {
    ThisWeek   int  // Simple, appropriate
    ThisMonth  int  // Simple, appropriate
    Last30Days int  // Simple, appropriate
}
```

### Checklist

- [ ] Don't use strong types for simple counts/totals from database
- [ ] Only use strong types when there are validation rules

---

## 16. SQL Performance: Unbounded Aggregate Queries

### Severity: Major (Scalability / Performance)

### Problem

Aggregate queries without time bounds require scanning all rows.

### Good Example

```sql
-- GOOD: LEAST() bounds scan to earliest time window needed
SELECT COUNT(*) FILTER (WHERE moment_date >= NOW() - INTERVAL '7 days') as this_week
FROM moments
WHERE user_id = :user_id
  AND moment_date >= LEAST(
      DATE_TRUNC('month', NOW()),
      NOW() - INTERVAL '7 days',
      NOW() - INTERVAL '1 day' * :days
  )
```

### Checklist

- [ ] Aggregate queries have explicit time bounds in WHERE clause
- [ ] Use LEAST() to find the earliest boundary across multiple time windows

---

## Quick Reference Checklist

- [ ] **Security**: Child entities validate parent ownership
- [ ] **UTF-8**: String length uses `utf8.RuneCountInString()`
- [ ] **Strong Types**: Include both `Value()` and `String()` methods
- [ ] **Strong Types**: Don't over-engineer simple counts
- [ ] **SQL**: Use `strings.Join()` for WHERE clause construction
- [ ] **Errors**: Define domain-specific sentinel errors
- [ ] **Logging**: Business layer includes structured logging
- [ ] **DRY**: Extract helpers for patterns repeated 3+ times
- [ ] **Thread Safety**: Use `math/rand/v2` (Go 1.22+)
- [ ] **Timezones**: Use UTC, never `time.Local`
- [ ] **Idempotency**: Use unique constraints for scheduled operations
- [ ] **Validation**: Business layer validates inputs before persisting
- [ ] **API Parameters**: Numeric parameters have max limits
- [ ] **SQL Performance**: Aggregate queries have time bounds

---

## 17. Error Handling: Fragile String-Based Error Detection

### Severity: Major (Integration Bug, Fragile Code)

### Problem

Detecting database errors by searching for substrings in error messages is fragile and can break if the message format changes.

### Bad Example

```go
// BAD: Fragile string-based error detection
if err := b.storer.Update(ctx, value); err != nil {
    if strings.Contains(err.Error(), "active life visions") {
        return Value{}, ErrHasActiveLifeVisions
    }
    return Value{}, fmt.Errorf("update: %w", err)
}
```

### Good Example

```go
// GOOD: Store layer detects specific PG error and returns sentinel error
// In store layer (valuedb.go):
if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbValue); err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23503" &&
       strings.Contains(pgErr.Message, "active life visions") {
        return valuebus.ErrHasActiveLifeVisions
    }
    return fmt.Errorf("namedexeccontext: %w", err)
}

// In business layer (valuebus.go):
if err := b.storer.Update(ctx, value); err != nil {
    if errors.Is(err, ErrHasActiveLifeVisions) {
        return Value{}, ErrHasActiveLifeVisions
    }
    return Value{}, fmt.Errorf("update: %w", err)
}
```

### Checklist

- [ ] Use `errors.As` to cast to specific error types (e.g., `*pgconn.PgError`)
- [ ] Check error codes, not just message text
- [ ] Define sentinel errors at the domain layer and return them from the store layer
- [ ] Use `errors.Is` in business layer to check for sentinel errors

---

## 18. Business Logic: Missing Parent Entity Validation on Restore

### Severity: Major (Data Integrity Issue)

### Problem

When restoring an archived child entity, failing to validate that its parent entity is still active can lead to orphaned or inconsistent data states.

### Bad Example

```go
// BAD: Restoring life vision without checking if parent value is active
func (b *Business) Restore(ctx context.Context, lifeVision LifeVision) (LifeVision, error) {
    if !lifeVision.Status.IsArchived() {
        return LifeVision{}, ErrNotArchived
    }

    // Missing: Check if parent value is still active!

    lifeVision.Status = entitystatus.Active
    lifeVision.ArchivedAt = nil
    lifeVision.DateUpdated = time.Now().UTC()

    if err := b.storer.Update(ctx, lifeVision); err != nil {
        return LifeVision{}, fmt.Errorf("update: %w", err)
    }

    return lifeVision, nil
}
```

### Good Example

```go
// GOOD: Validate parent value is active before restoring
func (b *Business) Restore(ctx context.Context, lifeVision LifeVision) (LifeVision, error) {
    if !lifeVision.Status.IsArchived() {
        return LifeVision{}, ErrNotArchived
    }

    // Validate parent value is still active before restoring
    value, err := b.valueBus.QueryByID(ctx, lifeVision.ValueID)
    if err != nil {
        return LifeVision{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", lifeVision.ValueID, err)
    }

    if !value.Status.IsActive() {
        return LifeVision{}, ErrTargetValueNotActive
    }

    lifeVision.Status = entitystatus.Active
    lifeVision.ArchivedAt = nil
    lifeVision.DateUpdated = time.Now().UTC()

    if err := b.storer.Update(ctx, lifeVision); err != nil {
        return LifeVision{}, fmt.Errorf("update: %w", err)
    }

    return lifeVision, nil
}
```

### Checklist

- [ ] Before restoring a child entity, verify the parent entity exists and is in valid state
- [ ] Define specific error types for invalid parent states (e.g., `ErrTargetValueNotActive`)
- [ ] Handle the new error in the app layer and return appropriate HTTP status codes

---

## 19. Error Handling: Conflating Query Errors with NotFound

### Severity: Major (Error Handling Bug)

### Problem

Combining Query error check with empty result check in a single condition discards the original error information, making debugging difficult and returning incorrect HTTP status codes.

### Bad Example

```go
// BAD: Query error discarded, returns 404 even for database errors
objetivos, err := a.objetivoBus.Query(ctx, filter, orderBy, page)
if err != nil || len(objetivos) == 0 {
    return errs.New(errs.NotFound, objetivobus.ErrNotFound)
}
```

### Good Example

```go
// GOOD: Handle error first, then check empty result
objetivos, err := a.objetivoBus.Query(ctx, filter, orderBy, page)
if err != nil {
    return errs.Newf(errs.Internal, "query: objetivoID[%s]: %s", objetivoID, err)
}
if len(objetivos) == 0 {
    return errs.New(errs.NotFound, objetivobus.ErrNotFound)
}
```

### Checklist

- [ ] Always check and handle errors before checking for empty results
- [ ] Preserve error context when wrapping errors
- [ ] Return appropriate HTTP status codes (500 for errors, 404 for not found)

---

## 20. State Transitions: Allowing Same-State Transitions

### Severity: Medium (Logic Bug)

### Problem

State machine implementations that allow transitions from a state to itself create confusing behavior and unnecessary database writes.

### Bad Example

```go
// BAD: activo→activo is allowed (wasteful/confusing)
func validateStatusTransition(from, to Status) error {
    if from.IsTerminal() {
        return ErrStatusTransitionNotAllowed
    }
    if from.IsActivo() {
        // Can go anywhere - but this includes staying activo!
        return nil
    }
    return ErrStatusTransitionNotAllowed
}
```

### Good Example

```go
// GOOD: Reject same-state transitions first
func validateStatusTransition(from, to Status) error {
    // Cannot transition to the same status (no-op)
    if from.Equal(to) {
        return ErrStatusTransitionNotAllowed
    }

    // Cannot transition from terminal statuses
    if from.IsTerminal() {
        return ErrStatusTransitionNotAllowed
    }

    if from.IsActivo() {
        // Can go anywhere except stay activo (already checked above)
        return nil
    }
    return ErrStatusTransitionNotAllowed
}
```

### Checklist

- [ ] State machines reject same-state transitions early
- [ ] Comments accurately describe allowed transitions
- [ ] Consider whether no-op transitions should be silent success vs error

---

## 21. Delegate Handlers: Returning Errors Instead of Logging

### Severity: Medium (Operational Issue)

### Problem

When calling delegate handlers after completing a primary operation (like delete), returning the delegate error aborts the operation even though the primary action succeeded. Delegate handlers should be idempotent and non-blocking.

### Bad Example

```go
// BAD: Delegate failure aborts the delete
func (b *Business) Delete(ctx context.Context, entity Entity) error {
    if err := b.storer.Delete(ctx, entity); err != nil {
        return fmt.Errorf("delete: %w", err)
    }

    if b.delegate != nil {
        if err := b.delegate.Call(ctx, ActionDeletedData(entity)); err != nil {
            return fmt.Errorf("delegate call: %w", err) // Blocks delete success!
        }
    }
    return nil
}
```

### Good Example

```go
// GOOD: Log delegate failures, don't abort the primary operation
func (b *Business) Delete(ctx context.Context, entity Entity) error {
    if err := b.storer.Delete(ctx, entity); err != nil {
        return fmt.Errorf("delete: %w", err)
    }

    // Delegate handlers are idempotent, so failures are logged but do not abort the delete flow.
    if b.delegate != nil {
        if err := b.delegate.Call(ctx, ActionDeletedData(entity)); err != nil {
            b.log.Error(ctx, "entity.delete.delegate", "entityID", entity.ID, "err", err)
        }
    }
    return nil
}
```

### Checklist

- [ ] Delegate calls after successful primary operations log errors instead of returning them
- [ ] Add comment explaining delegate handlers are idempotent
- [ ] Include entity ID and error in log message

---

## 22. Update Structs: Double Pointers for Nullable Optional Fields

### Severity: Medium (Code Quality / Maintainability)

### Problem

Using double pointers (`**Type`) to distinguish between "no update" and "set to null" is inconsistent with other Update structs and confusing for developers. Use a separate `ClearField` bool instead.

### Bad Example

```go
// BAD: Double pointer is confusing and inconsistent
type UpdateEntity struct {
    Status  *Status
    Notes   **notes.Notes // Double pointer to distinguish "not updating" from "setting to null"
}
```

### Good Example

```go
// GOOD: Use ClearField bool pattern for clarity
type UpdateEntity struct {
    Status     *Status
    Notes      *notes.Notes
    ClearNotes bool // When true, sets Notes to NULL (takes precedence over Notes field)
}

// Usage in business layer:
if ur.ClearNotes {
    entity.Notes = nil
} else if ur.Notes != nil {
    entity.Notes = ur.Notes
}
```

### Checklist

- [ ] Use single pointers for optional fields in Update structs
- [ ] Add `ClearField` bool for fields that can be explicitly set to null
- [ ] Document precedence: ClearField takes precedence over the value field

---

## 23. Bulk Operations: Missing Context in Decryption Errors

### Severity: Medium (Debugging Difficulty)

### Problem

When processing multiple records in a loop, errors that lack context about which record failed make debugging difficult.

### Bad Example

```go
// BAD: No context about which record failed
func toBusRecordsDecrypted(dbs []record) ([]Record, error) {
    records := make([]Record, len(dbs))
    for i, db := range dbs {
        var err error
        records[i], err = toBusRecordDecrypted(db)
        if err != nil {
            return nil, err // Which record failed?
        }
    }
    return records, nil
}
```

### Good Example

```go
// GOOD: Include index and record ID in error
func toBusRecordsDecrypted(dbs []record) ([]Record, error) {
    records := make([]Record, len(dbs))
    for i, db := range dbs {
        var err error
        records[i], err = toBusRecordDecrypted(db)
        if err != nil {
            return nil, fmt.Errorf("failed to decrypt record at index %d (recordID[%s]): %w", i, db.ID, err)
        }
    }
    return records, nil
}
```

### Checklist

- [ ] Include loop index in error messages for bulk operations
- [ ] Include record ID when available
- [ ] Wrap original error with fmt.Errorf and %w

---

## 24. API Parameters: Missing Date Range Validation

### Severity: Medium (Poor UX / Silent Failure)

### Problem

When API endpoints accept date range filters (startDate, endDate), failing to validate that startDate <= endDate can return empty results without clear feedback to the user.

### Bad Example

```go
// BAD: Parses dates independently without range validation
if qp.StartDate != "" {
    startDate, err := time.Parse("2006-01-02", qp.StartDate)
    if err != nil {
        return errs.NewFieldErrors("startDate", err)
    }
    filter.StartDate = &startDate
}

if qp.EndDate != "" {
    endDate, err := time.Parse("2006-01-02", qp.EndDate)
    if err != nil {
        return errs.NewFieldErrors("endDate", err)
    }
    filter.EndDate = &endDate
}
// Missing: validation that startDate <= endDate
```

### Good Example

```go
// GOOD: Validate date range after parsing both dates
var startDate, endDate time.Time
if qp.StartDate != "" {
    startDate, err = time.Parse("2006-01-02", qp.StartDate)
    if err != nil {
        return errs.NewFieldErrors("startDate", err)
    }
    filter.StartDate = &startDate
}

if qp.EndDate != "" {
    endDate, err = time.Parse("2006-01-02", qp.EndDate)
    if err != nil {
        return errs.NewFieldErrors("endDate", err)
    }
    filter.EndDate = &endDate
}

// Validate date range: startDate must be <= endDate
if filter.StartDate != nil && filter.EndDate != nil && startDate.After(endDate) {
    return errs.NewFieldErrors("startDate", errors.New("startDate must be before or equal to endDate"))
}
```

### Checklist

- [ ] All date range parameters validate startDate <= endDate
- [ ] Return clear error message identifying the invalid parameter

---

## 25. Code Duplication: Repeated Ownership Validation

### Severity: Medium (Maintainability)

### Problem

When multiple handlers need to fetch an entity and validate ownership, duplicating this logic creates maintenance burden and inconsistency risk.

### Bad Example

```go
// BAD: Ownership validation duplicated across handlers
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
    // ... parse inputs ...

    objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
    if err != nil {
        if errors.Is(err, objectivebus.ErrNotFound) {
            return errs.New(errs.NotFound, errors.New("objective not found"))
        }
        return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
    }
    if objective.UserID != userID {
        return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
    }
    // ... continue ...
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
    // Same ownership validation duplicated here...
}
```

### Good Example

```go
// GOOD: Extract helper method for ownership validation
func (a *app) getObjectiveWithOwnership(ctx context.Context, objectiveID, userID uuid.UUID) (objectivebus.Objective, error) {
    objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
    if err != nil {
        if errors.Is(err, objectivebus.ErrNotFound) {
            return objectivebus.Objective{}, errs.New(errs.NotFound, errors.New("objective not found"))
        }
        return objectivebus.Objective{}, errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
    }
    if objective.UserID != userID {
        return objectivebus.Objective{}, errs.New(errs.PermissionDenied, errors.New("user not authorized"))
    }
    return objective, nil
}

// Usage in handlers:
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
    if _, err := a.getObjectiveWithOwnership(ctx, objectiveID, userID); err != nil {
        return err.(web.Encoder)
    }
    // ... continue ...
}
```

### Checklist

- [ ] Extract helper methods for patterns repeated across handlers
- [ ] Helper returns appropriate error types for HTTP status mapping
- [ ] Use type assertion to return errors as web.Encoder

---

## 26. SQL Syntax: Non-PostgreSQL Pagination

### Severity: CRITICAL (Query Failure in Production)

### Problem

Using SQL Server/Oracle pagination syntax (`OFFSET ... ROWS FETCH NEXT ... ROWS ONLY`) in a PostgreSQL database causes query failures.

### Bad Example

```go
// BAD: SQL Server/Oracle syntax - will fail on PostgreSQL
q := fmt.Sprintf(`
SELECT id, name FROM users
%s
%s
OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)
```

### Good Example

```go
// GOOD: PostgreSQL-compatible syntax
q := fmt.Sprintf(`
SELECT id, name FROM users
%s
%s
LIMIT :rows_per_page OFFSET :offset`, whereClause, orderByClause)
```

### Checklist

- [ ] All pagination queries use `LIMIT :rows_per_page OFFSET :offset`
- [ ] Never use `OFFSET ... ROWS FETCH NEXT ... ROWS ONLY`

---

## 27. Filter Logic: Conflicting WHERE Conditions

### Severity: MAJOR (Logic Bug)

### Problem

When filter structs have mutually exclusive options (e.g., `ObjectiveID` and `InboxOnly`), building WHERE conditions independently can produce conflicting clauses like `objective_id = :objective_id AND objective_id IS NULL`.

### Bad Example

```go
// BAD: Can produce conflicting conditions
func buildWhereClause(filter QueryFilter, data map[string]any) string {
    var conditions []string

    if filter.ObjectiveID != nil {
        data["objective_id"] = *filter.ObjectiveID
        conditions = append(conditions, "objective_id = :objective_id")
    }

    if filter.InboxOnly {
        conditions = append(conditions, "objective_id IS NULL")
    }
    // If both are set, WHERE clause becomes:
    // objective_id = :objective_id AND objective_id IS NULL (impossible!)
}
```

### Good Example

```go
// GOOD: InboxOnly takes precedence to prevent conflicting conditions
func buildWhereClause(filter QueryFilter, data map[string]any) string {
    var conditions []string

    // InboxOnly takes precedence over ObjectiveID to prevent conflicting conditions.
    if filter.InboxOnly {
        conditions = append(conditions, "objective_id IS NULL")
    } else if filter.ObjectiveID != nil {
        data["objective_id"] = *filter.ObjectiveID
        conditions = append(conditions, "objective_id = :objective_id")
    }
}
```

### Checklist

- [ ] Identify mutually exclusive filter options
- [ ] Add comments explaining precedence rules
- [ ] Consider validating at API layer and returning error for conflicting filters

---

## 28. Boolean Parsing: Silent Failures

### Severity: MAJOR (User Input Validation)

### Problem

Parsing boolean query parameters and silently returning `false` for invalid values masks user errors and makes debugging difficult.

### Bad Example

```go
// BAD: Silently returns false for invalid values like "yes" or "nope"
func parseBool(s string) bool {
    if s == "" {
        return false
    }
    v, err := strconv.ParseBool(s)
    if err != nil {
        return false  // User typed "yes" but got false silently
    }
    return v
}
```

### Good Example

```go
// GOOD: Returns error for invalid non-empty values
func parseBool(s string) (bool, error) {
    if s == "" {
        return false, nil  // Empty is acceptable, defaults to false
    }
    v, err := strconv.ParseBool(s)
    if err != nil {
        return false, fmt.Errorf("invalid boolean value %q", s)
    }
    return v, nil
}

// Caller handles error:
inboxOnly, err := parseBool(qp.InboxOnly)
if err != nil {
    return errs.NewFieldErrors("inboxOnly", err)
}
```

### Checklist

- [ ] Boolean parsing returns `(bool, error)` for non-empty invalid values
- [ ] Callers check and return field-specific errors
- [ ] Empty string is treated as `false` (or whatever the default should be)

---

## 29. Time Formatting: Hardcoded Format Strings

### Severity: LOW (Code Quality)

### Problem

Using hardcoded time format strings instead of standard constants reduces readability and risks typos.

### Bad Example

```go
// BAD: Hardcoded format string
DateCreated: t.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
DateUpdated: t.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
```

### Good Example

```go
import "time"

// GOOD: Use standard library constant
DateCreated: t.DateCreated.Format(time.RFC3339),
DateUpdated: t.DateUpdated.Format(time.RFC3339),
```

### Checklist

- [ ] Use `time.RFC3339` for ISO 8601/RFC 3339 format
- [ ] Import `time` package when using time format constants

---

## 30. Transaction Safety: Non-Atomic Multi-Operation Business Logic

### Severity: CRITICAL (Data Integrity)

### Problem

When business logic requires multiple database operations (e.g., create + delete for a "move" operation), executing them without a transaction can leave data in an inconsistent state if one operation fails.

### Bad Example

```go
// BAD: Non-atomic create + delete - partial failure leaves orphaned data
func (b *Business) MoveTask(ctx context.Context, task Task, targetID uuid.UUID) (Task, error) {
    // If this succeeds but delete fails, we have duplicates
    newTask := Task{...}
    if err := b.storer.Create(ctx, newTask); err != nil {
        return Task{}, err
    }

    // If this fails, newTask exists but original wasn't deleted
    if err := b.storer.Delete(ctx, task); err != nil {
        return Task{}, err
    }

    return newTask, nil
}
```

### Good Example

```go
// GOOD: Atomic transaction ensures all-or-nothing semantics
func (b *Business) MoveTask(ctx context.Context, task Task, targetID uuid.UUID) (Task, error) {
    // Begin transaction to ensure atomicity
    tx, err := b.beginner.Begin()
    if err != nil {
        return Task{}, fmt.Errorf("begin transaction: %w", err)
    }

    // Create transaction-bound storer
    txStorer, err := b.storer.NewWithTx(tx)
    if err != nil {
        tx.Rollback()
        return Task{}, fmt.Errorf("new tx storer: %w", err)
    }

    // Create new task within transaction
    newTask := Task{...}
    if err := txStorer.Create(ctx, newTask); err != nil {
        tx.Rollback()
        return Task{}, fmt.Errorf("create: %w", err)
    }

    // Delete original within same transaction
    if err := txStorer.Delete(ctx, task); err != nil {
        tx.Rollback()
        return Task{}, fmt.Errorf("delete: %w", err)
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return Task{}, fmt.Errorf("commit: %w", err)
    }

    return newTask, nil
}
```

### Checklist

- [ ] Business struct includes `beginner sqldb.Beginner` field for transaction support
- [ ] Multi-operation logic uses `beginner.Begin()` to start transaction
- [ ] Create transaction-bound storer with `storer.NewWithTx(tx)`
- [ ] All rollbacks log errors if they fail
- [ ] Commit only after all operations succeed

---

## 31. Defense-in-Depth: Dropping Constraints Without Trigger Replacement

### Severity: MAJOR (Data Integrity)

### Problem

When moving validation from database constraints to the business layer (for flexibility), removing the constraint without adding a defense-in-depth trigger leaves the database unprotected against direct inserts or bugs in the business layer.

### Bad Example

```sql
-- BAD: Dropping constraint without replacement
DROP CONSTRAINT tasks_contribution_required;
-- Now there's no database-level protection!
```

### Good Example

```sql
-- GOOD: Replace constraint with flexible trigger for defense-in-depth
CREATE OR REPLACE FUNCTION validate_task_contribution()
RETURNS TRIGGER AS $
DECLARE
    objective_tracking_type TEXT;
BEGIN
    -- Rule 1: Inbox tasks must have NULL contribution
    IF NEW.objective_id IS NULL THEN
        IF NEW.contribution IS NOT NULL THEN
            RAISE EXCEPTION 'Inbox tasks cannot have contribution';
        END IF;
        RETURN NEW;
    END IF;

    -- Get objective tracking type
    SELECT tracking_type INTO objective_tracking_type
    FROM objectives WHERE objective_id = NEW.objective_id;

    -- Rule 2: Result objectives require contribution 1-10
    IF objective_tracking_type = 'result' THEN
        IF NEW.contribution IS NULL OR NEW.contribution < 1 OR NEW.contribution > 10 THEN
            RAISE EXCEPTION 'Result tasks require contribution between 1 and 10';
        END IF;
    END IF;

    -- Rule 3: Frequency objectives must have NULL contribution
    IF objective_tracking_type = 'frequency' THEN
        IF NEW.contribution IS NOT NULL THEN
            RAISE EXCEPTION 'Frequency tasks cannot have contribution';
        END IF;
    END IF;

    RETURN NEW;
END;
$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS validate_task_contribution_trigger ON tasks;
CREATE TRIGGER validate_task_contribution_trigger
    BEFORE INSERT OR UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION validate_task_contribution();
```

### Checklist

- [ ] Never drop a constraint without understanding why it existed
- [ ] If business layer handles validation, add trigger for defense-in-depth
- [ ] Triggers should use RAISE EXCEPTION with SQLSTATE for proper error handling
- [ ] Document the business layer validation that the trigger backs up
