# Moments Feature - Backend Implementation Guide

**Feature Name:** Emotional Moments Tracking ("Momentos")
**API Endpoint:** `/v1/moments`
**Database Table:** `moments`
**Version:** 1.03 (Migration)

---

## Table of Contents
1. [Overview](#overview)
2. [Business Types Layer](#business-types-layer)
3. [Database Migration](#database-migration)
4. [Business Layer Implementation](#business-layer-implementation)
5. [Database Layer Implementation](#database-layer-implementation)
6. [Application Layer Implementation](#application-layer-implementation)
7. [Integration & Route Registration](#integration--route-registration)
8. [Testing](#testing)
9. [Deployment](#deployment)

---

## Overview

This feature implements psychological self-observation tracking based on the "Registro Funcional Diario" therapeutic tool. Users can record difficult emotional moments with 8 fields:

| Field | Type | Description |
|-------|------|-------------|
| `moment_date` | TIMESTAMP | When the moment occurred |
| `situation` | TEXT | Where you were, what happened before |
| `thoughts` | TEXT | Thoughts that appeared in your mind |
| `physical_symptoms` | TEXT | Physical symptoms or emotions felt |
| `behavior` | TEXT | What you did in response (conduct) |
| `consequences` | TEXT | Immediate consequences |
| `values_reflection` | TEXT | Did you avoid or approach something important |
| `intensity` | INTEGER | Distress intensity (0-10 scale) |

**User Isolation:** All queries filtered by `user_id` from JWT token.
**Authentication:** Required on all endpoints.
**Privacy:** Personal use, no sharing features.

---

## Business Types Layer

**IMPORTANT**: Following the codebase pattern, we create strong types with validation for all domain values. Never use primitive types directly in business models.

### File: `business/types/intensity/intensity.go`

```go
// Package intensity represents a validated intensity value (0-10 scale) in the system.
package intensity

import (
	"fmt"
)

// Intensity represents a validated intensity value on a 0-10 scale.
// This is commonly used for measuring distress, emotion, or pain intensity.
type Intensity struct {
	value int
}

// Value returns the int value of the intensity.
func (i Intensity) Value() int {
	return i.value
}

// String returns the string representation of the intensity.
func (i Intensity) String() string {
	return fmt.Sprintf("%d", i.value)
}

// Equal provides support for the go-cmp package and testing.
func (i Intensity) Equal(i2 Intensity) bool {
	return i.value == i2.value
}

// MarshalText provides support for logging and any marshal needs.
func (i Intensity) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// =============================================================================

// Parse validates the int value and returns an Intensity if the value complies
// with the rules for intensity (0-10 scale).
func Parse(value int) (Intensity, error) {
	if value < 0 || value > 10 {
		return Intensity{}, fmt.Errorf("intensity must be between 0 and 10, got %d", value)
	}

	return Intensity{value}, nil
}

// MustParse parses the int value and returns an Intensity if the value
// complies with the rules for intensity. If an error occurs the function panics.
// Use this only in tests or when you're certain the value is valid.
func MustParse(value int) Intensity {
	intensity, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return intensity
}
```

**Why This Pattern:**
- ✅ Validation happens once at parse time
- ✅ Type system enforces valid data throughout the application
- ✅ Impossible to construct invalid Intensity values
- ✅ Follows existing patterns: `content.Content`, `name.Name`, `quantity.Quantity`

---

## Database Migration

### File: `business/sdk/migrate/sql/migrate.sql`

Add this migration **after** Version 1.02 (thinks table):

```sql
-- Version: 1.03
-- Description: Create table moments for emotional tracking

CREATE TABLE moments (
    moment_id         UUID        NOT NULL,
    user_id           UUID        NOT NULL,
    moment_date       TIMESTAMP   NOT NULL,
    situation         TEXT        NOT NULL,
    thoughts          TEXT        NOT NULL,
    physical_symptoms TEXT        NOT NULL,
    behavior          TEXT        NOT NULL,
    consequences      TEXT        NOT NULL,
    values_reflection TEXT        NOT NULL,
    intensity         INTEGER     NOT NULL CHECK (intensity >= 0 AND intensity <= 10),
    date_created      TIMESTAMP   NOT NULL,
    date_updated      TIMESTAMP   NOT NULL,

    PRIMARY KEY (moment_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX moments_user_id_idx ON moments(user_id);
CREATE INDEX moments_user_date_idx ON moments(user_id, moment_date DESC);
CREATE INDEX moments_date_created_idx ON moments(date_created DESC);
CREATE INDEX moments_intensity_idx ON moments(intensity);

COMMENT ON TABLE moments IS 'Tracks emotional/difficult moments for psychological self-observation';
COMMENT ON COLUMN moments.moment_date IS 'When the observed moment actually occurred (user can backdate)';
COMMENT ON COLUMN moments.intensity IS 'Distress intensity on 0-10 scale';
```

**Migration Process:**
- Darwin migration tool runs automatically on service startup
- Idempotent: safe to run multiple times
- Transactional: either succeeds completely or rolls back

---

## Business Layer Implementation

### 1. File: `business/domain/momentbus/model.go`

```go
package momentbus

import (
	"errors"
	"time"

	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound    = errors.New("moment not found")
	ErrFutureDate  = errors.New("moment_date cannot be in the future")
	ErrUniqueEntry = errors.New("moment entry already exists")
)

// Moment represents a tracked emotional/difficult moment in the system.
type Moment struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	MomentDate       time.Time
	Situation        content.Content
	Thoughts         content.Content
	PhysicalSymptoms content.Content
	Behavior         content.Content
	Consequences     content.Content
	ValuesReflection content.Content
	Intensity        intensity.Intensity
	DateCreated      time.Time
	DateUpdated      time.Time
}

// NewMoment contains information needed to create a new moment.
type NewMoment struct {
	UserID           uuid.UUID
	MomentDate       time.Time
	Situation        content.Content
	Thoughts         content.Content
	PhysicalSymptoms content.Content
	Behavior         content.Content
	Consequences     content.Content
	ValuesReflection content.Content
	Intensity        intensity.Intensity
}

// UpdateMoment contains information needed to update a moment.
type UpdateMoment struct {
	MomentDate       *time.Time
	Situation        *content.Content
	Thoughts         *content.Content
	PhysicalSymptoms *content.Content
	Behavior         *content.Content
	Consequences     *content.Content
	ValuesReflection *content.Content
	Intensity        *intensity.Intensity
}
```

### 2. File: `business/domain/momentbus/momentbus.go`

```go
package momentbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, moment Moment) error
	Update(ctx context.Context, moment Moment) error
	Delete(ctx context.Context, moment Moment) error
	Query(ctx context.Context, filter QueryFilter) ([]Moment, error)
	QueryByID(ctx context.Context, momentID uuid.UUID) (Moment, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for moment access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a moment business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// Create adds a new moment to the system.
func (b *Business) Create(ctx context.Context, nm NewMoment) (Moment, error) {
	// Validate moment_date is not in future
	now := time.Now()
	if nm.MomentDate.After(now) {
		return Moment{}, ErrFutureDate
	}

	moment := Moment{
		ID:               uuid.New(),
		UserID:           nm.UserID,
		MomentDate:       nm.MomentDate,
		Situation:        nm.Situation,
		Thoughts:         nm.Thoughts,
		PhysicalSymptoms: nm.PhysicalSymptoms,
		Behavior:         nm.Behavior,
		Consequences:     nm.Consequences,
		ValuesReflection: nm.ValuesReflection,
		Intensity:        nm.Intensity,
		DateCreated:      now,
		DateUpdated:      now,
	}

	if err := b.storer.Create(ctx, moment); err != nil {
		return Moment{}, fmt.Errorf("create: %w", err)
	}

	return moment, nil
}

// Update modifies information about a moment.
func (b *Business) Update(ctx context.Context, moment Moment, um UpdateMoment) (Moment, error) {
	if um.MomentDate != nil {
		// Validate not in future
		if um.MomentDate.After(time.Now()) {
			return Moment{}, ErrFutureDate
		}
		moment.MomentDate = *um.MomentDate
	}

	if um.Situation != nil {
		moment.Situation = *um.Situation
	}

	if um.Thoughts != nil {
		moment.Thoughts = *um.Thoughts
	}

	if um.PhysicalSymptoms != nil {
		moment.PhysicalSymptoms = *um.PhysicalSymptoms
	}

	if um.Behavior != nil {
		moment.Behavior = *um.Behavior
	}

	if um.Consequences != nil {
		moment.Consequences = *um.Consequences
	}

	if um.ValuesReflection != nil {
		moment.ValuesReflection = *um.ValuesReflection
	}

	if um.Intensity != nil {
		moment.Intensity = *um.Intensity
	}

	moment.DateUpdated = time.Now()

	if err := b.storer.Update(ctx, moment); err != nil {
		return Moment{}, fmt.Errorf("update: %w", err)
	}

	return moment, nil
}

// Delete removes a moment from the system.
func (b *Business) Delete(ctx context.Context, moment Moment) error {
	if err := b.storer.Delete(ctx, moment); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing moments from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter) ([]Moment, error) {
	moments, err := b.storer.Query(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return moments, nil
}

// QueryByID finds the moment by the specified ID.
func (b *Business) QueryByID(ctx context.Context, momentID uuid.UUID) (Moment, error) {
	moment, err := b.storer.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Moment{}, ErrNotFound
		}
		return Moment{}, fmt.Errorf("query: %w", err)
	}

	return moment, nil
}

// Count returns the total number of moments that match the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
```

### 3. File: `business/domain/momentbus/order.go`

```go
package momentbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy("moment_date", order.DESC)

// Set of fields that the results can be ordered by.
const (
	OrderByMomentID    = "moment_id"
	OrderByMomentDate  = "moment_date"
	OrderByIntensity   = "intensity"
	OrderByDateCreated = "date_created"
	OrderByDateUpdated = "date_updated"
)

var orderByFields = map[string]string{
	OrderByMomentID:    "moment_id",
	OrderByMomentDate:  "moment_date",
	OrderByIntensity:   "intensity",
	OrderByDateCreated: "date_created",
	OrderByDateUpdated: "date_updated",
}

// OrderBy validates and returns the order by field and direction.
func OrderBy(orderBy order.By) (order.By, error) {
	if err := order.Validate(orderBy, orderByFields); err != nil {
		return order.By{}, err
	}

	return orderBy, nil
}
```

### 4. File: `business/domain/momentbus/filter.go`

```go
package momentbus

import (
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID      *uuid.UUID
	UserID  *uuid.UUID
	Page    page.Page
	OrderBy order.By
}
```

---

## Database Layer Implementation

### 1. File: `business/domain/momentbus/stores/momentdb/model.go`

```go
package momentdb

import (
	"fmt"
	"time"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/types/content"
	"github.com/google/uuid"
)

type moment struct {
	ID               uuid.UUID `db:"moment_id"`
	UserID           uuid.UUID `db:"user_id"`
	MomentDate       time.Time `db:"moment_date"`
	Situation        string    `db:"situation"`
	Thoughts         string    `db:"thoughts"`
	PhysicalSymptoms string    `db:"physical_symptoms"`
	Behavior         string    `db:"behavior"`
	Consequences     string    `db:"consequences"`
	ValuesReflection string    `db:"values_reflection"`
	Intensity        int       `db:"intensity"`
	DateCreated      time.Time `db:"date_created"`
	DateUpdated      time.Time `db:"date_updated"`
}

func toDBMoment(bus momentbus.Moment) moment {
	return moment{
		ID:               bus.ID,
		UserID:           bus.UserID,
		MomentDate:       bus.MomentDate.UTC(),
		Situation:        bus.Situation.String(),
		Thoughts:         bus.Thoughts.String(),
		PhysicalSymptoms: bus.PhysicalSymptoms.String(),
		Behavior:         bus.Behavior.String(),
		Consequences:     bus.Consequences.String(),
		ValuesReflection: bus.ValuesReflection.String(),
		Intensity:        bus.Intensity.Value(),
		DateCreated:      bus.DateCreated.UTC(),
		DateUpdated:      bus.DateUpdated.UTC(),
	}
}

func toBusMoment(db moment) (momentbus.Moment, error) {
	situation, err := content.Parse(db.Situation)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse situation: %w", err)
	}

	thoughts, err := content.Parse(db.Thoughts)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse thoughts: %w", err)
	}

	physicalSymptoms, err := content.Parse(db.PhysicalSymptoms)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse physical_symptoms: %w", err)
	}

	behavior, err := content.Parse(db.Behavior)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse behavior: %w", err)
	}

	consequences, err := content.Parse(db.Consequences)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse consequences: %w", err)
	}

	valuesReflection, err := content.Parse(db.ValuesReflection)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse values_reflection: %w", err)
	}

	intensity, err := intensity.Parse(db.Intensity)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse intensity: %w", err)
	}

	return momentbus.Moment{
		ID:               db.ID,
		UserID:           db.UserID,
		MomentDate:       db.MomentDate.In(time.Local),
		Situation:        situation,
		Thoughts:         thoughts,
		PhysicalSymptoms: physicalSymptoms,
		Behavior:         behavior,
		Consequences:     consequences,
		ValuesReflection: valuesReflection,
		Intensity:        intensity,
		DateCreated:      db.DateCreated.In(time.Local),
		DateUpdated:      db.DateUpdated.In(time.Local),
	}, nil
}

func toBusMoments(dbs []moment) ([]momentbus.Moment, error) {
	moments := make([]momentbus.Moment, len(dbs))

	for i, db := range dbs {
		var err error
		moments[i], err = toBusMoment(db)
		if err != nil {
			return nil, err
		}
	}

	return moments, nil
}
```

### 2. File: `business/domain/momentbus/stores/momentdb/momentdb.go`

```go
package momentdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for moment database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create adds a new moment to the database.
func (s *Store) Create(ctx context.Context, moment momentbus.Moment) error {
	const q = `
	INSERT INTO moments (
		moment_id, user_id, moment_date, situation, thoughts,
		physical_symptoms, behavior, consequences, values_reflection,
		intensity, date_created, date_updated
	) VALUES (
		:moment_id, :user_id, :moment_date, :situation, :thoughts,
		:physical_symptoms, :behavior, :consequences, :values_reflection,
		:intensity, :date_created, :date_updated
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMoment(moment)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", momentbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about a moment in the database.
func (s *Store) Update(ctx context.Context, moment momentbus.Moment) error {
	const q = `
	UPDATE moments SET
		moment_date = :moment_date,
		situation = :situation,
		thoughts = :thoughts,
		physical_symptoms = :physical_symptoms,
		behavior = :behavior,
		consequences = :consequences,
		values_reflection = :values_reflection,
		intensity = :intensity,
		date_updated = :date_updated
	WHERE
		moment_id = :moment_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMoment(moment)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", momentbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a moment from the database.
func (s *Store) Delete(ctx context.Context, moment momentbus.Moment) error {
	const q = `
	DELETE FROM moments
	WHERE moment_id = :moment_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMoment(moment)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing moments from the database.
func (s *Store) Query(ctx context.Context, filter momentbus.QueryFilter) ([]momentbus.Moment, error) {
	data := map[string]any{}

	const q = `
	SELECT
		moment_id, user_id, moment_date, situation, thoughts,
		physical_symptoms, behavior, consequences, values_reflection,
		intensity, date_created, date_updated
	FROM
		moments`

	buf := sqldb.NewBuf(q)

	if filter.ID != nil {
		data["moment_id"] = *filter.ID
		buf.Add("moment_id = :moment_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		buf.Add("user_id = :user_id")
	}

	orderByClause, err := orderByClause(filter.OrderBy)
	if err != nil {
		return nil, err
	}

	buf.Add(orderByClause)
	buf.Add(sqldb.Pagination(filter.Page))

	var dbMoms []moment
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbMoms); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusMoments(dbMoms)
}

// QueryByID retrieves a single moment by its ID.
func (s *Store) QueryByID(ctx context.Context, momentID uuid.UUID) (momentbus.Moment, error) {
	data := map[string]any{
		"moment_id": momentID,
	}

	const q = `
	SELECT
		moment_id, user_id, moment_date, situation, thoughts,
		physical_symptoms, behavior, consequences, values_reflection,
		intensity, date_created, date_updated
	FROM
		moments
	WHERE
		moment_id = :moment_id`

	var dbMom moment
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbMom); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return momentbus.Moment{}, fmt.Errorf("db: %w", momentbus.ErrNotFound)
		}
		return momentbus.Moment{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusMoment(dbMom)
}

// Count returns the total number of moments in the database.
func (s *Store) Count(ctx context.Context, filter momentbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*) AS count
	FROM
		moments`

	buf := sqldb.NewBuf(q)

	if filter.ID != nil {
		data["moment_id"] = *filter.ID
		buf.Add("moment_id = :moment_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		buf.Add("user_id = :user_id")
	}

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}
```

### 3. File: `business/domain/momentbus/stores/momentdb/order.go`

```go
package momentdb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	momentbus.OrderByMomentID:    "moment_id",
	momentbus.OrderByMomentDate:  "moment_date",
	momentbus.OrderByIntensity:   "intensity",
	momentbus.OrderByDateCreated: "date_created",
	momentbus.OrderByDateUpdated: "date_updated",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

---

## Application Layer Implementation

### 1. File: `app/domain/momentapp/model.go`

```go
package momentapp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/types/content"
)

// Moment represents an API moment.
type Moment struct {
	ID               string `json:"id"`
	MomentDate       string `json:"momentDate"`
	Situation        string `json:"situation"`
	Thoughts         string `json:"thoughts"`
	PhysicalSymptoms string `json:"physicalSymptoms"`
	Behavior         string `json:"behavior"`
	Consequences     string `json:"consequences"`
	ValuesReflection string `json:"valuesReflection"`
	Intensity        int    `json:"intensity"`
	DateCreated      string `json:"dateCreated"`
	DateUpdated      string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (m Moment) Encode() ([]byte, string, error) {
	data, err := json.Marshal(m)
	return data, "application/json", err
}

// NewMoment contains information needed to create a new moment.
type NewMoment struct {
	MomentDate       string `json:"momentDate" validate:"required"`
	Situation        string `json:"situation" validate:"required"`
	Thoughts         string `json:"thoughts" validate:"required"`
	PhysicalSymptoms string `json:"physicalSymptoms" validate:"required"`
	Behavior         string `json:"behavior" validate:"required"`
	Consequences     string `json:"consequences" validate:"required"`
	ValuesReflection string `json:"valuesReflection" validate:"required"`
	Intensity        int    `json:"intensity" validate:"required,min=0,max=10"`
}

// Decode implements the decoder interface.
func (nm *NewMoment) Decode(data []byte) error {
	return json.Unmarshal(data, nm)
}

// Validate checks the data in the model is considered clean.
func (nm NewMoment) Validate() error {
	if err := errs.Check(nm); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateMoment contains information needed to update a moment.
type UpdateMoment struct {
	MomentDate       *string `json:"momentDate"`
	Situation        *string `json:"situation"`
	Thoughts         *string `json:"thoughts"`
	PhysicalSymptoms *string `json:"physicalSymptoms"`
	Behavior         *string `json:"behavior"`
	Consequences     *string `json:"consequences"`
	ValuesReflection *string `json:"valuesReflection"`
	Intensity        *int    `json:"intensity" validate:"omitempty,min=0,max=10"`
}

// Decode implements the decoder interface.
func (um *UpdateMoment) Decode(data []byte) error {
	return json.Unmarshal(data, um)
}

// Validate checks the data in the model is considered clean.
func (um UpdateMoment) Validate() error {
	if err := errs.Check(um); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// =============================================================================

func toAppMoment(moment momentbus.Moment) Moment {
	return Moment{
		ID:               moment.ID.String(),
		MomentDate:       moment.MomentDate.Format(time.RFC3339),
		Situation:        moment.Situation.String(),
		Thoughts:         moment.Thoughts.String(),
		PhysicalSymptoms: moment.PhysicalSymptoms.String(),
		Behavior:         moment.Behavior.String(),
		Consequences:     moment.Consequences.String(),
		ValuesReflection: moment.ValuesReflection.String(),
		Intensity:        moment.Intensity.Value(),
		DateCreated:      moment.DateCreated.Format(time.RFC3339),
		DateUpdated:      moment.DateUpdated.Format(time.RFC3339),
	}
}

func toAppMoments(moments []momentbus.Moment) []Moment {
	app := make([]Moment, len(moments))
	for i, moment := range moments {
		app[i] = toAppMoment(moment)
	}
	return app
}

// =============================================================================

func toBusNewMoment(ctx context.Context, nm NewMoment) (momentbus.NewMoment, error) {
	var errors errs.FieldErrors

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		errors.Add("userID", err)
	}

	momentDate, err := time.Parse(time.RFC3339, nm.MomentDate)
	if err != nil {
		errors.Add("momentDate", fmt.Errorf("must be RFC3339 format"))
	}

	situation, err := content.Parse(nm.Situation)
	if err != nil {
		errors.Add("situation", err)
	}

	thoughts, err := content.Parse(nm.Thoughts)
	if err != nil {
		errors.Add("thoughts", err)
	}

	physicalSymptoms, err := content.Parse(nm.PhysicalSymptoms)
	if err != nil {
		errors.Add("physicalSymptoms", err)
	}

	behavior, err := content.Parse(nm.Behavior)
	if err != nil {
		errors.Add("behavior", err)
	}

	consequences, err := content.Parse(nm.Consequences)
	if err != nil {
		errors.Add("consequences", err)
	}

	valuesReflection, err := content.Parse(nm.ValuesReflection)
	if err != nil {
		errors.Add("valuesReflection", err)
	}

	intensity, err := intensity.Parse(nm.Intensity)
	if err != nil {
		errors.Add("intensity", err)
	}

	if len(errors) > 0 {
		return momentbus.NewMoment{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return momentbus.NewMoment{
		UserID:           userID,
		MomentDate:       momentDate,
		Situation:        situation,
		Thoughts:         thoughts,
		PhysicalSymptoms: physicalSymptoms,
		Behavior:         behavior,
		Consequences:     consequences,
		ValuesReflection: valuesReflection,
		Intensity:        intensity,
	}, nil
}

func toBusUpdateMoment(ctx context.Context, um UpdateMoment) (momentbus.UpdateMoment, error) {
	var errors errs.FieldErrors
	var bus momentbus.UpdateMoment

	if um.MomentDate != nil {
		momentDate, err := time.Parse(time.RFC3339, *um.MomentDate)
		if err != nil {
			errors.Add("momentDate", fmt.Errorf("must be RFC3339 format"))
		} else {
			bus.MomentDate = &momentDate
		}
	}

	if um.Situation != nil {
		situation, err := content.Parse(*um.Situation)
		if err != nil {
			errors.Add("situation", err)
		} else {
			bus.Situation = &situation
		}
	}

	if um.Thoughts != nil {
		thoughts, err := content.Parse(*um.Thoughts)
		if err != nil {
			errors.Add("thoughts", err)
		} else {
			bus.Thoughts = &thoughts
		}
	}

	if um.PhysicalSymptoms != nil {
		physicalSymptoms, err := content.Parse(*um.PhysicalSymptoms)
		if err != nil {
			errors.Add("physicalSymptoms", err)
		} else {
			bus.PhysicalSymptoms = &physicalSymptoms
		}
	}

	if um.Behavior != nil {
		behavior, err := content.Parse(*um.Behavior)
		if err != nil {
			errors.Add("behavior", err)
		} else {
			bus.Behavior = &behavior
		}
	}

	if um.Consequences != nil {
		consequences, err := content.Parse(*um.Consequences)
		if err != nil {
			errors.Add("consequences", err)
		} else {
			bus.Consequences = &consequences
		}
	}

	if um.ValuesReflection != nil {
		valuesReflection, err := content.Parse(*um.ValuesReflection)
		if err != nil {
			errors.Add("valuesReflection", err)
		} else {
			bus.ValuesReflection = &valuesReflection
		}
	}

	if um.Intensity != nil {
		intensity, err := intensity.Parse(*um.Intensity)
		if err != nil {
			errors.Add("intensity", err)
		} else {
			bus.Intensity = &intensity
		}
	}

	if len(errors) > 0 {
		return momentbus.UpdateMoment{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}
```

### 2. File: `app/domain/momentapp/momentapp.go`

```go
package momentapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
	"github.com/google/uuid"
)

type api struct {
	momentBus *momentbus.Business
}

func newAPI(momentBus *momentbus.Business) *api {
	return &api{
		momentBus: momentBus,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app NewMoment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nm, err := toBusNewMoment(ctx, app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err := api.momentBus.Create(ctx, nm)
	if err != nil {
		if errors.Is(err, momentbus.ErrFutureDate) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: %s", err)
	}

	return toAppMoment(moment)
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app UpdateMoment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	momentID, err := uuid.Parse(web.Param(r, "moment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err := api.momentBus.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, momentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "getuserid: %s", err)
	}

	if moment.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	um, err := toBusUpdateMoment(ctx, app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err = api.momentBus.Update(ctx, moment, um)
	if err != nil {
		if errors.Is(err, momentbus.ErrFutureDate) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppMoment(moment)
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	momentID, err := uuid.Parse(web.Param(r, "moment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err := api.momentBus.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, momentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "getuserid: %s", err)
	}

	if moment.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	if err := api.momentBus.Delete(ctx, moment); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "getuserid: %s", err)
	}

	filter := momentbus.QueryFilter{
		UserID:  &userID,
		Page:    qp.page,
		OrderBy: qp.orderBy,
	}

	moments, err := api.momentBus.Query(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := api.momentBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppMoments(moments), total, qp.page)
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	momentID, err := uuid.Parse(web.Param(r, "moment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err := api.momentBus.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, momentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "getuserid: %s", err)
	}

	if moment.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	return toAppMoment(moment)
}
```

### 3. File: `app/domain/momentapp/order.go`

```go
package momentapp

import (
	"errors"
	"net/http"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
)

var orderByFields = map[string]string{
	"moment_id":    momentbus.OrderByMomentID,
	"moment_date":  momentbus.OrderByMomentDate,
	"intensity":    momentbus.OrderByIntensity,
	"date_created": momentbus.OrderByDateCreated,
	"date_updated": momentbus.OrderByDateUpdated,
}

type queryParams struct {
	page    page.Page
	orderBy order.By
}

func parseQueryParams(r *http.Request) (queryParams, error) {
	values := r.URL.Query()

	filter := queryParams{
		page:    page.MustParse(values.Get("page"), values.Get("rows")),
		orderBy: order.NewBy("moment_date", order.DESC),
	}

	if orderBy := values.Get("orderBy"); orderBy != "" {
		orderByField, exists := orderByFields[orderBy]
		if !exists {
			return queryParams{}, errors.New("invalid orderBy field")
		}

		direction := order.ASC
		if values.Get("orderDirection") == "desc" {
			direction = order.DESC
		}

		filter.orderBy = order.NewBy(orderByField, direction)
	}

	if err := filter.page.Validate(); err != nil {
		return queryParams{}, err
	}

	return filter, nil
}
```

### 4. File: `app/domain/momentapp/route.go`

```go
package momentapp

import (
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	MomentBus *momentbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg.MomentBus)

	app.HandlerFunc("POST /"+version+"/moments", api.create, mid.Authenticate())
	app.HandlerFunc("GET /"+version+"/moments", api.query, mid.Authenticate())
	app.HandlerFunc("GET /"+version+"/moments/{moment_id}", api.queryByID, mid.Authenticate())
	app.HandlerFunc("PUT /"+version+"/moments/{moment_id}", api.update, mid.Authenticate())
	app.HandlerFunc("DELETE /"+version+"/moments/{moment_id}", api.delete, mid.Authenticate())
}
```

---

## Integration & Route Registration

### 1. Modify: `app/sdk/mux/mux.go`

Add `MomentBus` to the `BusConfig` struct:

```go
type BusConfig struct {
	UserBus   *userbus.Business
	ThinkBus  *thinkbus.Business
	MomentBus *momentbus.Business  // ADD THIS LINE
}
```

### 2. Modify: `api/services/partners/all/all.go`

Add moment routes:

```go
package all

import (
	"github.com/francowini/rafiki/api/services/api/debug"
	"github.com/francowini/rafiki/api/services/api/health"
	"github.com/francowini/rafiki/app/domain/authapp"
	"github.com/francowini/rafiki/app/domain/momentapp"  // ADD THIS
	"github.com/francowini/rafiki/app/domain/thinkapp"
	"github.com/francowini/rafiki/app/sdk/mux"
	"github.com/francowini/rafiki/foundation/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() mux.RouteAdder {
	return mux.RouteAdderFunc(func(app *web.App, cfg mux.Config) {
		health.Routes(app, health.Config{
			Build: cfg.Build,
			Log:   cfg.Log,
			DB:    cfg.DB,
		})

		debug.Routes(app)

		authapp.Routes(app, authapp.Config{
			UserBus: cfg.BusCfg.UserBus,
			Auth:    cfg.Auth,
		})

		thinkapp.Routes(app, thinkapp.Config{
			ThinkBus: cfg.BusCfg.ThinkBus,
		})

		// ADD THIS BLOCK
		momentapp.Routes(app, momentapp.Config{
			MomentBus: cfg.BusCfg.MomentBus,
		})
	})
}
```

### 3. Modify: `api/services/partners/main.go`

Initialize MomentBus in the main function. Add this code after ThinkBus initialization (around line 135):

```go
// =========================================================================
// Business Domain

// ... existing UserBus and ThinkBus initialization ...

// Construct the business domain package for moments.
log.Info(ctx, "startup", "status", "initializing moment business")

momentBus := momentbus.NewBusiness(log, momentdb.NewStore(log, db))

// ... rest of the code ...

// In the mux.WebAPI call (around line 180), add MomentBus:
busCfg := mux.BusConfig{
	UserBus:   userBus,
	ThinkBus:  thinkBus,
	MomentBus: momentBus,  // ADD THIS LINE
}
```

**Also add the import:**

```go
import (
	// ... existing imports ...
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/domain/momentbus/stores/momentdb"
)
```

---

## Testing

### Local Testing Commands

```bash
# 1. Start services
make down
make up

# 2. Check migration logs
make logs | grep "database migrations completed"

# 3. Verify table exists
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "\d moments"

# 4. Get auth token (login first)
curl -X POST http://localhost:3000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@rafiki.lat",
    "password": "gophers"
  }'

# Save the token from response

# 5. Create a moment
curl -X POST http://localhost:3000/v1/moments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "momentDate": "2025-11-15T14:30:00Z",
    "situation": "Estaba en el trabajo, en una reunión con mi jefe",
    "thoughts": "No soy suficientemente bueno, van a despedirme",
    "physicalSymptoms": "Opresión en el pecho, corazón acelerado",
    "behavior": "Evité el contacto visual, hablé en voz baja",
    "consequences": "Sentí alivio temporal pero más ansiedad después",
    "valuesReflection": "Evité una conversación difícil que era importante",
    "intensity": 8
  }'

# 6. Query moments
curl -X GET "http://localhost:3000/v1/moments?page=1&rows=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 7. Get specific moment
curl -X GET http://localhost:3000/v1/moments/{moment_id} \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 8. Update moment
curl -X PUT http://localhost:3000/v1/moments/{moment_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "intensity": 6,
    "valuesReflection": "Después reflexioné que pude manejar mejor la situación"
  }'

# 9. Delete moment
curl -X DELETE http://localhost:3000/v1/moments/{moment_id} \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Validation Tests

```bash
# Test intensity validation (should fail)
curl -X POST http://localhost:3000/v1/moments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "momentDate": "2025-11-15T14:30:00Z",
    "situation": "Test",
    "thoughts": "Test",
    "physicalSymptoms": "Test",
    "behavior": "Test",
    "consequences": "Test",
    "valuesReflection": "Test",
    "intensity": 15
  }'
# Expected: 400 Bad Request - intensity must be 0-10

# Test future date (should fail)
curl -X POST http://localhost:3000/v1/moments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "momentDate": "2026-11-15T14:30:00Z",
    "situation": "Test",
    "thoughts": "Test",
    "physicalSymptoms": "Test",
    "behavior": "Test",
    "consequences": "Test",
    "valuesReflection": "Test",
    "intensity": 5
  }'
# Expected: 400 Bad Request - moment_date cannot be in future

# Test authentication (should fail)
curl -X GET http://localhost:3000/v1/moments
# Expected: 401 Unauthorized
```

---

## Deployment

### Production Deployment Commands

```bash
# 1. Commit all changes
git add .
git commit -m "feat: add moments tracking feature

- Add database migration (version 1.03)
- Add momentbus domain layer
- Add momentapp HTTP handlers
- Add authentication and user isolation
- Add validation for intensity and future dates"

git push origin main

# 2. Deploy to production (one command)
make deploy

# 3. Monitor deployment
make deploy-logs

# 4. Verify migration succeeded
ssh root@178.156.170.37
cd /opt/rafiki
docker compose logs partner-service | grep "database migrations completed"

# 5. Verify table exists
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "\d moments"

# 6. Test health check
curl https://api.rafiki.lat/v1/readiness

# 7. Test API endpoint (with real auth token)
curl https://api.rafiki.lat/v1/moments \
  -H "Authorization: Bearer YOUR_PROD_TOKEN"
```

### Rollback Procedure

```bash
# If deployment fails:
ssh root@178.156.170.37
cd /opt/rafiki

# Option 1: Revert code
git log  # Find previous commit
git reset --hard <previous-commit-hash>
./devops/deploy.sh

# Option 2: Manual table drop (if needed)
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
# DROP TABLE moments CASCADE;
```

---

## Implementation Checklist

### Phase 0: Business Types Layer
- [ ] Create `business/types/intensity/intensity.go`

### Phase 1: Business & Database Layers
- [ ] Create `business/domain/momentbus/model.go`
- [ ] Create `business/domain/momentbus/momentbus.go`
- [ ] Create `business/domain/momentbus/order.go`
- [ ] Create `business/domain/momentbus/filter.go`
- [ ] Create `business/domain/momentbus/stores/momentdb/model.go`
- [ ] Create `business/domain/momentbus/stores/momentdb/momentdb.go`
- [ ] Create `business/domain/momentbus/stores/momentdb/order.go`

### Phase 2: Application Layer
- [ ] Create `app/domain/momentapp/model.go`
- [ ] Create `app/domain/momentapp/momentapp.go`
- [ ] Create `app/domain/momentapp/order.go`
- [ ] Create `app/domain/momentapp/route.go`

### Phase 3: Database Migration
- [ ] Add migration to `business/sdk/migrate/sql/migrate.sql`

### Phase 4: Integration
- [ ] Modify `app/sdk/mux/mux.go` - add MomentBus to BusConfig
- [ ] Modify `api/services/partners/all/all.go` - add momentapp routes
- [ ] Modify `api/services/partners/main.go` - initialize MomentBus

### Phase 5: Testing
- [ ] Test migration locally (`make up`)
- [ ] Test CRUD operations with curl
- [ ] Test validation (intensity, future date)
- [ ] Test authentication
- [ ] Test user isolation

### Phase 6: Deployment
- [ ] Commit and push to main
- [ ] Deploy to production (`make deploy`)
- [ ] Verify migration and endpoints
- [ ] Document rollback procedure

---

## Summary

This implementation provides:

- ✅ **Complete CRUD API** for emotional moments tracking
- ✅ **User isolation** - users can only access their own moments
- ✅ **Authentication required** on all endpoints
- ✅ **Data validation** - intensity 0-10, no future dates, required fields
- ✅ **Pagination support** for listing moments
- ✅ **Flexible ordering** by date, intensity, or creation time
- ✅ **Production-ready** with proper error handling and logging
- ✅ **Zero-downtime deployment** with idempotent migrations

**Total Files Created:** 12 new files (1 business type + 7 business layer + 4 app layer)
**Total Files Modified:** 3 existing files
**Database Tables Added:** 1 (moments)
**API Endpoints:** 5 (POST, GET, GET/:id, PUT/:id, DELETE/:id)
**Business Types Created:** 1 (`intensity.Intensity` for validated 0-10 scale values)
