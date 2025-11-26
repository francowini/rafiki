# Value Reorder - Backend Implementation

## Overview

This document specifies the backend implementation for atomic value reordering, addressing:
- **Issue #14**: Support drag-and-drop reordering when all 10 value slots are filled
- **Issue #13**: Rollback inconsistency where client state reverts but backend may remain partially reordered

## Architecture Compliance

- **Domain Type**: Child (of userbus)
- **Parent Domain**: userbus
- **Imports**: `userbus.ExtBusiness` (interface-based)
- **Status**: ALIGNED with business-model-dependencies.md

## Solution Summary

Add a new atomic `POST /v1/values/reorder` endpoint that updates multiple values' display orders in a single database transaction. This eliminates the "parking slot" requirement and ensures all-or-nothing semantics.

## Database Schema

**No schema changes required.** The existing `values` table already supports this:

```sql
-- Existing schema (no changes)
CREATE TABLE values (
    value_id       UUID PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(user_id),
    content        TEXT NOT NULL,
    facet          TEXT NOT NULL,
    display_order  INT NOT NULL,
    date_created   TIMESTAMP NOT NULL DEFAULT NOW(),
    date_updated   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX values_user_order_unique_idx ON values(user_id, display_order);
```

## API Endpoint

### POST /v1/values/reorder

Atomically reorders multiple values in a single transaction.

**Request:**
```json
{
  "items": [
    { "id": "uuid-1", "displayOrder": 1 },
    { "id": "uuid-2", "displayOrder": 2 },
    { "id": "uuid-3", "displayOrder": 3 }
  ]
}
```

**Response (200 OK):**
```json
{
  "items": [
    {
      "id": "uuid-1",
      "content": "Health",
      "facet": "physical",
      "displayOrder": 1,
      "dateCreated": "2024-11-20T10:00:00Z",
      "dateUpdated": "2024-11-26T15:30:00Z"
    }
  ],
  "total": 3
}
```

**Error Responses:**
- `400 InvalidArgument`: Invalid UUID or displayOrder out of range (1-10)
- `401 Unauthenticated`: Missing or invalid auth token
- `409 Conflict`: Duplicate displayOrder values in request

**Behavior:**
- Invalid/non-existent value IDs are silently ignored
- Only values belonging to authenticated user are updated
- Transaction rolls back entirely on any failure

## Implementation Files

### 1. Route Registration

**File:** `app/domain/valueapp/route.go`

```go
// Add to Routes function
app.HandlerFunc(http.MethodPost, version, "/values/reorder", api.reorder, bearer)
```

### 2. App Layer Models

**File:** `app/domain/valueapp/model.go`

Add these types:

```go
// ReorderRequest represents a batch reorder request.
type ReorderRequest struct {
    Items []ReorderItem `json:"items" validate:"required,min=1,max=10,dive"`
}

// ReorderItem represents a single item to be reordered.
type ReorderItem struct {
    ID           string `json:"id" validate:"required,uuid"`
    DisplayOrder int    `json:"displayOrder" validate:"required,min=1,max=10"`
}

// Decode implements the decoder interface.
func (rr *ReorderRequest) Decode(data []byte) error {
    return json.Unmarshal(data, rr)
}

// Validate checks the data for correctness.
func (rr ReorderRequest) Validate() error {
    if err := errs.Check(rr); err != nil {
        return fmt.Errorf("validate: %w", err)
    }

    // Check for duplicate displayOrders in request
    seen := make(map[int]bool)
    for _, item := range rr.Items {
        if seen[item.DisplayOrder] {
            return fmt.Errorf("duplicate displayOrder: %d", item.DisplayOrder)
        }
        seen[item.DisplayOrder] = true
    }

    return nil
}

// toBusReorderRequest converts app layer to business domain type.
func toBusReorderRequest(appReorder ReorderRequest) (valuebus.ReorderRequest, error) {
    items := make([]valuebus.ReorderItem, len(appReorder.Items))

    for i, item := range appReorder.Items {
        valueID, err := uuid.Parse(item.ID)
        if err != nil {
            return valuebus.ReorderRequest{}, fmt.Errorf("parse value id: %w", err)
        }

        displayOrderVal, err := displayorder.Parse(item.DisplayOrder)
        if err != nil {
            return valuebus.ReorderRequest{}, fmt.Errorf("parse display order: %w", err)
        }

        items[i] = valuebus.ReorderItem{
            ID:           valueID,
            DisplayOrder: displayOrderVal,
        }
    }

    return valuebus.ReorderRequest{Items: items}, nil
}
```

### 3. HTTP Handler

**File:** `app/domain/valueapp/valueapp.go`

```go
// reorder handles POST /v1/values/reorder
func (a *app) reorder(ctx context.Context, r *http.Request) web.Encoder {
    var appReorder ReorderRequest
    if err := web.Decode(r, &appReorder); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    if err := appReorder.Validate(); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return errs.New(errs.Unauthenticated, err)
    }

    busReorder, err := toBusReorderRequest(appReorder)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    values, err := a.valueBus.Reorder(ctx, userID, busReorder)
    if err != nil {
        if errors.Is(err, valuebus.ErrUserDisabled) {
            return errs.New(errs.PermissionDenied, err)
        }
        if errors.Is(err, valuebus.ErrDuplicateOrder) {
            return errs.New(errs.AlreadyExists, err)
        }
        return errs.Newf(errs.Internal, "reorder: %s", err)
    }

    return query.NewResult(toAppValues(values), len(values), page.Page{})
}
```

### 4. Business Layer Models

**File:** `business/domain/valuebus/model.go`

Add these types:

```go
// ReorderItem represents a value ID with its new display order.
type ReorderItem struct {
    ID           uuid.UUID
    DisplayOrder displayorder.DisplayOrder
}

// ReorderRequest contains items to be reordered.
type ReorderRequest struct {
    Items []ReorderItem
}
```

### 5. Update Interfaces

**File:** `business/domain/valuebus/valuebus.go`

Update `Storer` interface (add BatchUpdate):

```go
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, value Value) error
    Update(ctx context.Context, value Value) error
    Delete(ctx context.Context, value Value) error
    DeleteByUserID(ctx context.Context, userID uuid.UUID) error
    BatchUpdate(ctx context.Context, values []Value) error  // ADD THIS
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Value, error)
    QueryByID(ctx context.Context, valueID uuid.UUID) (Value, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
}
```

Update `ExtBusiness` interface (add Reorder):

```go
type ExtBusiness interface {
    NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
    Create(ctx context.Context, nv NewValue) (Value, error)
    Update(ctx context.Context, value Value, uv UpdateValue) (Value, error)
    Delete(ctx context.Context, value Value) error
    Reorder(ctx context.Context, userID uuid.UUID, rr ReorderRequest) ([]Value, error)  // ADD THIS
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Value, error)
    QueryByID(ctx context.Context, valueID uuid.UUID) (Value, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
}
```

### 6. Business Layer Method

**File:** `business/domain/valuebus/valuebus.go`

Add the Reorder method:

```go
// Reorder atomically updates display order for multiple values.
// Invalid/non-existent value IDs are silently ignored.
// Returns updated values sorted by display order.
func (b *Business) Reorder(ctx context.Context, userID uuid.UUID, rr ReorderRequest) ([]Value, error) {
    // Validate parent user exists and is enabled
    usr, err := b.userBus.QueryByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user.querybyid: %s: %w", userID, err)
    }

    if !usr.Enabled {
        return nil, ErrUserDisabled
    }

    // Get current values for the user
    filter := QueryFilter{UserID: &userID}
    orderBy := order.By{Field: OrderByDisplayOrder, Direction: order.ASC}
    pg := page.NewPage(1, 10)

    currentValues, err := b.storer.Query(ctx, filter, orderBy, pg)
    if err != nil {
        return nil, fmt.Errorf("query current values: %w", err)
    }

    // Build update map from request
    updateMap := make(map[string]displayorder.DisplayOrder)
    for _, item := range rr.Items {
        updateMap[item.ID.String()] = item.DisplayOrder
    }

    // Prepare values to update (only user's values that are in request)
    now := time.Now()
    valuesToUpdate := make([]Value, 0, len(currentValues))
    for _, current := range currentValues {
        if newOrder, found := updateMap[current.ID.String()]; found {
            current.DisplayOrder = newOrder
            current.DateUpdated = now
            valuesToUpdate = append(valuesToUpdate, current)
        }
    }

    // Silent ignore: if no matching values, return current state
    if len(valuesToUpdate) == 0 {
        return currentValues, nil
    }

    // Atomic batch update
    if err := b.storer.BatchUpdate(ctx, valuesToUpdate); err != nil {
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return nil, ErrDuplicateOrder
        }
        return nil, fmt.Errorf("batch update: %w", err)
    }

    // Return updated values
    updatedValues, err := b.storer.Query(ctx, filter, orderBy, pg)
    if err != nil {
        return nil, fmt.Errorf("query updated values: %w", err)
    }

    return updatedValues, nil
}
```

### 7. Store Layer Method

**File:** `business/domain/valuebus/stores/valuedb/valuedb.go`

Add BatchUpdate method:

```go
// BatchUpdate updates multiple values atomically.
// All updates happen within the caller's transaction context.
func (s *Store) BatchUpdate(ctx context.Context, values []valuebus.Value) error {
    if len(values) == 0 {
        return nil
    }

    const q = `
    UPDATE values SET
        display_order = :display_order,
        date_updated = :date_updated
    WHERE
        value_id = :value_id`

    for _, value := range values {
        data := struct {
            ValueID      string    `db:"value_id"`
            DisplayOrder int       `db:"display_order"`
            DateUpdated  time.Time `db:"date_updated"`
        }{
            ValueID:      value.ID.String(),
            DisplayOrder: value.DisplayOrder.Value(),
            DateUpdated:  value.DateUpdated,
        }

        if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
            if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
                return valuebus.ErrDuplicateOrder
            }
            return fmt.Errorf("namedexeccontext: %w", err)
        }
    }

    return nil
}
```

## Transaction Handling

The batch update executes within the database connection's implicit transaction. For explicit transaction control when needed:

```go
// In HTTP handler or middleware, wrap with transaction:
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

busWithTx, err := valueBus.NewWithTx(tx)
if err != nil {
    return err
}

values, err := busWithTx.Reorder(ctx, userID, request)
if err != nil {
    return err // Rollback happens via defer
}

return tx.Commit()
```

## Error Handling

| Error | HTTP Status | Condition |
|-------|-------------|-----------|
| `ErrUserDisabled` | 403 Forbidden | User account is disabled |
| `ErrDuplicateOrder` | 409 Conflict | Two values assigned same displayOrder |
| Validation error | 400 Bad Request | Invalid UUID or displayOrder out of range |
| `ErrNotFound` | Silent ignore | Value ID doesn't exist (per user decision) |

## Testing Scenarios

1. **Happy path**: Reorder 3 values among 10 filled slots
2. **All 10 slots filled**: Swap positions 1 and 10
3. **Invalid ID**: Include non-existent UUID (should be silently ignored)
4. **Duplicate order**: Send two values with same displayOrder (should return 409)
5. **Unauthorized**: Try to reorder another user's values (should be ignored)
6. **Empty request**: Send empty items array (should return current values)

## Deployment Notes

1. Deploy backend first with new endpoint
2. Endpoint is backward compatible (existing PUT still works)
3. No database migration required
4. Frontend can be deployed after backend is live
