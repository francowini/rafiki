package objetivoregistroapp

import (
	"net/http"

	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/domain/objetivoregistrobus"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains all dependencies needed for route setup.
type Config struct {
	ObjetivoBus         objetivobus.ExtBusiness
	ObjetivoRegistroBus objetivoregistrobus.ExtBusiness
	Auth                *auth.Auth
}

// Routes registers all objetivo record endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	api := newApp(cfg.ObjetivoBus, cfg.ObjetivoRegistroBus)

	// Record operations (nested under objetivos)
	app.HandlerFunc(http.MethodPost, version, "/objetivos/{objetivo_id}/records", api.create, bearer)
	app.HandlerFunc(http.MethodGet, version, "/objetivos/{objetivo_id}/records", api.query, bearer)

	// Delete record (flat endpoint for simpler access)
	app.HandlerFunc(http.MethodDelete, version, "/objetivos/records/{record_id}", api.delete, bearer)
}
