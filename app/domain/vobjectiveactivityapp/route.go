package vobjectiveactivityapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	Log                   *logger.Logger
	VObjectiveActivityBus vobjectiveactivitybus.ExtBusiness
	Auth                  *auth.Auth
}

// Routes registers all objective activity endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.Log, cfg.VObjectiveActivityBus)

	// Activity endpoint (nested under objectives)
	app.HandlerFunc(http.MethodGet, version, "/objectives/{objective_id}/activity", api.query, bearer)
}
