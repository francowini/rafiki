// Package vobjectiveactivitydb provides database access for objective activity queries.
package vobjectiveactivitydb

import (
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
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
	return vobjectiveactivitybus.ActivityItem{
		ID:           row.ItemID,
		ActivityType: vobjectiveactivitybus.ActivityType(row.ActivityType),
		DateCreated:  row.DateCreated,
		RecordStatus: row.RecordStatus,
		Notes:        row.Notes,
		TaskTitle:    row.TaskTitle,
		Contribution: row.Contribution,
	}
}
