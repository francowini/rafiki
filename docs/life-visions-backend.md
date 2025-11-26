# Life Visions - Backend Implementation

## Overview

Life Visions are aspirational states of being that represent how users want to live continuously. Each life vision belongs to exactly one value (many-to-one relationship).

## Architecture Compliance

- **Domain Type**: Child (of valuebus)
- **Parent Domain**: valuebus
- **Imports**: valuebus.ExtBusiness (one-directional)
- **Status**: ALIGNED with business-model-dependencies.md

## Database Schema

**File**: `business/sdk/migrate/sql/migrate.sql`

Add Version 1.06:

```sql
-- Version: 1.06
-- Description: Create life_visions table for aspirational states of being

CREATE TABLE life_visions (
    life_vision_id  UUID        NOT NULL,
    user_id         UUID        NOT NULL,
    value_id        UUID        NOT NULL,
    content         TEXT        NOT NULL,  -- encrypted (10-500 chars validation)
    date_created    TIMESTAMP   NOT NULL,
    date_updated    TIMESTAMP   NOT NULL,

    PRIMARY KEY (life_vision_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (value_id) REFERENCES values(value_id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX life_visions_user_id_idx ON life_visions(user_id);
CREATE INDEX life_visions_value_id_idx ON life_visions(value_id);
CREATE INDEX life_visions_date_created_idx ON life_visions(date_created DESC);
```

## Business Types

### LifeVisionContent

**File**: `business/types/lifevisioncontent/lifevisioncontent.go`

```go
package lifevisioncontent

import (
	"fmt"
	"strings"
)

// Set of known errors for life vision content validation.
var (
	ErrEmpty    = fmt.Errorf("life vision content cannot be empty")
	ErrTooShort = fmt.Errorf("life vision content must be at least 10 characters")
	ErrTooLong  = fmt.Errorf("life vision content must be less than 500 characters")
)

// LifeVisionContent represents a validated life vision statement.
type LifeVisionContent struct {
	value string
}

// Value returns the string value of the life vision content.
func (c LifeVisionContent) Value() string {
	return c.value
}

// String returns the string representation.
func (c LifeVisionContent) String() string {
	return c.value
}

// Equal provides support for the go-cmp package and testing.
func (c LifeVisionContent) Equal(c2 LifeVisionContent) bool {
	return c.value == c2.value
}

// MarshalText provides support for logging and any marshal needs.
func (c LifeVisionContent) MarshalText() ([]byte, error) {
	return []byte(c.value), nil
}

// Parse validates and creates a LifeVisionContent.
func Parse(value string) (LifeVisionContent, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return LifeVisionContent{}, ErrEmpty
	}

	if len(value) < 10 {
		return LifeVisionContent{}, ErrTooShort
	}

	if len(value) > 500 {
		return LifeVisionContent{}, ErrTooLong
	}

	return LifeVisionContent{value: value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) LifeVisionContent {
	content, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return content
}
```

## Domain Model

### Model

**File**: `business/domain/lifevisionbus/model.go`

```go
package lifevisionbus

import (
	"time"

	"github.com/francowini/rafiki/business/types/lifevisioncontent"
	"github.com/google/uuid"
)

// LifeVision represents an aspirational state of being.
type LifeVision struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ValueID     uuid.UUID
	Content     lifevisioncontent.LifeVisionContent
	DateCreated time.Time
	DateUpdated time.Time
}

// NewLifeVision contains information needed to create a new life vision.
type NewLifeVision struct {
	ValueID uuid.UUID
	Content lifevisioncontent.LifeVisionContent
}

// UpdateLifeVision contains information needed to update a life vision.
type UpdateLifeVision struct {
	Content *lifevisioncontent.LifeVisionContent
	ValueID *uuid.UUID
}
```

### Filter

**File**: `business/domain/lifevisionbus/filter.go`

```go
package lifevisionbus

import "github.com/google/uuid"

// QueryFilter holds the available fields to filter life visions.
type QueryFilter struct {
	ID      *uuid.UUID
	UserID  *uuid.UUID
	ValueID *uuid.UUID
}
```

### Order

**File**: `business/domain/lifevisionbus/order.go`

```go
package lifevisionbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default order for queries.
var DefaultOrderBy = order.NewBy(OrderByDateCreated, order.DESC)

// Order field names for life visions.
const (
	OrderByID          = "life_vision_id"
	OrderByDateCreated = "date_created"
	OrderByDateUpdated = "date_updated"
)
```

### Events

**File**: `business/domain/lifevisionbus/event.go`

```go
package lifevisionbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/google/uuid"
)

// Domain name and actions for delegate events.
const (
	DomainName    = "lifevision"
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	LifeVisionID uuid.UUID
	ValueID      uuid.UUID
	UserID       uuid.UUID
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(lv LifeVision) delegate.Data {
	params := ActionDeletedParms{
		LifeVisionID: lv.ID,
		ValueID:      lv.ValueID,
		UserID:       lv.UserID,
	}

	rawParams, _ := json.Marshal(params)

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}

// registerDelegateFunctions registers delegate handlers.
func (b *Business) registerDelegateFunctions() {
	if b.delegate != nil {
		b.delegate.Register(valuebus.DomainName, valuebus.ActionDeleted, b.actionValueDeleted)
	}
}

// actionValueDeleted handles value deletion events.
func (b *Business) actionValueDeleted(ctx context.Context, data delegate.Data) error {
	var params valuebus.ActionDeletedParms
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-valuedeleted",
		"value_id", params.ValueID,
		"status", "life visions deleted via CASCADE")

	return nil
}
```

### Business Logic

**File**: `business/domain/lifevisionbus/lifevisionbus.go`

```go
package lifevisionbus

import (
	"context"
	"fmt"
	"time"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs for data storage.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, lifeVision LifeVision) error
	Update(ctx context.Context, lifeVision LifeVision) error
	Delete(ctx context.Context, lifeVision LifeVision) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LifeVision, error)
	QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (LifeVision, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness represents the external business API for life visions.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error)
	Update(ctx context.Context, lifeVision LifeVision, ulv UpdateLifeVision) (LifeVision, error)
	Delete(ctx context.Context, lifeVision LifeVision) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LifeVision, error)
	QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (LifeVision, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for life vision access.
type Business struct {
	log      *logger.Logger
	valueBus valuebus.ExtBusiness
	delegate *delegate.Delegate
	storer   Storer
}

// NewBusiness constructs a life vision business API for use.
func NewBusiness(
	log *logger.Logger,
	valueBus valuebus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
) ExtBusiness {
	b := &Business{
		log:      log,
		valueBus: valueBus,
		delegate: dlg,
		storer:   storer,
	}

	b.registerDelegateFunctions()

	return b
}

// NewWithTx creates a new Business with transaction support.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	valueBus, err := b.valueBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		valueBus: valueBus,
		delegate: b.delegate,
		storer:   storer,
	}, nil
}

// Create adds a new life vision.
func (b *Business) Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error) {
	// Validate parent value exists and user owns it
	value, err := b.valueBus.QueryByID(ctx, nlv.ValueID)
	if err != nil {
		return LifeVision{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", nlv.ValueID, err)
	}

	now := time.Now()

	lifeVision := LifeVision{
		ID:          uuid.New(),
		UserID:      value.UserID,
		ValueID:     nlv.ValueID,
		Content:     nlv.Content,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := b.storer.Create(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("create: %w", err)
	}

	return lifeVision, nil
}

// Update modifies an existing life vision.
func (b *Business) Update(ctx context.Context, lifeVision LifeVision, ulv UpdateLifeVision) (LifeVision, error) {
	if ulv.Content != nil {
		lifeVision.Content = *ulv.Content
	}

	if ulv.ValueID != nil {
		// Validate new value exists and user owns it
		value, err := b.valueBus.QueryByID(ctx, *ulv.ValueID)
		if err != nil {
			return LifeVision{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", *ulv.ValueID, err)
		}

		// Ensure the new value belongs to the same user
		if value.UserID != lifeVision.UserID {
			return LifeVision{}, fmt.Errorf("cannot move life vision to value owned by different user")
		}

		lifeVision.ValueID = *ulv.ValueID
	}

	lifeVision.DateUpdated = time.Now()

	if err := b.storer.Update(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("update: %w", err)
	}

	return lifeVision, nil
}

// Delete removes a life vision.
func (b *Business) Delete(ctx context.Context, lifeVision LifeVision) error {
	if err := b.storer.Delete(ctx, lifeVision); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Publish event for potential future child domains
	if b.delegate != nil {
		if err := b.delegate.Call(ctx, ActionDeletedData(lifeVision)); err != nil {
			return fmt.Errorf("delegate call: %w", err)
		}
	}

	return nil
}

// Query retrieves life visions based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LifeVision, error) {
	lifeVisions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return lifeVisions, nil
}

// QueryByID retrieves a single life vision by ID.
func (b *Business) QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (LifeVision, error) {
	lifeVision, err := b.storer.QueryByID(ctx, lifeVisionID)
	if err != nil {
		return LifeVision{}, fmt.Errorf("query: lifeVisionID[%s]: %w", lifeVisionID, err)
	}

	return lifeVision, nil
}

// Count returns the number of life visions matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return b.storer.Count(ctx, filter)
}
```

## API Endpoints

**File**: `app/domain/lifevisionapp/route.go`

```go
package lifevisionapp

import (
	"net/http"

	"github.com/francowini/rafiki/api/sdk/http/mid"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains dependencies for the lifevision routes.
type Config struct {
	LifeVisionBus lifevisionbus.ExtBusiness
	Auth          *mid.Auth
}

// Routes registers the life vision API endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.LifeVisionBus)

	// CRUD operations
	app.HandlerFunc(http.MethodPost, version, "/lifevisions", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/lifevisions", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/lifevisions/{lifevision_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/lifevisions/{lifevision_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/lifevisions/{lifevision_id}", api.delete, bearer)

	// Query by value (nested endpoint)
	app.HandlerFunc(http.MethodGet, version, "/values/{value_id}/lifevisions", api.queryByValue, bearer)
}
```

## API Models

**File**: `app/domain/lifevisionapp/model.go`

```go
package lifevisionapp

import (
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/types/lifevisioncontent"
	"github.com/google/uuid"
)

// LifeVision represents a life vision in the API response.
type LifeVision struct {
	ID          string `json:"id"`
	ValueID     string `json:"valueId"`
	Content     string `json:"content"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

// NewLifeVision represents data for creating a life vision.
type NewLifeVision struct {
	ValueID string `json:"valueId" validate:"required,uuid"`
	Content string `json:"content" validate:"required"`
}

// UpdateLifeVision represents data for updating a life vision.
type UpdateLifeVision struct {
	ValueID *string `json:"valueId" validate:"omitempty,uuid"`
	Content *string `json:"content"`
}

// Decode validates and converts NewLifeVision to business layer type.
func (app NewLifeVision) Decode() (lifevisionbus.NewLifeVision, error) {
	valueID, err := uuid.Parse(app.ValueID)
	if err != nil {
		return lifevisionbus.NewLifeVision{}, err
	}

	content, err := lifevisioncontent.Parse(app.Content)
	if err != nil {
		return lifevisionbus.NewLifeVision{}, err
	}

	return lifevisionbus.NewLifeVision{
		ValueID: valueID,
		Content: content,
	}, nil
}

// Decode validates and converts UpdateLifeVision to business layer type.
func (app UpdateLifeVision) Decode() (lifevisionbus.UpdateLifeVision, error) {
	var ulv lifevisionbus.UpdateLifeVision

	if app.ValueID != nil {
		valueID, err := uuid.Parse(*app.ValueID)
		if err != nil {
			return lifevisionbus.UpdateLifeVision{}, err
		}
		ulv.ValueID = &valueID
	}

	if app.Content != nil {
		content, err := lifevisioncontent.Parse(*app.Content)
		if err != nil {
			return lifevisionbus.UpdateLifeVision{}, err
		}
		ulv.Content = &content
	}

	return ulv, nil
}

// Encode converts a business life vision to API response.
func Encode(lv lifevisionbus.LifeVision) LifeVision {
	return LifeVision{
		ID:          lv.ID.String(),
		ValueID:     lv.ValueID.String(),
		Content:     lv.Content.String(),
		DateCreated: lv.DateCreated.Format("2006-01-02T15:04:05Z"),
		DateUpdated: lv.DateUpdated.Format("2006-01-02T15:04:05Z"),
	}
}
```

## Files to Create

```
business/types/lifevisioncontent/
└── lifevisioncontent.go

business/domain/lifevisionbus/
├── lifevisionbus.go
├── model.go
├── event.go
├── filter.go
├── order.go
└── stores/
    └── lifevisiondb/
        ├── lifevisiondb.go
        ├── model.go
        └── order.go

app/domain/lifevisionapp/
├── lifevisionapp.go
├── model.go
├── order.go
└── route.go
```

## Wiring (main.go)

After valueBus initialization, add:

```go
// Create life vision domain (child of value)
lifeVisionStore := lifevisiondb.NewStore(log, db, encryptor)
lifeVisionBus := lifevisionbus.NewBusiness(log, valueBus, dlg, lifeVisionStore)
```

## Implementation Checklist

- [ ] Create `lifevisioncontent` strong type package
- [ ] Add migration Version 1.06
- [ ] Create `lifevisionbus` domain package
- [ ] Create `lifevisiondb` store package
- [ ] Create `lifevisionapp` API package
- [ ] Wire in main.go
- [ ] Register routes in all/all.go
- [ ] Test cascade delete (delete value → visions deleted)
- [ ] Test API endpoints
