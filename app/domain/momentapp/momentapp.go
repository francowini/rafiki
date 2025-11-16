package momentapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
	"github.com/google/uuid"
)

type app struct {
	momentBus *momentbus.Business
}

func newApp(momentBus *momentbus.Business) *app {
	return &app{
		momentBus: momentBus,
	}
}

func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appMoment NewMoment
	if err := web.Decode(r, &appMoment); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nm, err := toBusNewMoment(ctx, appMoment)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err := a.momentBus.Create(ctx, nm)
	if err != nil {
		if errors.Is(err, momentbus.ErrFutureDate) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: moment[%+v]: %s", moment, err)
	}

	return toAppMoment(moment)
}

func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appMoment UpdateMoment
	if err := web.Decode(r, &appMoment); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	momentID, err := uuid.Parse(web.Param(r, "moment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	moment, err := a.momentBus.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, momentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: momentID[%s]: %s", momentID, err)
	}

	if moment.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	um, err := toBusUpdateMoment(ctx, appMoment)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err = a.momentBus.Update(ctx, moment, um)
	if err != nil {
		if errors.Is(err, momentbus.ErrFutureDate) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppMoment(moment)
}

func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	momentID, err := uuid.Parse(web.Param(r, "moment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	moment, err := a.momentBus.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, momentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: momentID[%s]: %s", momentID, err)
	}

	if moment.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	if err := a.momentBus.Delete(ctx, moment); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, momentbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := momentbus.QueryFilter{
		UserID:  &userID,
		Page:    pg,
		OrderBy: orderBy,
	}

	moments, err := a.momentBus.Query(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.momentBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppMoments(moments), total, pg)
}

func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	momentID, err := uuid.Parse(web.Param(r, "moment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	moment, err := a.momentBus.QueryByID(ctx, momentID)
	if err != nil {
		if errors.Is(err, momentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: momentID[%s]: %s", momentID, err)
	}

	if moment.UserID != userID {
		return errs.New(errs.Unauthenticated, errors.New("user not authorized"))
	}

	return toAppMoment(moment)
}

// =============================================================================

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	return queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
	}
}
