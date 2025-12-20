package objectiverecordbus

import (
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/recordstatus"
)

// QueryFilter defines filter criteria for querying objective records.
type QueryFilter struct {
	ID          *uuid.UUID
	ObjectiveID *uuid.UUID
	UserID      *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	Status      *recordstatus.RecordStatus
}
