package roleapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/types/rolefacet"
	"github.com/francowini/rafiki/business/types/roletype"
	"github.com/francowini/rafiki/foundation/web"
)

type app struct {
	roleBus  rolebus.ExtBusiness
	valueBus valuebus.ExtBusiness
}

func newApp(roleBus rolebus.ExtBusiness, valueBus valuebus.ExtBusiness) *app {
	return &app{
		roleBus:  roleBus,
		valueBus: valueBus,
	}
}

// =============================================================================
// CRUD Operations
// =============================================================================

// create handles POST /v1/roles
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var appRole NewRole
	if err := web.Decode(r, &appRole); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appRole.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nr, err := toBusNewRole(ctx, appRole)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err := a.roleBus.Create(ctx, nr)
	if err != nil {
		if errors.Is(err, rolebus.ErrMaxRoles) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrDuplicateOrder) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrUserDisabled) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, rolebus.ErrValueNotOwned) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, rolebus.ErrValueArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrInvalidRoleType) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrInvalidFacet) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "create: userID[%s]: %s", nr.UserID, err)
	}

	// Return role with value IDs
	return toAppRoleWithValues(role, nr.ValueIDs)
}

// update handles PUT /v1/roles/{role_id}
func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var appRole UpdateRole
	if err := web.Decode(r, &appRole); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := appRole.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	role, err := a.roleBus.QueryByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: roleID[%s]: %s", roleID, err)
	}

	if role.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	ur, err := toBusUpdateRole(appRole)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err = a.roleBus.Update(ctx, role, ur)
	if err != nil {
		if errors.Is(err, rolebus.ErrDuplicateOrder) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrInvalidRoleType) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrInvalidFacet) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	return toAppRole(role)
}

// delete handles DELETE /v1/roles/{role_id}
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	role, err := a.roleBus.QueryByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: roleID[%s]: %s", roleID, err)
	}

	if role.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	if err := a.roleBus.Delete(ctx, role); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// =============================================================================
// Archive/Restore Operations
// =============================================================================

// archive handles PUT /v1/roles/{role_id}/archive
func (a *app) archive(ctx context.Context, r *http.Request) web.Encoder {
	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	role, err := a.roleBus.QueryByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: roleID[%s]: %s", roleID, err)
	}

	if role.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	role, err = a.roleBus.Archive(ctx, role)
	if err != nil {
		if errors.Is(err, rolebus.ErrAlreadyArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "archive: %s", err)
	}

	return toAppRole(role)
}

// restore handles PUT /v1/roles/{role_id}/restore
func (a *app) restore(ctx context.Context, r *http.Request) web.Encoder {
	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Query with includeArchived=true to find archived roles
	filter := rolebus.QueryFilter{
		ID:              &roleID,
		UserID:          &userID,
		IncludeArchived: true,
	}
	roles, err := a.roleBus.Query(ctx, filter, order.NewBy("", ""), page.MustParse("1", "1"))
	if err != nil || len(roles) == 0 {
		return errs.New(errs.NotFound, rolebus.ErrNotFound)
	}

	role := roles[0]

	role, err = a.roleBus.Restore(ctx, role)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrTargetValueNotActive) {
			return errs.New(errs.Aborted, err)
		}
		return errs.Newf(errs.Internal, "restore: %s", err)
	}

	return toAppRole(role)
}

// =============================================================================
// Query Operations
// =============================================================================

// query handles GET /v1/roles
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

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, rolebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter := rolebus.QueryFilter{
		UserID:          &userID,
		IncludeArchived: qp.IncludeArchived == "true",
	}

	// Filter by role type if provided
	if qp.RoleType != "" {
		roleTypeVal, err := roletype.Parse(qp.RoleType)
		if err != nil {
			return errs.NewFieldErrors("roleType", err)
		}
		filter.RoleType = &roleTypeVal
	}

	// Filter by facet if provided
	if qp.Facet != "" {
		facetVal, err := rolefacet.Parse(qp.Facet)
		if err != nil {
			return errs.NewFieldErrors("facet", err)
		}
		filter.Facet = &facetVal
	}

	roles, err := a.roleBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.roleBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppRoles(roles), total, pg)
}

// queryByID handles GET /v1/roles/{role_id}
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err := a.roleBus.QueryByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: roleID[%s]: %s", roleID, err)
	}

	if role.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	// Get connected values
	valueIDs, err := a.roleBus.GetValueIDs(ctx, roleID)
	if err != nil {
		return errs.Newf(errs.Internal, "getvalueids: %s", err)
	}

	return toAppRoleWithValues(role, valueIDs)
}

// =============================================================================
// Reorder Operation
// =============================================================================

// reorder handles POST /v1/roles/reorder
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

	roles, err := a.roleBus.Reorder(ctx, userID, busReorder)
	if err != nil {
		if errors.Is(err, rolebus.ErrUserDisabled) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, rolebus.ErrDuplicateOrder) {
			return errs.New(errs.AlreadyExists, err)
		}
		return errs.Newf(errs.Internal, "reorder: %s", err)
	}

	return query.NewResult(toAppRoles(roles), len(roles), page.Page{})
}

// =============================================================================
// Value Connection Operations
// =============================================================================

// connectValue handles POST /v1/roles/{role_id}/values
func (a *app) connectValue(ctx context.Context, r *http.Request) web.Encoder {
	var req ConnectValueRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	valueID, err := uuid.Parse(req.ValueID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	err = a.roleBus.ConnectValue(ctx, roleID, valueID, userID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		if errors.Is(err, rolebus.ErrValueNotOwned) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, rolebus.ErrValueArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		if errors.Is(err, rolebus.ErrDuplicateValueLink) {
			return errs.New(errs.AlreadyExists, err)
		}
		return errs.Newf(errs.Internal, "connectvalue: %s", err)
	}

	return RoleValue{
		RoleID:  roleID.String(),
		ValueID: valueID.String(),
	}
}

// disconnectValue handles DELETE /v1/roles/{role_id}/values/{value_id}
func (a *app) disconnectValue(ctx context.Context, r *http.Request) web.Encoder {
	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
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

	err = a.roleBus.DisconnectValue(ctx, roleID, valueID, userID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "disconnectvalue: %s", err)
	}

	return nil
}

// setValues handles PUT /v1/roles/{role_id}/values
func (a *app) setValues(ctx context.Context, r *http.Request) web.Encoder {
	var req SetValuesRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Parse value IDs
	valueIDs := make([]uuid.UUID, len(req.ValueIDs))
	for i, idStr := range req.ValueIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return errs.NewFieldErrors(fmt.Sprintf("value_ids[%d]", i), err)
		}
		valueIDs[i] = id
	}

	err = a.roleBus.SetValues(ctx, roleID, valueIDs, userID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		if errors.Is(err, rolebus.ErrValueNotOwned) {
			return errs.New(errs.PermissionDenied, err)
		}
		if errors.Is(err, rolebus.ErrValueArchived) {
			return errs.New(errs.InvalidArgument, err)
		}
		return errs.Newf(errs.Internal, "setvalues: %s", err)
	}

	// Return updated role with values
	role, err := a.roleBus.QueryByID(ctx, roleID)
	if err != nil {
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppRoleWithValues(role, valueIDs)
}

// getValues handles GET /v1/roles/{role_id}/values
func (a *app) getValues(ctx context.Context, r *http.Request) web.Encoder {
	roleID, err := uuid.Parse(web.Param(r, "role_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify role exists and belongs to user
	role, err := a.roleBus.QueryByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: roleID[%s]: %s", roleID, err)
	}

	if role.UserID != userID {
		return errs.New(errs.PermissionDenied, errors.New("user not authorized"))
	}

	valueIDs, err := a.roleBus.GetValueIDs(ctx, roleID)
	if err != nil {
		return errs.Newf(errs.Internal, "getvalueids: %s", err)
	}

	// Convert to string array for response
	ids := make([]string, len(valueIDs))
	for i, id := range valueIDs {
		ids[i] = id.String()
	}

	return valueIDsResponse{ValueIDs: ids}
}

// getRolesForValue handles GET /v1/values/{value_id}/roles
func (a *app) getRolesForValue(ctx context.Context, r *http.Request) web.Encoder {
	valueID, err := uuid.Parse(web.Param(r, "value_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.roleBus.GetRoleIDsForValue(ctx, valueID, userID)
	if err != nil {
		if errors.Is(err, rolebus.ErrValueNotOwned) {
			return errs.New(errs.PermissionDenied, err)
		}
		return errs.Newf(errs.Internal, "getroleidsforvalue: %s", err)
	}

	// Convert to string array for response
	ids := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		ids[i] = id.String()
	}

	return roleIDsResponse{RoleIDs: ids}
}

// =============================================================================
// Metadata Operations
// =============================================================================

// getRoleFacets handles GET /v1/metadata/role-facets
func (a *app) getRoleFacets(ctx context.Context, r *http.Request) web.Encoder {
	facets := rolefacet.All()
	infos := make([]RoleFacetInfo, len(facets))

	for i, f := range facets {
		infos[i] = RoleFacetInfo{
			Value: f.String(),
			Label: getRoleFacetLabel(f.String()),
		}
	}

	return RoleFacetList{Facets: infos}
}

// =============================================================================
// Helper Types
// =============================================================================

type valueIDsResponse struct {
	ValueIDs []string `json:"valueIds"`
}

func (v valueIDsResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(v)
	return data, "application/json", err
}

type roleIDsResponse struct {
	RoleIDs []string `json:"roleIds"`
}

func (r roleIDsResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// =============================================================================
// Query Params Parsing
// =============================================================================

type queryParams struct {
	Page            string
	Rows            string
	OrderBy         string
	RoleType        string
	Facet           string
	IncludeArchived string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()
	return queryParams{
		Page:            values.Get("page"),
		Rows:            values.Get("rows"),
		OrderBy:         values.Get("orderBy"),
		RoleType:        values.Get("roleType"),
		Facet:           values.Get("facet"),
		IncludeArchived: values.Get("includeArchived"),
	}
}
