package objectiverecordbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/recordstatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, record ObjectiveRecord) error
	Update(ctx context.Context, record ObjectiveRecord) error
	Delete(ctx context.Context, record ObjectiveRecord) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ObjectiveRecord, error)
	QueryByID(ctx context.Context, recordID uuid.UUID) (ObjectiveRecord, error)
	QueryByObjectiveAndDate(ctx context.Context, objectiveID uuid.UUID, recordDate time.Time) (ObjectiveRecord, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nr NewObjectiveRecord) (ObjectiveRecord, error)
	Update(ctx context.Context, record ObjectiveRecord, ur UpdateObjectiveRecord) (ObjectiveRecord, error)
	Delete(ctx context.Context, record ObjectiveRecord) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ObjectiveRecord, error)
	QueryByID(ctx context.Context, recordID uuid.UUID) (ObjectiveRecord, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	CalculateCompliance(ctx context.Context, objective objectivebus.Objective) (int, error)
}

// Business manages objective record operations.
type Business struct {
	log          *logger.Logger
	objectiveBus objectivebus.ExtBusiness
	delegate     *delegate.Delegate
	storer       Storer
}

// NewBusiness constructs a Business for objective record domain.
func NewBusiness(
	log *logger.Logger,
	objectiveBus objectivebus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
) ExtBusiness {
	b := &Business{
		log:          log,
		objectiveBus: objectiveBus,
		delegate:     dlg,
		storer:       storer,
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

	objectiveBus, err := b.objectiveBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	// Create business without re-registering delegate functions
	// to avoid duplicate handler registration on the shared delegate.
	return &Business{
		log:          b.log,
		objectiveBus: objectiveBus,
		delegate:     b.delegate,
		storer:       storer,
	}, nil
}

// Create adds a new objective record or updates if one exists for the date (upsert).
func (b *Business) Create(ctx context.Context, nr NewObjectiveRecord) (ObjectiveRecord, error) {
	b.log.Info(ctx, "objectiverecordbus.create", "userID", nr.UserID, "objectiveID", nr.ObjectiveID)

	// Validate parent objective exists
	objective, err := b.objectiveBus.QueryByID(ctx, nr.ObjectiveID)
	if err != nil {
		return ObjectiveRecord{}, fmt.Errorf("objective.querybyid: objectiveID[%s]: %w", nr.ObjectiveID, err)
	}

	// Security: Verify authenticated user owns the objective
	if objective.UserID != nr.UserID {
		return ObjectiveRecord{}, ErrNotObjectiveOwner
	}

	// Verify objective is frequency type
	if !objective.TrackingType.IsFrequency() {
		return ObjectiveRecord{}, ErrNotFrequencyType
	}

	// Validate grace period (today or yesterday only)
	if err := validateGracePeriod(nr.RecordDate); err != nil {
		return ObjectiveRecord{}, err
	}

	// Check if record already exists for this date (for upsert decision)
	existingRecord, err := b.storer.QueryByObjectiveAndDate(ctx, nr.ObjectiveID, nr.RecordDate)
	isUpdate := err == nil
	if err != nil && !errors.Is(err, sqldb.ErrDBNotFound) {
		return ObjectiveRecord{}, fmt.Errorf("check existing: %w", err)
	}

	now := time.Now().UTC()

	var record ObjectiveRecord
	if isUpdate {
		// Update existing record - preserve ID and DateCreated
		record = ObjectiveRecord{
			ID:          existingRecord.ID,
			ObjectiveID: nr.ObjectiveID,
			UserID:      objective.UserID,
			RecordDate:  normalizeToDate(nr.RecordDate),
			Status:      nr.Status,
			Notes:       nr.Notes,
			DateCreated: existingRecord.DateCreated,
		}
	} else {
		// Create new record
		record = ObjectiveRecord{
			ID:          uuid.New(),
			ObjectiveID: nr.ObjectiveID,
			UserID:      objective.UserID,
			RecordDate:  normalizeToDate(nr.RecordDate),
			Status:      nr.Status,
			Notes:       nr.Notes,
			DateCreated: now,
		}
	}

	// Perform upsert (insert or update)
	if err := b.storer.Create(ctx, record); err != nil {
		b.log.Error(ctx, "objectiverecordbus.create", "err", err)
		return ObjectiveRecord{}, fmt.Errorf("create: %w", err)
	}

	if isUpdate {
		b.log.Info(ctx, "objectiverecordbus.create.success", "recordID", record.ID, "action", "updated")
	} else {
		b.log.Info(ctx, "objectiverecordbus.create.success", "recordID", record.ID, "action", "created")
	}

	return record, nil
}

// Update modifies an existing objective record.
func (b *Business) Update(ctx context.Context, record ObjectiveRecord, ur UpdateObjectiveRecord) (ObjectiveRecord, error) {
	b.log.Info(ctx, "objectiverecordbus.update", "recordID", record.ID)

	if ur.Status != nil {
		record.Status = *ur.Status
	}

	if ur.Notes != nil {
		record.Notes = *ur.Notes
	}

	if err := b.storer.Update(ctx, record); err != nil {
		b.log.Error(ctx, "objectiverecordbus.update", "err", err)
		return ObjectiveRecord{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectiverecordbus.update.success", "recordID", record.ID)
	return record, nil
}

// Delete removes an objective record.
func (b *Business) Delete(ctx context.Context, record ObjectiveRecord) error {
	b.log.Info(ctx, "objectiverecordbus.delete", "recordID", record.ID)

	if err := b.storer.Delete(ctx, record); err != nil {
		b.log.Error(ctx, "objectiverecordbus.delete", "err", err)
		return fmt.Errorf("delete: %w", err)
	}

	b.log.Info(ctx, "objectiverecordbus.delete.success", "recordID", record.ID)
	return nil
}

// Query retrieves objective records based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ObjectiveRecord, error) {
	records, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return records, nil
}

// QueryByID finds an objective record by its ID.
func (b *Business) QueryByID(ctx context.Context, recordID uuid.UUID) (ObjectiveRecord, error) {
	record, err := b.storer.QueryByID(ctx, recordID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return ObjectiveRecord{}, ErrNotFound
		}
		return ObjectiveRecord{}, fmt.Errorf("query: recordID[%s]: %w", recordID, err)
	}

	return record, nil
}

// Count returns the total number of objective records matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// CalculateCompliance calculates the compliance percentage for the current period.
func (b *Business) CalculateCompliance(ctx context.Context, objective objectivebus.Objective) (int, error) {
	if !objective.TrackingType.IsFrequency() {
		return 0, errors.New("compliance only applies to frequency tracking")
	}

	if objective.FrequencyType == nil {
		return 0, errors.New("frequency_type not set")
	}

	now := time.Now().UTC()
	var startDate time.Time
	var targetCount int

	switch {
	case objective.FrequencyType.IsDaily():
		// Current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		// Go idiom: day=0 of the next month returns the last day of the current month
		daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		targetCount = daysInMonth

	case objective.FrequencyType.IsNPerWeek():
		// Current week (Monday-Sunday)
		// Convert time.Weekday (Sunday=0..Saturday=6) to Monday=1..Sunday=7
		// so that Monday is treated as the first day of the week.
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		// Calculate startDate as the Monday of the current week at UTC midnight
		startDate = now.AddDate(0, 0, -(weekday - 1))
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		if objective.FrequencyCount == nil {
			return 0, ErrFrequencyCountRequired
		}
		targetCount = objective.FrequencyCount.Value()

	case objective.FrequencyType.IsNPerMonth():
		// Current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		if objective.FrequencyCount == nil {
			return 0, ErrFrequencyCountRequired
		}
		targetCount = objective.FrequencyCount.Value()

	default:
		return 0, fmt.Errorf("unknown frequency type: %s", objective.FrequencyType.String())
	}

	// Count completed records in period
	completedStatus := recordstatus.Completed
	filter := QueryFilter{
		ObjectiveID: &objective.ID,
		StartDate:   &startDate,
		EndDate:     &now,
		Status:      &completedStatus,
	}

	completedCount, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count records: %w", err)
	}

	if targetCount == 0 {
		return 0, nil
	}

	compliance := (completedCount * 100) / targetCount
	if compliance > 100 {
		compliance = 100
	}

	return compliance, nil
}

// validateGracePeriod ensures records can only be logged for today or yesterday.
func validateGracePeriod(recordDate time.Time) error {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	normalizedDate := normalizeToDate(recordDate)

	if normalizedDate.After(today) {
		return ErrFutureDateNotAllowed
	}

	if normalizedDate.Before(yesterday) {
		return ErrGracePeriodExceeded
	}

	return nil
}

// normalizeToDate strips time components and normalizes to UTC midnight.
func normalizeToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
