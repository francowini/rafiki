package vexportbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

// Error variables for export operations.
var (
	ErrNotFound         = errors.New("export items not found")
	ErrInvalidDateRange = errors.New("start_date must be before end_date")
)

// ItemType represents the type of export item (moment or think).
type ItemType string

// Item type constants.
const (
	ItemTypeMoment ItemType = "moment"
	ItemTypeThink  ItemType = "think"
)

func (it ItemType) String() string {
	return string(it)
}

// ExportItem represents a unified export item (moment or think).
type ExportItem struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	ItemType ItemType
	ItemDate time.Time

	// Moment-specific fields (nil for thinks)
	Situation        *content.Content
	Thoughts         *content.Content
	PhysicalSymptoms *content.Content
	Behavior         *content.Content
	Consequences     *content.Content
	ValuesReflection *content.Content
	Intensity        *intensity.Intensity

	// Think-specific fields (empty/nil for moments)
	Category string
	Content  *content.Content

	DateCreated time.Time
}
