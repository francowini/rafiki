# Export Feature - Backend Implementation

## Overview

Export moments and thinks from a date range as JSON for client-side markdown generation. This feature enables users to download their journal entries to share with their psychologist.

## Architecture Compliance

- **Domain Type**: Query Domain (read-only, aggregates data from multiple domains)
- **Parent Domain**: None (isolated island)
- **Imports**: No business domain imports (momentbus, thinkbus) - only SDK types
- **Pattern Reference**: `vproductbus/` (canonical Query Domain implementation)
- **Status**: ALIGNED with business-model-dependencies.md

### Intentional Differences from vproductbus

1. **Required UserID Filter**: Unlike vproductbus where all filters are optional, `UserID` is always required for security (user data isolation)
2. **Encryption/Decryption**: Store layer requires `encrypt.Encryptor` because export data contains sensitive journal entries
3. **Date Range Validation**: Business layer validates `StartDate < EndDate` (domain-specific business rule)

## Database View

Add to `/business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 1.07
-- Description: Create view_export_items for combined moments and thinks export

CREATE VIEW view_export_items AS
-- Select moments with type identifier
SELECT
    moment_id AS item_id,
    user_id,
    'moment' AS item_type,
    moment_date AS item_date,
    situation,
    thoughts,
    physical_symptoms,
    behavior,
    consequences,
    values_reflection,
    intensity,
    NULL AS category,
    NULL AS content,
    date_created
FROM moments

UNION ALL

-- Select thinks with type identifier
SELECT
    think_id AS item_id,
    user_id,
    'think' AS item_type,
    date_created AS item_date,
    NULL AS situation,
    NULL AS thoughts,
    NULL AS physical_symptoms,
    NULL AS behavior,
    NULL AS consequences,
    NULL AS values_reflection,
    NULL AS intensity,
    category,
    content,
    date_created
FROM thinks;
```

**Note**: PlanetScale (MySQL 8.0) supports views. No index on view needed - underlying table indexes will be used.

## Domain Structure

```
business/domain/vexportbus/
├── vexportbus.go     # Business logic (read-only) with ExtBusiness interface
├── model.go          # ExportItem model
├── filter.go         # QueryFilter with dates
├── order.go          # Ordering constants (NEW)
└── stores/
    └── vexportdb/
        ├── vexportdb.go  # View queries with pagination
        ├── model.go      # DB model and conversion
        ├── filter.go     # Filter application (NEW)
        └── order.go      # Order field mapping (NEW)
```

## Business Types

### File: `business/domain/vexportbus/model.go`

```go
package vexportbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

var (
	ErrNotFound         = errors.New("export items not found")
	ErrInvalidDateRange = errors.New("start_date must be before end_date")
)

type ItemType string

const (
	ItemTypeMoment ItemType = "moment"
	ItemTypeThink  ItemType = "think"
)

func (it ItemType) String() string {
	return string(it)
}

// ExportItem represents a unified export item (moment or think).
type ExportItem struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ItemType         ItemType
	ItemDate         time.Time

	// Moment-specific fields (nil for thinks)
	Situation        *content.Content
	Thoughts         *content.Content
	PhysicalSymptoms *content.Content
	Behavior         *content.Content
	Consequences     *content.Content
	ValuesReflection *content.Content
	Intensity        *intensity.Intensity

	// Think-specific fields (empty/nil for moments)
	Category string
	Content  *content.Content

	DateCreated time.Time
}
```

### File: `business/domain/vexportbus/filter.go`

```go
package vexportbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// NOTE: Unlike vproductbus, UserID is REQUIRED (not pointer) for security.
// Export data must always be scoped to the authenticated user.
type QueryFilter struct {
	UserID    uuid.UUID  // REQUIRED - always set for user isolation
	StartDate *time.Time // Filter items >= this date (inclusive)
	EndDate   *time.Time // Filter items <= this date (inclusive)
}
```

### File: `business/domain/vexportbus/order.go`

```go
package vexportbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByItemDate, order.DESC)

// Set of fields that the results can be ordered by.
const (
	OrderByItemID      = "item_id"
	OrderByItemType    = "item_type"
	OrderByItemDate    = "item_date"
	OrderByDateCreated = "date_created"
)
```

### File: `business/domain/vexportbus/vexportbus.go`

```go
package vexportbus

import (
	"context"
	"fmt"

	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data. READ ONLY (no Create/Update/Delete).
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ExportItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra
// functionality around the core business logic.
type ExtBusiness interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ExportItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Extension is a function that wraps a new layer of business logic
// around the existing business logic.
type Extension func(ExtBusiness) ExtBusiness

// Business manages the set of APIs for view export access.
type Business struct {
	storer Storer
}

// NewBusiness constructs a vexport business API for use.
func NewBusiness(storer Storer, extensions ...Extension) ExtBusiness {
	b := ExtBusiness(&Business{
		storer: storer,
	})

	for i := len(extensions) - 1; i >= 0; i-- {
		ext := extensions[i]
		if ext != nil {
			b = ext(b)
		}
	}

	return b
}

// Query retrieves a list of export items based on the filter.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ExportItem, error) {
	// Domain-specific validation: date range must be valid
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			return nil, ErrInvalidDateRange
		}
	}

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of export items matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return b.storer.Count(ctx, filter)
}
```

## Database Store

### File: `business/domain/vexportbus/stores/vexportdb/model.go`

```go
package vexportdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

type exportItem struct {
	ID               uuid.UUID      `db:"item_id"`
	UserID           uuid.UUID      `db:"user_id"`
	ItemType         string         `db:"item_type"`
	ItemDate         time.Time      `db:"item_date"`
	Situation        sql.NullString `db:"situation"`
	Thoughts         sql.NullString `db:"thoughts"`
	PhysicalSymptoms sql.NullString `db:"physical_symptoms"`
	Behavior         sql.NullString `db:"behavior"`
	Consequences     sql.NullString `db:"consequences"`
	ValuesReflection sql.NullString `db:"values_reflection"`
	Intensity        sql.NullInt32  `db:"intensity"`
	Category         sql.NullString `db:"category"`
	Content          sql.NullString `db:"content"`
	DateCreated      time.Time      `db:"date_created"`
}

func toBusExportItemDecrypted(db exportItem, enc encrypt.Encryptor) (vexportbus.ExportItem, error) {
	item := vexportbus.ExportItem{
		ID:          db.ID,
		UserID:      db.UserID,
		ItemType:    vexportbus.ItemType(db.ItemType),
		ItemDate:    db.ItemDate.In(time.Local),
		DateCreated: db.DateCreated.In(time.Local),
	}

	// Decrypt moment-specific fields
	if db.Situation.Valid {
		decrypted, err := enc.Decrypt(db.Situation.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt situation: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse situation: %w", err)
		}
		item.Situation = &parsed
	}

	if db.Thoughts.Valid {
		decrypted, err := enc.Decrypt(db.Thoughts.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt thoughts: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse thoughts: %w", err)
		}
		item.Thoughts = &parsed
	}

	if db.PhysicalSymptoms.Valid {
		decrypted, err := enc.Decrypt(db.PhysicalSymptoms.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt physical_symptoms: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse physical_symptoms: %w", err)
		}
		item.PhysicalSymptoms = &parsed
	}

	if db.Behavior.Valid {
		decrypted, err := enc.Decrypt(db.Behavior.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt behavior: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse behavior: %w", err)
		}
		item.Behavior = &parsed
	}

	if db.Consequences.Valid {
		decrypted, err := enc.Decrypt(db.Consequences.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt consequences: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse consequences: %w", err)
		}
		item.Consequences = &parsed
	}

	if db.ValuesReflection.Valid {
		decrypted, err := enc.Decrypt(db.ValuesReflection.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt values_reflection: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse values_reflection: %w", err)
		}
		item.ValuesReflection = &parsed
	}

	if db.Intensity.Valid {
		parsed, err := intensity.Parse(int(db.Intensity.Int32))
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse intensity: %w", err)
		}
		item.Intensity = &parsed
	}

	// Decrypt think-specific fields
	if db.Category.Valid {
		item.Category = db.Category.String
	}

	if db.Content.Valid {
		decrypted, err := enc.Decrypt(db.Content.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt content: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse content: %w", err)
		}
		item.Content = &parsed
	}

	return item, nil
}

func toBusExportItemsDecrypted(dbs []exportItem, enc encrypt.Encryptor) ([]vexportbus.ExportItem, error) {
	items := make([]vexportbus.ExportItem, len(dbs))

	for i, db := range dbs {
		var err error
		items[i], err = toBusExportItemDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}
```

### File: `business/domain/vexportbus/stores/vexportdb/filter.go`

```go
package vexportdb

import (
	"bytes"
	"strings"

	"github.com/francowini/rafiki/business/domain/vexportbus"
)

func (s *Store) applyFilter(filter vexportbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	// UserID is always required (security: user data isolation)
	data["user_id"] = filter.UserID
	wc = append(wc, "user_id = :user_id")

	if filter.StartDate != nil {
		data["start_date"] = filter.StartDate.UTC()
		wc = append(wc, "item_date >= :start_date")
	}

	if filter.EndDate != nil {
		data["end_date"] = filter.EndDate.UTC()
		wc = append(wc, "item_date <= :end_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
```

### File: `business/domain/vexportbus/stores/vexportdb/order.go`

```go
package vexportdb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	vexportbus.OrderByItemID:      "item_id",
	vexportbus.OrderByItemType:    "item_type",
	vexportbus.OrderByItemDate:    "item_date",
	vexportbus.OrderByDateCreated: "date_created",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

### File: `business/domain/vexportbus/stores/vexportdb/vexportdb.go`

```go
package vexportdb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages the set of APIs for export view database access.
type Store struct {
	log       *logger.Logger
	db        sqlx.ExtContext
	encryptor encrypt.Encryptor
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {
	return &Store{
		log:       log,
		db:        db,
		encryptor: encryptor,
	}
}

// Query retrieves a list of export items from the database view.
func (s *Store) Query(ctx context.Context, filter vexportbus.QueryFilter, orderBy order.By, pg page.Page) ([]vexportbus.ExportItem, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		item_id, user_id, item_type, item_date,
		situation, thoughts, physical_symptoms, behavior,
		consequences, values_reflection, intensity,
		category, content, date_created
	FROM
		view_export_items`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dbItems []exportItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbItems); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusExportItemsDecrypted(dbItems, s.encryptor)
}

// Count returns the total number of export items matching the filter.
func (s *Store) Count(ctx context.Context, filter vexportbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		count(1) AS count
	FROM
		view_export_items`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}
```

## API Endpoint

### File: `app/domain/vexportapp/model.go`

```go
package vexportapp

import (
	"time"

	"github.com/francowini/rafiki/business/domain/vexportbus"
)

type ExportItem struct {
	ID               string  `json:"id"`
	ItemType         string  `json:"itemType"`
	ItemDate         string  `json:"itemDate"`

	// Moment-specific fields
	Situation        *string `json:"situation,omitempty"`
	Thoughts         *string `json:"thoughts,omitempty"`
	PhysicalSymptoms *string `json:"physicalSymptoms,omitempty"`
	Behavior         *string `json:"behavior,omitempty"`
	Consequences     *string `json:"consequences,omitempty"`
	ValuesReflection *string `json:"valuesReflection,omitempty"`
	Intensity        *int    `json:"intensity,omitempty"`

	// Think-specific fields
	Category string  `json:"category,omitempty"`
	Content  *string `json:"content,omitempty"`

	DateCreated string `json:"dateCreated"`
}

func toAppExportItem(item vexportbus.ExportItem) ExportItem {
	appItem := ExportItem{
		ID:          item.ID.String(),
		ItemType:    item.ItemType.String(),
		ItemDate:    item.ItemDate.Format(time.RFC3339),
		DateCreated: item.DateCreated.Format(time.RFC3339),
	}

	if item.Situation != nil {
		s := item.Situation.String()
		appItem.Situation = &s
	}
	if item.Thoughts != nil {
		s := item.Thoughts.String()
		appItem.Thoughts = &s
	}
	if item.PhysicalSymptoms != nil {
		s := item.PhysicalSymptoms.String()
		appItem.PhysicalSymptoms = &s
	}
	if item.Behavior != nil {
		s := item.Behavior.String()
		appItem.Behavior = &s
	}
	if item.Consequences != nil {
		s := item.Consequences.String()
		appItem.Consequences = &s
	}
	if item.ValuesReflection != nil {
		s := item.ValuesReflection.String()
		appItem.ValuesReflection = &s
	}
	if item.Intensity != nil {
		i := item.Intensity.Value()
		appItem.Intensity = &i
	}

	appItem.Category = item.Category
	if item.Content != nil {
		s := item.Content.String()
		appItem.Content = &s
	}

	return appItem
}

func toAppExportItems(items []vexportbus.ExportItem) []ExportItem {
	app := make([]ExportItem, len(items))
	for i, item := range items {
		app[i] = toAppExportItem(item)
	}
	return app
}
```

### File: `app/domain/vexportapp/vexportapp.go`

```go
package vexportapp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

// Ordering fields mapping for API
var orderByFields = map[string]string{
	"item_date":    vexportbus.OrderByItemDate,
	"item_type":    vexportbus.OrderByItemType,
	"date_created": vexportbus.OrderByDateCreated,
}

type app struct {
	vexportBus vexportbus.ExtBusiness
}

func newApp(vexportBus vexportbus.ExtBusiness) *app {
	return &app{
		vexportBus: vexportBus,
	}
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Parse pagination (default: page 1, 1000 rows for exports)
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page/rows", err)
	}

	// Parse ordering (default: item_date DESC)
	orderBy, err := order.Parse(orderByFields, qp.OrderBy, vexportbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("orderBy", err)
	}

	// Parse date filters
	var startDate, endDate *time.Time

	if qp.StartDate != "" {
		parsed, err := time.Parse(time.RFC3339, qp.StartDate)
		if err != nil {
			return errs.NewFieldErrors("start_date", fmt.Errorf("must be RFC3339 format"))
		}
		startDate = &parsed
	}

	if qp.EndDate != "" {
		parsed, err := time.Parse(time.RFC3339, qp.EndDate)
		if err != nil {
			return errs.NewFieldErrors("end_date", fmt.Errorf("must be RFC3339 format"))
		}
		endDate = &parsed
	}

	filter := vexportbus.QueryFilter{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Execute query
	items, err := a.vexportBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		if err == vexportbus.ErrInvalidDateRange {
			return errs.NewFieldErrors("date_range", err)
		}
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	// Get total count
	total, err := a.vexportBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppExportItems(items), total, pg)
}

type queryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	StartDate string
	EndDate   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	return queryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}
```

### File: `app/domain/vexportapp/route.go`

```go
package vexportapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/foundation/web"
)

type Config struct {
	VExportBus vexportbus.ExtBusiness
	Auth       *auth.Auth
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.VExportBus)

	app.HandlerFunc(http.MethodGet, version, "/export", api.query, bearer)
}
```

## Integration

### Update `app/sdk/mux/mux.go`

Add to `BusConfig` struct:

```go
type BusConfig struct {
	ThinkBus      *thinkbus.Business
	MomentBus     *momentbus.Business
	VExportBus    vexportbus.ExtBusiness  // ADD THIS (note: ExtBusiness interface)
	ValueBus      valuebus.ExtBusiness
	LifeVisionBus lifevisionbus.ExtBusiness
	UserBus       userbus.ExtBusiness
	Auth          *auth.Auth
}
```

### Update `api/services/partners/all/all.go`

Add import and route:

```go
import (
	"github.com/francowini/rafiki/app/domain/vexportapp"
	// ... other imports
)

func (Add) Add(app *web.App, cfg mux.Config) {
	// ... existing routes ...

	vexportapp.Routes(app, vexportapp.Config{
		VExportBus: cfg.BusConfig.VExportBus,
		Auth:       cfg.BusConfig.Auth,
	})
}
```

### Update `api/services/partners/main.go`

Initialize business in `run` function:

```go
// Initialize vexport business
vexportStore := vexportdb.NewStore(log, db, encryptor)
vexportBus := vexportbus.NewBusiness(log, vexportStore)

// Add to BusConfig
BusConfig: mux.BusConfig{
	ThinkBus:      thinkBus,
	MomentBus:     momentBus,
	VExportBus:    vexportBus,  // ADD THIS
	ValueBus:      valueBus,
	LifeVisionBus: lifeVisionBus,
	UserBus:       userBus,
	Auth:          auth,
},
```

## API Specification

### Endpoint

```
GET /v1/export?start_date=2024-11-20T00:00:00Z&end_date=2024-11-27T23:59:59Z&page=1&rows=1000
```

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| start_date | string (RFC3339) | No | - | Filter items >= this date |
| end_date | string (RFC3339) | No | - | Filter items <= this date |
| page | int | No | 1 | Page number (1-based) |
| rows | int | No | 10 | Items per page (use 1000 for exports) |
| orderBy | string | No | item_date,DESC | Sort field and direction |

### Available Order Fields

| Field | Description |
|-------|-------------|
| item_date | Date of the moment/think |
| item_type | Type: "moment" or "think" |
| date_created | When the entry was created |

### Authentication

- Bearer token required (JWT)
- User ID extracted from token for data isolation

### Response

```json
{
  "items": [
    {
      "id": "uuid",
      "itemType": "moment",
      "itemDate": "2024-11-27T10:30:00Z",
      "situation": "Had a difficult conversation",
      "thoughts": "Felt undervalued",
      "physicalSymptoms": "Tight chest",
      "behavior": "Became defensive",
      "consequences": "Meeting ended awkwardly",
      "valuesReflection": "Violated calm communication",
      "intensity": 7,
      "dateCreated": "2024-11-27T10:45:00Z"
    },
    {
      "id": "uuid",
      "itemType": "think",
      "itemDate": "2024-11-26T15:20:00Z",
      "category": "reflection",
      "content": "Need to practice mindful breathing",
      "dateCreated": "2024-11-26T15:20:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "rowsPerPage": 1000
}
```

### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 401 | Unauthenticated | Missing or invalid bearer token |
| 400 | Bad Request | Invalid date format, start_date > end_date, or invalid pagination |
| 500 | Internal | Database or encryption errors |

## Implementation Checklist

- [ ] Add SQL migration (Version 1.07) to migrate.sql
- [ ] Create `business/domain/vexportbus/model.go`
- [ ] Create `business/domain/vexportbus/filter.go`
- [ ] Create `business/domain/vexportbus/order.go` (NEW)
- [ ] Create `business/domain/vexportbus/vexportbus.go`
- [ ] Create `business/domain/vexportbus/stores/vexportdb/model.go`
- [ ] Create `business/domain/vexportbus/stores/vexportdb/filter.go` (NEW)
- [ ] Create `business/domain/vexportbus/stores/vexportdb/order.go` (NEW)
- [ ] Create `business/domain/vexportbus/stores/vexportdb/vexportdb.go`
- [ ] Create `app/domain/vexportapp/model.go`
- [ ] Create `app/domain/vexportapp/vexportapp.go`
- [ ] Create `app/domain/vexportapp/route.go`
- [ ] Update `app/sdk/mux/mux.go` BusConfig
- [ ] Update `api/services/partners/all/all.go`
- [ ] Update `api/services/partners/main.go`
- [ ] Test with Bruno/curl
- [ ] Deploy to production
