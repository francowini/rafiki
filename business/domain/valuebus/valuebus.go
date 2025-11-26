package valuebus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/displayorder"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, value Value) error
	Update(ctx context.Context, value Value) error
	Delete(ctx context.Context, value Value) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	BatchUpdate(ctx context.Context, values []Value) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Value, error)
	QueryByID(ctx context.Context, valueID uuid.UUID) (Value, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nv NewValue) (Value, error)
	Update(ctx context.Context, value Value, uv UpdateValue) (Value, error)
	Delete(ctx context.Context, value Value) error
	Reorder(ctx context.Context, userID uuid.UUID, rr ReorderRequest) ([]Value, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Value, error)
	QueryByID(ctx context.Context, valueID uuid.UUID) (Value, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Extension is a function that wraps a new layer of business logic
// around the existing business logic.
type Extension func(ExtBusiness) ExtBusiness

// Business manages value operations.
type Business struct {
	log        *logger.Logger
	userBus    userbus.ExtBusiness
	delegate   *delegate.Delegate
	storer     Storer
	extensions []Extension
}

// NewBusiness constructs a Business for value domain.
func NewBusiness(
	log *logger.Logger,
	userBus userbus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
	extensions ...Extension,
) ExtBusiness {
	b := &Business{
		log:        log,
		userBus:    userBus,
		delegate:   dlg,
		storer:     storer,
		extensions: extensions,
	}

	// Only register delegate functions on the root business instance.
	b.registerDelegateFunctions()

	return applyExtensions(b, extensions)
}

// applyExtensions wraps the business with the provided extensions.
func applyExtensions(b *Business, extensions []Extension) ExtBusiness {
	extBus := ExtBusiness(b)

	for i := len(extensions) - 1; i >= 0; i-- {
		ext := extensions[i]
		if ext != nil {
			extBus = ext(extBus)
		}
	}

	return extBus
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	userBus, err := b.userBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	// Create business without re-registering delegate functions
	// to avoid duplicate handler registration on the shared delegate.
	nb := &Business{
		log:        b.log,
		userBus:    userBus,
		delegate:   b.delegate,
		storer:     storer,
		extensions: b.extensions,
	}

	return applyExtensions(nb, b.extensions), nil
}

// Create adds a new value to the system.
func (b *Business) Create(ctx context.Context, nv NewValue) (Value, error) {
	// Validate parent user exists and is enabled
	usr, err := b.userBus.QueryByID(ctx, nv.UserID)
	if err != nil {
		return Value{}, fmt.Errorf("user.querybyid: %s: %w", nv.UserID, err)
	}

	if !usr.Enabled {
		return Value{}, ErrUserDisabled
	}

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

	// Publish event for child domains (e.g., goals)
	if b.delegate != nil {
		if err := b.delegate.Call(ctx, ActionDeletedData(value.ID)); err != nil {
			return fmt.Errorf("delegate call: %w", err)
		}
	}

	return nil
}

// Query retrieves a list of values based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Value, error) {
	values, err := b.storer.Query(ctx, filter, orderBy, page)
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

// Reorder atomically updates display order for multiple values.
// Invalid/non-existent value IDs are silently ignored.
// Returns updated values sorted by display order.
func (b *Business) Reorder(ctx context.Context, userID uuid.UUID, rr ReorderRequest) ([]Value, error) {
	// Validate parent user exists and is enabled
	usr, err := b.userBus.QueryByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user.querybyid: %s: %w", userID, err)
	}

	if !usr.Enabled {
		return nil, ErrUserDisabled
	}

	// Get current values for the user
	filter := QueryFilter{UserID: &userID}
	orderBy := order.By{Field: OrderByDisplayOrder, Direction: order.ASC}
	pg := page.MustParse("1", "10")

	currentValues, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query current values: %w", err)
	}

	// Build update map from request
	updateMap := make(map[string]displayorder.DisplayOrder)
	for _, item := range rr.Items {
		updateMap[item.ID.String()] = item.DisplayOrder
	}

	// Prepare values to update (only user's values that are in request)
	now := time.Now()
	valuesToUpdate := make([]Value, 0, len(currentValues))
	for _, current := range currentValues {
		if newOrder, found := updateMap[current.ID.String()]; found {
			current.DisplayOrder = newOrder
			current.DateUpdated = now
			valuesToUpdate = append(valuesToUpdate, current)
		}
	}

	// Silent ignore: if no matching values, return current state
	if len(valuesToUpdate) == 0 {
		return currentValues, nil
	}

	// Atomic batch update
	if err := b.storer.BatchUpdate(ctx, valuesToUpdate); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return nil, ErrDuplicateOrder
		}
		return nil, fmt.Errorf("batch update: %w", err)
	}

	// Return updated values
	updatedValues, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query updated values: %w", err)
	}

	return updatedValues, nil
}
