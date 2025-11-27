package vexportbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/logger"
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
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a vexport business API for use.
func NewBusiness(log *logger.Logger, storer Storer, extensions ...Extension) ExtBusiness {
	b := ExtBusiness(&Business{
		log:    log,
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
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ExportItem, error) {
	b.log.Info(ctx, "vexport.query", "userID", filter.UserID, "startDate", filter.StartDate, "endDate", filter.EndDate, "page", pg.Number(), "rows", pg.RowsPerPage(), "orderBy", orderBy)

	// Validate UserID is provided (security: user data isolation)
	if filter.UserID == uuid.Nil {
		b.log.Error(ctx, "vexport.query", "err", ErrInvalidUserID)
		return nil, ErrInvalidUserID
	}

	// Domain-specific validation: date range must be valid
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			b.log.Error(ctx, "vexport.query", "err", ErrInvalidDateRange, "startDate", filter.StartDate, "endDate", filter.EndDate)
			return nil, ErrInvalidDateRange
		}
	}

	items, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		b.log.Error(ctx, "vexport.query", "err", err, "userID", filter.UserID)
		return nil, fmt.Errorf("storer.query: %w", err)
	}

	b.log.Info(ctx, "vexport.query.success", "userID", filter.UserID, "count", len(items))

	return items, nil
}

// Count returns the total number of export items matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	b.log.Info(ctx, "vexport.count", "userID", filter.UserID, "startDate", filter.StartDate, "endDate", filter.EndDate)

	// Validate UserID is provided (security: user data isolation)
	if filter.UserID == uuid.Nil {
		b.log.Error(ctx, "vexport.count", "err", ErrInvalidUserID)
		return 0, ErrInvalidUserID
	}

	// Domain-specific validation: date range must be valid
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			b.log.Error(ctx, "vexport.count", "err", ErrInvalidDateRange, "startDate", filter.StartDate, "endDate", filter.EndDate)
			return 0, ErrInvalidDateRange
		}
	}

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		err = fmt.Errorf("storer.count: %w", err)
		b.log.Error(ctx, "vexport.count", "err", err, "userID", filter.UserID)
		return 0, err
	}

	b.log.Info(ctx, "vexport.count.success", "userID", filter.UserID, "count", count)

	return count, nil
}
