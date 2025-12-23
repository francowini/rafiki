// Package vobjectiveactivitydb provides database access for objective activity queries.
package vobjectiveactivitydb

import (
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
	"github.com/francowini/rafiki/business/types/contribution"
	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/recordstatus"
	"github.com/francowini/rafiki/business/types/tasktitle"
)

// activityRow represents a single activity row from the database.
type activityRow struct {
	ObjectiveID  uuid.UUID `db:"objective_id"`
	UserID       uuid.UUID `db:"user_id"`
	ActivityDate time.Time `db:"activity_date"`
	ActivityType string    `db:"activity_type"`
	ItemID       uuid.UUID `db:"item_id"`
	TaskTitle    *string   `db:"task_title"`
	Contribution *int      `db:"contribution"`
	RecordStatus *string   `db:"record_status"`
	Notes        *string   `db:"notes"`
	DateCreated  time.Time `db:"date_created"`
}

// toBusActivityItem converts a database row to a business activity item.
func toBusActivityItem(row activityRow) vobjectiveactivitybus.ActivityItem {
	item := vobjectiveactivitybus.ActivityItem{
		ID:           row.ItemID,
		ActivityType: vobjectiveactivitybus.ActivityType(row.ActivityType),
		DateCreated:  row.DateCreated,
	}

	// Convert RecordStatus if present
	if row.RecordStatus != nil {
		if rs, err := recordstatus.Parse(*row.RecordStatus); err == nil {
			item.RecordStatus = &rs
		}
	}

	// Convert Notes if present
	if row.Notes != nil {
		if n, err := notes.Parse(*row.Notes); err == nil {
			item.Notes = &n
		}
	}

	// Convert TaskTitle if present
	if row.TaskTitle != nil {
		if tt, err := tasktitle.Parse(*row.TaskTitle); err == nil {
			item.TaskTitle = &tt
		}
	}

	// Convert Contribution if present
	if row.Contribution != nil {
		if c, err := contribution.Parse(*row.Contribution); err == nil {
			item.Contribution = &c
		}
	}

	return item
}
