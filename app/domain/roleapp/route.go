package roleapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	RoleBus  rolebus.ExtBusiness
	ValueBus valuebus.ExtBusiness
	Auth     *auth.Auth
}

// Routes registers all role endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.RoleBus, cfg.ValueBus)

	// CRUD operations
	app.HandlerFunc(http.MethodPost, version, "/roles", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/roles", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/roles/{role_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/roles/{role_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/roles/{role_id}", api.delete, bearer)

	// Archive/Restore
	app.HandlerFunc(http.MethodPut, version, "/roles/{role_id}/archive", api.archive, bearer)
	app.HandlerFunc(http.MethodPut, version, "/roles/{role_id}/restore", api.restore, bearer)

	// Reorder
	app.HandlerFunc(http.MethodPost, version, "/roles/reorder", api.reorder, bearer)

	// Value connections
	app.HandlerFunc(http.MethodPost, version, "/roles/{role_id}/values", api.connectValue, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/roles/{role_id}/values/{value_id}", api.disconnectValue, bearer)
	app.HandlerFunc(http.MethodPut, version, "/roles/{role_id}/values", api.setValues, bearer)
	app.HandlerFunc(http.MethodGet, version, "/roles/{role_id}/values", api.getValues, bearer)

	// Reverse lookup: roles for a value
	app.HandlerFunc(http.MethodGet, version, "/values/{value_id}/roles", api.getRolesForValue, bearer)

	// Metadata
	app.HandlerFunc(http.MethodGet, version, "/metadata/role-facets", api.getRoleFacets, bearer)
}
