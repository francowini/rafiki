// Package telegramnotify provides a job worker for sending Telegram notifications.
package telegramnotify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// Worker processes telegram notification jobs.
type Worker struct {
	river.WorkerDefaults[Args]
	log              *logger.Logger
	notificationBus  *notificationbus.Business
	vnotificationBus *vnotificationbus.Business
	telegramClient   *telegram.Client
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
		telegramClient:   telegramClient,
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

	// 1. Schedule new messages if it's time
	if err := w.scheduleMessages(ctx, now); err != nil {
		w.log.Error(ctx, "telegram_notify", "msg", "failed to schedule messages", "err", err)
		// Don't return error - continue to process pending messages
	}

	// 2. Process pending messages
	if err := w.sendPendingMessages(ctx); err != nil {
		return fmt.Errorf("send pending messages: %w", err)
	}

	return nil
}

func (w *Worker) scheduleMessages(ctx context.Context, now time.Time) error {
	// Get all users with Telegram enabled
	users, err := w.notificationBus.QueryTelegramUsers(ctx)
	if err != nil {
		w.log.Error(ctx, "telegram_notify", "msg", "failed to query telegram users", "err", err)
		return fmt.Errorf("query telegram users: %w", err)
	}

	if len(users) == 0 {
		w.log.Info(ctx, "telegram_notify", "msg", "no telegram users found")
		return nil
	}

	currentTime := now.Format("15:04")
	morningWindow := w.isInTimeWindow(currentTime, w.config.MorningTime)
	eveningWindow := w.isInTimeWindow(currentTime, w.config.EveningTime)

	// Always log the time check for debugging
	w.log.Info(ctx, "telegram_notify", "msg", "checking time windows",
		"current_time", currentTime,
		"morning_target", w.config.MorningTime,
		"evening_target", w.config.EveningTime,
		"morning_window", morningWindow,
		"evening_window", eveningWindow,
		"users_count", len(users))

	if !morningWindow && !eveningWindow {
		return nil
	}

	for _, user := range users {
		if morningWindow {
			if err := w.scheduleMorningMessage(ctx, user.UserID, now); err != nil {
				w.log.Error(ctx, "telegram_notify", "msg", "failed to schedule morning message", "user_id", user.UserID, "err", err)
			}
		}

		if eveningWindow {
			if err := w.scheduleEveningMessage(ctx, user.UserID, now); err != nil {
				w.log.Error(ctx, "telegram_notify", "msg", "failed to schedule evening message", "user_id", user.UserID, "err", err)
			}
		}
	}

	return nil
}

func (w *Worker) isInTimeWindow(current, target string) bool {
	// Parse times and check if current is within 10 minutes of target
	currentTime, err := time.Parse("15:04", current)
	if err != nil {
		return false
	}
	targetTime, err := time.Parse("15:04", target)
	if err != nil {
		return false
	}

	// Calculate difference in minutes
	diff := currentTime.Sub(targetTime).Minutes()

	// Within 10 minute window (0 to +9 minutes after target)
	return diff >= 0 && diff < 10
}

func (w *Worker) scheduleMorningMessage(ctx context.Context, userID uuid.UUID, now time.Time) error {
	// Get user's values with visions
	values, err := w.vnotificationBus.QueryByUserID(ctx, userID)
	if err != nil {
		w.log.Error(ctx, "telegram_notify", "msg", "failed to query values for morning message", "user_id", userID, "err", err)
		return fmt.Errorf("query values: %w", err)
	}

	// Generate message content
	content := notificationbus.GenerateMorningMessage(values)

	// Create message (idempotent - duplicate schedule is not an error)
	_, err = w.notificationBus.Create(ctx, notificationbus.NewMessage{
		UserID:      userID,
		MessageType: notificationbus.MessageTypeMorning,
		Content:     content,
		ScheduledAt: now,
	})
	if err != nil {
		if errors.Is(err, notificationbus.ErrDuplicateSchedule) {
			// Already scheduled for today - this is expected behavior
			w.log.Info(ctx, "telegram_notify", "msg", "morning message already scheduled", "user_id", userID)
			return nil
		}
		w.log.Error(ctx, "telegram_notify", "msg", "failed to create morning message", "user_id", userID, "err", err)
		return fmt.Errorf("create message: %w", err)
	}

	w.log.Info(ctx, "telegram_notify", "msg", "scheduled morning message", "user_id", userID)
	return nil
}

func (w *Worker) scheduleEveningMessage(ctx context.Context, userID uuid.UUID, now time.Time) error {
	// Get user's values with visions
	values, err := w.vnotificationBus.QueryByUserID(ctx, userID)
	if err != nil {
		w.log.Error(ctx, "telegram_notify", "msg", "failed to query values for evening message", "user_id", userID, "err", err)
		return fmt.Errorf("query values: %w", err)
	}

	// Generate message content
	content := notificationbus.GenerateEveningMessage(values)

	// Create message (idempotent - duplicate schedule is not an error)
	_, err = w.notificationBus.Create(ctx, notificationbus.NewMessage{
		UserID:      userID,
		MessageType: notificationbus.MessageTypeEvening,
		Content:     content,
		ScheduledAt: now,
	})
	if err != nil {
		if errors.Is(err, notificationbus.ErrDuplicateSchedule) {
			// Already scheduled for today - this is expected behavior
			w.log.Info(ctx, "telegram_notify", "msg", "evening message already scheduled", "user_id", userID)
			return nil
		}
		w.log.Error(ctx, "telegram_notify", "msg", "failed to create evening message", "user_id", userID, "err", err)
		return fmt.Errorf("create message: %w", err)
	}

	w.log.Info(ctx, "telegram_notify", "msg", "scheduled evening message", "user_id", userID)
	return nil
}

func (w *Worker) sendPendingMessages(ctx context.Context) error {
	// Get pending messages using configured timezone for consistency with scheduling
	now := time.Now().In(w.location)
	messages, err := w.notificationBus.QueryPending(ctx, now)
	if err != nil {
		w.log.Error(ctx, "telegram_notify", "msg", "failed to query pending messages", "err", err)
		return fmt.Errorf("query pending: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	w.log.Info(ctx, "telegram_notify", "msg", "processing pending messages", "count", len(messages))

	// Get all Telegram users for chat ID lookup
	users, err := w.notificationBus.QueryTelegramUsers(ctx)
	if err != nil {
		w.log.Error(ctx, "telegram_notify", "msg", "failed to query telegram users for sending", "err", err)
		return fmt.Errorf("query telegram users: %w", err)
	}

	// Create lookup map
	chatIDMap := make(map[uuid.UUID]notificationbus.TelegramChatID)
	for _, u := range users {
		chatIDMap[u.UserID] = u.TelegramChatID
	}

	// Process each message
	for _, msg := range messages {
		chatID, ok := chatIDMap[msg.UserID]
		if !ok {
			w.log.Error(ctx, "telegram_notify", "msg", "user telegram not enabled", "message_id", msg.ID, "user_id", msg.UserID)
			errMsg := notificationbus.FailureReason("user telegram not enabled")
			if _, err := w.notificationBus.MarkFailed(ctx, msg, errMsg); err != nil {
				w.log.Error(ctx, "telegram_notify", "msg", "failed to mark message as failed", "message_id", msg.ID, "err", err)
			}
			continue
		}

		// Send via Telegram
		w.log.Info(ctx, "telegram_notify", "msg", "sending telegram message", "message_id", msg.ID, "chat_id", chatID.Value())
		resp, err := w.telegramClient.SendMessage(ctx, chatID.Value(), msg.Content)
		if err != nil {
			w.log.Error(ctx, "telegram_notify", "msg", "telegram send failed", "message_id", msg.ID, "err", err)
			errMsg := notificationbus.FailureReason(err.Error())
			if _, err := w.notificationBus.MarkFailed(ctx, msg, errMsg); err != nil {
				w.log.Error(ctx, "telegram_notify", "msg", "failed to mark message as failed", "message_id", msg.ID, "err", err)
			}
			continue
		}

		// Mark as sent
		telegramMsgID := notificationbus.TelegramMessageID(resp.Result.MessageID)
		if _, err := w.notificationBus.MarkSent(ctx, msg, telegramMsgID); err != nil {
			w.log.Error(ctx, "telegram_notify", "msg", "failed to mark message as sent", "message_id", msg.ID, "err", err)
			continue
		}

		w.log.Info(ctx, "telegram_notify", "msg", "sent message", "message_id", msg.ID, "telegram_msg_id", resp.Result.MessageID)
	}

	return nil
}
