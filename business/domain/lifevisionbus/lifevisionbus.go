package lifevisionbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, lifeVision LifeVision) error
	Update(ctx context.Context, lifeVision LifeVision) error
	Delete(ctx context.Context, lifeVision LifeVision) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LifeVision, error)
	QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (LifeVision, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error)
	Update(ctx context.Context, lifeVision LifeVision, ulv UpdateLifeVision) (LifeVision, error)
	Delete(ctx context.Context, lifeVision LifeVision) error
	Archive(ctx context.Context, lifeVision LifeVision) (LifeVision, error)
	Restore(ctx context.Context, lifeVision LifeVision) (LifeVision, error)
	Reassign(ctx context.Context, lifeVision LifeVision, newValueID uuid.UUID) (LifeVision, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LifeVision, error)
	QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (LifeVision, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages life vision operations.
type Business struct {
	log      *logger.Logger
	valueBus valuebus.ExtBusiness
	delegate *delegate.Delegate
	storer   Storer
}

// NewBusiness constructs a Business for life vision domain.
func NewBusiness(
	log *logger.Logger,
	valueBus valuebus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
) ExtBusiness {
	b := &Business{
		log:      log,
		valueBus: valueBus,
		delegate: dlg,
		storer:   storer,
	}

	// Only register delegate functions on the root business instance.
	b.registerDelegateFunctions()

	return b
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	valueBus, err := b.valueBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	// Create business without re-registering delegate functions
	// to avoid duplicate handler registration on the shared delegate.
	return &Business{
		log:      b.log,
		valueBus: valueBus,
		delegate: b.delegate,
		storer:   storer,
	}, nil
}

// Create adds a new life vision.
func (b *Business) Create(ctx context.Context, nlv NewLifeVision) (LifeVision, error) {
	// Validate parent value exists
	value, err := b.valueBus.QueryByID(ctx, nlv.ValueID)
	if err != nil {
		return LifeVision{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", nlv.ValueID, err)
	}

	// Security: Verify authenticated user owns the value
	if value.UserID != nlv.UserID {
		return LifeVision{}, ErrNotValueOwner
	}

	now := time.Now()

	lifeVision := LifeVision{
		ID:          uuid.New(),
		UserID:      value.UserID,
		ValueID:     nlv.ValueID,
		Content:     nlv.Content,
		Status:      entitystatus.Active,
		ArchivedAt:  nil,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := b.storer.Create(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("create: %w", err)
	}

	return lifeVision, nil
}

// Update modifies an existing life vision.
func (b *Business) Update(ctx context.Context, lifeVision LifeVision, ulv UpdateLifeVision) (LifeVision, error) {
	if ulv.Content != nil {
		lifeVision.Content = *ulv.Content
	}

	if ulv.ValueID != nil {
		// Validate new value exists and user owns it
		value, err := b.valueBus.QueryByID(ctx, *ulv.ValueID)
		if err != nil {
			return LifeVision{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", *ulv.ValueID, err)
		}

		// Security: Verify authenticated user owns the new value
		if value.UserID != lifeVision.UserID {
			return LifeVision{}, ErrNotValueOwner
		}

		lifeVision.ValueID = *ulv.ValueID
	}

	lifeVision.DateUpdated = time.Now()

	if err := b.storer.Update(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("update: %w", err)
	}

	return lifeVision, nil
}

// Delete removes a life vision.
func (b *Business) Delete(ctx context.Context, lifeVision LifeVision) error {
	if err := b.storer.Delete(ctx, lifeVision); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Publish event for potential future child domains
	if b.delegate != nil {
		if err := b.delegate.Call(ctx, ActionDeletedData(lifeVision)); err != nil {
			return fmt.Errorf("delegate call: %w", err)
		}
	}

	return nil
}

// Archive sets a life vision's status to archived.
func (b *Business) Archive(ctx context.Context, lifeVision LifeVision) (LifeVision, error) {
	if lifeVision.Status.IsArchived() {
		return LifeVision{}, ErrAlreadyArchived
	}

	now := time.Now().UTC()
	lifeVision.Status = entitystatus.Archived
	lifeVision.ArchivedAt = &now
	lifeVision.DateUpdated = now

	if err := b.storer.Update(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("update: %w", err)
	}

	return lifeVision, nil
}

// Restore sets an archived life vision's status back to active.
func (b *Business) Restore(ctx context.Context, lifeVision LifeVision) (LifeVision, error) {
	if !lifeVision.Status.IsArchived() {
		return LifeVision{}, ErrNotArchived
	}

	lifeVision.Status = entitystatus.Active
	lifeVision.ArchivedAt = nil
	lifeVision.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("update: %w", err)
	}

	return lifeVision, nil
}

// Reassign changes the value_id of a life vision to a different value.
func (b *Business) Reassign(ctx context.Context, lifeVision LifeVision, newValueID uuid.UUID) (LifeVision, error) {
	// Validate new value exists and user owns it
	value, err := b.valueBus.QueryByID(ctx, newValueID)
	if err != nil {
		return LifeVision{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", newValueID, err)
	}

	if value.UserID != lifeVision.UserID {
		return LifeVision{}, ErrNotValueOwner
	}

	if !value.Status.IsActive() {
		return LifeVision{}, ErrTargetValueNotActive
	}

	lifeVision.ValueID = newValueID
	lifeVision.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, lifeVision); err != nil {
		return LifeVision{}, fmt.Errorf("update: %w", err)
	}

	return lifeVision, nil
}

// Query retrieves life visions based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LifeVision, error) {
	lifeVisions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return lifeVisions, nil
}

// QueryByID finds a life vision by its ID.
func (b *Business) QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (LifeVision, error) {
	lifeVision, err := b.storer.QueryByID(ctx, lifeVisionID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return LifeVision{}, ErrNotFound
		}
		return LifeVision{}, fmt.Errorf("query: lifeVisionID[%s]: %w", lifeVisionID, err)
	}

	return lifeVision, nil
}

// Count returns the total number of life visions matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
