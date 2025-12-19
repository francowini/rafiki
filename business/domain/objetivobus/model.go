// Package objetivobus provides business logic for objectives management.
package objetivobus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/cumplimientopct"
	"github.com/francowini/rafiki/business/types/frecuenciatype"
	"github.com/francowini/rafiki/business/types/objetivostatus"
	"github.com/francowini/rafiki/business/types/objetivotitle"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// Domain errors
var (
	ErrNotFound                    = errors.New("objetivo not found")
	ErrMissingUserID               = errors.New("userID is required for querying objetivos")
	ErrNotLifeVisionOwner          = errors.New("user does not own the specified life vision")
	ErrTargetLifeVisionNotActive   = errors.New("target life vision must be active")
	ErrStatusTransitionNotAllowed  = errors.New("status transition not allowed (completado/abandonado are terminal)")
	ErrTrackingTypeImmutable       = errors.New("tracking type cannot be changed after creation")
	ErrInvalidFrecuenciaConfig     = errors.New("frecuencia type requires frecuencia_n >= 1")
	ErrMissingCumplimientoPct      = errors.New("cumplimiento_target_pct required for frecuencia tracking")
	ErrInvalidResultadoConfig      = errors.New("metrica_objetivo must be > 0 for resultado tracking")
	ErrProgressExceedsTarget       = errors.New("metrica_actual cannot exceed metrica_objetivo")
	ErrOnlyResultadoAllowsProgress = errors.New("progress increment only allowed for resultado tracking type")
	ErrAlreadyArchived             = errors.New("objetivo is already archived")
	ErrNotArchived                 = errors.New("objetivo is not archived")
	ErrTerminalStatusNoArchive     = errors.New("cannot archive objetivo with terminal status (completado/abandonado)")
)

// Objetivo represents a measurable objective linked to a life vision.
type Objetivo struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	LifeVisionID uuid.UUID
	Titulo       objetivotitle.ObjetivoTitle
	TipoTracking trackingtype.TrackingType // IMMUTABLE after creation
	Status       objetivostatus.ObjetivoStatus

	// Resultado tracking fields (NULL if frecuencia)
	MetricaObjetivo *int // Target metric (e.g., 35 books)
	MetricaActual   *int // Current progress (auto-calculated from tasks in future)

	// Frecuencia tracking fields (NULL if resultado)
	FrecuenciaTipo        *frecuenciatype.FrecuenciaType   // daily, n_por_semana, n_por_mes
	FrecuenciaN           *int                             // e.g., 5 for "5 days per week"
	CumplimientoTargetPct *cumplimientopct.CumplimientoPct // REQUIRED for frecuencia

	ArchivedAt  *time.Time
	DateCreated time.Time
	DateUpdated time.Time
}

// NewObjetivo contains information needed to create a new objetivo.
type NewObjetivo struct {
	UserID       uuid.UUID
	LifeVisionID uuid.UUID
	Titulo       objetivotitle.ObjetivoTitle
	TipoTracking trackingtype.TrackingType

	// Resultado fields (required if tipo_tracking = resultado)
	MetricaObjetivo *int

	// Frecuencia fields (required if tipo_tracking = frecuencia)
	FrecuenciaTipo        *frecuenciatype.FrecuenciaType
	FrecuenciaN           *int
	CumplimientoTargetPct *cumplimientopct.CumplimientoPct
}

// UpdateObjetivo contains information needed to update an objetivo.
type UpdateObjetivo struct {
	Titulo *objetivotitle.ObjetivoTitle
	Status *objetivostatus.ObjetivoStatus

	// Resultado fields
	MetricaObjetivo *int

	// Frecuencia fields
	FrecuenciaTipo        *frecuenciatype.FrecuenciaType
	FrecuenciaN           *int
	CumplimientoTargetPct *cumplimientopct.CumplimientoPct
}

// IncrementProgressRequest for MVP manual progress tracking (legacy - use UpdateProgressRequest).
type IncrementProgressRequest struct {
	Increment int // Can be negative for decrement
}

// UpdateProgressRequest for setting progress by increment or direct value.
type UpdateProgressRequest struct {
	Increment *int // Increment/decrement by this value (optional)
	Value     *int // Set to this exact value (optional)
}

// ChangeStatusRequest for changing objetivo status.
type ChangeStatusRequest struct {
	Status objetivostatus.ObjetivoStatus
}

type (
	// ObjetivoStatus is an alias for external usage.
	ObjetivoStatus = objetivostatus.ObjetivoStatus
	// TrackingType is an alias for external usage.
	TrackingType = trackingtype.TrackingType
)

// ParseStatus parses a string into an ObjetivoStatus.
func ParseStatus(s string) (objetivostatus.ObjetivoStatus, error) {
	return objetivostatus.Parse(s)
}

// ParseTrackingType parses a string into a TrackingType.
func ParseTrackingType(s string) (trackingtype.TrackingType, error) {
	return trackingtype.Parse(s)
}
