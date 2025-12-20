package objectiverecordapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/domain/objectiverecordbus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	ObjectiveBus       objectivebus.ExtBusiness
	ObjectiveRecordBus objectiverecordbus.ExtBusiness
	Auth               *auth.Auth
}

// Routes registers all objective record endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.ObjectiveBus, cfg.ObjectiveRecordBus)

	// Record operations (nested under objectives)
	app.HandlerFunc(http.MethodPost, version, "/objectives/{objective_id}/records", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/objectives/{objective_id}/records", api.query, bearer)

	// Delete record (flat endpoint for simpler access)
	app.HandlerFunc(http.MethodDelete, version, "/objectives/records/{record_id}", api.delete, bearer)
}
