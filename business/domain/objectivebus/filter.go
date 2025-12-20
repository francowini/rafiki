package objectivebus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/objectivestatus"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// QueryFilter defines filter criteria for querying objectives.
type QueryFilter struct {
	ID              *uuid.UUID
	UserID          *uuid.UUID
	LifeVisionID    *uuid.UUID
	Status          *objectivestatus.ObjectiveStatus
	TrackingType    *trackingtype.TrackingType
	IncludeArchived bool
}
