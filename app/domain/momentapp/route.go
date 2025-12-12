package momentapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all the configuration for the moment API
type Config struct {
	MomentBus *momentbus.Business
	Auth      *auth.Auth
}

// Routes adds the routes for the moment api
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Create Bearer auth middleware
	bearer := mid.Bearer(cfg.Auth)

	api := newApp(cfg.MomentBus)

	// Apply bearer middleware to all routes
	app.HandlerFunc(http.MethodPost, version, "/moments", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/moments", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/moments/stats", api.queryStats, bearer)
	app.HandlerFunc(http.MethodGet, version, "/moments/{moment_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/moments/{moment_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/moments/{moment_id}", api.delete, bearer)
}
