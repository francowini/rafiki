// Package vobjectiveactivitybus provides read-only business logic for objective activity heatmaps.
package vobjectiveactivitybus

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors
var (
	ErrNotFound          = errors.New("activity not found")
	ErrInvalidYear       = errors.New("year must be between 2000 and 2100")
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
	RecordStatus *string
	Notes        *string

	// Task fields (nil for records)
	TaskTitle    *string
	Contribution *int
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
