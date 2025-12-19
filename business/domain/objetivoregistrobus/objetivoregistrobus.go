package objetivoregistrobus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/registrostatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, record ObjetivoRecord) error
	Update(ctx context.Context, record ObjetivoRecord) error
	Delete(ctx context.Context, record ObjetivoRecord) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ObjetivoRecord, error)
	QueryByID(ctx context.Context, recordID uuid.UUID) (ObjetivoRecord, error)
	QueryByObjetivoAndDate(ctx context.Context, objetivoID uuid.UUID, fechaRegistro time.Time) (ObjetivoRecord, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nr NewObjetivoRecord) (ObjetivoRecord, error)
	Update(ctx context.Context, record ObjetivoRecord, ur UpdateObjetivoRecord) (ObjetivoRecord, error)
	Delete(ctx context.Context, record ObjetivoRecord) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ObjetivoRecord, error)
	QueryByID(ctx context.Context, recordID uuid.UUID) (ObjetivoRecord, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	CalculateCompliance(ctx context.Context, objetivo objetivobus.Objetivo) (int, error)
}

// Business manages objetivo record operations.
type Business struct {
	log         *logger.Logger
	objetivoBus objetivobus.ExtBusiness
	delegate    *delegate.Delegate
	storer      Storer
}

// NewBusiness constructs a Business for objetivo record domain.
func NewBusiness(
	log *logger.Logger,
	objetivoBus objetivobus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
) ExtBusiness {
	b := &Business{
		log:         log,
		objetivoBus: objetivoBus,
		delegate:    dlg,
		storer:      storer,
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

	objetivoBus, err := b.objetivoBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	// Create business without re-registering delegate functions
	// to avoid duplicate handler registration on the shared delegate.
	return &Business{
		log:         b.log,
		objetivoBus: objetivoBus,
		delegate:    b.delegate,
		storer:      storer,
	}, nil
}

// Create adds a new objetivo record or updates if one exists for the date (upsert).
func (b *Business) Create(ctx context.Context, nr NewObjetivoRecord) (ObjetivoRecord, error) {
	b.log.Info(ctx, "objetivoregistrobus.create", "userID", nr.UserID, "objetivoID", nr.ObjetivoID)

	// Validate parent objetivo exists
	objetivo, err := b.objetivoBus.QueryByID(ctx, nr.ObjetivoID)
	if err != nil {
		return ObjetivoRecord{}, fmt.Errorf("objetivo.querybyid: objetivoID[%s]: %w", nr.ObjetivoID, err)
	}

	// Security: Verify authenticated user owns the objetivo
	if objetivo.UserID != nr.UserID {
		return ObjetivoRecord{}, ErrNotObjetivoOwner
	}

	// Verify objetivo is frecuencia type
	if !objetivo.TipoTracking.IsFrecuencia() {
		return ObjetivoRecord{}, ErrNotFrecuenciaType
	}

	// Validate grace period (today or yesterday only)
	if err := validateGracePeriod(nr.FechaRegistro); err != nil {
		return ObjetivoRecord{}, err
	}

	// Check if record already exists for this date (for upsert decision)
	existingRecord, err := b.storer.QueryByObjetivoAndDate(ctx, nr.ObjetivoID, nr.FechaRegistro)
	isUpdate := err == nil
	if err != nil && !errors.Is(err, sqldb.ErrDBNotFound) {
		return ObjetivoRecord{}, fmt.Errorf("check existing: %w", err)
	}

	now := time.Now().UTC()

	var record ObjetivoRecord
	if isUpdate {
		// Update existing record - preserve ID and DateCreated
		record = ObjetivoRecord{
			ID:            existingRecord.ID,
			ObjetivoID:    nr.ObjetivoID,
			UserID:        objetivo.UserID,
			FechaRegistro: normalizeToDate(nr.FechaRegistro),
			Status:        nr.Status,
			Notes:         nr.Notes,
			DateCreated:   existingRecord.DateCreated,
		}
	} else {
		// Create new record
		record = ObjetivoRecord{
			ID:            uuid.New(),
			ObjetivoID:    nr.ObjetivoID,
			UserID:        objetivo.UserID,
			FechaRegistro: normalizeToDate(nr.FechaRegistro),
			Status:        nr.Status,
			Notes:         nr.Notes,
			DateCreated:   now,
		}
	}

	// Perform upsert (insert or update)
	if err := b.storer.Create(ctx, record); err != nil {
		b.log.Error(ctx, "objetivoregistrobus.create", "err", err)
		return ObjetivoRecord{}, fmt.Errorf("create: %w", err)
	}

	if isUpdate {
		b.log.Info(ctx, "objetivoregistrobus.create.success", "recordID", record.ID, "action", "updated")
	} else {
		b.log.Info(ctx, "objetivoregistrobus.create.success", "recordID", record.ID, "action", "created")
	}

	return record, nil
}

// Update modifies an existing objetivo record.
func (b *Business) Update(ctx context.Context, record ObjetivoRecord, ur UpdateObjetivoRecord) (ObjetivoRecord, error) {
	b.log.Info(ctx, "objetivoregistrobus.update", "recordID", record.ID)

	if ur.Status != nil {
		record.Status = *ur.Status
	}

	if ur.Notes != nil {
		record.Notes = *ur.Notes
	}

	if err := b.storer.Update(ctx, record); err != nil {
		b.log.Error(ctx, "objetivoregistrobus.update", "err", err)
		return ObjetivoRecord{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivoregistrobus.update.success", "recordID", record.ID)
	return record, nil
}

// Delete removes an objetivo record.
func (b *Business) Delete(ctx context.Context, record ObjetivoRecord) error {
	b.log.Info(ctx, "objetivoregistrobus.delete", "recordID", record.ID)

	if err := b.storer.Delete(ctx, record); err != nil {
		b.log.Error(ctx, "objetivoregistrobus.delete", "err", err)
		return fmt.Errorf("delete: %w", err)
	}

	b.log.Info(ctx, "objetivoregistrobus.delete.success", "recordID", record.ID)
	return nil
}

// Query retrieves objetivo records based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ObjetivoRecord, error) {
	records, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return records, nil
}

// QueryByID finds an objetivo record by its ID.
func (b *Business) QueryByID(ctx context.Context, recordID uuid.UUID) (ObjetivoRecord, error) {
	record, err := b.storer.QueryByID(ctx, recordID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return ObjetivoRecord{}, ErrNotFound
		}
		return ObjetivoRecord{}, fmt.Errorf("query: recordID[%s]: %w", recordID, err)
	}

	return record, nil
}

// Count returns the total number of objetivo records matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// CalculateCompliance calculates the compliance percentage for the current period.
func (b *Business) CalculateCompliance(ctx context.Context, objetivo objetivobus.Objetivo) (int, error) {
	if !objetivo.TipoTracking.IsFrecuencia() {
		return 0, errors.New("compliance only applies to frecuencia tracking")
	}

	if objetivo.FrecuenciaTipo == nil {
		return 0, errors.New("frecuencia_tipo not set")
	}

	now := time.Now().UTC()
	var startDate time.Time
	var targetCount int

	switch {
	case objetivo.FrecuenciaTipo.IsDaily():
		// Current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		// Go idiom: day=0 of the next month returns the last day of the current month
		daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		targetCount = daysInMonth

	case objetivo.FrecuenciaTipo.IsNPorSemana():
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
		if objetivo.FrecuenciaN == nil {
			return 0, ErrFrecuenciaNRequired
		}
		targetCount = objetivo.FrecuenciaN.Value()

	case objetivo.FrecuenciaTipo.IsNPorMes():
		// Current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		if objetivo.FrecuenciaN == nil {
			return 0, ErrFrecuenciaNRequired
		}
		targetCount = objetivo.FrecuenciaN.Value()

	default:
		return 0, fmt.Errorf("unknown frecuencia type: %s", objetivo.FrecuenciaTipo.String())
	}

	// Count completed records in period
	completadoStatus := registrostatus.Completado
	filter := QueryFilter{
		ObjetivoID: &objetivo.ID,
		StartDate:  &startDate,
		EndDate:    &now,
		Status:     &completadoStatus,
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
func validateGracePeriod(fechaRegistro time.Time) error {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	recordDate := normalizeToDate(fechaRegistro)

	if recordDate.After(today) {
		return ErrFutureDateNotAllowed
	}

	if recordDate.Before(yesterday) {
		return ErrGracePeriodExceeded
	}

	return nil
}

// normalizeToDate strips time components and normalizes to UTC midnight.
func normalizeToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
