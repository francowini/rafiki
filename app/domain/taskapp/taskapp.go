package taskapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/domain/taskbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	taskBus      taskbus.ExtBusiness
	objectiveBus objectivebus.ExtBusiness
	db           sqldb.Beginner
}

func newApp(taskBus taskbus.ExtBusiness, objectiveBus objectivebus.ExtBusiness, db sqldb.Beginner) *app {
	return &app{
		taskBus:      taskBus,
		objectiveBus: objectiveBus,
		db:           db,
	}
}

// create handles POST /v1/tasks
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appTask NewTask
	if err := web.Decode(r, &appTask); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appTask.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	nt, err := toBusNewTask(appTask, userID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err := a.taskBus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, taskbus.ErrObjectiveNotFound) {
			return errs.New(errs.NotFound, errors.New("objective not found"))
		}
		if errors.Is(err, taskbus.ErrNotObjectiveOwner) {
			return errs.New(errs.PermissionDenied, errors.New("user does not own the specified objective"))
		}
		if errors.Is(err, taskbus.ErrContributionRequiredForLink) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, taskbus.ErrOnlyResultAllowsContribution) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: %s", err)
	}

	return toAppTask(task)
}

// update handles PUT /v1/tasks/{task_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appTask UpdateTask
	if err := web.Decode(r, &appTask); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appTask.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, respErr := a.getTaskWithOwnership(ctx, r)
	if respErr != nil {
		return respErr
	}

	ut, err := toBusUpdateTask(appTask)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err = a.taskBus.Update(ctx, task, ut)
	if err != nil {
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppTask(task)
}

// delete handles DELETE /v1/tasks/{task_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	task, respErr := a.getTaskWithOwnership(ctx, r)
	if respErr != nil {
		return respErr
	}

	if err := a.taskBus.Delete(ctx, task); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// complete handles PUT /v1/tasks/{task_id}/complete
func (a *app) complete(ctx context.Context, r *http.Request) web.Encoder {
	task, respErr := a.getTaskWithOwnership(ctx, r)
	if respErr != nil {
		return respErr
	}

	// For inbox tasks, just complete without transaction
	if task.ObjectiveID == nil {
		task, err := a.taskBus.Complete(ctx, task)
		if err != nil {
			if errors.Is(err, taskbus.ErrAlreadyCompleted) {
				return errs.New(errs.FailedPrecondition, err)
			}
			if errors.Is(err, taskbus.ErrAlreadyCanceled) {
				return errs.New(errs.FailedPrecondition, err)
			}
			return errs.Newf(errs.Internal, "complete: %s", err)
		}
		return CompleteTaskResponse{
			Task: toAppTask(task),
		}
	}

	// For objective-linked tasks, use transaction to update progress
	tx, err := a.db.Begin()
	if err != nil {
		return errs.Newf(errs.Internal, "begin transaction: %s", err)
	}
	defer func() {
		//nolint:errcheck // Rollback error cannot be handled in defer after earlier error
		tx.Rollback()
	}()

	taskBusTx, err := a.taskBus.NewWithTx(tx)
	if err != nil {
		return errs.Newf(errs.Internal, "taskbus.newwithtx: %s", err)
	}

	objectiveBusTx, err := a.objectiveBus.NewWithTx(tx)
	if err != nil {
		return errs.Newf(errs.Internal, "objectivebus.newwithtx: %s", err)
	}

	// Complete the task
	task, err = taskBusTx.Complete(ctx, task)
	if err != nil {
		if errors.Is(err, taskbus.ErrAlreadyCompleted) {
			return errs.New(errs.FailedPrecondition, err)
		}
		if errors.Is(err, taskbus.ErrAlreadyCanceled) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "complete: %s", err)
	}

	// Get the objective and update progress
	objective, err := objectiveBusTx.QueryByID(ctx, *task.ObjectiveID)
	if err != nil {
		return errs.Newf(errs.Internal, "objective.querybyid: %s", err)
	}

	// Only update progress for result tracking type
	var objectiveProgress *int
	if objective.TrackingType.IsResult() && task.Contribution != nil {
		increment := task.Contribution.Value()
		objective, err = objectiveBusTx.UpdateProgress(ctx, objective, objectivebus.UpdateProgressRequest{
			Increment: &increment,
		})
		if err != nil {
			return errs.Newf(errs.Internal, "updateprogress: %s", err)
		}
		if objective.CurrentMetric != nil {
			v := objective.CurrentMetric.Value()
			objectiveProgress = &v
		}
	}

	if err := tx.Commit(); err != nil {
		return errs.Newf(errs.Internal, "commit: %s", err)
	}

	return CompleteTaskResponse{
		Task:              toAppTask(task),
		ObjectiveProgress: objectiveProgress,
	}
}

// uncomplete handles PUT /v1/tasks/{task_id}/uncomplete
func (a *app) uncomplete(ctx context.Context, r *http.Request) web.Encoder {
	task, respErr := a.getTaskWithOwnership(ctx, r)
	if respErr != nil {
		return respErr
	}

	// For objective-linked tasks, use transaction to reverse progress
	if task.ObjectiveID != nil {
		tx, err := a.db.Begin()
		if err != nil {
			return errs.Newf(errs.Internal, "begin transaction: %s", err)
		}
		defer func() {
			//nolint:errcheck // Rollback error cannot be handled in defer after earlier error
			tx.Rollback()
		}()

		taskBusTx, err := a.taskBus.NewWithTx(tx)
		if err != nil {
			return errs.Newf(errs.Internal, "taskbus.newwithtx: %s", err)
		}

		objectiveBusTx, err := a.objectiveBus.NewWithTx(tx)
		if err != nil {
			return errs.Newf(errs.Internal, "objectivebus.newwithtx: %s", err)
		}

		// Get the objective and reverse progress first
		objective, err := objectiveBusTx.QueryByID(ctx, *task.ObjectiveID)
		if err != nil {
			return errs.Newf(errs.Internal, "objective.querybyid: %s", err)
		}

		// Reverse progress for result tracking type
		if objective.TrackingType.IsResult() && task.Contribution != nil {
			decrement := -task.Contribution.Value()
			_, err = objectiveBusTx.UpdateProgress(ctx, objective, objectivebus.UpdateProgressRequest{
				Increment: &decrement,
			})
			if err != nil && !errors.Is(err, objectivebus.ErrProgressExceedsTarget) {
				return errs.Newf(errs.Internal, "updateprogress: %s", err)
			}
		}

		// Uncomplete the task
		task, err = taskBusTx.Uncomplete(ctx, task)
		if err != nil {
			if errors.Is(err, taskbus.ErrNotCompleted) {
				return errs.New(errs.FailedPrecondition, err)
			}
			if errors.Is(err, taskbus.ErrCannotUncompleteInboxTask) {
				return errs.New(errs.FailedPrecondition, err)
			}
			return errs.Newf(errs.Internal, "uncomplete: %s", err)
		}

		if err := tx.Commit(); err != nil {
			return errs.Newf(errs.Internal, "commit: %s", err)
		}

		return toAppTask(task)
	}

	// Inbox tasks (ObjectiveID == nil) cannot be uncompleted.
	// The business layer will reject this, but we handle it defensively.
	task, err := a.taskBus.Uncomplete(ctx, task)
	if err != nil {
		if errors.Is(err, taskbus.ErrNotCompleted) {
			return errs.New(errs.FailedPrecondition, err)
		}
		// ErrCannotUncompleteInboxTask is expected here for inbox tasks
		if errors.Is(err, taskbus.ErrCannotUncompleteInboxTask) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "uncomplete unexpected: %s", err)
	}

	return toAppTask(task)
}

// cancel handles PUT /v1/tasks/{task_id}/cancel
func (a *app) cancel(ctx context.Context, r *http.Request) web.Encoder {
	task, respErr := a.getTaskWithOwnership(ctx, r)
	if respErr != nil {
		return respErr
	}

	task, err := a.taskBus.Cancel(ctx, task)
	if err != nil {
		if errors.Is(err, taskbus.ErrAlreadyCanceled) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "cancel: %s", err)
	}

	return toAppTask(task)
}

// query handles GET /v1/tasks
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("pagination", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, taskbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	inboxOnly, err := parseBool(qp.InboxOnly)
	if err != nil {
		return errs.NewFieldErrors("inboxOnly", err)
	}

	filter := taskbus.QueryFilter{
		UserID:    &userID,
		InboxOnly: inboxOnly,
	}

	// Filter by objective if provided
	if qp.ObjectiveID != "" {
		objectiveID, err := uuid.Parse(qp.ObjectiveID)
		if err != nil {
			return errs.NewFieldErrors("objectiveId", err)
		}
		filter.ObjectiveID = &objectiveID
	}

	// Filter by status if provided
	if qp.Status != "" {
		status, err := parseStatus(qp.Status)
		if err != nil {
			return errs.NewFieldErrors("status", err)
		}
		filter.Status = status
	}

	tasks, err := a.taskBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.taskBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppTasks(tasks), total, pg)
}

// queryByID handles GET /v1/tasks/{task_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	task, respErr := a.getTaskWithOwnership(ctx, r)
	if respErr != nil {
		return respErr
	}

	return toAppTask(task)
}

// queryByObjective handles GET /v1/objectives/{objective_id}/tasks
func (a *app) queryByObjective(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("pagination", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, taskbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := taskbus.QueryFilter{
		UserID:      &userID,
		ObjectiveID: &objectiveID,
	}

	// Filter by status if provided
	if qp.Status != "" {
		status, err := parseStatus(qp.Status)
		if err != nil {
			return errs.NewFieldErrors("status", err)
		}
		filter.Status = status
	}

	tasks, err := a.taskBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.taskBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppTasks(tasks), total, pg)
}

// ===== Helper functions =====

// getTaskWithOwnership fetches a task and verifies ownership.
func (a *app) getTaskWithOwnership(ctx context.Context, r *http.Request) (taskbus.Task, web.Encoder) {
	taskID, err := uuid.Parse(web.Param(r, "task_id"))
	if err != nil {
		return taskbus.Task{}, errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return taskbus.Task{}, errs.New(errs.Unauthenticated, err)
	}

	task, err := a.taskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, taskbus.ErrNotFound) {
			return taskbus.Task{}, errs.New(errs.NotFound, err)
		}
		return taskbus.Task{}, errs.Newf(errs.Internal, "querybyid: taskID[%s]: %s", taskID, err)
	}

	if task.UserID != userID {
		return taskbus.Task{}, errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	return task, nil
}

// ===== Query params parsing =====

type queryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ObjectiveID string
	Status      string
	InboxOnly   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ObjectiveID: values.Get("objectiveId"),
		Status:      values.Get("status"),
		InboxOnly:   values.Get("inboxOnly"),
	}
}

func parseStatus(s string) (*taskbus.TaskStatus, error) {
	status, err := taskbus.ParseStatus(s)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// parseBool parses a boolean string. Empty string returns false with no error.
// Invalid non-empty values return an error.
func parseBool(s string) (bool, error) {
	if s == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value %q", s)
	}
	return v, nil
}
