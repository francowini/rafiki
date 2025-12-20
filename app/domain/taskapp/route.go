package taskapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/domain/taskbus"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	TaskBus      taskbus.ExtBusiness
	ObjectiveBus objectivebus.ExtBusiness
	Auth         *auth.Auth
	DB           sqldb.Beginner
}

// Routes registers all task endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.TaskBus, cfg.ObjectiveBus, cfg.DB)

	// CRUD operations
	app.HandlerFunc(http.MethodPost, version, "/tasks", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/tasks", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/tasks/{task_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/tasks/{task_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/tasks/{task_id}", api.delete, bearer)

	// Status operations
	app.HandlerFunc(http.MethodPut, version, "/tasks/{task_id}/complete", api.complete, bearer)
	app.HandlerFunc(http.MethodPut, version, "/tasks/{task_id}/uncomplete", api.uncomplete, bearer)
	app.HandlerFunc(http.MethodPut, version, "/tasks/{task_id}/cancel", api.cancel, bearer)

	// Query by objective (nested endpoint)
	app.HandlerFunc(http.MethodGet, version, "/objectives/{objective_id}/tasks", api.queryByObjective, bearer)
}
