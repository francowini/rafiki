// Package mux provides support to bind domain level routes
// to the application mux.
package mux

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/domain/vexportbus"
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
	ThinkBus      *thinkbus.Business
	MomentBus     *momentbus.Business
	ValueBus      valuebus.ExtBusiness
	LifeVisionBus lifevisionbus.ExtBusiness
	RoleBus       rolebus.ExtBusiness
	VExportBus    vexportbus.ExtBusiness
	UserBus       userbus.ExtBusiness
	Auth          *auth.Auth
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
// Returns an error if file server configuration fails.
func WebAPI(cfg Config, routeAdder RouteAdder, options ...func(opts *Options)) (http.Handler, error) {
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
		var err error
		if site.react {
			err = app.FileServerReact(site.static, site.staticDir, site.staticPath)
		} else {
			err = app.FileServer(site.static, site.staticDir, site.staticPath)
		}
		if err != nil {
			return nil, fmt.Errorf("configuring file server for %s: %w", site.staticPath, err)
		}
	}

	return app, nil
}
