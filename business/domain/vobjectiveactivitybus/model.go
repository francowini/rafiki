// Package vobjectiveactivitybus provides read-only business logic for objective activity heatmaps.
package vobjectiveactivitybus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/contribution"
	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/recordstatus"
	"github.com/francowini/rafiki/business/types/tasktitle"
)

// Domain errors
var (
	ErrNotFound          = errors.New("activity not found")
	ErrNotObjectiveOwner = errors.New("user does not own objective")
)

// ActivityType represents the type of activity item.
type ActivityType string

// ActivityType values.
const (
	ActivityTypeRecord ActivityType = "record"
	ActivityTypeTask   ActivityType = "task"
)

// StreakUnit represents the unit of streak measurement.
type StreakUnit string

// StreakUnit values.
const (
	StreakUnitDays   StreakUnit = "days"
	StreakUnitWeeks  StreakUnit = "weeks"
	StreakUnitMonths StreakUnit = "months"
)

// ActivityItem represents a single activity (record or task completion) on a day.
type ActivityItem struct {
	ID           uuid.UUID
	ActivityType ActivityType
	DateCreated  time.Time

	// Record fields (nil for tasks)
	RecordStatus *recordstatus.RecordStatus
	Notes        *notes.Notes

	// Task fields (nil for records)
	TaskTitle    *tasktitle.TaskTitle
	Contribution *contribution.Contribution
}

// DayActivity represents activity data for a single day.
type DayActivity struct {
	Date        time.Time
	HasActivity bool
	Items       []ActivityItem
}

// Activity represents the complete activity data for an objective within a year.
type Activity struct {
	ObjectiveID      uuid.UUID
	UserID           uuid.UUID
	Year             int
	Days             []DayActivity
	TotalCompletions int
	CurrentStreak    int
	LongestStreak    int
	StreakUnit       StreakUnit
}
