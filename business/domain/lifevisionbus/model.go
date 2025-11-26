// Package lifevisionbus provides business logic for life visions management.
package lifevisionbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/lifevisioncontent"
)

// Domain errors
var (
	ErrNotFound      = errors.New("life vision not found")
	ErrMissingUserID = errors.New("userID is required for querying life visions")
)

// LifeVision represents an aspirational state of being.
type LifeVision struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ValueID     uuid.UUID
	Content     lifevisioncontent.LifeVisionContent
	DateCreated time.Time
	DateUpdated time.Time
}

// NewLifeVision contains information needed to create a new life vision.
type NewLifeVision struct {
	ValueID uuid.UUID
	Content lifevisioncontent.LifeVisionContent
}

// UpdateLifeVision contains information needed to update a life vision.
type UpdateLifeVision struct {
	Content *lifevisioncontent.LifeVisionContent
	ValueID *uuid.UUID
}
