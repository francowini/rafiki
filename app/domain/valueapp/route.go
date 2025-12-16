package valueapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	ValueBus valuebus.ExtBusiness
	Auth     *auth.Auth
}

// Routes registers all value endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.ValueBus)

	app.HandlerFunc(http.MethodPost, version, "/values", api.create, bearer)
	app.HandlerFunc(http.MethodPost, version, "/values/reorder", api.reorder, bearer)
	app.HandlerFunc(http.MethodGet, version, "/values", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/values/{value_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/values/{value_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/values/{value_id}", api.delete, bearer)
	app.HandlerFunc(http.MethodPut, version, "/values/{value_id}/archive", api.archive, bearer)
	app.HandlerFunc(http.MethodPut, version, "/values/{value_id}/restore", api.restore, bearer)
}
