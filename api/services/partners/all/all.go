// Package all binds all the routes into the specified app.
package all

import (
	"github.com/francowini/rafiki/app/domain/authapp"
	"github.com/francowini/rafiki/app/domain/checkapp"
	"github.com/francowini/rafiki/app/domain/momentapp"
	"github.com/francowini/rafiki/app/domain/thinkapp"
	"github.com/francowini/rafiki/app/domain/valueapp"
	"github.com/francowini/rafiki/app/sdk/mux"
	"github.com/francowini/rafiki/foundation/web"
)

// Routes constructs the Add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() Add {
	return Add{}
}

// Add implements the RouterAdder interface for binding all routes.
type Add struct{}

// Add implements the RouterAdder interface.
func (Add) Add(app *web.App, cfg mux.Config) {
	checkapp.Routes(app, checkapp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	authapp.Routes(app, authapp.Config{
		Auth:    cfg.BusConfig.Auth,
		UserBus: cfg.BusConfig.UserBus,
	})

	thinkapp.Routes(app, thinkapp.Config{
		Log:      cfg.Log,
		ThinkBus: cfg.BusConfig.ThinkBus,
		Auth:     cfg.BusConfig.Auth,
	})

	momentapp.Routes(app, momentapp.Config{
		MomentBus: cfg.BusConfig.MomentBus,
		Auth:      cfg.BusConfig.Auth,
	})

	valueapp.Routes(app, valueapp.Config{
		ValueBus: cfg.BusConfig.ValueBus,
		Auth:     cfg.BusConfig.Auth,
	})
}
