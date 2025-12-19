package objetivobus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/metricaactual"
	"github.com/francowini/rafiki/business/types/objetivostatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, objetivo Objetivo) error
	Update(ctx context.Context, objetivo Objetivo) error
	Delete(ctx context.Context, objetivo Objetivo) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Objetivo, error)
	QueryByID(ctx context.Context, objetivoID uuid.UUID) (Objetivo, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, no NewObjetivo) (Objetivo, error)
	Update(ctx context.Context, objetivo Objetivo, uo UpdateObjetivo) (Objetivo, error)
	Delete(ctx context.Context, objetivo Objetivo) error
	Archive(ctx context.Context, objetivo Objetivo) (Objetivo, error)
	Restore(ctx context.Context, objetivo Objetivo) (Objetivo, error)
	ChangeStatus(ctx context.Context, objetivo Objetivo, req ChangeStatusRequest) (Objetivo, error)
	IncrementProgress(ctx context.Context, objetivo Objetivo, req IncrementProgressRequest) (Objetivo, error)
	UpdateProgress(ctx context.Context, objetivo Objetivo, req UpdateProgressRequest) (Objetivo, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Objetivo, error)
	QueryByID(ctx context.Context, objetivoID uuid.UUID) (Objetivo, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages objetivo operations.
type Business struct {
	log           *logger.Logger
	lifeVisionBus lifevisionbus.ExtBusiness
	delegate      *delegate.Delegate
	storer        Storer
}

// NewBusiness constructs a Business for objetivo domain.
func NewBusiness(
	log *logger.Logger,
	lifeVisionBus lifevisionbus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
) ExtBusiness {
	b := &Business{
		log:           log,
		lifeVisionBus: lifeVisionBus,
		delegate:      dlg,
		storer:        storer,
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

	lifeVisionBus, err := b.lifeVisionBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	// Create business without re-registering delegate functions
	// to avoid duplicate handler registration on the shared delegate.
	return &Business{
		log:           b.log,
		lifeVisionBus: lifeVisionBus,
		delegate:      b.delegate,
		storer:        storer,
	}, nil
}

// Create adds a new objetivo.
func (b *Business) Create(ctx context.Context, no NewObjetivo) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.create", "userID", no.UserID, "lifeVisionID", no.LifeVisionID)

	// Validate parent life vision exists
	lifeVision, err := b.lifeVisionBus.QueryByID(ctx, no.LifeVisionID)
	if err != nil {
		return Objetivo{}, fmt.Errorf("lifevision.querybyid: lifeVisionID[%s]: %w", no.LifeVisionID, err)
	}

	// Security: Verify authenticated user owns the life vision
	if lifeVision.UserID != no.UserID {
		return Objetivo{}, ErrNotLifeVisionOwner
	}

	// Verify life vision is active
	if !lifeVision.Status.IsActive() {
		return Objetivo{}, ErrTargetLifeVisionNotActive
	}

	// Validate tracking type configuration
	if err := validateTrackingConfig(no); err != nil {
		return Objetivo{}, err
	}

	now := time.Now().UTC()

	objetivo := Objetivo{
		ID:                    uuid.New(),
		UserID:                lifeVision.UserID,
		LifeVisionID:          no.LifeVisionID,
		Titulo:                no.Titulo,
		TipoTracking:          no.TipoTracking,
		Status:                objetivostatus.Activo,
		MetricaObjetivo:       no.MetricaObjetivo,
		MetricaActual:         nil,
		FrecuenciaTipo:        no.FrecuenciaTipo,
		FrecuenciaN:           no.FrecuenciaN,
		CumplimientoTargetPct: no.CumplimientoTargetPct,
		ArchivedAt:            nil,
		DateCreated:           now,
		DateUpdated:           now,
	}

	// Initialize MetricaActual to 0 for resultado tracking
	if no.TipoTracking.IsResultado() {
		zero := metricaactual.Zero()
		objetivo.MetricaActual = &zero
	}

	if err := b.storer.Create(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.create", "err", err)
		return Objetivo{}, fmt.Errorf("create: %w", err)
	}

	b.log.Info(ctx, "objetivobus.create.success", "objetivoID", objetivo.ID)
	return objetivo, nil
}

// Update modifies an existing objetivo.
func (b *Business) Update(ctx context.Context, objetivo Objetivo, uo UpdateObjetivo) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.update", "objetivoID", objetivo.ID)

	// Status cannot be changed via Update (use ChangeStatus)
	if uo.Status != nil {
		return Objetivo{}, errors.New("status changes must use ChangeStatus method")
	}

	if uo.Titulo != nil {
		objetivo.Titulo = *uo.Titulo
	}

	if uo.MetricaObjetivo != nil {
		if !objetivo.TipoTracking.IsResultado() {
			return Objetivo{}, errors.New("metrica_objetivo only valid for resultado tracking")
		}
		// Validation (> 0) is handled by metricaobjetivo.Parse at app layer
		objetivo.MetricaObjetivo = uo.MetricaObjetivo
	}

	if uo.FrecuenciaTipo != nil {
		if !objetivo.TipoTracking.IsFrecuencia() {
			return Objetivo{}, errors.New("frecuencia_tipo only valid for frecuencia tracking")
		}
		objetivo.FrecuenciaTipo = uo.FrecuenciaTipo
	}

	if uo.FrecuenciaN != nil {
		if !objetivo.TipoTracking.IsFrecuencia() {
			return Objetivo{}, errors.New("frecuencia_n only valid for frecuencia tracking")
		}
		// Validation (>= 1) is handled by frecuencian.Parse at app layer
		objetivo.FrecuenciaN = uo.FrecuenciaN
	}

	if uo.CumplimientoTargetPct != nil {
		if !objetivo.TipoTracking.IsFrecuencia() {
			return Objetivo{}, errors.New("cumplimiento_target_pct only valid for frecuencia tracking")
		}
		objetivo.CumplimientoTargetPct = uo.CumplimientoTargetPct
	}

	objetivo.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.update", "err", err)
		return Objetivo{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivobus.update.success", "objetivoID", objetivo.ID)
	return objetivo, nil
}

// Delete removes an objetivo.
func (b *Business) Delete(ctx context.Context, objetivo Objetivo) error {
	b.log.Info(ctx, "objetivobus.delete", "objetivoID", objetivo.ID)

	if err := b.storer.Delete(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.delete", "err", err)
		return fmt.Errorf("delete: %w", err)
	}

	// Publish event for child domains (objetivoregistrobus)
	if b.delegate != nil {
		if err := b.delegate.Call(ctx, ActionDeletedData(objetivo)); err != nil {
			return fmt.Errorf("delegate call: %w", err)
		}
	}

	b.log.Info(ctx, "objetivobus.delete.success", "objetivoID", objetivo.ID)
	return nil
}

// Archive sets an objetivo's archived_at timestamp (soft delete).
func (b *Business) Archive(ctx context.Context, objetivo Objetivo) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.archive", "objetivoID", objetivo.ID)

	if objetivo.ArchivedAt != nil {
		return Objetivo{}, ErrAlreadyArchived
	}

	// Cannot archive terminal status objetivos
	if objetivo.Status.IsTerminal() {
		return Objetivo{}, ErrTerminalStatusNoArchive
	}

	now := time.Now().UTC()
	objetivo.ArchivedAt = &now
	objetivo.DateUpdated = now

	if err := b.storer.Update(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.archive", "err", err)
		return Objetivo{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivobus.archive.success", "objetivoID", objetivo.ID)
	return objetivo, nil
}

// Restore removes the archived_at timestamp from an objetivo.
func (b *Business) Restore(ctx context.Context, objetivo Objetivo) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.restore", "objetivoID", objetivo.ID)

	if objetivo.ArchivedAt == nil {
		return Objetivo{}, ErrNotArchived
	}

	// Validate parent life vision is still active before restoring
	lifeVision, err := b.lifeVisionBus.QueryByID(ctx, objetivo.LifeVisionID)
	if err != nil {
		return Objetivo{}, fmt.Errorf("lifevision.querybyid: lifeVisionID[%s]: %w", objetivo.LifeVisionID, err)
	}

	if !lifeVision.Status.IsActive() {
		return Objetivo{}, ErrTargetLifeVisionNotActive
	}

	objetivo.ArchivedAt = nil
	objetivo.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.restore", "err", err)
		return Objetivo{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivobus.restore.success", "objetivoID", objetivo.ID)
	return objetivo, nil
}

// ChangeStatus changes the status of an objetivo with transition rules.
func (b *Business) ChangeStatus(ctx context.Context, objetivo Objetivo, req ChangeStatusRequest) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.changestatus", "objetivoID", objetivo.ID, "from", objetivo.Status, "to", req.Status)

	// Validate status transition
	if err := validateStatusTransition(objetivo.Status, req.Status); err != nil {
		return Objetivo{}, err
	}

	objetivo.Status = req.Status
	objetivo.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.changestatus", "err", err)
		return Objetivo{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivobus.changestatus.success", "objetivoID", objetivo.ID)
	return objetivo, nil
}

// IncrementProgress increments or decrements MetricaActual for resultado tracking (MVP).
func (b *Business) IncrementProgress(ctx context.Context, objetivo Objetivo, req IncrementProgressRequest) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.incrementprogress", "objetivoID", objetivo.ID, "increment", req.Increment)

	if !objetivo.TipoTracking.IsResultado() {
		return Objetivo{}, ErrOnlyResultadoAllowsProgress
	}

	if objetivo.MetricaActual == nil || objetivo.MetricaObjetivo == nil {
		return Objetivo{}, errors.New("metrica values not initialized")
	}

	newValue := objetivo.MetricaActual.Value() + req.Increment
	if newValue < 0 {
		newValue = 0
	}
	if newValue > objetivo.MetricaObjetivo.Value() {
		return Objetivo{}, ErrProgressExceedsTarget
	}

	// Parse validates non-negative (which we've already ensured above)
	newMetricaActual, err := metricaactual.Parse(newValue)
	if err != nil {
		return Objetivo{}, fmt.Errorf("parse metrica_actual: %w", err)
	}
	objetivo.MetricaActual = &newMetricaActual
	objetivo.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.incrementprogress", "err", err)
		return Objetivo{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivobus.incrementprogress.success", "objetivoID", objetivo.ID, "newValue", newValue)
	return objetivo, nil
}

// UpdateProgress updates MetricaActual for resultado tracking (supports both increment and direct value).
func (b *Business) UpdateProgress(ctx context.Context, objetivo Objetivo, req UpdateProgressRequest) (Objetivo, error) {
	b.log.Info(ctx, "objetivobus.updateprogress", "objetivoID", objetivo.ID, "increment", req.Increment, "value", req.Value)

	if !objetivo.TipoTracking.IsResultado() {
		return Objetivo{}, ErrOnlyResultadoAllowsProgress
	}

	if objetivo.MetricaActual == nil || objetivo.MetricaObjetivo == nil {
		return Objetivo{}, errors.New("metrica values not initialized")
	}

	var newValue int
	if req.Value != nil {
		// Direct value set
		newValue = *req.Value
	} else if req.Increment != nil {
		// Increment mode
		newValue = objetivo.MetricaActual.Value() + *req.Increment
	} else {
		return Objetivo{}, errors.New("either increment or value must be provided")
	}

	if newValue < 0 {
		newValue = 0
	}
	if newValue > objetivo.MetricaObjetivo.Value() {
		return Objetivo{}, ErrProgressExceedsTarget
	}

	// Parse validates non-negative (which we've already ensured above)
	newMetricaActual, err := metricaactual.Parse(newValue)
	if err != nil {
		return Objetivo{}, fmt.Errorf("parse metrica_actual: %w", err)
	}
	objetivo.MetricaActual = &newMetricaActual
	objetivo.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objetivo); err != nil {
		b.log.Error(ctx, "objetivobus.updateprogress", "err", err)
		return Objetivo{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objetivobus.updateprogress.success", "objetivoID", objetivo.ID, "newValue", newValue)
	return objetivo, nil
}

// Query retrieves objetivos based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Objetivo, error) {
	objetivos, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return objetivos, nil
}

// QueryByID finds an objetivo by its ID.
func (b *Business) QueryByID(ctx context.Context, objetivoID uuid.UUID) (Objetivo, error) {
	objetivo, err := b.storer.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Objetivo{}, ErrNotFound
		}
		return Objetivo{}, fmt.Errorf("query: objetivoID[%s]: %w", objetivoID, err)
	}

	return objetivo, nil
}

// Count returns the total number of objetivos matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// validateTrackingConfig validates the tracking type configuration.
func validateTrackingConfig(no NewObjetivo) error {
	if no.TipoTracking.IsResultado() {
		// MetricaObjetivo must be provided and > 0 (validated by metricaobjetivo.Parse)
		if no.MetricaObjetivo == nil {
			return ErrInvalidResultadoConfig
		}
	}

	if no.TipoTracking.IsFrecuencia() {
		if no.FrecuenciaTipo == nil {
			return errors.New("frecuencia_tipo required for frecuencia tracking")
		}

		// For n_por_semana and n_por_mes, frecuencia_n must be provided (>= 1 validated by frecuencian.Parse)
		if !no.FrecuenciaTipo.IsDaily() {
			if no.FrecuenciaN == nil {
				return ErrInvalidFrecuenciaConfig
			}
		}

		if no.CumplimientoTargetPct == nil {
			return ErrMissingCumplimientoPct
		}
	}

	return nil
}

// validateStatusTransition validates if a status transition is allowed.
func validateStatusTransition(from, to objetivostatus.ObjetivoStatus) error {
	// Cannot transition to the same status (no-op)
	if from.Equal(to) {
		return ErrStatusTransitionNotAllowed
	}

	// Cannot transition from terminal statuses
	if from.IsTerminal() {
		return ErrStatusTransitionNotAllowed
	}

	// Valid transitions:
	// activo → completado, pausado, abandonado
	// pausado → activo, abandonado
	if from.IsActivo() {
		// Can go anywhere except stay activo (already checked above)
		return nil
	}

	if from.IsPausado() {
		// Can only go to activo or abandonado
		if to.IsActivo() || to.Equal(objetivostatus.Abandonado) {
			return nil
		}
		return ErrStatusTransitionNotAllowed
	}

	return ErrStatusTransitionNotAllowed
}
