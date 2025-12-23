package vobjectiveactivitybus

import "github.com/google/uuid"

// QueryFilter contains the criteria for querying activity data.
type QueryFilter struct {
	ObjectiveID uuid.UUID
	UserID      uuid.UUID
	Year        int
}
