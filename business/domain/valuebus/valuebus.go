package valuebus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
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
