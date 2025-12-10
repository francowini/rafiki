package jobqueue

import (
	"github.com/riverqueue/river"
)

// NewWorkers creates a new worker registry.
// Register workers before creating the Client.
//
// Example:
//
//	workers := jobqueue.NewWorkers()
//	jobqueue.AddWorker(workers, &mypackage.MyWorker{})
//	jobqueue.AddWorker(workers, &otherpackage.OtherWorker{})
//
//	client, err := jobqueue.NewClient(ctx, log, dbPool, cfg, workers)
func NewWorkers() *river.Workers {
	return river.NewWorkers()
}

// AddWorker registers a worker for a specific job type.
// The worker must implement river.Worker[T] where T is the job args type.
func AddWorker[T river.JobArgs](workers *river.Workers, worker river.Worker[T]) {
	river.AddWorker(workers, worker)
}
