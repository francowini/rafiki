package objectiveapp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	objectiveBus objectivebus.ExtBusiness
}

func newApp(objectiveBus objectivebus.ExtBusiness) *app {
	return &app{
		objectiveBus: objectiveBus,
	}
}

// create handles POST /v1/objectives
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appObjective NewObjective
	if err := web.Decode(r, &appObjective); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appObjective.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	no, err := toBusNewObjective(appObjective, userID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objective, err := a.objectiveBus.Create(ctx, no)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, errors.New("life vision not found"))
		}
		if errors.Is(err, objectivebus.ErrNotLifeVisionOwner) {
			return errs.New(errs.PermissionDenied, errors.New("user does not own the specified life vision"))
		}
		if errors.Is(err, objectivebus.ErrTargetLifeVisionNotActive) {
			return errs.New(errs.FailedPrecondition, errors.New("life vision must be active"))
		}
		if errors.Is(err, objectivebus.ErrInvalidResultConfig) ||
			errors.Is(err, objectivebus.ErrInvalidFrequencyConfig) ||
			errors.Is(err, objectivebus.ErrMissingCompliancePct) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: lifeVisionID[%s]: %s", no.LifeVisionID, err)
	}

	return toAppObjective(objective)
}

// update handles PUT /v1/objectives/{objective_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appObjective UpdateObjective
	if err := web.Decode(r, &appObjective); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appObjective.Validate(); err != nil {
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

	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	uo, err := toBusUpdateObjective(appObjective)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objective, err = a.objectiveBus.Update(ctx, objective, uo)
	if err != nil {
		if errors.Is(err, objectivebus.ErrInvalidResultConfig) ||
			errors.Is(err, objectivebus.ErrInvalidFrequencyConfig) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppObjective(objective)
}

// delete handles DELETE /v1/objectives/{objective_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.objectiveBus.Delete(ctx, objective); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// archive handles PUT /v1/objectives/{objective_id}/archive
func (a *app) archive(ctx context.Context, r *http.Request) web.Encoder {
	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	objective, err = a.objectiveBus.Archive(ctx, objective)
	if err != nil {
		if errors.Is(err, objectivebus.ErrAlreadyArchived) {
			return errs.New(errs.FailedPrecondition, err)
		}
		if errors.Is(err, objectivebus.ErrTerminalStatusNoArchive) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "archive: %s", err)
	}

	return toAppObjective(objective)
}

// restore handles PUT /v1/objectives/{objective_id}/restore
func (a *app) restore(ctx context.Context, r *http.Request) web.Encoder {
	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Query with includeArchived=true to find archived objectives
	filter := objectivebus.QueryFilter{
		ID:              &objectiveID,
		UserID:          &userID,
		IncludeArchived: true,
	}
	objectives, err := a.objectiveBus.Query(ctx, filter, order.NewBy("", ""), page.MustParse("1", "1"))
	if err != nil {
		return errs.Newf(errs.Internal, "query: objectiveID[%s]: %s", objectiveID, err)
	}
	if len(objectives) == 0 {
		return errs.New(errs.NotFound, objectivebus.ErrNotFound)
	}

	objective := objectives[0]

	objective, err = a.objectiveBus.Restore(ctx, objective)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, objectivebus.ErrTargetLifeVisionNotActive) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "restore: %s", err)
	}

	return toAppObjective(objective)
}

// changeStatus handles PUT /v1/objectives/{objective_id}/status
func (a *app) changeStatus(ctx context.Context, r *http.Request) web.Encoder {
	var req ChangeStatusRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
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

	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	busReq, err := toBusChangeStatusRequest(req)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objective, err = a.objectiveBus.ChangeStatus(ctx, objective, busReq)
	if err != nil {
		if errors.Is(err, objectivebus.ErrStatusTransitionNotAllowed) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "changestatus: %s", err)
	}

	return toAppObjective(objective)
}

// incrementProgress handles PUT /v1/objectives/{objective_id}/progress
// Supports both increment (legacy) and direct value setting.
func (a *app) incrementProgress(ctx context.Context, r *http.Request) web.Encoder {
	var req UpdateProgressRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
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

	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	busReq := toBusUpdateProgressRequest(req)

	objective, err = a.objectiveBus.UpdateProgress(ctx, objective, busReq)
	if err != nil {
		if errors.Is(err, objectivebus.ErrOnlyResultAllowsProgress) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, objectivebus.ErrProgressExceedsTarget) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "updateprogress: %s", err)
	}

	return toAppObjective(objective)
}

// query handles GET /v1/objectives
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

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, objectivebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := objectivebus.QueryFilter{
		UserID:          &userID,
		IncludeArchived: parseBool(qp.IncludeArchived),
	}

	// Filter by life vision if provided
	if qp.LifeVisionID != "" {
		lifeVisionID, err := uuid.Parse(qp.LifeVisionID)
		if err != nil {
			return errs.NewFieldErrors("lifeVisionId", err)
		}
		filter.LifeVisionID = &lifeVisionID
	}

	// Filter by status if provided
	if qp.Status != "" {
		status, err := parseStatus(qp.Status)
		if err != nil {
			return errs.NewFieldErrors("status", err)
		}
		filter.Status = status
	}

	// Filter by tracking type if provided
	if qp.TrackingType != "" {
		trackingType, err := parseTrackingType(qp.TrackingType)
		if err != nil {
			return errs.NewFieldErrors("trackingType", err)
		}
		filter.TrackingType = trackingType
	}

	objectives, err := a.objectiveBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.objectiveBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppObjectives(objectives), total, pg)
}

// queryByID handles GET /v1/objectives/{objective_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objectiveID, err := uuid.Parse(web.Param(r, "objective_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objective, err := a.objectiveBus.QueryByID(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, objectivebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objectiveID[%s]: %s", objectiveID, err)
	}

	if objective.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	return toAppObjective(objective)
}

// queryByLifeVision handles GET /v1/lifevisions/{lifevision_id}/objectives
func (a *app) queryByLifeVision(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	lifeVisionID, err := uuid.Parse(web.Param(r, "lifevision_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("pagination", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, objectivebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := objectivebus.QueryFilter{
		UserID:          &userID,
		LifeVisionID:    &lifeVisionID,
		IncludeArchived: parseBool(qp.IncludeArchived),
	}

	objectives, err := a.objectiveBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.objectiveBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppObjectives(objectives), total, pg)
}

// ===== Query params parsing =====

type queryParams struct {
	Page            string
	Rows            string
	OrderBy         string
	LifeVisionID    string
	Status          string
	TrackingType    string
	IncludeArchived string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:            values.Get("page"),
		Rows:            values.Get("rows"),
		OrderBy:         values.Get("orderBy"),
		LifeVisionID:    values.Get("lifeVisionId"),
		Status:          values.Get("status"),
		TrackingType:    values.Get("trackingType"),
		IncludeArchived: values.Get("includeArchived"),
	}
}

func parseStatus(s string) (*objectivebus.ObjectiveStatus, error) {
	status, err := objectivebus.ParseStatus(s)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func parseTrackingType(s string) (*objectivebus.TrackingType, error) {
	tt, err := objectivebus.ParseTrackingType(s)
	if err != nil {
		return nil, err
	}
	return &tt, nil
}

// parseBool parses a boolean string, returning false for empty or invalid values.
func parseBool(s string) bool {
	if s == "" {
		return false
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return v
}
