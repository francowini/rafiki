package objetivoregistrobus

import (
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/registrostatus"
)

// QueryFilter defines filter criteria for querying objetivo records.
type QueryFilter struct {
	ID         *uuid.UUID
	ObjetivoID *uuid.UUID
	UserID     *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Status     *registrostatus.RegistroStatus
}
