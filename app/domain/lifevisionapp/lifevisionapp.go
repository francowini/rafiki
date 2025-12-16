package lifevisionapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	lifeVisionBus lifevisionbus.ExtBusiness
}

func newApp(lifeVisionBus lifevisionbus.ExtBusiness) *app {
	return &app{
		lifeVisionBus: lifeVisionBus,
	}
}

// create handles POST /v1/lifevisions
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appLifeVision NewLifeVision
	if err := web.Decode(r, &appLifeVision); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appLifeVision.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	nlv, err := toBusNewLifeVision(appLifeVision, userID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lifeVision, err := a.lifeVisionBus.Create(ctx, nlv)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, errors.New("value not found"))
		}
		if errors.Is(err, lifevisionbus.ErrNotValueOwner) {
			return errs.New(errs.PermissionDenied, errors.New("user does not own the specified value"))
		}
		return errs.Newf(errs.Internal, "create: valueID[%s]: %s", nlv.ValueID, err)
	}

	return toAppLifeVision(lifeVision)
}

// update handles PUT /v1/lifevisions/{lifevision_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appLifeVision UpdateLifeVision
	if err := web.Decode(r, &appLifeVision); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appLifeVision.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	lifeVision, err := a.lifeVisionBus.QueryByID(ctx, lifeVisionID)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: lifeVisionID[%s]: %s", lifeVisionID, err)
	}

	if lifeVision.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	ulv, err := toBusUpdateLifeVision(appLifeVision)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lifeVision, err = a.lifeVisionBus.Update(ctx, lifeVision, ulv)
	if err != nil {
		if errors.Is(err, valuebus.ErrNotFound) {
			return errs.New(errs.NotFound, errors.New("value not found"))
		}
		if errors.Is(err, lifevisionbus.ErrNotValueOwner) {
			return errs.New(errs.PermissionDenied, errors.New("user does not own the specified value"))
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppLifeVision(lifeVision)
}

// delete handles DELETE /v1/lifevisions/{lifevision_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	lifeVision, err := a.lifeVisionBus.QueryByID(ctx, lifeVisionID)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: lifeVisionID[%s]: %s", lifeVisionID, err)
	}

	if lifeVision.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.lifeVisionBus.Delete(ctx, lifeVision); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// archive handles PUT /v1/lifevisions/{lifevision_id}/archive
func (a *app) archive(ctx context.Context, r *http.Request) web.Encoder {
	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	lifeVision, err := a.lifeVisionBus.QueryByID(ctx, lifeVisionID)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: lifeVisionID[%s]: %s", lifeVisionID, err)
	}

	if lifeVision.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	lifeVision, err = a.lifeVisionBus.Archive(ctx, lifeVision)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrAlreadyArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "archive: %s", err)
	}

	return toAppLifeVision(lifeVision)
}

// restore handles PUT /v1/lifevisions/{lifevision_id}/restore
func (a *app) restore(ctx context.Context, r *http.Request) web.Encoder {
	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Query with includeArchived=true to find archived life visions
	filter := lifevisionbus.QueryFilter{
		ID:              &lifeVisionID,
		UserID:          &userID,
		IncludeArchived: true,
	}
	lifeVisions, err := a.lifeVisionBus.Query(ctx, filter, order.NewBy("", ""), page.MustParse("1", "1"))
	if err != nil || len(lifeVisions) == 0 {
		return errs.New(errs.NotFound, lifevisionbus.ErrNotFound)
	}

	lifeVision := lifeVisions[0]

	lifeVision, err = a.lifeVisionBus.Restore(ctx, lifeVision)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, lifevisionbus.ErrTargetValueNotActive) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "restore: %s", err)
	}

	return toAppLifeVision(lifeVision)
}

// reassign handles PUT /v1/lifevisions/{lifevision_id}/reassign
func (a *app) reassign(ctx context.Context, r *http.Request) web.Encoder {
	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req ReassignRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newValueID, err := uuid.Parse(req.ValueID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	lifeVision, err := a.lifeVisionBus.QueryByID(ctx, lifeVisionID)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: lifeVisionID[%s]: %s", lifeVisionID, err)
	}

	if lifeVision.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	lifeVision, err = a.lifeVisionBus.Reassign(ctx, lifeVision, newValueID)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotValueOwner) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, lifevisionbus.ErrTargetValueNotActive) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "reassign: %s", err)
	}

	return toAppLifeVision(lifeVision)
}

// query handles GET /v1/lifevisions
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

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, lifevisionbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := lifevisionbus.QueryFilter{
		UserID:          &userID,
		IncludeArchived: qp.IncludeArchived == "true",
	}

	// Filter by value if provided
	if qp.ValueID != "" {
		valueID, err := uuid.Parse(qp.ValueID)
		if err != nil {
			return errs.NewFieldErrors("valueId", err)
		}
		filter.ValueID = &valueID
	}

	lifeVisions, err := a.lifeVisionBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.lifeVisionBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppLifeVisions(lifeVisions), total, pg)
}

// queryByID handles GET /v1/lifevisions/{lifevision_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lifeVision, err := a.lifeVisionBus.QueryByID(ctx, lifeVisionID)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: lifeVisionID[%s]: %s", lifeVisionID, err)
	}

	if lifeVision.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	return toAppLifeVision(lifeVision)
}

// queryByValue handles GET /v1/values/{value_id}/lifevisions
func (a *app) queryByValue(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("pagination", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, lifevisionbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := lifevisionbus.QueryFilter{
		UserID:  &userID,
		ValueID: &valueID,
	}

	lifeVisions, err := a.lifeVisionBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.lifeVisionBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppLifeVisions(lifeVisions), total, pg)
}

// ===== Query params parsing =====

type queryParams struct {
	Page            string
	Rows            string
	OrderBy         string
	ValueID         string
	IncludeArchived string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:            values.Get("page"),
		Rows:            values.Get("rows"),
		OrderBy:         values.Get("orderBy"),
		ValueID:         values.Get("valueId"),
		IncludeArchived: values.Get("includeArchived"),
	}
}
