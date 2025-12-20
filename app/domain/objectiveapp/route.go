package objectiveapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	ObjectiveBus objectivebus.ExtBusiness
	Auth         *auth.Auth
}

// Routes registers all objective endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.ObjectiveBus)

	// CRUD operations
	app.HandlerFunc(http.MethodPost, version, "/objectives", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/objectives", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/objectives/{objective_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/objectives/{objective_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/objectives/{objective_id}", api.delete, bearer)

	// Status and progress operations
	app.HandlerFunc(http.MethodPut, version, "/objectives/{objective_id}/status", api.changeStatus, bearer)
	app.HandlerFunc(http.MethodPut, version, "/objectives/{objective_id}/progress", api.incrementProgress, bearer)

	// Archive and restore
	app.HandlerFunc(http.MethodPut, version, "/objectives/{objective_id}/archive", api.archive, bearer)
	app.HandlerFunc(http.MethodPut, version, "/objectives/{objective_id}/restore", api.restore, bearer)

	// Query by life vision (nested endpoint)
	app.HandlerFunc(http.MethodGet, version, "/lifevisions/{lifevision_id}/objectives", api.queryByLifeVision, bearer)
}
