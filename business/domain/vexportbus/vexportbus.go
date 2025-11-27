package vexportbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"

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
	// Validate UserID is provided (security: user data isolation)
	if filter.UserID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

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
	// Validate UserID is provided (security: user data isolation)
	if filter.UserID == uuid.Nil {
		return 0, ErrInvalidUserID
	}

	// Domain-specific validation: date range must be valid
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			return 0, ErrInvalidDateRange
		}
	}

	return b.storer.Count(ctx, filter)
}
