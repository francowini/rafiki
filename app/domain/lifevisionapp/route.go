package lifevisionapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	LifeVisionBus lifevisionbus.ExtBusiness
	Auth          *auth.Auth
}

// Routes registers all life vision endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.LifeVisionBus)

	// CRUD operations
	app.HandlerFunc(http.MethodPost, version, "/lifevisions", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/lifevisions", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/lifevisions/{lifevision_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/lifevisions/{lifevision_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/lifevisions/{lifevision_id}", api.delete, bearer)

	// Query by value (nested endpoint)
	app.HandlerFunc(http.MethodGet, version, "/values/{value_id}/lifevisions", api.queryByValue, bearer)
}
