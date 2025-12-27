package telegramapp

import (
	"net/http"

	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/foundation/jobqueue"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/telegram"
	"github.com/francowini/rafiki/foundation/web"
)

// Config contains dependencies for Telegram webhook handlers.
type Config struct {
	Log                *logger.Logger
	UserBus            userbus.ExtBusiness
	TelegramSessionBus *telegramsessionbus.Business
	TelegramClient     *telegram.Client
	JobQueue           *jobqueue.Client
	WebhookSecret      string
}

// Routes registers the Telegram webhook endpoint.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newApp(
		cfg.Log,
		cfg.UserBus,
		cfg.TelegramSessionBus,
		cfg.TelegramClient,
		cfg.JobQueue,
	)

	// Webhook endpoint with signature verification middleware
	app.HandlerFunc(
		http.MethodPost,
		version,
		"/telegram/webhook",
		api.webhook,
		verifyTelegramSignature(cfg.WebhookSecret),
	)
}
