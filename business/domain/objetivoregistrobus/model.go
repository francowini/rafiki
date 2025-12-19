// Package objetivoregistrobus provides business logic for objective records management.
package objetivoregistrobus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/registrostatus"
)

// Domain errors
var (
	ErrNotFound             = errors.New("objetivo record not found")
	ErrGracePeriodExceeded  = errors.New("can only log records for today or yesterday")
	ErrNotFrecuenciaType    = errors.New("records only allowed for frecuencia tracking type")
	ErrFutureDateNotAllowed = errors.New("cannot log record for future date")
	ErrMissingUserID        = errors.New("userID is required for querying objetivo records")
	ErrNotObjetivoOwner     = errors.New("user does not own the specified objetivo")
)

// ObjetivoRecord represents a single completion record for frequency tracking.
type ObjetivoRecord struct {
	ID            uuid.UUID
	ObjetivoID    uuid.UUID
	UserID        uuid.UUID
	FechaRegistro time.Time // Date of the record (can be today or yesterday only)
	Status        registrostatus.RegistroStatus
	Notes         *string
	DateCreated   time.Time
}

// NewObjetivoRecord contains information needed to create a record.
type NewObjetivoRecord struct {
	ObjetivoID    uuid.UUID
	UserID        uuid.UUID
	FechaRegistro time.Time // Must be today or yesterday only
	Status        registrostatus.RegistroStatus
	Notes         *string
}

// UpdateObjetivoRecord contains information to update a record.
type UpdateObjetivoRecord struct {
	Status *registrostatus.RegistroStatus
	Notes  **string // Double pointer to distinguish between "not updating" and "setting to null"
}
