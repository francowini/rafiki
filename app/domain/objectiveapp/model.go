// Package objectiveapp provides HTTP handlers for objectives.
package objectiveapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/types/compliancepct"
	"github.com/francowini/rafiki/business/types/frequencycount"
	"github.com/francowini/rafiki/business/types/frequencytype"
	"github.com/francowini/rafiki/business/types/objectivestatus"
	"github.com/francowini/rafiki/business/types/objectivetitle"
	"github.com/francowini/rafiki/business/types/targetmetric"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// Objective represents an objective for API responses.
type Objective struct {
	ID                  string  `json:"id"`
	LifeVisionID        string  `json:"lifeVisionId"`
	Title               string  `json:"title"`
	TrackingType        string  `json:"trackingType"`
	Status              string  `json:"status"`
	TargetMetric        *int    `json:"targetMetric,omitempty"`
	CurrentMetric       *int    `json:"currentMetric,omitempty"`
	FrequencyType       *string `json:"frequencyType,omitempty"`
	FrequencyCount      *int    `json:"frequencyCount,omitempty"`
	ComplianceTargetPct *int    `json:"complianceTargetPct,omitempty"`
	ArchivedAt          *string `json:"archivedAt,omitempty"`
	DateCreated         string  `json:"dateCreated"`
	DateUpdated         string  `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (o Objective) Encode() ([]byte, string, error) {
	data, err := json.Marshal(o)
	return data, "application/json", err
}

// NewObjective represents data for creating a new objective.
type NewObjective struct {
	LifeVisionID        string  `json:"lifeVisionId" validate:"required,uuid"`
	Title               string  `json:"title" validate:"required"`
	TrackingType        string  `json:"trackingType" validate:"required,oneof=result frequency"`
	TargetMetric        *int    `json:"targetMetric"`
	FrequencyType       *string `json:"frequencyType" validate:"omitempty,oneof=daily n_per_week n_per_month"`
	FrequencyCount      *int    `json:"frequencyCount"`
	ComplianceTargetPct *int    `json:"complianceTargetPct" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (no *NewObjective) Decode(data []byte) error {
	return json.Unmarshal(data, no)
}

// Validate checks the data for correctness.
func (no NewObjective) Validate() error {
	if err := errs.Check(no); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateObjective represents data for updating an objective.
type UpdateObjective struct {
	Title               *string `json:"title"`
	TargetMetric        *int    `json:"targetMetric"`
	FrequencyType       *string `json:"frequencyType" validate:"omitempty,oneof=daily n_per_week n_per_month"`
	FrequencyCount      *int    `json:"frequencyCount"`
	ComplianceTargetPct *int    `json:"complianceTargetPct" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (uo *UpdateObjective) Decode(data []byte) error {
	return json.Unmarshal(data, uo)
}

// Validate checks the data for correctness.
func (uo UpdateObjective) Validate() error {
	if err := errs.Check(uo); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ChangeStatusRequest represents data for changing objective status.
type ChangeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active completed abandoned paused"`
}

// Decode implements the decoder interface.
func (csr *ChangeStatusRequest) Decode(data []byte) error {
	return json.Unmarshal(data, csr)
}

// Validate checks the data for correctness.
func (csr ChangeStatusRequest) Validate() error {
	if err := errs.Check(csr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateProgressRequest represents data for updating progress (increment or set value).
type UpdateProgressRequest struct {
	Increment *int `json:"increment"` // Increment/decrement by this value
	Value     *int `json:"value"`     // Set to this exact value
}

// Decode implements the decoder interface.
func (upr *UpdateProgressRequest) Decode(data []byte) error {
	return json.Unmarshal(data, upr)
}

// Validate checks the data for correctness.
func (upr UpdateProgressRequest) Validate() error {
	if upr.Increment == nil && upr.Value == nil {
		return fmt.Errorf("validate: either increment or value must be provided")
	}
	if upr.Increment != nil && upr.Value != nil {
		return fmt.Errorf("validate: cannot provide both increment and value")
	}
	if upr.Value != nil && *upr.Value < 0 {
		return fmt.Errorf("validate: value must be non-negative")
	}
	return nil
}

// ===== Business → App conversions =====

func toAppObjective(o objectivebus.Objective) Objective {
	var archivedAt *string
	if o.ArchivedAt != nil {
		t := o.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		archivedAt = &t
	}

	var frequencyType *string
	if o.FrequencyType != nil {
		ft := o.FrequencyType.String()
		frequencyType = &ft
	}

	var complianceTargetPct *int
	if o.ComplianceTargetPct != nil {
		cp := o.ComplianceTargetPct.Value()
		complianceTargetPct = &cp
	}

	var targetMetric *int
	if o.TargetMetric != nil {
		v := o.TargetMetric.Value()
		targetMetric = &v
	}

	var currentMetric *int
	if o.CurrentMetric != nil {
		v := o.CurrentMetric.Value()
		currentMetric = &v
	}

	var frequencyCount *int
	if o.FrequencyCount != nil {
		v := o.FrequencyCount.Value()
		frequencyCount = &v
	}

	return Objective{
		ID:                  o.ID.String(),
		LifeVisionID:        o.LifeVisionID.String(),
		Title:               o.Title.String(),
		TrackingType:        o.TrackingType.String(),
		Status:              o.Status.String(),
		TargetMetric:        targetMetric,
		CurrentMetric:       currentMetric,
		FrequencyType:       frequencyType,
		FrequencyCount:      frequencyCount,
		ComplianceTargetPct: complianceTargetPct,
		ArchivedAt:          archivedAt,
		DateCreated:         o.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated:         o.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppObjectives(objectives []objectivebus.Objective) []Objective {
	app := make([]Objective, len(objectives))
	for i, o := range objectives {
		app[i] = toAppObjective(o)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewObjective(no NewObjective, userID uuid.UUID) (objectivebus.NewObjective, error) {
	var errors errs.FieldErrors

	// Parse life vision ID
	lifeVisionID, err := uuid.Parse(no.LifeVisionID)
	if err != nil {
		errors.Add("lifeVisionId", err)
	}

	// Parse title
	title, err := objectivetitle.Parse(no.Title)
	if err != nil {
		errors.Add("title", err)
	}

	// Parse tracking_type
	tt, err := trackingtype.Parse(no.TrackingType)
	if err != nil {
		errors.Add("trackingType", err)
	}

	// Parse frequency_type if provided
	var freqType *frequencytype.FrequencyType
	if no.FrequencyType != nil {
		ft, err := frequencytype.Parse(*no.FrequencyType)
		if err != nil {
			errors.Add("frequencyType", err)
		} else {
			freqType = &ft
		}
	}

	// Parse compliance_target_pct if provided
	var complianceTargetPct *compliancepct.CompliancePct
	if no.ComplianceTargetPct != nil {
		cp, err := compliancepct.Parse(*no.ComplianceTargetPct)
		if err != nil {
			errors.Add("complianceTargetPct", err)
		} else {
			complianceTargetPct = &cp
		}
	}

	// Parse target_metric if provided
	var targetMetricPtr *targetmetric.TargetMetric
	if no.TargetMetric != nil {
		tm, err := targetmetric.Parse(*no.TargetMetric)
		if err != nil {
			errors.Add("targetMetric", err)
		} else {
			targetMetricPtr = &tm
		}
	}

	// Parse frequency_count if provided
	var frequencyCountPtr *frequencycount.FrequencyCount
	if no.FrequencyCount != nil {
		fc, err := frequencycount.Parse(*no.FrequencyCount)
		if err != nil {
			errors.Add("frequencyCount", err)
		} else {
			frequencyCountPtr = &fc
		}
	}

	if len(errors) > 0 {
		return objectivebus.NewObjective{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return objectivebus.NewObjective{
		UserID:              userID,
		LifeVisionID:        lifeVisionID,
		Title:               title,
		TrackingType:        tt,
		TargetMetric:        targetMetricPtr,
		FrequencyType:       freqType,
		FrequencyCount:      frequencyCountPtr,
		ComplianceTargetPct: complianceTargetPct,
	}, nil
}

func toBusUpdateObjective(uo UpdateObjective) (objectivebus.UpdateObjective, error) {
	var errors errs.FieldErrors
	var bus objectivebus.UpdateObjective

	if uo.Title != nil {
		title, err := objectivetitle.Parse(*uo.Title)
		if err != nil {
			errors.Add("title", err)
		} else {
			bus.Title = &title
		}
	}

	if uo.TargetMetric != nil {
		tm, err := targetmetric.Parse(*uo.TargetMetric)
		if err != nil {
			errors.Add("targetMetric", err)
		} else {
			bus.TargetMetric = &tm
		}
	}

	if uo.FrequencyType != nil {
		ft, err := frequencytype.Parse(*uo.FrequencyType)
		if err != nil {
			errors.Add("frequencyType", err)
		} else {
			bus.FrequencyType = &ft
		}
	}

	if uo.FrequencyCount != nil {
		fc, err := frequencycount.Parse(*uo.FrequencyCount)
		if err != nil {
			errors.Add("frequencyCount", err)
		} else {
			bus.FrequencyCount = &fc
		}
	}

	if uo.ComplianceTargetPct != nil {
		cp, err := compliancepct.Parse(*uo.ComplianceTargetPct)
		if err != nil {
			errors.Add("complianceTargetPct", err)
		} else {
			bus.ComplianceTargetPct = &cp
		}
	}

	if len(errors) > 0 {
		return objectivebus.UpdateObjective{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}

func toBusChangeStatusRequest(csr ChangeStatusRequest) (objectivebus.ChangeStatusRequest, error) {
	status, err := objectivestatus.Parse(csr.Status)
	if err != nil {
		return objectivebus.ChangeStatusRequest{}, fmt.Errorf("validate: status: %w", err)
	}

	return objectivebus.ChangeStatusRequest{
		Status: status,
	}, nil
}

func toBusUpdateProgressRequest(upr UpdateProgressRequest) objectivebus.UpdateProgressRequest {
	return objectivebus.UpdateProgressRequest{
		Increment: upr.Increment,
		Value:     upr.Value,
	}
}
