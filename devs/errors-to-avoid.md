# Critical Errors to Avoid

This document catalogs critical errors discovered during code review that should be avoided in future development. Use this as a checklist when implementing new features.

---

## Table of Contents

1. [Security: Child Entity Ownership Validation](#1-security-child-entity-ownership-validation)
2. [Error Handling: Sentinel vs Generic Errors](#2-error-handling-sentinel-vs-generic-errors)
3. [String Length: UTF-8 Rune Count vs Byte Count](#3-string-length-utf-8-rune-count-vs-byte-count)
4. [Strong Types: Missing Value() Method](#4-strong-types-missing-value-method)
5. [SQL: WHERE Clause Building](#5-sql-where-clause-building)

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

## Quick Reference Checklist

When implementing new features, verify:

- [ ] **Security**: Child entities validate parent ownership against authenticated user
- [ ] **UTF-8**: String length validation uses `utf8.RuneCountInString()`
- [ ] **Strong Types**: Include both `Value()` and `String()` methods
- [ ] **SQL**: Use `strings.Join()` for WHERE clause construction
- [ ] **Errors**: Define domain-specific errors (e.g., `ErrNotValueOwner`)
- [ ] **App Layer**: Handle all business errors with appropriate HTTP status codes
