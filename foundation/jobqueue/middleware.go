package jobqueue

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"

	"github.com/francowini/rafiki/foundation/logger"
)

// slogHandler adapts Rafiki's logger to slog.Handler interface for River.
type slogHandler struct {
	log    *logger.Logger
	attrs  []slog.Attr
	groups []string // Nested group names for prefixing attributes
}

func newSlogHandler(log *logger.Logger) *slogHandler {
	return &slogHandler{log: log}
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	// Convert slog attributes to key-value pairs
	args := make([]any, 0, r.NumAttrs()*2+len(h.attrs)*2)

	// Build group prefix if any groups are set
	groupPrefix := ""
	if len(h.groups) > 0 {
		for _, g := range h.groups {
			groupPrefix += g + "."
		}
	}

	// Add handler-level attributes first (with group prefix)
	for _, attr := range h.attrs {
		args = append(args, groupPrefix+attr.Key, attr.Value.Any())
	}

	// Add record attributes (with group prefix)
	r.Attrs(func(a slog.Attr) bool {
		args = append(args, groupPrefix+a.Key, a.Value.Any())
		return true
	})

	switch r.Level {
	case slog.LevelDebug:
		h.log.Debug(ctx, r.Message, args...)
	case slog.LevelInfo:
		h.log.Info(ctx, r.Message, args...)
	case slog.LevelWarn:
		h.log.Warn(ctx, r.Message, args...)
	case slog.LevelError:
		h.log.Error(ctx, r.Message, args...)
	default:
		h.log.Info(ctx, r.Message, args...)
	}

	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &slogHandler{
		log:    h.log,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &slogHandler{
		log:    h.log,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// JobMiddleware provides common job execution logging.
// Wrap your worker's Work method with this for consistent logging.
//
// Example:
//
//	func (w *MyWorker) Work(ctx context.Context, job *river.Job[MyArgs]) error {
//	    return jobqueue.JobMiddleware(ctx, w.log, job, func() error {
//	        // Actual work here
//	        return nil
//	    })
//	}
func JobMiddleware[T river.JobArgs](ctx context.Context, log *logger.Logger, job *river.Job[T], work func() error) error {
	log.Info(ctx, "job_started",
		"job_id", job.ID,
		"kind", job.Kind,
		"queue", job.Queue,
		"attempt", job.Attempt,
	)

	err := work()

	if err != nil {
		log.Error(ctx, "job_failed",
			"job_id", job.ID,
			"kind", job.Kind,
			"attempt", job.Attempt,
			"err", err,
		)
	} else {
		log.Info(ctx, "job_completed",
			"job_id", job.ID,
			"kind", job.Kind,
			"attempt", job.Attempt,
		)
	}

	return err
}
