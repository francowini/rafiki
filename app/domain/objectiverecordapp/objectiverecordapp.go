package objectiverecordapp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/domain/objectiverecordbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	objectiveBus       objectivebus.ExtBusiness
	objectiveRecordBus objectiverecordbus.ExtBusiness
}

func newApp(objectiveBus objectivebus.ExtBusiness, objectiveRecordBus objectiverecordbus.ExtBusiness) *app {
	return &app{
		objectiveBus:       objectiveBus,
		objectiveRecordBus: objectiveRecordBus,
	}
}

// getObjectiveWithOwnership fetches an objective and validates the user owns it.
// Returns the objective or an errs.Error with appropriate status code.
func (a *app) getObjectiveWithOwnership(ctx context.Context, objectiveID, userID uuid.UUID) (objectivebus.Objective, error) {
	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return objectivebus.Objective{}, errs.New(errs.NotFound, errors.New("objective not found"))
		}
		return objectivebus.Objective{}, errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return objectivebus.Objective{}, errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	return objective, nil
}

// create handles POST /v1/objectives/{objective_id}/records
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appRecord NewObjectiveRecord
	if err := web.Decode(r, &appRecord); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appRecord.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify objective exists and user owns it
	if _, err := a.getObjectiveWithOwnership(ctx, objectiveID, userID); err != nil {
		return err.(web.Encoder)
	}

	nr, err := toBusNewRecord(appRecord, objectiveID, userID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	record, err := a.objectiveRecordBus.Create(ctx, nr)
	if err != nil {
		if errors.Is(err, objectiverecordbus.ErrNotFrequencyType) {
			return errs.New(errs.InvalidArgument, errors.New("records only allowed for frequency tracking type"))
		}
		if errors.Is(err, objectiverecordbus.ErrGracePeriodExceeded) {
			return errs.New(errs.InvalidArgument, errors.New("can only log records for today or yesterday"))
		}
		if errors.Is(err, objectiverecordbus.ErrFutureDateNotAllowed) {
			return errs.New(errs.InvalidArgument, errors.New("cannot log record for future date"))
		}
		return errs.Newf(errs.Internal, "create: objectiveID[%s]: %s", objectiveID, err)
	}

	return toAppRecord(record)
}

// delete handles DELETE /v1/objectives/records/{record_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	recordID, err := uuid.Parse(web.Param(r, "record_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	record, err := a.objectiveRecordBus.QueryByID(ctx, recordID)
	if err != nil {
		if errors.Is(err, objectiverecordbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: recordID[%s]: %s", recordID, err)
	}

	if record.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.objectiveRecordBus.Delete(ctx, record); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// query handles GET /v1/objectives/{objective_id}/records
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify objective exists and user owns it
	if _, err := a.getObjectiveWithOwnership(ctx, objectiveID, userID); err != nil {
		return err.(web.Encoder)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("pagination", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, objectiverecordbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := objectiverecordbus.QueryFilter{
		ObjectiveID: &objectiveID,
	}

	// Parse date filters if provided
	var startDate, endDate time.Time
	if qp.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", qp.StartDate)
		if err != nil {
			return errs.NewFieldErrors("startDate", err)
		}
		filter.StartDate = &startDate
	}

	if qp.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", qp.EndDate)
		if err != nil {
			return errs.NewFieldErrors("endDate", err)
		}
		filter.EndDate = &endDate
	}

	// Validate date range: startDate must be <= endDate
	if filter.StartDate != nil && filter.EndDate != nil && startDate.After(endDate) {
		return errs.NewFieldErrors("startDate", errors.New("startDate must be before or equal to endDate"))
	}

	records, err := a.objectiveRecordBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.objectiveRecordBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppRecords(records), total, pg)
}

// ===== Query params parsing =====

type queryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	StartDate string
	EndDate   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		StartDate: values.Get("startDate"),
		EndDate:   values.Get("endDate"),
	}
}
