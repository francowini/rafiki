package vobjectiveactivitybus

import (
	"context"
	"fmt"

	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations for activity queries.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter) (Activity, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality.
type ExtBusiness interface {
	Query(ctx context.Context, filter QueryFilter) (Activity, error)
}

// Extension is a function that wraps ExtBusiness with additional functionality.
type Extension func(ExtBusiness) ExtBusiness

// Business manages read-only objective activity operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a Business for objective activity domain.
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

// Query retrieves activity data for an objective within a year.
// Year validation is handled by the year.Year type at parse time.
func (b *Business) Query(ctx context.Context, filter QueryFilter) (Activity, error) {
	b.log.Info(ctx, "vobjectiveactivity.query", "objectiveID", filter.ObjectiveID, "year", filter.Year)

	activity, err := b.storer.Query(ctx, filter)
	if err != nil {
		b.log.Error(ctx, "vobjectiveactivity.query.error", "err", err)
		return Activity{}, fmt.Errorf("storer.query: %w", err)
	}

	return activity, nil
}
