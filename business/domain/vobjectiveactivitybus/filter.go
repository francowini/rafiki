package vobjectiveactivitybus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/year"
)

// QueryFilter contains the criteria for querying activity data.
type QueryFilter struct {
	ObjectiveID uuid.UUID
	UserID      uuid.UUID
	Year        year.Year
}
