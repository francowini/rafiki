package vexportapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains the dependencies for the vexport app handlers.
type Config struct {
	VExportBus vexportbus.ExtBusiness
	Auth       *auth.Auth
}

// Routes adds the export routes to the application.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.VExportBus)

	app.HandlerFunc(http.MethodGet, version, "/export", api.query, bearer)
}
