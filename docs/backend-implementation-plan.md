# Backend Implementation Plan - Values Feature

**Project:** Rafiki Habits Tracker
**Feature:** Personal Values Tracking with Facet Categorization
**Date:** November 22, 2025
**Status:** Ready for Implementation

---

## Overview

This document provides complete, copy-paste ready code for implementing the values feature backend following the existing codebase patterns discovered in the Rafiki project.

---

## Phase 1: Database Migration

### File: `business/sdk/migrate/sql/migrate.sql`

Add the following migration at the end of the file:

```sql
-- Version: 1.04
-- Description: Create facet_type ENUM and values table for personal values tracking

-- Create ENUM type for life facets
CREATE TYPE facet_type AS ENUM (
    'health',
    'relationships',
    'career',
    'personal_growth',
    'family',
    'creativity',
    'community',
    'spirituality'
);

-- Create values table
CREATE TABLE values (
    value_id       UUID        NOT NULL,
    user_id        UUID        NOT NULL,
    content        TEXT        NOT NULL,
    facet          facet_type  NOT NULL,
    display_order  INTEGER     NOT NULL,
    date_created   TIMESTAMP   NOT NULL,
    date_updated   TIMESTAMP   NOT NULL,

    PRIMARY KEY (value_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX values_user_id_idx ON values(user_id);
CREATE INDEX values_user_order_idx ON values(user_id, display_order ASC);
CREATE INDEX values_facet_idx ON values(facet);

-- Unique constraint to prevent duplicate display_order per user
CREATE UNIQUE INDEX values_user_order_unique_idx ON values(user_id, display_order);

-- Documentation
COMMENT ON TABLE values IS 'Stores user core personal values with life facet categorization (max 10 per user)';
COMMENT ON COLUMN values.content IS 'Encrypted value statement (3-200 characters)';
COMMENT ON COLUMN values.facet IS 'Life domain categorization based on ACT therapy';
COMMENT ON COLUMN values.display_order IS 'User-controlled priority ranking (1=highest)';
```

---

## Phase 2: Business Types

### File: `business/types/facet/facet.go` (NEW)

```go
// Package facet provides a type for life facet categorization.
package facet

import (
	"fmt"
)

// Predefined facet values
var (
	Health          = newFacet("health")
	Relationships   = newFacet("relationships")
	Career          = newFacet("career")
	PersonalGrowth  = newFacet("personal_growth")
	Family          = newFacet("family")
	Creativity      = newFacet("creativity")
	Community       = newFacet("community")
	Spirituality    = newFacet("spirituality")
)

var facets = make(map[string]Facet)

// Facet represents a validated life domain category.
type Facet struct {
	value string
}

func newFacet(facet string) Facet {
	f := Facet{value: facet}
	facets[facet] = f
	return f
}

// String returns the string value of the facet.
func (f Facet) String() string {
	return f.value
}

// Equal provides support for the go-cmp package and testing.
func (f Facet) Equal(f2 Facet) bool {
	return f.value == f2.value
}

// MarshalText provides support for logging and any marshal needs.
func (f Facet) MarshalText() ([]byte, error) {
	return []byte(f.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (f *Facet) UnmarshalText(data []byte) error {
	facet, err := Parse(string(data))
	if err != nil {
		return err
	}

	*f = facet
	return nil
}

// Parse validates and returns a Facet.
func Parse(value string) (Facet, error) {
	facet, exists := facets[value]
	if !exists {
		return Facet{}, fmt.Errorf("invalid facet %q", value)
	}
	return facet, nil
}

// MustParse parses the facet string or panics on error. Use in tests only.
func MustParse(value string) Facet {
	facet, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return facet
}

// All returns all valid facet values.
func All() []Facet {
	return []Facet{
		Health,
		Relationships,
		Career,
		PersonalGrowth,
		Family,
		Creativity,
		Community,
		Spirituality,
	}
}
```

### File: `business/types/valuecontent/valuecontent.go` (NEW)

```go
// Package valuecontent provides validation for value content strings.
package valuecontent

import (
	"errors"
	"strings"
)

// ValueContent represents validated value content (3-200 characters).
type ValueContent struct {
	value string
}

// Validation errors
var (
	ErrEmpty    = errors.New("value content cannot be empty")
	ErrTooShort = errors.New("value content must be at least 3 characters")
	ErrTooLong  = errors.New("value content must be at most 200 characters")
)

// String returns the string value of the content.
func (vc ValueContent) String() string {
	return vc.value
}

// Equal provides support for the go-cmp package and testing.
func (vc ValueContent) Equal(vc2 ValueContent) bool {
	return vc.value == vc2.value
}

// MarshalText provides support for logging and any marshal needs.
func (vc ValueContent) MarshalText() ([]byte, error) {
	return []byte(vc.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (vc *ValueContent) UnmarshalText(data []byte) error {
	content, err := Parse(string(data))
	if err != nil {
		return err
	}

	*vc = content
	return nil
}

// Parse validates and creates a ValueContent.
func Parse(value string) (ValueContent, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return ValueContent{}, ErrEmpty
	}

	if len(value) < 3 {
		return ValueContent{}, ErrTooShort
	}

	if len(value) > 200 {
		return ValueContent{}, ErrTooLong
	}

	return ValueContent{value: value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) ValueContent {
	content, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return content
}
```

---

## Phase 3: Domain Layer (Business Logic)

### File: `business/domain/valuebus/model.go` (NEW)

```go
// Package valuebus provides business logic for personal values management.
package valuebus

import (
	"errors"
	"time"

	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/business/types/valuecontent"
	"github.com/google/uuid"
)

// Domain errors
var (
	ErrNotFound      = errors.New("value not found")
	ErrMaxValues     = errors.New("maximum 10 values allowed per user")
	ErrDuplicateOrder = errors.New("display order already exists for this user")
)

// Value represents a personal value with facet categorization.
type Value struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Content      valuecontent.ValueContent
	Facet        facet.Facet
	DisplayOrder int
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewValue contains information needed to create a new value.
type NewValue struct {
	UserID       uuid.UUID
	Content      valuecontent.ValueContent
	Facet        facet.Facet
	DisplayOrder int
}

// UpdateValue contains information needed to update a value.
// All fields are optional (use pointers for partial updates).
type UpdateValue struct {
	Content      *valuecontent.ValueContent
	Facet        *facet.Facet
	DisplayOrder *int
}
```

### File: `business/domain/valuebus/filter.go` (NEW)

```go
package valuebus

import (
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/google/uuid"
)

// QueryFilter holds filtering options for querying values.
type QueryFilter struct {
	ID     *uuid.UUID
	UserID *uuid.UUID
	Facet  *facet.Facet
	Page   page.Page
	OrderBy order.By
}
```

### File: `business/domain/valuebus/order.go` (NEW)

```go
package valuebus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default ordering for values.
var DefaultOrderBy = order.NewBy(OrderByDisplayOrder, order.ASC)

// Order field constants (short codes for API).
const (
	OrderByValueID      = "value_id"
	OrderByDisplayOrder = "display_order"
	OrderByFacet        = "facet"
	OrderByDateCreated  = "date_created"
	OrderByDateUpdated  = "date_updated"
)
```

### File: `business/domain/valuebus/valuebus.go` (NEW)

```go
package valuebus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
)

// Storer interface defines required database operations.
type Storer interface {
	Create(ctx context.Context, value Value) error
	Update(ctx context.Context, value Value) error
	Delete(ctx context.Context, value Value) error
	Query(ctx context.Context, filter QueryFilter) ([]Value, error)
	QueryByID(ctx context.Context, valueID uuid.UUID) (Value, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages value operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a Business for value domain.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// Create adds a new value to the system.
func (b *Business) Create(ctx context.Context, nv NewValue) (Value, error) {
	// Enforce 10 values max per user
	filter := QueryFilter{
		UserID: &nv.UserID,
	}
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return Value{}, fmt.Errorf("count: %w", err)
	}

	if count >= 10 {
		return Value{}, ErrMaxValues
	}

	now := time.Now()

	value := Value{
		ID:           uuid.New(),
		UserID:       nv.UserID,
		Content:      nv.Content,
		Facet:        nv.Facet,
		DisplayOrder: nv.DisplayOrder,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := b.storer.Create(ctx, value); err != nil {
		return Value{}, fmt.Errorf("create: %w", err)
	}

	return value, nil
}

// Update modifies an existing value.
func (b *Business) Update(ctx context.Context, value Value, uv UpdateValue) (Value, error) {
	if uv.Content != nil {
		value.Content = *uv.Content
	}

	if uv.Facet != nil {
		value.Facet = *uv.Facet
	}

	if uv.DisplayOrder != nil {
		value.DisplayOrder = *uv.DisplayOrder
	}

	value.DateUpdated = time.Now()

	if err := b.storer.Update(ctx, value); err != nil {
		return Value{}, fmt.Errorf("update: %w", err)
	}

	return value, nil
}

// Delete removes a value from the system.
func (b *Business) Delete(ctx context.Context, value Value) error {
	if err := b.storer.Delete(ctx, value); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of values based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter) ([]Value, error) {
	values, err := b.storer.Query(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return values, nil
}

// QueryByID finds a value by its ID.
func (b *Business) QueryByID(ctx context.Context, valueID uuid.UUID) (Value, error) {
	value, err := b.storer.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Value{}, ErrNotFound
		}
		return Value{}, fmt.Errorf("query: %w", err)
	}

	return value, nil
}

// Count returns the total number of values matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
```

---

## Phase 4: Store Layer (Database)

### File: `business/domain/valuebus/stores/valuedb/model.go` (NEW)

```go
// Package valuedb provides database access for values.
package valuedb

import (
	"fmt"
	"time"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/business/types/valuecontent"
	"github.com/google/uuid"
)

// value represents the database model.
type value struct {
	ID           uuid.UUID `db:"value_id"`
	UserID       uuid.UUID `db:"user_id"`
	Content      string    `db:"content"`        // encrypted
	Facet        string    `db:"facet"`
	DisplayOrder int       `db:"display_order"`
	DateCreated  time.Time `db:"date_created"`
	DateUpdated  time.Time `db:"date_updated"`
}

// toDBValueEncrypted converts business model to DB model with encryption.
func toDBValueEncrypted(bus valuebus.Value, enc encrypt.Encryptor) (value, error) {
	// Encrypt content
	content, err := enc.Encrypt(bus.Content.String())
	if err != nil {
		return value{}, fmt.Errorf("encrypt content: %w", err)
	}

	return value{
		ID:           bus.ID,
		UserID:       bus.UserID,
		Content:      content,
		Facet:        bus.Facet.String(),
		DisplayOrder: bus.DisplayOrder,
		DateCreated:  bus.DateCreated.UTC(),
		DateUpdated:  bus.DateUpdated.UTC(),
	}, nil
}

// toBusValueDecrypted converts DB model to business model with decryption.
func toBusValueDecrypted(db value, enc encrypt.Encryptor) (valuebus.Value, error) {
	// Decrypt and parse content
	contentStr, err := enc.Decrypt(db.Content)
	if err != nil {
		return valuebus.Value{}, fmt.Errorf("decrypt content: %w", err)
	}

	content, err := valuecontent.Parse(contentStr)
	if err != nil {
		return valuebus.Value{}, fmt.Errorf("parse content: %w", err)
	}

	// Parse facet
	facetVal, err := facet.Parse(db.Facet)
	if err != nil {
		return valuebus.Value{}, fmt.Errorf("parse facet: %w", err)
	}

	return valuebus.Value{
		ID:           db.ID,
		UserID:       db.UserID,
		Content:      content,
		Facet:        facetVal,
		DisplayOrder: db.DisplayOrder,
		DateCreated:  db.DateCreated.In(time.Local),
		DateUpdated:  db.DateUpdated.In(time.Local),
	}, nil
}

// toBusValuesDecrypted converts a slice of DB models to business models.
func toBusValuesDecrypted(dbs []value, enc encrypt.Encryptor) ([]valuebus.Value, error) {
	values := make([]valuebus.Value, len(dbs))

	for i, db := range dbs {
		var err error
		values[i], err = toBusValueDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}
```

### File: `business/domain/valuebus/stores/valuedb/order.go` (NEW)

```go
package valuedb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	valuebus.OrderByValueID:      "value_id",
	valuebus.OrderByDisplayOrder: "display_order",
	valuebus.OrderByFacet:        "facet",
	valuebus.OrderByDateCreated:  "date_created",
	valuebus.OrderByDateUpdated:  "date_updated",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

### File: `business/domain/valuebus/stores/valuedb/valuedb.go` (NEW)

```go
package valuedb

import (
	"context"
	"errors"
	"fmt"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages database operations for values.
type Store struct {
	log       *logger.Logger
	db        sqlx.ExtContext
	encryptor encrypt.Encryptor
}

// NewStore constructs a Store for database access.
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {
	return &Store{
		log:       log,
		db:        db,
		encryptor: encryptor,
	}
}

// Create inserts a new value into the database.
func (s *Store) Create(ctx context.Context, value valuebus.Value) error {
	const q = `
	INSERT INTO values (
		value_id, user_id, content, facet, display_order,
		date_created, date_updated
	) VALUES (
		:value_id, :user_id, :content, :facet, :display_order,
		:date_created, :date_updated
	)`

	dbValue, err := toDBValueEncrypted(value, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbValue); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", valuebus.ErrDuplicateOrder)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing value in the database.
func (s *Store) Update(ctx context.Context, value valuebus.Value) error {
	const q = `
	UPDATE values SET
		content = :content,
		facet = :facet,
		display_order = :display_order,
		date_updated = :date_updated
	WHERE
		value_id = :value_id`

	dbValue, err := toDBValueEncrypted(value, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbValue); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", valuebus.ErrDuplicateOrder)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a value from the database.
func (s *Store) Delete(ctx context.Context, value valuebus.Value) error {
	const q = `
	DELETE FROM values
	WHERE value_id = :value_id`

	data := struct {
		ID string `db:"value_id"`
	}{
		ID: value.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves values based on filter criteria.
func (s *Store) Query(ctx context.Context, filter valuebus.QueryFilter) ([]valuebus.Value, error) {
	data := map[string]any{
		"offset":        (filter.Page.Number() - 1) * filter.Page.RowsPerPage(),
		"rows_per_page": filter.Page.RowsPerPage(),
	}

	whereClause := buildWhereClause(filter, data)
	orderByClause, err := orderByClause(filter.OrderBy)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
	SELECT
		value_id, user_id, content, facet, display_order,
		date_created, date_updated
	FROM values
	%s
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)

	var dbValues []value
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbValues); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusValuesDecrypted(dbValues, s.encryptor)
}

// QueryByID finds a value by its ID.
func (s *Store) QueryByID(ctx context.Context, valueID uuid.UUID) (valuebus.Value, error) {
	const q = `
	SELECT
		value_id, user_id, content, facet, display_order,
		date_created, date_updated
	FROM values
	WHERE value_id = :value_id`

	data := struct {
		ID string `db:"value_id"`
	}{
		ID: valueID.String(),
	}

	var dbValue value
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbValue); err != nil {
		return valuebus.Value{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusValueDecrypted(dbValue, s.encryptor)
}

// Count returns the total number of values matching the filter.
func (s *Store) Count(ctx context.Context, filter valuebus.QueryFilter) (int, error) {
	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM values
	%s`, whereClause)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// buildWhereClause constructs the WHERE clause dynamically.
func buildWhereClause(filter valuebus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["value_id"] = *filter.ID
		conditions = append(conditions, "value_id = :value_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.Facet != nil {
		data["facet"] = filter.Facet.String()
		conditions = append(conditions, "facet = :facet")
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

---

## Phase 5: App Layer (HTTP/JSON)

### File: `app/domain/valueapp/model.go` (NEW)

```go
// Package valueapp provides HTTP handlers for values.
package valueapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/business/types/valuecontent"
)

// Value represents a value for API responses.
type Value struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	Facet        string `json:"facet"`
	DisplayOrder int    `json:"displayOrder"`
	DateCreated  string `json:"dateCreated"`
	DateUpdated  string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (v Value) Encode() ([]byte, string, error) {
	data, err := json.Marshal(v)
	return data, "application/json", err
}

// NewValue represents data for creating a new value.
type NewValue struct {
	Content      string `json:"content" validate:"required"`
	Facet        string `json:"facet" validate:"required"`
	DisplayOrder int    `json:"displayOrder" validate:"required,min=1,max=10"`
}

// Decode implements the decoder interface.
func (nv *NewValue) Decode(data []byte) error {
	return json.Unmarshal(data, nv)
}

// Validate checks the data for correctness.
func (nv NewValue) Validate() error {
	if err := errs.Check(nv); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateValue represents data for updating a value.
type UpdateValue struct {
	Content      *string `json:"content"`
	Facet        *string `json:"facet"`
	DisplayOrder *int    `json:"displayOrder" validate:"omitempty,min=1,max=10"`
}

// Decode implements the decoder interface.
func (uv *UpdateValue) Decode(data []byte) error {
	return json.Unmarshal(data, uv)
}

// Validate checks the data for correctness.
func (uv UpdateValue) Validate() error {
	if err := errs.Check(uv); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ===== Business → App conversions =====

func toAppValue(value valuebus.Value) Value {
	return Value{
		ID:           value.ID.String(),
		Content:      value.Content.String(),
		Facet:        value.Facet.String(),
		DisplayOrder: value.DisplayOrder,
		DateCreated:  value.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated:  value.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppValues(values []valuebus.Value) []Value {
	app := make([]Value, len(values))
	for i, value := range values {
		app[i] = toAppValue(value)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewValue(ctx context.Context, nv NewValue) (valuebus.NewValue, error) {
	var errors errs.FieldErrors

	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		errors.Add("userID", err)
	}

	// Parse content
	content, err := valuecontent.Parse(nv.Content)
	if err != nil {
		errors.Add("content", err)
	}

	// Parse facet
	facetVal, err := facet.Parse(nv.Facet)
	if err != nil {
		errors.Add("facet", err)
	}

	if len(errors) > 0 {
		return valuebus.NewValue{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return valuebus.NewValue{
		UserID:       userID,
		Content:      content,
		Facet:        facetVal,
		DisplayOrder: nv.DisplayOrder,
	}, nil
}

func toBusUpdateValue(ctx context.Context, uv UpdateValue) (valuebus.UpdateValue, error) {
	var errors errs.FieldErrors
	var bus valuebus.UpdateValue

	if uv.Content != nil {
		content, err := valuecontent.Parse(*uv.Content)
		if err != nil {
			errors.Add("content", err)
		} else {
			bus.Content = &content
		}
	}

	if uv.Facet != nil {
		facetVal, err := facet.Parse(*uv.Facet)
		if err != nil {
			errors.Add("facet", err)
		} else {
			bus.Facet = &facetVal
		}
	}

	if uv.DisplayOrder != nil {
		bus.DisplayOrder = uv.DisplayOrder
	}

	if len(errors) > 0 {
		return valuebus.UpdateValue{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}
```

### File: `app/domain/valueapp/order.go` (NEW)

```go
package valueapp

import "github.com/francowini/rafiki/business/domain/valuebus"

var orderByFields = map[string]string{
	"value_id":      valuebus.OrderByValueID,
	"display_order": valuebus.OrderByDisplayOrder,
	"facet":         valuebus.OrderByFacet,
	"date_created":  valuebus.OrderByDateCreated,
	"date_updated":  valuebus.OrderByDateUpdated,
}
```

### File: `app/domain/valueapp/valueapp.go` (NEW)

```go
package valueapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
	"github.com/google/uuid"
)

type app struct {
	valueBus *valuebus.Business
}

func newApp(valueBus *valuebus.Business) *app {
	return &app{
		valueBus: valueBus,
	}
}

// create handles POST /v1/values
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appValue NewValue
	if err := web.Decode(r, &appValue); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nv, err := toBusNewValue(ctx, appValue)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	value, err := a.valueBus.Create(ctx, nv)
	if err != nil {
		if errors.Is(err, valuebus.ErrMaxValues) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, valuebus.ErrDuplicateOrder) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: value[%+v]: %s", value, err)
	}

	return toAppValue(value)
}

// update handles PUT /v1/values/{value_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appValue UpdateValue
	if err := web.Decode(r, &appValue); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	value, err := a.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: valueID[%s]: %s", valueID, err)
	}

	if value.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	uv, err := toBusUpdateValue(ctx, appValue)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	value, err = a.valueBus.Update(ctx, value, uv)
	if err != nil {
		if errors.Is(err, valuebus.ErrDuplicateOrder) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppValue(value)
}

// delete handles DELETE /v1/values/{value_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	value, err := a.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: valueID[%s]: %s", valueID, err)
	}

	if value.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	if err := a.valueBus.Delete(ctx, value); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// query handles GET /v1/values
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, valuebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := valuebus.QueryFilter{
		UserID:  &userID,
		Page:    pg,
		OrderBy: orderBy,
	}

	// Filter by facet if provided
	if qp.Facet != "" {
		facetVal, err := facet.Parse(qp.Facet)
		if err != nil {
			return errs.NewFieldErrors("facet", err)
		}
		filter.Facet = &facetVal
	}

	values, err := a.valueBus.Query(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.valueBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppValues(values), total, pg)
}

// queryByID handles GET /v1/values/{value_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	value, err := a.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: valueID[%s]: %s", valueID, err)
	}

	if value.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	return toAppValue(value)
}

// ===== Query params parsing =====

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Facet   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		Facet:   values.Get("facet"),
	}
}
```

### File: `app/domain/valueapp/route.go` (NEW)

```go
package valueapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	ValueBus *valuebus.Business
	Auth     *auth.Auth
}

// Routes registers all value endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.ValueBus)

	app.HandlerFunc(http.MethodPost, version, "/values", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/values", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/values/{value_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/values/{value_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/values/{value_id}", api.delete, bearer)
}
```

---

## Phase 6: Integration

### File: `app/sdk/mux/mux.go` (MODIFIED)

Add `ValueBus` to the `BusConfig` struct:

```go
type BusConfig struct {
	ThinkBus  *thinkbus.Business
	MomentBus *momentbus.Business
	ValueBus  *valuebus.Business  // ADD THIS LINE
	UserBus   userbus.ExtBusiness
	Auth      *auth.Auth
}
```

### File: `api/services/partners/all/all.go` (MODIFIED)

Add value routes to the `Add` method:

```go
import (
	// ... existing imports ...
	"github.com/francowini/rafiki/app/domain/valueapp"
)

func (add) Add(app *web.App, cfg mux.Config) {
	// ... existing routes ...

	valueapp.Routes(app, valueapp.Config{
		ValueBus: cfg.BusConfig.ValueBus,
		Auth:     cfg.BusConfig.Auth,
	})
}
```

### File: `api/services/partners/main.go` (MODIFIED)

Add ValueBus initialization in the `run` function:

```go
import (
	// ... existing imports ...
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/domain/valuebus/stores/valuedb"
)

// In the run() function, after creating encryptor:

// =========================================================================
// Business Domain

// ... existing business layers ...

// ValueBus
valueStore := valuedb.NewStore(log, db, encryptor)
valueBus := valuebus.NewBusiness(log, valueStore)

// ... later in the function ...

cfgMux := mux.Config{
	Build:  build,
	Log:    log,
	DB:     db,
	Tracer: tracer,
	BusConfig: mux.BusConfig{
		ThinkBus:  thinkBus,
		MomentBus: momentBus,
		ValueBus:  valueBus,  // ADD THIS LINE
		UserBus:   userBus,
		Auth:      authInstance,
	},
}
```

---

## Testing Checklist

After implementation, test the following:

### Local Testing

1. **Run migration:**
   ```bash
   make up
   # Check logs to verify migration 1.04 applied
   ```

2. **Test all endpoints with curl:**

   ```bash
   # Get auth token
   TOKEN="your_jwt_token_here"

   # Create value
   curl -X POST http://localhost:3000/v1/values \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "content": "Live with integrity and authenticity",
       "facet": "personal_growth",
       "displayOrder": 1
     }'

   # List values
   curl -X GET http://localhost:3000/v1/values \
     -H "Authorization: Bearer $TOKEN"

   # Get single value
   curl -X GET http://localhost:3000/v1/values/{value_id} \
     -H "Authorization: Bearer $TOKEN"

   # Update value
   curl -X PUT http://localhost:3000/v1/values/{value_id} \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "content": "UPDATED: Live with complete integrity",
       "displayOrder": 2
     }'

   # Delete value
   curl -X DELETE http://localhost:3000/v1/values/{value_id} \
     -H "Authorization: Bearer $TOKEN"

   # Filter by facet
   curl -X GET "http://localhost:3000/v1/values?facet=personal_growth" \
     -H "Authorization: Bearer $TOKEN"
   ```

3. **Test max 10 values limit:**
   - Create 10 values successfully
   - Attempt to create 11th value
   - Should return error: "maximum 10 values allowed per user"

4. **Test validation:**
   - Empty content → error
   - Content < 3 chars → error
   - Content > 200 chars → error
   - Invalid facet → error
   - displayOrder < 1 or > 10 → error

5. **Test encryption:**
   - Query database directly
   - Verify `content` field is encrypted (not plain text)

6. **Run linter:**
   ```bash
   golangci-lint run --fix
   ```

---

## Implementation Order

Follow this order to minimize errors:

1. ✅ **Database Migration** (`migrate.sql`)
2. ✅ **Business Types** (`facet.go`, `valuecontent.go`)
3. ✅ **Domain Models** (`valuebus/model.go`, `filter.go`, `order.go`)
4. ✅ **Business Logic** (`valuebus/valuebus.go`)
5. ✅ **Store Layer** (`valuedb/model.go`, `order.go`, `valuedb.go`)
6. ✅ **App Layer** (`valueapp/model.go`, `order.go`, `valueapp.go`, `route.go`)
7. ✅ **Integration** (`mux.go`, `all.go`, `main.go`)
8. ✅ **Testing** (local endpoints, validation, encryption)

---

## Success Criteria

- ✅ All 5 endpoints working (POST, GET list, GET by ID, PUT, DELETE)
- ✅ Facet filtering working
- ✅ Max 10 values enforced
- ✅ Content validation (3-200 chars)
- ✅ Field-level encryption verified
- ✅ User isolation enforced (can't access other users' values)
- ✅ golangci-lint passes with no errors
- ✅ No errors in application logs

---

## Next Steps

After backend is complete and tested:

1. **Deploy Backend** (see `DEPLOYMENT_GUIDE.md`)
2. **Implement Frontend** (see `values-feature-implementation-plan.md` Phase 2)
3. **Deploy Frontend** (Vercel)
4. **Monitor** (24-hour monitoring period)

---

**Document Version:** 1.0
**Last Updated:** November 22, 2025
**Status:** Ready for Implementation
