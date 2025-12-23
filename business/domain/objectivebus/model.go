// Package objectivebus provides business logic for objectives management.
package objectivebus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/beginmetric"
	"github.com/francowini/rafiki/business/types/compliancepct"
	"github.com/francowini/rafiki/business/types/currentmetric"
	"github.com/francowini/rafiki/business/types/frequencycount"
	"github.com/francowini/rafiki/business/types/frequencytype"
	"github.com/francowini/rafiki/business/types/objectivestatus"
	"github.com/francowini/rafiki/business/types/objectivetitle"
	"github.com/francowini/rafiki/business/types/targetmetric"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// Domain errors
var (
	ErrNotFound                   = errors.New("objective not found")
	ErrMissingUserID              = errors.New("userID is required for querying objectives")
	ErrNotLifeVisionOwner         = errors.New("user does not own the specified life vision")
	ErrTargetLifeVisionNotActive  = errors.New("target life vision must be active")
	ErrStatusTransitionNotAllowed = errors.New("status transition not allowed (completed/abandoned are terminal)")
	ErrTrackingTypeImmutable      = errors.New("tracking type cannot be changed after creation")
	ErrInvalidFrequencyConfig     = errors.New("frequency type requires frequency_count >= 1")
	ErrMissingCompliancePct       = errors.New("compliance_target_pct required for frequency tracking")
	ErrInvalidResultConfig        = errors.New("target_metric must be > 0 for result tracking")
	ErrProgressExceedsTarget      = errors.New("current_metric cannot exceed target_metric")
	ErrOnlyResultAllowsProgress   = errors.New("progress increment only allowed for result tracking type")
	ErrAlreadyArchived            = errors.New("objective is already archived")
	ErrNotArchived                = errors.New("objective is not archived")
	ErrTerminalStatusNoArchive    = errors.New("cannot archive objective with terminal status (completed/abandoned)")
	ErrStatusChangeMustUseMethod  = errors.New("status changes must use ChangeStatus method")
	ErrMetricNotInitialized       = errors.New("metric values not initialized")
	ErrIncrementOrValueRequired   = errors.New("either increment or value must be provided")
	ErrFrequencyTypeRequired      = errors.New("frequency_type required for frequency tracking")
	ErrMetricOnlyForResult        = errors.New("target_metric only valid for result tracking")
	ErrFrequencyOnlyForFrequency  = errors.New("frequency fields only valid for frequency tracking")
)

// Objective represents a measurable objective linked to a life vision.
type Objective struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	LifeVisionID uuid.UUID
	Title        objectivetitle.ObjectiveTitle
	TrackingType trackingtype.TrackingType // IMMUTABLE after creation
	Status       objectivestatus.ObjectiveStatus

	// Result tracking fields (NULL if frequency)
	TargetMetric  *targetmetric.TargetMetric   // Target metric (e.g., 35 books)
	CurrentMetric *currentmetric.CurrentMetric // Current progress (auto-calculated from tasks in future)
	BeginMetric   *beginmetric.BeginMetric     // Optional starting metric for progress calculation

	// Frequency tracking fields (NULL if result)
	FrequencyType       *frequencytype.FrequencyType   // daily, n_per_week, n_per_month
	FrequencyCount      *frequencycount.FrequencyCount // e.g., 5 for "5 days per week"
	ComplianceTargetPct *compliancepct.CompliancePct   // REQUIRED for frequency

	ArchivedAt  *time.Time
	DateCreated time.Time
	DateUpdated time.Time
}

// NewObjective contains information needed to create a new objective.
type NewObjective struct {
	UserID       uuid.UUID
	LifeVisionID uuid.UUID
	Title        objectivetitle.ObjectiveTitle
	TrackingType trackingtype.TrackingType

	// Result fields (required if tracking_type = result)
	TargetMetric *targetmetric.TargetMetric
	BeginMetric  *beginmetric.BeginMetric // Optional starting metric for progress calculation

	// Frequency fields (required if tracking_type = frequency)
	FrequencyType       *frequencytype.FrequencyType
	FrequencyCount      *frequencycount.FrequencyCount
	ComplianceTargetPct *compliancepct.CompliancePct
}

// UpdateObjective contains information needed to update an objective.
type UpdateObjective struct {
	Title  *objectivetitle.ObjectiveTitle
	Status *objectivestatus.ObjectiveStatus

	// Result fields
	TargetMetric     *targetmetric.TargetMetric
	BeginMetric      *beginmetric.BeginMetric
	ClearBeginMetric bool // Use ClearNotes pattern for nullable fields

	// Frequency fields
	FrequencyType       *frequencytype.FrequencyType
	FrequencyCount      *frequencycount.FrequencyCount
	ComplianceTargetPct *compliancepct.CompliancePct
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

// ChangeStatusRequest for changing objective status.
type ChangeStatusRequest struct {
	Status objectivestatus.ObjectiveStatus
}

type (
	// ObjectiveStatus is an alias for external usage.
	ObjectiveStatus = objectivestatus.ObjectiveStatus
	// TrackingType is an alias for external usage.
	TrackingType = trackingtype.TrackingType
)

// ParseStatus parses a string into an ObjectiveStatus.
func ParseStatus(s string) (objectivestatus.ObjectiveStatus, error) {
	return objectivestatus.Parse(s)
}

// ParseTrackingType parses a string into a TrackingType.
func ParseTrackingType(s string) (trackingtype.TrackingType, error) {
	return trackingtype.Parse(s)
}
