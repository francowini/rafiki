package objetivobus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/objetivostatus"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// QueryFilter defines filter criteria for querying objetivos.
type QueryFilter struct {
	ID              *uuid.UUID
	UserID          *uuid.UUID
	LifeVisionID    *uuid.UUID
	Status          *objetivostatus.ObjetivoStatus
	TipoTracking    *trackingtype.TrackingType
	IncludeArchived bool
}
