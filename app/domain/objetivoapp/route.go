package objetivoapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	ObjetivoBus objetivobus.ExtBusiness
	Auth        *auth.Auth
}

// Routes registers all objetivo endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.ObjetivoBus)

	// CRUD operations
	app.HandlerFunc(http.MethodPost, version, "/objetivos", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/objetivos", api.query, bearer)
	app.HandlerFunc(http.MethodGet, version, "/objetivos/{objetivo_id}", api.queryByID, bearer)
	app.HandlerFunc(http.MethodPut, version, "/objetivos/{objetivo_id}", api.update, bearer)
	app.HandlerFunc(http.MethodDelete, version, "/objetivos/{objetivo_id}", api.delete, bearer)

	// Status and progress operations
	app.HandlerFunc(http.MethodPut, version, "/objetivos/{objetivo_id}/status", api.changeStatus, bearer)
	app.HandlerFunc(http.MethodPut, version, "/objetivos/{objetivo_id}/progress", api.incrementProgress, bearer)

	// Archive and restore
	app.HandlerFunc(http.MethodPut, version, "/objetivos/{objetivo_id}/archive", api.archive, bearer)
	app.HandlerFunc(http.MethodPut, version, "/objetivos/{objetivo_id}/restore", api.restore, bearer)

	// Query by life vision (nested endpoint)
	app.HandlerFunc(http.MethodGet, version, "/lifevisions/{lifevision_id}/objetivos", api.queryByLifeVision, bearer)
}
