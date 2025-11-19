package momentbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
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
