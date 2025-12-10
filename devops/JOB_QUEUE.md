# Job Queue (River)

Background job processing using [River Queue](https://riverqueue.com/).

## Overview

River is a PostgreSQL-based job queue. Jobs are stored in `river_job` table and processed by workers embedded in the service.

**Key benefits:**
- Uses existing PostgreSQL (no Redis)
- Transactional job insertion (jobs commit with business data)
- Type-safe Go generics

## Configuration

Default configuration in `foundation/jobqueue/config.go`:

```go
Queues: map[string]river.QueueConfig{
    "default": {MaxWorkers: 4},  // General purpose
    "high":    {MaxWorkers: 3},  // User-facing, fast jobs
    "low":     {MaxWorkers: 1},  // Background maintenance
}
JobTimeout:           5 * time.Minute
RescueStuckJobsAfter: 1 * time.Hour
```

**Environment variables** (optional overrides):
```bash
PARTNER_JOBQUEUE_ENABLED=true
PARTNER_JOBQUEUE_DEFAULT_WORKERS=4
PARTNER_JOBQUEUE_HIGH_WORKERS=3
PARTNER_JOBQUEUE_LOW_WORKERS=1
```

## Creating a Job

### 1. Define Job Args

```go
// business/jobs/myjob/myjob.go
package myjob

type Args struct {
    UserID string `json:"user_id"`
}

func (Args) Kind() string { return "my_job" }

func (Args) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        Queue:       jobqueue.QueueDefault,
        MaxAttempts: 5,
    }
}
```

### 2. Define Worker

```go
type Worker struct {
    river.WorkerDefaults[Args]
    log *logger.Logger
}

func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
    return jobqueue.JobMiddleware(ctx, w.log, job, func() error {
        // Do work here
        return nil
    })
}
```

### 3. Register Worker

In `main.go`:
```go
workers := jobqueue.NewWorkers()
jobqueue.AddWorker(workers, myjob.NewWorker(log))

jq, err := jobqueue.NewClient(ctx, log, pgxPool, jqConfig, workers)
```

### 4. Insert Jobs

```go
// Simple insert
_, err := jq.Insert(ctx, myjob.Args{UserID: "123"}, nil)

// Transactional insert (job commits with business data)
tx, _ := db.Begin(ctx)
_, err = jq.InsertTx(ctx, tx, myjob.Args{UserID: "123"}, nil)
tx.Commit(ctx)

// Scheduled job
_, err = jq.ScheduleAt(ctx, myjob.Args{}, time.Now().Add(1*time.Hour), nil)
```

## Monitoring

```sql
-- Job counts by state
SELECT state, COUNT(*) FROM river_job GROUP BY state;

-- Pending jobs per queue
SELECT queue, COUNT(*) FROM river_job WHERE state = 'available' GROUP BY queue;

-- Recent failures
SELECT id, kind, errors[array_upper(errors, 1)] as last_error
FROM river_job WHERE state = 'discarded'
ORDER BY finalized_at DESC LIMIT 10;
```

## Job States

| State | Description |
|-------|-------------|
| `available` | Ready to be processed |
| `running` | Currently being processed |
| `completed` | Successfully finished |
| `retryable` | Failed, will retry |
| `discarded` | Failed all retries |
| `cancelled` | Manually cancelled |
| `scheduled` | Waiting for scheduled time |

## Transaction Strategies

The codebase uses `sqlx` for business logic but River requires `pgx.Tx`. These are incompatible transaction types.

### Choose by Job Type

| Job Type | Strategy | Trade-off |
|----------|----------|-----------|
| **Fire & forget** (notifications, analytics, emails) | Non-transactional | Job may not exist if insert fails after commit |
| **Critical** (payments, exports) | Dual pool | Extra complexity, but atomic guarantees |

### Strategy A: Non-transactional (Recommended for most cases)

Insert job **after** business transaction commits:

```go
func (b *Business) Create(ctx context.Context, new NewLifeVision) (LifeVision, error) {
    // 1. Business logic with sqlx (existing code, no changes)
    lv, err := b.storer.Create(ctx, new)
    if err != nil {
        return LifeVision{}, err
    }

    // 2. Job after commit (non-transactional)
    if _, err := b.jq.Insert(ctx, myJob.Args{ID: lv.ID}, nil); err != nil {
        b.log.Error(ctx, "job_insert_failed", "err", err)
        // Don't return error - business data is already saved
    }

    return lv, nil
}
```

**When to use**: Notifications, analytics, emails, background processing where occasional loss is acceptable.

### Strategy B: Dual Pool (For critical jobs)

Use `pgxpool` directly when you need atomic job insertion:

```go
func (b *Business) CreateWithCriticalJob(ctx context.Context, data Data) error {
    // Start pgx transaction
    tx, err := b.pgxPool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // 1. Insert business data (using pgx, not sqlx)
    _, err = tx.Exec(ctx, `INSERT INTO items (id, data) VALUES ($1, $2)`, id, data)
    if err != nil {
        return err
    }

    // 2. Insert job in SAME transaction
    _, err = b.jq.InsertTx(ctx, tx, criticalJob.Args{ID: id}, nil)
    if err != nil {
        return err
    }

    // 3. Both commit together or both rollback
    return tx.Commit(ctx)
}
```

**When to use**: Payment processing, critical exports, anything where job loss means data inconsistency.

**Cost**: Requires writing pgx-style queries (`$1` params instead of `:name`).

### Why Not Migrate Everything to pgx?

| Approach | Effort | Benefit |
|----------|--------|---------|
| Keep sqlx + non-transactional jobs | Zero | Works for 90% of cases |
| Dual pool for critical paths | Low | Atomic where needed |
| Full migration to pgx | High (~2-3 days) | Marginal performance gain |

**Recommendation**: Start with non-transactional (Strategy A). Add dual pool (Strategy B) only for specific critical jobs.

## Troubleshooting

**Jobs not processing:**
```bash
# Check leader election
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT * FROM river_leader;"

# Check logs
docker compose logs partner-service | grep jobqueue
```

**High failure rate:**
```sql
SELECT kind, COUNT(*) FROM river_job
WHERE state = 'discarded' GROUP BY kind;
```

## Files

- `foundation/jobqueue/` - Core SDK
- `business/sdk/migrate/sql/migrate.sql` - River tables (version 1.08)
- `docs/river-queue-foundation.md` - Detailed implementation guide
