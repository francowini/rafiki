package thinkapp

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	thinkBus *thinkbus.Business
}

func newApp(thinkBus *thinkbus.Business) *app {
	return &app{
		thinkBus: thinkBus,
	}
}

// create adds a new think to the system
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var app NewThink
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validation happens in conversion via Parse functions
	// UserID extracted from context in toBusNewThink
	nt, err := toBusNewThink(ctx, app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	think, err := a.thinkBus.Create(ctx, nt)
	if err != nil {
		return errs.Newf(errs.Internal, "create: think[%+v]: %s", think, err)
	}

	return toAppThink(think)
}

// query returns a list of all thinks with paging
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, thinkbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	thinks, err := a.thinkBus.Query(ctx, userID, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.thinkBus.Count(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppThinks(thinks), total, page)
}

// queryByID returns a think by its ID
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	thinkID, err := uuid.Parse(web.Param(r, "think_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	think, err := a.thinkBus.QueryByID(ctx, thinkID, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "querybyid: thinkID[%s]: %s", thinkID, err)
	}

	return toAppThink(think)
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
