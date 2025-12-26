// Package sessioncleanup provides a job worker for cleaning up expired Telegram sessions.
package sessioncleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"

	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/foundation/jobqueue"
	"github.com/francowini/rafiki/foundation/logger"
)

// SessionTTL is the duration after which inactive sessions are considered expired.
const SessionTTL = 15 * time.Minute

// Args contains the data needed to process this job.
type Args struct{}

// Kind returns the unique identifier for this job type.
func (Args) Kind() string {
	return "session_cleanup"
}

// InsertOpts returns default insertion options.
func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       jobqueue.QueueLowPriority,
		MaxAttempts: 1,
		UniqueOpts: river.UniqueOpts{
			ByPeriod: 5 * time.Minute,
		},
	}
}

// Worker processes session cleanup jobs.
type Worker struct {
	river.WorkerDefaults[Args]
	log        *logger.Logger
	sessionBus *telegramsessionbus.Business
}

// NewWorker creates a new session cleanup worker.
func NewWorker(log *logger.Logger, sessionBus *telegramsessionbus.Business) *Worker {
	return &Worker{
		log:        log,
		sessionBus: sessionBus,
	}
}

// Work processes a single cleanup job.
func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	return jobqueue.JobMiddleware(ctx, w.log, job, func() error {
		return w.cleanupSessions(ctx)
	})
}

func (w *Worker) cleanupSessions(ctx context.Context) error {
	count, err := w.sessionBus.CleanupExpired(ctx, SessionTTL)
	if err != nil {
		return fmt.Errorf("cleanup expired sessions: %w", err)
	}

	w.log.Info(ctx, "session_cleanup.completed",
		"sessions_deleted", count,
		"ttl_minutes", int(SessionTTL.Minutes()),
	)

	return nil
}
