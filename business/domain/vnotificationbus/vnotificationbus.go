package vnotificationbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface declares the behavior this package needs to retrieve data.
// READ ONLY - no Create/Update/Delete.
type Storer interface {
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]ValueWithVision, error)
}

// Business manages the set of APIs for vnotification access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a vnotification business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// QueryByUserID retrieves all values with their life visions for a user.
func (b *Business) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]ValueWithVision, error) {
	values, err := b.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query by user id: %w", err)
	}

	return values, nil
}
