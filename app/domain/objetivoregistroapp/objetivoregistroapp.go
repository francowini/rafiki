package objetivoregistroapp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/domain/objetivoregistrobus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	objetivoBus         objetivobus.ExtBusiness
	objetivoRegistroBus objetivoregistrobus.ExtBusiness
}

func newApp(objetivoBus objetivobus.ExtBusiness, objetivoRegistroBus objetivoregistrobus.ExtBusiness) *app {
	return &app{
		objetivoBus:         objetivoBus,
		objetivoRegistroBus: objetivoRegistroBus,
	}
}

// create handles POST /v1/objetivos/{objetivo_id}/records
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appRecord NewObjetivoRecord
	if err := web.Decode(r, &appRecord); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appRecord.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objetivoID, err := uuid.Parse(web.Param(r, "objetivo_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify objetivo exists and user owns it
	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, errors.New("objetivo not found"))
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	nr, err := toBusNewRecord(appRecord, objetivoID, userID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	record, err := a.objetivoRegistroBus.Create(ctx, nr)
	if err != nil {
		if errors.Is(err, objetivoregistrobus.ErrNotFrecuenciaType) {
			return errs.New(errs.InvalidArgument, errors.New("records only allowed for frecuencia tracking type"))
		}
		if errors.Is(err, objetivoregistrobus.ErrGracePeriodExceeded) {
			return errs.New(errs.InvalidArgument, errors.New("can only log records for today or yesterday"))
		}
		if errors.Is(err, objetivoregistrobus.ErrFutureDateNotAllowed) {
			return errs.New(errs.InvalidArgument, errors.New("cannot log record for future date"))
		}
		return errs.Newf(errs.Internal, "create: objetivoID[%s]: %s", objetivoID, err)
	}

	return toAppRecord(record)
}

// delete handles DELETE /v1/objetivos/records/{record_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	recordID, err := uuid.Parse(web.Param(r, "record_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	record, err := a.objetivoRegistroBus.QueryByID(ctx, recordID)
	if err != nil {
		if errors.Is(err, objetivoregistrobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: recordID[%s]: %s", recordID, err)
	}

	if record.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.objetivoRegistroBus.Delete(ctx, record); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// query handles GET /v1/objetivos/{objetivo_id}/records
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	objetivoID, err := uuid.Parse(web.Param(r, "objetivo_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify objetivo exists and user owns it
	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, errors.New("objetivo not found"))
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("pagination", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, objetivoregistrobus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := objetivoregistrobus.QueryFilter{
		ObjetivoID: &objetivoID,
	}

	// Parse date filters if provided
	if qp.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", qp.StartDate)
		if err != nil {
			return errs.NewFieldErrors("startDate", err)
		}
		filter.StartDate = &startDate
	}

	if qp.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", qp.EndDate)
		if err != nil {
			return errs.NewFieldErrors("endDate", err)
		}
		filter.EndDate = &endDate
	}

	records, err := a.objetivoRegistroBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.objetivoRegistroBus.Count(ctx, filter)
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
