// Package objetivoregistrobus provides business logic for objective records management.
package objetivoregistrobus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/notes"
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
	ErrFrecuenciaNRequired  = errors.New("frecuencia_n required for n_por_semana and n_por_mes types")
)

// ObjetivoRecord represents a single completion record for frequency tracking.
type ObjetivoRecord struct {
	ID            uuid.UUID
	ObjetivoID    uuid.UUID
	UserID        uuid.UUID
	FechaRegistro time.Time // Date of the record (can be today or yesterday only)
	Status        registrostatus.RegistroStatus
	Notes         *notes.Notes
	DateCreated   time.Time
}

// NewObjetivoRecord contains information needed to create a record.
type NewObjetivoRecord struct {
	ObjetivoID    uuid.UUID
	UserID        uuid.UUID
	FechaRegistro time.Time // Must be today or yesterday only
	Status        registrostatus.RegistroStatus
	Notes         *notes.Notes
}

// UpdateObjetivoRecord contains information to update a record.
type UpdateObjetivoRecord struct {
	Status *registrostatus.RegistroStatus
	Notes  **notes.Notes // Double pointer to distinguish between "not updating" and "setting to null"
}
