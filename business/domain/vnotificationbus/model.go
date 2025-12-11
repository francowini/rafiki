package vnotificationbus

import "github.com/google/uuid"

// ValueWithVision represents a value and its life vision for notifications.
type ValueWithVision struct {
	UserID            uuid.UUID
	ValueID           uuid.UUID
	ValueContent      string
	ValueFacet        string
	ValueOrder        int
	LifeVisionID      *uuid.UUID
	LifeVisionContent *string
}
