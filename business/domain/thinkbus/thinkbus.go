package thinkbus

import (
	"context"
	"fmt"
	"time"

	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, think Think) error
	Query(ctx context.Context, userID uuid.UUID, orderBy order.By, page page.Page) ([]Think, error)
	Count(ctx context.Context, userID uuid.UUID) (int, error)
	QueryByID(ctx context.Context, thinkID uuid.UUID, userID uuid.UUID) (Think, error)
}

// Business manages the set of APIs for think access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a think business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// Create adds a new think to the system.
// NO VALIDATION - All validation happens in app layer via Parse functions.
func (b *Business) Create(ctx context.Context, nt NewThink) (Think, error) {
	now := time.Now()

	think := Think{
		ID:          uuid.New(),
		UserID:      nt.UserID,
		Category:    nt.Category,
		Content:     nt.Content,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := b.storer.Create(ctx, think); err != nil {
		return Think{}, fmt.Errorf("create: %w", err)
	}

	return think, nil
}

// Query retrieves a list of all thinks with pagination
func (b *Business) Query(ctx context.Context, userID uuid.UUID, orderBy order.By, page page.Page) ([]Think, error) {
	thinks, err := b.storer.Query(ctx, userID, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return thinks, nil
}

// Count returns the total number of thinks
func (b *Business) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := b.storer.Count(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID finds the think by the specified ID
func (b *Business) QueryByID(ctx context.Context, thinkID uuid.UUID, userID uuid.UUID) (Think, error) {
	think, err := b.storer.QueryByID(ctx, thinkID, userID)
	if err != nil {
		return Think{}, fmt.Errorf("query: thinkID[%s]: %w", thinkID, err)
	}

	return think, nil
}
