package jobqueue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"github.com/francowini/rafiki/foundation/logger"
)

// Client wraps River with Rafiki-specific configuration and logging.
type Client struct {
	river  *river.Client[pgx.Tx]
	driver *riverpgxv5.Driver
	log    *logger.Logger
	config Config
}

// NewClient creates a new job queue client.
//
// Example:
//
//	workers := jobqueue.NewWorkers()
//	jobqueue.AddWorker(workers, &mypackage.MyWorker{})
//
//	client, err := jobqueue.NewClient(ctx, log, dbPool, jobqueue.DefaultConfig(), workers)
//	if err != nil {
//	    return fmt.Errorf("create job queue: %w", err)
//	}
//	defer client.Stop(ctx)
func NewClient(ctx context.Context, log *logger.Logger, dbPool *pgxpool.Pool, cfg Config, workers *river.Workers) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create a slog logger that wraps our logger
	slogLogger := slog.New(newSlogHandler(log))

	driver := riverpgxv5.New(dbPool)

	riverCfg := &river.Config{
		Queues:               cfg.Queues,
		Workers:              workers,
		Logger:               slogLogger,
		FetchCooldown:        cfg.FetchCooldown,
		FetchPollInterval:    cfg.FetchPollInterval,
		JobTimeout:           cfg.JobTimeout,
		RescueStuckJobsAfter: cfg.RescueStuckJobsAfter,
	}

	riverClient, err := river.NewClient(driver, riverCfg)
	if err != nil {
		return nil, fmt.Errorf("create river client: %w", err)
	}

	return &Client{
		river:  riverClient,
		driver: driver,
		log:    log,
		config: cfg,
	}, nil
}

// Start begins processing jobs. Call this after registering all workers.
func (c *Client) Start(ctx context.Context) error {
	c.log.Info(ctx, "jobqueue", "status", "starting", "queues", len(c.config.Queues))

	if err := c.river.Start(ctx); err != nil {
		return fmt.Errorf("start river: %w", err)
	}

	c.log.Info(ctx, "jobqueue", "status", "started")
	return nil
}

// Stop gracefully shuts down the job queue, waiting for in-flight jobs.
func (c *Client) Stop(ctx context.Context) error {
	c.log.Info(ctx, "jobqueue", "status", "stopping")

	if err := c.river.Stop(ctx); err != nil {
		return fmt.Errorf("stop river: %w", err)
	}

	c.log.Info(ctx, "jobqueue", "status", "stopped")
	return nil
}

// Insert enqueues a job for background processing.
//
// Example:
//
//	_, err := client.Insert(ctx, MyJobArgs{UserID: "123"}, nil)
func (c *Client) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if opts == nil {
		opts = &river.InsertOpts{}
	}

	// Apply retention policy if configured for this job type
	if policy, exists := c.config.RetentionPolicies[args.Kind()]; exists {
		// Note: River handles retention via maintenance services
		c.log.Debug(ctx, "jobqueue_insert", "kind", args.Kind(), "retention", policy.Duration)
	}

	result, err := c.river.Insert(ctx, args, opts)
	if err != nil {
		return nil, fmt.Errorf("insert job %s: %w", args.Kind(), err)
	}

	c.log.Debug(ctx, "jobqueue_insert",
		"job_id", result.Job.ID,
		"kind", args.Kind(),
		"queue", result.Job.Queue,
		"state", result.Job.State,
	)

	return result, nil
}

// InsertTx enqueues a job within a database transaction.
// The job is only visible after the transaction commits.
//
// Example:
//
//	tx, _ := db.Begin(ctx)
//	defer tx.Rollback(ctx)
//
//	// Create business entity
//	moment, err := momentStore.Create(ctx, tx, data)
//
//	// Enqueue notification (atomic with moment creation)
//	_, err = client.InsertTx(ctx, tx, NotifyArgs{MomentID: moment.ID}, nil)
//
//	tx.Commit(ctx) // Both moment AND job are committed together
func (c *Client) InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if opts == nil {
		opts = &river.InsertOpts{}
	}

	result, err := c.river.InsertTx(ctx, tx, args, opts)
	if err != nil {
		return nil, fmt.Errorf("insert job %s in tx: %w", args.Kind(), err)
	}

	c.log.Debug(ctx, "jobqueue_insert_tx",
		"job_id", result.Job.ID,
		"kind", args.Kind(),
		"queue", result.Job.Queue,
	)

	return result, nil
}

// InsertMany enqueues multiple jobs in a single database operation.
func (c *Client) InsertMany(ctx context.Context, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error) {
	results, err := c.river.InsertMany(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("insert many jobs: %w", err)
	}

	c.log.Debug(ctx, "jobqueue_insert_many", "count", len(results))
	return results, nil
}

// ScheduleAt enqueues a job to run at a specific time.
//
// Example:
//
//	// Send reminder tomorrow at 8 PM
//	tomorrow8PM := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour).Add(20 * time.Hour)
//	_, err := client.ScheduleAt(ctx, ReminderArgs{UserID: "123"}, tomorrow8PM, nil)
func (c *Client) ScheduleAt(ctx context.Context, args river.JobArgs, scheduledAt time.Time, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	// Create a new InsertOpts to avoid mutating the caller's opts
	newOpts := river.InsertOpts{
		ScheduledAt: scheduledAt,
	}

	// Copy fields from provided opts if non-nil
	if opts != nil {
		newOpts.MaxAttempts = opts.MaxAttempts
		newOpts.Priority = opts.Priority
		newOpts.Queue = opts.Queue
		newOpts.Tags = opts.Tags
		newOpts.UniqueOpts = opts.UniqueOpts
		newOpts.Pending = opts.Pending
		newOpts.Metadata = opts.Metadata
	}

	return c.Insert(ctx, args, &newOpts)
}

// River returns the underlying River client for advanced operations.
// Use sparingly - prefer the wrapped methods above.
func (c *Client) River() *river.Client[pgx.Tx] {
	return c.river
}

// Driver returns the underlying River driver for testing operations.
func (c *Client) Driver() *riverpgxv5.Driver {
	return c.driver
}
