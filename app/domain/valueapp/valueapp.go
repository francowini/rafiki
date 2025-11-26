package valueapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	valueBus valuebus.ExtBusiness
}

func newApp(valueBus valuebus.ExtBusiness) *app {
	return &app{
		valueBus: valueBus,
	}
}

// create handles POST /v1/values
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appValue NewValue
	if err := web.Decode(r, &appValue); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appValue.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nv, err := toBusNewValue(ctx, appValue)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	value, err := a.valueBus.Create(ctx, nv)
	if err != nil {
		if errors.Is(err, valuebus.ErrMaxValues) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, valuebus.ErrDuplicateOrder) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, valuebus.ErrUserDisabled) {
			return errs.New(errs.PermissionDenied, err)
		}
		return errs.Newf(errs.Internal, "create: userID[%s]: %s", nv.UserID, err)
	}

	return toAppValue(value)
}

// update handles PUT /v1/values/{value_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appValue UpdateValue
	if err := web.Decode(r, &appValue); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appValue.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	value, err := a.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: valueID[%s]: %s", valueID, err)
	}

	if value.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	uv, err := toBusUpdateValue(appValue)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	value, err = a.valueBus.Update(ctx, value, uv)
	if err != nil {
		if errors.Is(err, valuebus.ErrDuplicateOrder) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppValue(value)
}

// delete handles DELETE /v1/values/{value_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	value, err := a.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: valueID[%s]: %s", valueID, err)
	}

	if value.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.valueBus.Delete(ctx, value); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// query handles GET /v1/values
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

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, valuebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := valuebus.QueryFilter{
		UserID: &userID,
	}

	// Filter by facet if provided
	if qp.Facet != "" {
		facetVal, err := facet.Parse(qp.Facet)
		if err != nil {
			return errs.NewFieldErrors("facet", err)
		}
		filter.Facet = &facetVal
	}

	values, err := a.valueBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.valueBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppValues(values), total, pg)
}

// queryByID handles GET /v1/values/{value_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	value, err := a.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: valueID[%s]: %s", valueID, err)
	}

	if value.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	return toAppValue(value)
}

// reorder handles POST /v1/values/reorder
func (a *app) reorder(ctx context.Context, r *http.Request) web.Encoder {
	var appReorder ReorderRequest
	if err := web.Decode(r, &appReorder); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appReorder.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	busReorder, err := toBusReorderRequest(appReorder)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	values, err := a.valueBus.Reorder(ctx, userID, busReorder)
	if err != nil {
		if errors.Is(err, valuebus.ErrUserDisabled) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, valuebus.ErrDuplicateOrder) {
			return errs.New(errs.AlreadyExists, err)
		}
		return errs.Newf(errs.Internal, "reorder: %s", err)
	}

	return query.NewResult(toAppValues(values), len(values), page.Page{})
}

// ===== Query params parsing =====

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Facet   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		Facet:   values.Get("facet"),
	}
}
