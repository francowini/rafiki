package thinkapp

import (
	"net/http"

	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all the configuration for the think API
type Config struct {
	ThinkBus *thinkbus.Business
	Log      *logger.Logger
}

// Routes adds the routes for the think api
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newApp(cfg.ThinkBus)

	app.HandlerFunc(http.MethodGet, version, "/thinks", api.query)
	app.HandlerFunc(http.MethodGet, version, "/thinks/{think_id}", api.queryByID)
	app.HandlerFunc(http.MethodPost, version, "/thinks", api.create)
}
