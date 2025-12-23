package objectivebus

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
	"github.com/francowini/rafiki/business/types/currentmetric"
	"github.com/francowini/rafiki/business/types/objectivestatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, objective Objective) error
	Update(ctx context.Context, objective Objective) error
	Delete(ctx context.Context, objective Objective) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Objective, error)
	QueryByID(ctx context.Context, objectiveID uuid.UUID) (Objective, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, no NewObjective) (Objective, error)
	Update(ctx context.Context, objective Objective, uo UpdateObjective) (Objective, error)
	Delete(ctx context.Context, objective Objective) error
	Archive(ctx context.Context, objective Objective) (Objective, error)
	Restore(ctx context.Context, objective Objective) (Objective, error)
	ChangeStatus(ctx context.Context, objective Objective, req ChangeStatusRequest) (Objective, error)
	IncrementProgress(ctx context.Context, objective Objective, req IncrementProgressRequest) (Objective, error)
	UpdateProgress(ctx context.Context, objective Objective, req UpdateProgressRequest) (Objective, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Objective, error)
	QueryByID(ctx context.Context, objectiveID uuid.UUID) (Objective, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages objective operations.
type Business struct {
	log           *logger.Logger
	lifeVisionBus lifevisionbus.ExtBusiness
	delegate      *delegate.Delegate
	storer        Storer
}

// NewBusiness constructs a Business for objective domain.
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

// Create adds a new objective.
func (b *Business) Create(ctx context.Context, no NewObjective) (Objective, error) {
	b.log.Info(ctx, "objectivebus.create", "userID", no.UserID, "lifeVisionID", no.LifeVisionID)

	// Validate parent life vision exists
	lifeVision, err := b.lifeVisionBus.QueryByID(ctx, no.LifeVisionID)
	if err != nil {
		return Objective{}, fmt.Errorf("lifevision.querybyid: lifeVisionID[%s]: %w", no.LifeVisionID, err)
	}

	// Security: Verify authenticated user owns the life vision
	if lifeVision.UserID != no.UserID {
		return Objective{}, ErrNotLifeVisionOwner
	}

	// Verify life vision is active
	if !lifeVision.Status.IsActive() {
		return Objective{}, ErrTargetLifeVisionNotActive
	}

	// Validate tracking type configuration
	if err := validateTrackingConfig(no); err != nil {
		return Objective{}, err
	}

	now := time.Now().UTC()

	objective := Objective{
		ID:                  uuid.New(),
		UserID:              lifeVision.UserID,
		LifeVisionID:        no.LifeVisionID,
		Title:               no.Title,
		TrackingType:        no.TrackingType,
		Status:              objectivestatus.Active,
		TargetMetric:        no.TargetMetric,
		CurrentMetric:       nil,
		BeginMetric:         no.BeginMetric,
		FrequencyType:       no.FrequencyType,
		FrequencyCount:      no.FrequencyCount,
		ComplianceTargetPct: no.ComplianceTargetPct,
		ArchivedAt:          nil,
		DateCreated:         now,
		DateUpdated:         now,
	}

	// Initialize CurrentMetric for result tracking
	if no.TrackingType.IsResult() {
		// For inverse goals (begin > target), start at begin value
		// For normal goals, start at 0
		var initial currentmetric.CurrentMetric
		if no.BeginMetric != nil && no.TargetMetric != nil && no.BeginMetric.Value() > no.TargetMetric.Value() {
			// Decrease goal: start at begin value
			var err error
			initial, err = currentmetric.Parse(no.BeginMetric.Value())
			if err != nil {
				// BeginMetric was already validated at app layer, but log and use safe default
				b.log.Error(ctx, "objectivebus.create.parsecurrentmetric", "beginMetricValue", no.BeginMetric.Value(), "err", err)
				initial = currentmetric.Zero()
			}
		} else {
			// Increase goal: start at 0
			initial = currentmetric.Zero()
		}
		objective.CurrentMetric = &initial
	}

	if err := b.storer.Create(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.create", "err", err)
		return Objective{}, fmt.Errorf("create: %w", err)
	}

	b.log.Info(ctx, "objectivebus.create.success", "objectiveID", objective.ID)
	return objective, nil
}

// Update modifies an existing objective.
func (b *Business) Update(ctx context.Context, objective Objective, uo UpdateObjective) (Objective, error) {
	b.log.Info(ctx, "objectivebus.update", "objectiveID", objective.ID)

	// Status cannot be changed via Update (use ChangeStatus)
	if uo.Status != nil {
		return Objective{}, ErrStatusChangeMustUseMethod
	}

	if uo.Title != nil {
		objective.Title = *uo.Title
	}

	if uo.TargetMetric != nil {
		if !objective.TrackingType.IsResult() {
			return Objective{}, ErrMetricOnlyForResult
		}
		// Validation (> 0) is handled by targetmetric.Parse at app layer
		objective.TargetMetric = uo.TargetMetric
	}

	// Handle BeginMetric update (ClearBeginMetric takes precedence)
	if uo.ClearBeginMetric {
		objective.BeginMetric = nil
	} else if uo.BeginMetric != nil {
		if !objective.TrackingType.IsResult() {
			return Objective{}, ErrMetricOnlyForResult
		}
		// Validation (>= 0) is handled by beginmetric.Parse at app layer
		objective.BeginMetric = uo.BeginMetric
	}

	if uo.FrequencyType != nil {
		if !objective.TrackingType.IsFrequency() {
			return Objective{}, ErrFrequencyOnlyForFrequency
		}
		objective.FrequencyType = uo.FrequencyType
	}

	if uo.FrequencyCount != nil {
		if !objective.TrackingType.IsFrequency() {
			return Objective{}, ErrFrequencyOnlyForFrequency
		}
		// Validation (>= 1) is handled by frequencycount.Parse at app layer
		objective.FrequencyCount = uo.FrequencyCount
	}

	if uo.ComplianceTargetPct != nil {
		if !objective.TrackingType.IsFrequency() {
			return Objective{}, ErrFrequencyOnlyForFrequency
		}
		objective.ComplianceTargetPct = uo.ComplianceTargetPct
	}

	objective.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.update", "err", err)
		return Objective{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectivebus.update.success", "objectiveID", objective.ID)
	return objective, nil
}

// Delete removes an objective.
func (b *Business) Delete(ctx context.Context, objective Objective) error {
	b.log.Info(ctx, "objectivebus.delete", "objectiveID", objective.ID)

	if err := b.storer.Delete(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.delete", "err", err)
		return fmt.Errorf("delete: %w", err)
	}

	// Publish event for child domains (objectiverecordbus).
	// Delegate handlers are idempotent, so failures are logged but do not abort the delete flow.
	if b.delegate != nil {
		if err := b.delegate.Call(ctx, ActionDeletedData(objective)); err != nil {
			b.log.Error(ctx, "objectivebus.delete.delegate", "objectiveID", objective.ID, "err", err)
		}
	}

	b.log.Info(ctx, "objectivebus.delete.success", "objectiveID", objective.ID)
	return nil
}

// Archive sets an objective's archived_at timestamp (soft delete).
func (b *Business) Archive(ctx context.Context, objective Objective) (Objective, error) {
	b.log.Info(ctx, "objectivebus.archive", "objectiveID", objective.ID)

	if objective.ArchivedAt != nil {
		return Objective{}, ErrAlreadyArchived
	}

	// Cannot archive terminal status objectives
	if objective.Status.IsTerminal() {
		return Objective{}, ErrTerminalStatusNoArchive
	}

	now := time.Now().UTC()
	objective.ArchivedAt = &now
	objective.DateUpdated = now

	if err := b.storer.Update(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.archive", "err", err)
		return Objective{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectivebus.archive.success", "objectiveID", objective.ID)
	return objective, nil
}

// Restore removes the archived_at timestamp from an objective.
func (b *Business) Restore(ctx context.Context, objective Objective) (Objective, error) {
	b.log.Info(ctx, "objectivebus.restore", "objectiveID", objective.ID)

	if objective.ArchivedAt == nil {
		return Objective{}, ErrNotArchived
	}

	// Validate parent life vision is still active before restoring
	lifeVision, err := b.lifeVisionBus.QueryByID(ctx, objective.LifeVisionID)
	if err != nil {
		return Objective{}, fmt.Errorf("lifevision.querybyid: lifeVisionID[%s]: %w", objective.LifeVisionID, err)
	}

	if !lifeVision.Status.IsActive() {
		return Objective{}, ErrTargetLifeVisionNotActive
	}

	objective.ArchivedAt = nil
	objective.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.restore", "err", err)
		return Objective{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectivebus.restore.success", "objectiveID", objective.ID)
	return objective, nil
}

// ChangeStatus changes the status of an objective with transition rules.
func (b *Business) ChangeStatus(ctx context.Context, objective Objective, req ChangeStatusRequest) (Objective, error) {
	b.log.Info(ctx, "objectivebus.changestatus", "objectiveID", objective.ID, "from", objective.Status, "to", req.Status)

	// Validate status transition
	if err := validateStatusTransition(objective.Status, req.Status); err != nil {
		return Objective{}, err
	}

	objective.Status = req.Status
	objective.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.changestatus", "err", err)
		return Objective{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectivebus.changestatus.success", "objectiveID", objective.ID)
	return objective, nil
}

// IncrementProgress increments or decrements CurrentMetric for result tracking (MVP).
func (b *Business) IncrementProgress(ctx context.Context, objective Objective, req IncrementProgressRequest) (Objective, error) {
	b.log.Info(ctx, "objectivebus.incrementprogress", "objectiveID", objective.ID, "increment", req.Increment)

	if !objective.TrackingType.IsResult() {
		return Objective{}, ErrOnlyResultAllowsProgress
	}

	if objective.CurrentMetric == nil || objective.TargetMetric == nil {
		return Objective{}, ErrMetricNotInitialized
	}

	newValue := objective.CurrentMetric.Value() + req.Increment
	if newValue < 0 {
		newValue = 0
	}

	// Direction-aware validation: only block exceeding target for increasing goals
	if err := validateProgressNotExceedsTarget(objective, newValue); err != nil {
		return Objective{}, err
	}

	// Parse validates non-negative (which we've already ensured above)
	newCurrentMetric, err := currentmetric.Parse(newValue)
	if err != nil {
		return Objective{}, fmt.Errorf("parse current_metric: %w", err)
	}
	objective.CurrentMetric = &newCurrentMetric
	objective.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.incrementprogress", "err", err)
		return Objective{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectivebus.incrementprogress.success", "objectiveID", objective.ID, "newValue", newValue)
	return objective, nil
}

// UpdateProgress updates CurrentMetric for result tracking (supports both increment and direct value).
func (b *Business) UpdateProgress(ctx context.Context, objective Objective, req UpdateProgressRequest) (Objective, error) {
	b.log.Info(ctx, "objectivebus.updateprogress", "objectiveID", objective.ID, "increment", req.Increment, "value", req.Value)

	if !objective.TrackingType.IsResult() {
		return Objective{}, ErrOnlyResultAllowsProgress
	}

	if objective.CurrentMetric == nil || objective.TargetMetric == nil {
		return Objective{}, ErrMetricNotInitialized
	}

	var newValue int
	if req.Value != nil {
		// Direct value set
		newValue = *req.Value
	} else if req.Increment != nil {
		// Increment mode
		newValue = objective.CurrentMetric.Value() + *req.Increment
	} else {
		return Objective{}, ErrIncrementOrValueRequired
	}

	if newValue < 0 {
		newValue = 0
	}

	// Direction-aware validation: only block exceeding target for increasing goals
	if err := validateProgressNotExceedsTarget(objective, newValue); err != nil {
		return Objective{}, err
	}

	// Parse validates non-negative (which we've already ensured above)
	newCurrentMetric, err := currentmetric.Parse(newValue)
	if err != nil {
		return Objective{}, fmt.Errorf("parse current_metric: %w", err)
	}
	objective.CurrentMetric = &newCurrentMetric
	objective.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, objective); err != nil {
		b.log.Error(ctx, "objectivebus.updateprogress", "err", err)
		return Objective{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "objectivebus.updateprogress.success", "objectiveID", objective.ID, "newValue", newValue)
	return objective, nil
}

// Query retrieves objectives based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Objective, error) {
	objectives, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return objectives, nil
}

// QueryByID finds an objective by its ID.
func (b *Business) QueryByID(ctx context.Context, objectiveID uuid.UUID) (Objective, error) {
	objective, err := b.storer.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Objective{}, ErrNotFound
		}
		return Objective{}, fmt.Errorf("query: objectiveID[%s]: %w", objectiveID, err)
	}

	return objective, nil
}

// Count returns the total number of objectives matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// validateTrackingConfig validates the tracking type configuration.
func validateTrackingConfig(no NewObjective) error {
	if no.TrackingType.IsResult() {
		// TargetMetric must be provided and > 0 (validated by targetmetric.Parse)
		if no.TargetMetric == nil {
			return ErrInvalidResultConfig
		}
	}

	if no.TrackingType.IsFrequency() {
		if no.FrequencyType == nil {
			return ErrFrequencyTypeRequired
		}

		// For n_per_week and n_per_month, frequency_count must be provided (>= 1 validated by frequencycount.Parse)
		if !no.FrequencyType.IsDaily() {
			if no.FrequencyCount == nil {
				return ErrInvalidFrequencyConfig
			}
		}

		if no.ComplianceTargetPct == nil {
			return ErrMissingCompliancePct
		}
	}

	return nil
}

// validateStatusTransition validates if a status transition is allowed.
func validateStatusTransition(from, to objectivestatus.ObjectiveStatus) error {
	// Cannot transition to the same status (no-op)
	if from.Equal(to) {
		return ErrStatusTransitionNotAllowed
	}

	// Cannot transition from terminal statuses
	if from.IsTerminal() {
		return ErrStatusTransitionNotAllowed
	}

	// Valid transitions:
	// active → completed, paused, abandoned
	// paused → active, abandoned
	if from.IsActive() {
		// Can go anywhere except stay active (already checked above)
		return nil
	}

	if from.IsPaused() {
		// Can only go to active or abandoned
		if to.IsActive() || to.Equal(objectivestatus.Abandoned) {
			return nil
		}
		return ErrStatusTransitionNotAllowed
	}

	return ErrStatusTransitionNotAllowed
}

// isInverseGoal returns true if this is a decrease goal (begin > target).
// For example: reducing weight from 80kg to 70kg.
func isInverseGoal(obj Objective) bool {
	return obj.BeginMetric != nil && obj.TargetMetric != nil && obj.BeginMetric.Value() > obj.TargetMetric.Value()
}

// validateProgressNotExceedsTarget checks if the new progress value is valid.
// For increasing goals: newValue must not exceed target.
// For inverse goals: newValue CAN exceed target (that's 0-99% progress).
func validateProgressNotExceedsTarget(obj Objective, newValue int) error {
	if !isInverseGoal(obj) && obj.TargetMetric != nil && newValue > obj.TargetMetric.Value() {
		return ErrProgressExceedsTarget
	}
	return nil
}
