// Package jobqueue provides a job queue foundation using River Queue for background job processing.
package jobqueue

import (
	"fmt"
	"time"

	"github.com/riverqueue/river"
)

// Queue name constants for consistent usage across the application.
const (
	QueueDefault      = river.QueueDefault // "default"
	QueueHighPriority = "high"
	QueueLowPriority  = "low"
)

// Config holds job queue configuration.
type Config struct {
	// Queues defines worker pools per queue.
	// Key: queue name, Value: queue configuration.
	Queues map[string]river.QueueConfig

	// FetchCooldown is how long to wait between fetch attempts when the queue is empty.
	// Lower values = more responsive but more DB queries.
	// Default: 100ms
	FetchCooldown time.Duration

	// FetchPollInterval is the fallback polling interval when LISTEN/NOTIFY is unavailable.
	// Default: 1s
	FetchPollInterval time.Duration

	// JobTimeout is the maximum time a job can run before being marked as stuck.
	// Default: 5m
	JobTimeout time.Duration

	// RescueStuckJobsAfter is how long a job can be in "running" state before rescue.
	// Default: 1h
	RescueStuckJobsAfter time.Duration

	// RetentionPolicies defines how long to keep completed/failed jobs per job type.
	// Key: job kind (from JobArgs.Kind()), Value: retention policy.
	RetentionPolicies map[string]RetentionPolicy
}

// RetentionPolicy defines cleanup rules for a job type.
type RetentionPolicy struct {
	// Duration is how long to keep completed jobs before deletion.
	Duration time.Duration
}

// Validate checks configuration validity.
func (c Config) Validate() error {
	if len(c.Queues) == 0 {
		return fmt.Errorf("at least one queue must be configured")
	}

	totalWorkers := 0
	for name, q := range c.Queues {
		if q.MaxWorkers < 1 {
			return fmt.Errorf("queue %q must have at least 1 worker", name)
		}
		totalWorkers += q.MaxWorkers
	}

	if totalWorkers > 100 {
		return fmt.Errorf("total workers (%d) exceeds recommended maximum (100)", totalWorkers)
	}

	return nil
}

// DefaultConfig returns configuration optimized for Hetzner CPX11 (2 vCPU, 2GB RAM).
//
// Worker count calculation:
//   - Available CPU: 1.5 cores (0.5 reserved for main service)
//   - I/O multiplier: 2-3x for network-bound jobs
//   - Memory: ~10-20MB per worker
//   - Total: 8 workers = ~160MB, well within limits
func DefaultConfig() Config {
	return Config{
		Queues: map[string]river.QueueConfig{
			QueueDefault:      {MaxWorkers: 4}, // General purpose
			QueueHighPriority: {MaxWorkers: 3}, // User-facing, fast jobs
			QueueLowPriority:  {MaxWorkers: 1}, // Background maintenance
		},
		FetchCooldown:        100 * time.Millisecond,
		FetchPollInterval:    1 * time.Second,
		JobTimeout:           5 * time.Minute,
		RescueStuckJobsAfter: 1 * time.Hour,
		// RetentionPolicies: Empty by default. River's job cleaner handles
		// retention automatically. Configure per job type using WithRetention()
		// if you need custom retention periods.
		RetentionPolicies: map[string]RetentionPolicy{},
	}
}

// WithQueue returns a new Config with an additional or modified queue.
func (c Config) WithQueue(name string, maxWorkers int) Config {
	newQueues := make(map[string]river.QueueConfig, len(c.Queues)+1)
	for k, v := range c.Queues {
		newQueues[k] = v
	}
	newQueues[name] = river.QueueConfig{MaxWorkers: maxWorkers}
	c.Queues = newQueues
	return c
}

// WithRetention returns a new Config with a retention policy for a job type.
func (c Config) WithRetention(jobKind string, duration time.Duration) Config {
	if c.RetentionPolicies == nil {
		c.RetentionPolicies = make(map[string]RetentionPolicy)
	}
	newPolicies := make(map[string]RetentionPolicy, len(c.RetentionPolicies)+1)
	for k, v := range c.RetentionPolicies {
		newPolicies[k] = v
	}
	newPolicies[jobKind] = RetentionPolicy{Duration: duration}
	c.RetentionPolicies = newPolicies
	return c
}
