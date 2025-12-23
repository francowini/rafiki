// Package objectiverecordbus provides business logic for objective records management.
package objectiverecordbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/recordstatus"
)

// Domain errors
var (
	ErrNotFound               = errors.New("objective record not found")
	ErrGracePeriodExceeded    = errors.New("can only log records for today or yesterday")
	ErrNotFrequencyType       = errors.New("records only allowed for frequency tracking type")
	ErrFutureDateNotAllowed   = errors.New("cannot log record for future date")
	ErrMissingUserID          = errors.New("userID is required for querying objective records")
	ErrNotObjectiveOwner      = errors.New("user does not own the specified objective")
	ErrFrequencyCountRequired = errors.New("frequency_count required for n_per_week and n_per_month types")
)

// ObjectiveRecord represents a single completion record for frequency tracking.
type ObjectiveRecord struct {
	ID          uuid.UUID
	ObjectiveID uuid.UUID
	UserID      uuid.UUID
	RecordDate  time.Time // Date of the record (can be today or yesterday only)
	Status      recordstatus.RecordStatus
	Notes       *notes.Notes
	DateCreated time.Time
}

// NewObjectiveRecord contains information needed to create a record.
type NewObjectiveRecord struct {
	ObjectiveID uuid.UUID
	UserID      uuid.UUID
	RecordDate  time.Time // Must be today or yesterday only
	Status      recordstatus.RecordStatus
	Notes       *notes.Notes
}

// UpdateObjectiveRecord contains information to update a record.
// Nil pointer fields mean "no update".
// To clear Notes, pass a pointer to an empty notes.Notes value.
type UpdateObjectiveRecord struct {
	Status     *recordstatus.RecordStatus
	Notes      *notes.Notes
	ClearNotes bool // When true, sets Notes to NULL (takes precedence over Notes field)
}
