package thinkapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all the configuration for the think API
type Config struct {
	ThinkBus *thinkbus.Business
	Log      *logger.Logger
	Auth     *auth.Auth
}

// Routes adds the routes for the think api
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Create Bearer auth middleware
	bearer := mid.Bearer(cfg.Auth)

	api := newApp(cfg.ThinkBus)

	// Apply bearer middleware to all routes
	app.HandlerFunc(http.MethodGet, version, "/thinks", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/thinks/{think_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPost, version, "/thinks", api.create, bearer)
}
