package telegramapp

import (
	"time"

	"github.com/riverqueue/river"
	"github.com/francowini/rafiki/foundation/jobqueue"
)

// Kind returns the unique job type identifier.
func (TelegramMessageArgs) Kind() string {
	return "telegram_message"
}

// InsertOpts returns River insertion options.
func (TelegramMessageArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       jobqueue.QueueDefault,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByPeriod: 30 * time.Second, // 30s deduplication
		},
	}
}
