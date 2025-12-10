// Package healthcheck provides a simple job for testing the job queue.
package healthcheck

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/francowini/rafiki/foundation/jobqueue"
	"github.com/francowini/rafiki/foundation/logger"
)

// Args contains the data needed to process this job.
type Args struct {
	Message string `json:"message"`
}

// Kind returns the unique identifier for this job type.
func (Args) Kind() string {
	return "healthcheck"
}

// InsertOpts returns default insertion options.
func (Args) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       jobqueue.QueueDefault,
		MaxAttempts: 1,
	}
}

// Worker processes healthcheck jobs.
type Worker struct {
	river.WorkerDefaults[Args]
	log *logger.Logger
}

// NewWorker creates a new healthcheck worker.
func NewWorker(log *logger.Logger) *Worker {
	return &Worker{log: log}
}

// Work processes a single job.
func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	return jobqueue.JobMiddleware(ctx, w.log, job, func() error {
		w.log.Info(ctx, "healthcheck_job",
			"message", job.Args.Message,
		)
		return nil
	})
}
