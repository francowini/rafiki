package taskbus

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
	"github.com/francowini/rafiki/business/types/contribution"
	"github.com/francowini/rafiki/business/types/taskstatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, task Task) error
	Update(ctx context.Context, task Task) error
	Delete(ctx context.Context, task Task) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Task, error)
	QueryByID(ctx context.Context, taskID uuid.UUID) (Task, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// ExtBusiness interface provides support for extensions that wrap extra functionality
// around the core business logic.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nt NewTask) (Task, error)
	Update(ctx context.Context, task Task, ut UpdateTask) (Task, error)
	Delete(ctx context.Context, task Task) error
	Complete(ctx context.Context, task Task) (Task, error)
	Uncomplete(ctx context.Context, task Task) (Task, error)
	Cancel(ctx context.Context, task Task) (Task, error)
	MoveTask(ctx context.Context, task Task, targetObjectiveID uuid.UUID, cont *contribution.Contribution) (Task, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Task, error)
	QueryByID(ctx context.Context, taskID uuid.UUID) (Task, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages task operations.
type Business struct {
	log          *logger.Logger
	objectiveBus objectivebus.ExtBusiness
	delegate     *delegate.Delegate
	storer       Storer
}

// NewBusiness constructs a Business for task domain.
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

	// Register delegate functions on the root business instance.
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

// Create adds a new task.
func (b *Business) Create(ctx context.Context, nt NewTask) (Task, error) {
	b.log.Info(ctx, "taskbus.create", "userID", nt.UserID, "objectiveID", nt.ObjectiveID)

	// Inbox tasks (no objective) cannot have a contribution
	if nt.ObjectiveID == nil && nt.Contribution != nil {
		return Task{}, ErrContributionNotAllowedInbox
	}

	// Validate objective link if provided
	if nt.ObjectiveID != nil {
		objective, err := b.objectiveBus.QueryByID(ctx, *nt.ObjectiveID)
		if err != nil {
			if errors.Is(err, objectivebus.ErrNotFound) {
				return Task{}, ErrObjectiveNotFound
			}
			return Task{}, fmt.Errorf("objective.querybyid: objectiveID[%s]: %w", *nt.ObjectiveID, err)
		}

		// Security: Verify authenticated user owns the objective
		if objective.UserID != nt.UserID {
			return Task{}, ErrNotObjectiveOwner
		}

		// RESULT objectives: contribution REQUIRED (1-10)
		if objective.TrackingType.IsResult() {
			if nt.Contribution == nil {
				return Task{}, ErrContributionRequiredForLink
			}
		}

		// FREQUENCY objectives: contribution must be NULL
		if objective.TrackingType.IsFrequency() {
			if nt.Contribution != nil {
				return Task{}, ErrFrequencyTasksNoContribution
			}
		}
	}

	now := time.Now().UTC()

	task := Task{
		ID:           uuid.New(),
		UserID:       nt.UserID,
		ObjectiveID:  nt.ObjectiveID,
		Title:        nt.Title,
		Description:  nt.Description,
		Contribution: nt.Contribution,
		Status:       taskstatus.StatusPending,
		CompletedAt:  nil,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := b.storer.Create(ctx, task); err != nil {
		b.log.Error(ctx, "taskbus.create", "err", err)
		return Task{}, fmt.Errorf("create: %w", err)
	}

	b.log.Info(ctx, "taskbus.create.success", "taskID", task.ID)
	return task, nil
}

// Update modifies an existing task.
func (b *Business) Update(ctx context.Context, task Task, ut UpdateTask) (Task, error) {
	b.log.Info(ctx, "taskbus.update", "taskID", task.ID)

	// Cannot update completed or canceled tasks
	if task.Status.IsCompleted() || task.Status.IsCanceled() {
		return Task{}, ErrCannotUpdateTerminalTask
	}

	if ut.Title != nil {
		task.Title = *ut.Title
	}

	// ClearDescription takes precedence: if true, set Description to nil
	if ut.ClearDescription {
		task.Description = nil
	} else if ut.Description != nil {
		task.Description = ut.Description
	}

	if ut.Contribution != nil {
		// Only allow contribution update for objective-linked tasks
		if task.ObjectiveID == nil {
			return Task{}, ErrCannotSetContributionOnInbox
		}
		task.Contribution = ut.Contribution
	}

	task.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, task); err != nil {
		b.log.Error(ctx, "taskbus.update", "err", err)
		return Task{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "taskbus.update.success", "taskID", task.ID)
	return task, nil
}

// Delete removes a task.
func (b *Business) Delete(ctx context.Context, task Task) error {
	b.log.Info(ctx, "taskbus.delete", "taskID", task.ID)

	if err := b.storer.Delete(ctx, task); err != nil {
		b.log.Error(ctx, "taskbus.delete", "err", err)
		return fmt.Errorf("delete: %w", err)
	}

	b.log.Info(ctx, "taskbus.delete.success", "taskID", task.ID)
	return nil
}

// Complete marks a task as completed.
// For objective-linked tasks, this should be called within a transaction that also updates objective progress.
func (b *Business) Complete(ctx context.Context, task Task) (Task, error) {
	b.log.Info(ctx, "taskbus.complete", "taskID", task.ID)

	if task.Status.IsCompleted() {
		return Task{}, ErrAlreadyCompleted
	}

	if task.Status.IsCanceled() {
		return Task{}, ErrAlreadyCanceled
	}

	now := time.Now().UTC()
	task.Status = taskstatus.StatusCompleted
	task.CompletedAt = &now
	task.DateUpdated = now

	if err := b.storer.Update(ctx, task); err != nil {
		b.log.Error(ctx, "taskbus.complete", "err", err)
		return Task{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "taskbus.complete.success", "taskID", task.ID)
	return task, nil
}

// Uncomplete reverses task completion (only for objective-linked tasks).
// For objective-linked tasks, this should be called within a transaction that also updates objective progress.
func (b *Business) Uncomplete(ctx context.Context, task Task) (Task, error) {
	b.log.Info(ctx, "taskbus.uncomplete", "taskID", task.ID)

	if !task.Status.IsCompleted() {
		return Task{}, ErrNotCompleted
	}

	// Inbox tasks cannot be uncompleted (no objective progress to reverse)
	if task.ObjectiveID == nil {
		return Task{}, ErrCannotUncompleteInboxTask
	}

	task.Status = taskstatus.StatusPending
	task.CompletedAt = nil
	task.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, task); err != nil {
		b.log.Error(ctx, "taskbus.uncomplete", "err", err)
		return Task{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "taskbus.uncomplete.success", "taskID", task.ID)
	return task, nil
}

// Cancel marks a task as canceled.
func (b *Business) Cancel(ctx context.Context, task Task) (Task, error) {
	b.log.Info(ctx, "taskbus.cancel", "taskID", task.ID)

	if task.Status.IsCanceled() {
		return Task{}, ErrAlreadyCanceled
	}

	if task.Status.IsCompleted() {
		return Task{}, ErrCannotCancelCompletedTask
	}

	task.Status = taskstatus.StatusCanceled
	task.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, task); err != nil {
		b.log.Error(ctx, "taskbus.cancel", "err", err)
		return Task{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "taskbus.cancel.success", "taskID", task.ID)
	return task, nil
}

// MoveTask relocates an inbox task to an objective. This creates a new task
// with the target objective and deletes the original. This pattern preserves
// parent-reference immutability and maintains audit trails.
func (b *Business) MoveTask(ctx context.Context, task Task, targetObjectiveID uuid.UUID, cont *contribution.Contribution) (Task, error) {
	b.log.Info(ctx, "taskbus.move", "taskID", task.ID, "targetObjectiveID", targetObjectiveID)

	// Can only move inbox tasks
	if task.ObjectiveID != nil {
		b.log.Error(ctx, "taskbus.move", "err", "task already linked to objective")
		return Task{}, ErrTaskAlreadyLinked
	}

	// Validate target objective
	objective, err := b.objectiveBus.QueryByID(ctx, targetObjectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			b.log.Error(ctx, "taskbus.move.objective_not_found", "objectiveID", targetObjectiveID)
			return Task{}, ErrObjectiveNotFound
		}
		b.log.Error(ctx, "taskbus.move.query_objective", "err", err, "objectiveID", targetObjectiveID)
		return Task{}, fmt.Errorf("objective.querybyid: objectiveID[%s]: %w", targetObjectiveID, err)
	}

	// Security: Verify authenticated user owns both the task and target objective
	if objective.UserID != task.UserID {
		b.log.Error(ctx, "taskbus.move.ownership_violation", "taskUserID", task.UserID, "objectiveUserID", objective.UserID)
		return Task{}, ErrNotObjectiveOwner
	}

	// Validate contribution based on tracking type
	var contPtr *contribution.Contribution
	if objective.TrackingType.IsResult() {
		// Contribution required for result objectives
		if cont == nil {
			b.log.Error(ctx, "taskbus.move.contribution_required", "objectiveID", targetObjectiveID)
			return Task{}, ErrContributionRequiredForLink
		}
		contPtr = cont
	} else if objective.TrackingType.IsFrequency() {
		// Contribution not allowed for frequency objectives
		if cont != nil {
			b.log.Error(ctx, "taskbus.move.contribution_not_allowed", "objectiveID", targetObjectiveID)
			return Task{}, ErrFrequencyTasksNoContribution
		}
		// contPtr remains nil
	}

	// Create new task with target objective
	newTask := NewTask{
		UserID:       task.UserID,
		ObjectiveID:  &targetObjectiveID,
		Title:        task.Title,
		Description:  task.Description,
		Contribution: contPtr,
	}

	createdTask, err := b.Create(ctx, newTask)
	if err != nil {
		b.log.Error(ctx, "taskbus.move.create_failed", "err", err)
		return Task{}, fmt.Errorf("create new task: %w", err)
	}

	// Delete original inbox task
	if err := b.Delete(ctx, task); err != nil {
		// Delegate handler - log but don't fail since new task is already created
		b.log.Error(ctx, "taskbus.move.delete_original", "taskID", task.ID, "err", err)
	}

	b.log.Info(ctx, "taskbus.move.success", "oldTaskID", task.ID, "newTaskID", createdTask.ID)
	return createdTask, nil
}

// Query retrieves tasks based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Task, error) {
	tasks, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tasks, nil
}

// QueryByID finds a task by its ID.
func (b *Business) QueryByID(ctx context.Context, taskID uuid.UUID) (Task, error) {
	task, err := b.storer.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("query: taskID[%s]: %w", taskID, err)
	}

	return task, nil
}

// Count returns the total number of tasks matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
