// Package telegramnotify provides a job worker for sending Telegram notifications.
// This is a thin orchestration layer that delegates business logic to notificationbus.
package telegramnotify

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"

	"github.com/francowini/rafiki/business/domain/notificationbus"
	"github.com/francowini/rafiki/business/domain/vnotificationbus"
	"github.com/francowini/rafiki/foundation/jobqueue"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/telegram"
)

// Config holds the configuration for the Telegram notification worker.
type Config struct {
	MorningTime string // Format: "08:00"
	EveningTime string // Format: "21:00"
	Timezone    string // e.g., "America/Argentina/Buenos_Aires"
}

// Args contains the data needed to process this job.
type Args struct {
	// Empty struct - this job processes all pending messages
}

// Kind returns the unique identifier for this job type.
func (Args) Kind() string {
	return "telegram_notify"
}

// InsertOpts returns default insertion options.
func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       jobqueue.QueueDefault,
		MaxAttempts: 1, // This job should not retry - it handles retries internally
		UniqueOpts: river.UniqueOpts{
			ByPeriod: 5 * time.Minute, // Deduplicate jobs within 5 minute window
		},
	}
}

// telegramSenderAdapter adapts telegram.Client to notificationbus.TelegramSender.
type telegramSenderAdapter struct {
	client *telegram.Client
}

func (a *telegramSenderAdapter) SendMessage(ctx context.Context, chatID int64, content string) (notificationbus.TelegramSendResponse, error) {
	resp, err := a.client.SendMessage(ctx, chatID, content)
	if err != nil {
		return notificationbus.TelegramSendResponse{}, err
	}
	return notificationbus.TelegramSendResponse{
		MessageID: resp.Result.MessageID,
	}, nil
}

// Worker processes telegram notification jobs.
type Worker struct {
	river.WorkerDefaults[Args]
	log              *logger.Logger
	notificationBus  *notificationbus.Business
	vnotificationBus *vnotificationbus.Business
	telegramSender   notificationbus.TelegramSender
	config           Config
	location         *time.Location
}

// NewWorker creates a new telegram notification worker.
// Returns an error if the configuration is invalid (bad timezone or time format).
func NewWorker(
	log *logger.Logger,
	notificationBus *notificationbus.Business,
	vnotificationBus *vnotificationbus.Business,
	telegramClient *telegram.Client,
	cfg Config,
) (*Worker, error) {
	// Validate timezone
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", cfg.Timezone, err)
	}

	// Validate MorningTime format (HH:MM)
	if _, err := time.Parse("15:04", cfg.MorningTime); err != nil {
		return nil, fmt.Errorf("invalid morning time %q (expected HH:MM format): %w", cfg.MorningTime, err)
	}

	// Validate EveningTime format (HH:MM)
	if _, err := time.Parse("15:04", cfg.EveningTime); err != nil {
		return nil, fmt.Errorf("invalid evening time %q (expected HH:MM format): %w", cfg.EveningTime, err)
	}

	return &Worker{
		log:              log,
		notificationBus:  notificationBus,
		vnotificationBus: vnotificationBus,
		telegramSender:   &telegramSenderAdapter{client: telegramClient},
		config:           cfg,
		location:         loc,
	}, nil
}

// Work processes a single job.
func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	return jobqueue.JobMiddleware(ctx, w.log, job, func() error {
		return w.processNotifications(ctx)
	})
}

func (w *Worker) processNotifications(ctx context.Context) error {
	now := time.Now().In(w.location)

	// 1. Get all users with Telegram enabled
	users, err := w.notificationBus.QueryTelegramUsers(ctx)
	if err != nil {
		return fmt.Errorf("query telegram users: %w", err)
	}

	if len(users) == 0 {
		w.log.Info(ctx, "telegram_notify", "msg", "no telegram users found")
		return nil
	}

	// 2. Schedule messages (delegates to business layer)
	cfg := notificationbus.ScheduleConfig{
		MorningTime: w.config.MorningTime,
		EveningTime: w.config.EveningTime,
		Location:    w.location,
	}

	// Log time check for debugging
	w.log.Info(ctx, "telegram_notify", "msg", "checking time windows",
		"current_time", now.Format("15:04"),
		"morning_target", cfg.MorningTime,
		"evening_target", cfg.EveningTime,
		"users_count", len(users))

	for _, user := range users {
		// Business layer decides if it's time to schedule
		if err := w.notificationBus.ScheduleMorningMessageIfTime(ctx, user.UserID, cfg, now, w.vnotificationBus); err != nil {
			w.log.Error(ctx, "telegram_notify", "msg", "schedule morning failed", "user_id", user.UserID, "err", err)
		}
		if err := w.notificationBus.ScheduleEveningMessageIfTime(ctx, user.UserID, cfg, now, w.vnotificationBus); err != nil {
			w.log.Error(ctx, "telegram_notify", "msg", "schedule evening failed", "user_id", user.UserID, "err", err)
		}
	}

	// 3. Send pending messages (delegates to business layer)
	if err := w.notificationBus.SendPendingMessages(ctx, w.telegramSender, now); err != nil {
		return fmt.Errorf("send pending messages: %w", err)
	}

	return nil
}
