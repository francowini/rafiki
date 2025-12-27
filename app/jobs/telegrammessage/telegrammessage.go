// Package telegrammessage provides a job worker for processing Telegram messages
// with AI-guided conversation for moment tracking.
package telegrammessage

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/francowini/rafiki/app/domain/telegramapp"
	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/foundation/anthropic"
	"github.com/francowini/rafiki/foundation/jobqueue"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/telegram"
)

// Worker processes Telegram message jobs with AI validation.
type Worker struct {
	river.WorkerDefaults[telegramapp.TelegramMessageArgs]
	log             *logger.Logger
	sessionBus      *telegramsessionbus.Business
	anthropicClient *anthropic.Client
	telegramClient  *telegram.Client
}

// NewWorker creates a new Telegram message worker.
func NewWorker(
	log *logger.Logger,
	sessionBus *telegramsessionbus.Business,
	anthropicClient *anthropic.Client,
	telegramClient *telegram.Client,
) *Worker {
	return &Worker{
		log:             log,
		sessionBus:      sessionBus,
		anthropicClient: anthropicClient,
		telegramClient:  telegramClient,
	}
}

// Work processes a single job.
func (w *Worker) Work(ctx context.Context, job *river.Job[telegramapp.TelegramMessageArgs]) error {
	return jobqueue.JobMiddleware(ctx, w.log, job, func() error {
		return w.processMessage(ctx, job.Args)
	})
}
