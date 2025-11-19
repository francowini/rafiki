package momentbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound    = errors.New("moment not found")
	ErrFutureDate  = errors.New("moment_date cannot be in the future")
	ErrUniqueEntry = errors.New("moment entry already exists")
)

// Moment represents a tracked emotional/difficult moment in the system.
type Moment struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	MomentDate       time.Time
	Situation        content.Content
	Thoughts         content.Content
	PhysicalSymptoms content.Content
	Behavior         content.Content
	Consequences     content.Content
	ValuesReflection content.Content
	Intensity        intensity.Intensity
	DateCreated      time.Time
	DateUpdated      time.Time
}

// NewMoment contains information needed to create a new moment.
type NewMoment struct {
	UserID           uuid.UUID
	MomentDate       time.Time
	Situation        content.Content
	Thoughts         content.Content
	PhysicalSymptoms content.Content
	Behavior         content.Content
	Consequences     content.Content
	ValuesReflection content.Content
	Intensity        intensity.Intensity
}

// UpdateMoment contains information needed to update a moment.
type UpdateMoment struct {
	MomentDate       *time.Time
	Situation        *content.Content
	Thoughts         *content.Content
	PhysicalSymptoms *content.Content
	Behavior         *content.Content
	Consequences     *content.Content
	ValuesReflection *content.Content
	Intensity        *intensity.Intensity
}
