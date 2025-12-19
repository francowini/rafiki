package objetivoapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	objetivoBus objetivobus.ExtBusiness
}

func newApp(objetivoBus objetivobus.ExtBusiness) *app {
	return &app{
		objetivoBus: objetivoBus,
	}
}

// create handles POST /v1/objetivos
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appObjetivo NewObjetivo
	if err := web.Decode(r, &appObjetivo); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appObjetivo.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	no, err := toBusNewObjetivo(appObjetivo, userID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objetivo, err := a.objetivoBus.Create(ctx, no)
	if err != nil {
		if errors.Is(err, lifevisionbus.ErrNotFound) {
			return errs.New(errs.NotFound, errors.New("life vision not found"))
		}
		if errors.Is(err, objetivobus.ErrNotLifeVisionOwner) {
			return errs.New(errs.PermissionDenied, errors.New("user does not own the specified life vision"))
		}
		if errors.Is(err, objetivobus.ErrTargetLifeVisionNotActive) {
			return errs.New(errs.FailedPrecondition, errors.New("life vision must be active"))
		}
		if errors.Is(err, objetivobus.ErrInvalidResultadoConfig) ||
			errors.Is(err, objetivobus.ErrInvalidFrecuenciaConfig) ||
			errors.Is(err, objetivobus.ErrMissingCumplimientoPct) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: lifeVisionID[%s]: %s", no.LifeVisionID, err)
	}

	return toAppObjetivo(objetivo)
}

// update handles PUT /v1/objetivos/{objetivo_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appObjetivo UpdateObjetivo
	if err := web.Decode(r, &appObjetivo); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appObjetivo.Validate(); err != nil {
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

	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	uo, err := toBusUpdateObjetivo(appObjetivo)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objetivo, err = a.objetivoBus.Update(ctx, objetivo, uo)
	if err != nil {
		if errors.Is(err, objetivobus.ErrInvalidResultadoConfig) ||
			errors.Is(err, objetivobus.ErrInvalidFrecuenciaConfig) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppObjetivo(objetivo)
}

// delete handles DELETE /v1/objetivos/{objetivo_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	objetivoID, err := uuid.Parse(web.Param(r, "objetivo_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.objetivoBus.Delete(ctx, objetivo); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// archive handles PUT /v1/objetivos/{objetivo_id}/archive
func (a *app) archive(ctx context.Context, r *http.Request) web.Encoder {
	objetivoID, err := uuid.Parse(web.Param(r, "objetivo_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	objetivo, err = a.objetivoBus.Archive(ctx, objetivo)
	if err != nil {
		if errors.Is(err, objetivobus.ErrAlreadyArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, objetivobus.ErrTerminalStatusNoArchive) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "archive: %s", err)
	}

	return toAppObjetivo(objetivo)
}

// restore handles PUT /v1/objetivos/{objetivo_id}/restore
func (a *app) restore(ctx context.Context, r *http.Request) web.Encoder {
	objetivoID, err := uuid.Parse(web.Param(r, "objetivo_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Query with includeArchived=true to find archived objetivos
	filter := objetivobus.QueryFilter{
		ID:              &objetivoID,
		UserID:          &userID,
		IncludeArchived: true,
	}
	objetivos, err := a.objetivoBus.Query(ctx, filter, order.NewBy("", ""), page.MustParse("1", "1"))
	if err != nil || len(objetivos) == 0 {
		return errs.New(errs.NotFound, objetivobus.ErrNotFound)
	}

	objetivo := objetivos[0]

	objetivo, err = a.objetivoBus.Restore(ctx, objetivo)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, objetivobus.ErrTargetLifeVisionNotActive) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "restore: %s", err)
	}

	return toAppObjetivo(objetivo)
}

// changeStatus handles PUT /v1/objetivos/{objetivo_id}/status
func (a *app) changeStatus(ctx context.Context, r *http.Request) web.Encoder {
	var req ChangeStatusRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
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

	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	busReq, err := toBusChangeStatusRequest(req)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objetivo, err = a.objetivoBus.ChangeStatus(ctx, objetivo, busReq)
	if err != nil {
		if errors.Is(err, objetivobus.ErrStatusTransitionNotAllowed) {
			return errs.New(errs.FailedPrecondition, err)
		}
		return errs.Newf(errs.Internal, "changestatus: %s", err)
	}

	return toAppObjetivo(objetivo)
}

// incrementProgress handles PUT /v1/objetivos/{objetivo_id}/progress
// Supports both increment (legacy) and direct value setting.
func (a *app) incrementProgress(ctx context.Context, r *http.Request) web.Encoder {
	var req UpdateProgressRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
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

	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	busReq := toBusUpdateProgressRequest(req)

	objetivo, err = a.objetivoBus.UpdateProgress(ctx, objetivo, busReq)
	if err != nil {
		if errors.Is(err, objetivobus.ErrOnlyResultadoAllowsProgress) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, objetivobus.ErrProgressExceedsTarget) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "updateprogress: %s", err)
	}

	return toAppObjetivo(objetivo)
}

// query handles GET /v1/objetivos
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

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, objetivobus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := objetivobus.QueryFilter{
		UserID:          &userID,
		IncludeArchived: qp.IncludeArchived == "true",
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
	if qp.TipoTracking != "" {
		tipoTracking, err := parseTipoTracking(qp.TipoTracking)
		if err != nil {
			return errs.NewFieldErrors("tipoTracking", err)
		}
		filter.TipoTracking = tipoTracking
	}

	objetivos, err := a.objetivoBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.objetivoBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppObjetivos(objetivos), total, pg)
}

// queryByID handles GET /v1/objetivos/{objetivo_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	objetivoID, err := uuid.Parse(web.Param(r, "objetivo_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	objetivo, err := a.objetivoBus.QueryByID(ctx, objetivoID)
	if err != nil {
		if errors.Is(err, objetivobus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: objetivoID[%s]: %s", objetivoID, err)
	}

	if objetivo.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	return toAppObjetivo(objetivo)
}

// queryByLifeVision handles GET /v1/lifevisions/{lifevision_id}/objetivos
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

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, objetivobus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := objetivobus.QueryFilter{
		UserID:          &userID,
		LifeVisionID:    &lifeVisionID,
		IncludeArchived: qp.IncludeArchived == "true",
	}

	objetivos, err := a.objetivoBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.objetivoBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppObjetivos(objetivos), total, pg)
}

// ===== Query params parsing =====

type queryParams struct {
	Page            string
	Rows            string
	OrderBy         string
	LifeVisionID    string
	Status          string
	TipoTracking    string
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
		TipoTracking:    values.Get("tipoTracking"),
		IncludeArchived: values.Get("includeArchived"),
	}
}

func parseStatus(s string) (*objetivobus.ObjetivoStatus, error) {
	status, err := objetivobus.ParseStatus(s)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func parseTipoTracking(s string) (*objetivobus.TrackingType, error) {
	tt, err := objetivobus.ParseTrackingType(s)
	if err != nil {
		return nil, err
	}
	return &tt, nil
}
