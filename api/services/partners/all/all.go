// Package all binds all the routes into the specified app.
package all

import (
	"github.com/francowini/rafiki/app/domain/authapp"
	"github.com/francowini/rafiki/app/domain/checkapp"
	"github.com/francowini/rafiki/app/domain/lifevisionapp"
	"github.com/francowini/rafiki/app/domain/momentapp"
	"github.com/francowini/rafiki/app/domain/objectiveapp"
	"github.com/francowini/rafiki/app/domain/objectiverecordapp"
	"github.com/francowini/rafiki/app/domain/taskapp"
	"github.com/francowini/rafiki/app/domain/telegramapp"
	"github.com/francowini/rafiki/app/domain/thinkapp"
	"github.com/francowini/rafiki/app/domain/valueapp"
	"github.com/francowini/rafiki/app/domain/vexportapp"
	"github.com/francowini/rafiki/app/domain/vobjectiveactivityapp"
	"github.com/francowini/rafiki/app/sdk/mux"
	"github.com/francowini/rafiki/business/sdk/sqldb"
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

	lifevisionapp.Routes(app, lifevisionapp.Config{
		LifeVisionBus: cfg.BusConfig.LifeVisionBus,
		Auth:          cfg.BusConfig.Auth,
	})

	objectiveapp.Routes(app, objectiveapp.Config{
		ObjectiveBus: cfg.BusConfig.ObjectiveBus,
		Auth:         cfg.BusConfig.Auth,
	})

	objectiverecordapp.Routes(app, objectiverecordapp.Config{
		ObjectiveBus:       cfg.BusConfig.ObjectiveBus,
		ObjectiveRecordBus: cfg.BusConfig.ObjectiveRecordBus,
		Auth:               cfg.BusConfig.Auth,
	})

	taskapp.Routes(app, taskapp.Config{
		TaskBus:      cfg.BusConfig.TaskBus,
		ObjectiveBus: cfg.BusConfig.ObjectiveBus,
		Auth:         cfg.BusConfig.Auth,
		DB:           sqldb.NewBeginner(cfg.DB),
	})

	vexportapp.Routes(app, vexportapp.Config{
		VExportBus: cfg.BusConfig.VExportBus,
		Auth:       cfg.BusConfig.Auth,
	})

	vobjectiveactivityapp.Routes(app, vobjectiveactivityapp.Config{
		Log:                   cfg.Log,
		VObjectiveActivityBus: cfg.BusConfig.VObjectiveActivityBus,
		Auth:                  cfg.BusConfig.Auth,
	})

	if cfg.TelegramClient != nil && cfg.WebhookSecret != "" {
		telegramapp.Routes(app, telegramapp.Config{
			Log:                cfg.Log,
			UserBus:            cfg.BusConfig.UserBus,
			TelegramSessionBus: cfg.BusConfig.TelegramSessionBus,
			TelegramClient:     cfg.TelegramClient,
			JobQueue:           cfg.JobQueue,
			WebhookSecret:      cfg.WebhookSecret,
		})
	}
}
