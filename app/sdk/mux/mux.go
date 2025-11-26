// Package mux provides support to bind domain level routes
// to the application mux.
package mux

import (
	"embed"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/web"
)

// StaticSite represents a static site to run.
type StaticSite struct {
	react      bool
	static     embed.FS
	staticDir  string
	staticPath string
}

// Options represent optional parameters.
type Options struct {
	corsOrigin []string
	sites      []StaticSite
}

// WithCORS provides configuration options for CORS.
func WithCORS(origins []string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origins
	}
}

// WithFileServer provides configuration options for file server.
func WithFileServer(react bool, static embed.FS, dir, path string) func(opts *Options) {
	return func(opts *Options) {
		opts.sites = append(opts.sites, StaticSite{
			react:      react,
			static:     static,
			staticDir:  dir,
			staticPath: path,
		})
	}
}

// BusConfig contains the business layer dependencies for route handlers.
type BusConfig struct {
	ThinkBus  *thinkbus.Business
	MomentBus *momentbus.Business
	ValueBus  valuebus.ExtBusiness
	UserBus   userbus.ExtBusiness
	Auth      *auth.Auth
}

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build     string
	Log       *logger.Logger
	DB        *sqlx.DB
	Tracer    trace.Tracer
	BusConfig BusConfig
}

// RouteAdder defines behavior that sets the routes to bind for an instance
// of the service.
type RouteAdder interface {
	Add(app *web.App, cfg Config)
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg Config, routeAdder RouteAdder, options ...func(opts *Options)) http.Handler {
	app := web.NewApp(
		cfg.Log.Info,
		cfg.Tracer,
		mid.Otel(cfg.Tracer),
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	var opts Options
	for _, option := range options {
		option(&opts)
	}

	if len(opts.corsOrigin) > 0 {
		app.EnableCORS(opts.corsOrigin)
	}

	routeAdder.Add(app, cfg)

	for _, site := range opts.sites {
		if site.react {
			app.FileServerReact(site.static, site.staticDir, site.staticPath) //nolint:errcheck // Static files are embedded, paths are validated at compile time
		} else {
			app.FileServer(site.static, site.staticDir, site.staticPath) //nolint:errcheck // Static files are embedded, paths are validated at compile time
		}
	}

	return app
}
