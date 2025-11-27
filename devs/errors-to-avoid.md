# Critical Errors to Avoid

This document catalogs critical errors discovered during code review that should be avoided in future development. Use this as a checklist when implementing new features.

---

## Table of Contents

1. [Security: Child Entity Ownership Validation](#1-security-child-entity-ownership-validation)
2. [Error Handling: Sentinel vs Generic Errors](#2-error-handling-sentinel-vs-generic-errors)
3. [String Length: UTF-8 Rune Count vs Byte Count](#3-string-length-utf-8-rune-count-vs-byte-count)
4. [Strong Types: Missing Value() Method](#4-strong-types-missing-value-method)
5. [SQL: WHERE Clause Building](#5-sql-where-clause-building)
6. [Logging: Missing Structured Logging in Business Layer](#6-logging-missing-structured-logging-in-business-layer)
7. [Code Duplication: Repeated Decrypt-Parse Patterns](#7-code-duplication-repeated-decrypt-parse-patterns)

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

## Quick Reference Checklist

When implementing new features, verify:

- [ ] **Security**: Child entities validate parent ownership against authenticated user
- [ ] **UTF-8**: String length validation uses `utf8.RuneCountInString()`
- [ ] **Strong Types**: Include both `Value()` and `String()` methods
- [ ] **SQL**: Use `strings.Join()` for WHERE clause construction
- [ ] **Errors**: Define domain-specific errors (e.g., `ErrNotValueOwner`)
- [ ] **App Layer**: Handle all business errors with appropriate HTTP status codes
- [ ] **Logging**: Business layer methods include structured entry/error/success logging
- [ ] **DRY**: Extract helpers for patterns repeated 3+ times
