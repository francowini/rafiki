// Package lifevisionapp provides HTTP handlers for life visions.
package lifevisionapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/types/lifevisioncontent"
)

// LifeVision represents a life vision for API responses.
type LifeVision struct {
	ID          string `json:"id"`
	ValueID     string `json:"valueId"`
	Content     string `json:"content"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (lv LifeVision) Encode() ([]byte, string, error) {
	data, err := json.Marshal(lv)
	return data, "application/json", err
}

// NewLifeVision represents data for creating a new life vision.
type NewLifeVision struct {
	ValueID string `json:"valueId" validate:"required,uuid"`
	Content string `json:"content" validate:"required"`
}

// Decode implements the decoder interface.
func (nlv *NewLifeVision) Decode(data []byte) error {
	return json.Unmarshal(data, nlv)
}

// Validate checks the data for correctness.
func (nlv NewLifeVision) Validate() error {
	if err := errs.Check(nlv); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateLifeVision represents data for updating a life vision.
type UpdateLifeVision struct {
	ValueID *string `json:"valueId" validate:"omitempty,uuid"`
	Content *string `json:"content"`
}

// Decode implements the decoder interface.
func (ulv *UpdateLifeVision) Decode(data []byte) error {
	return json.Unmarshal(data, ulv)
}

// Validate checks the data for correctness.
func (ulv UpdateLifeVision) Validate() error {
	if err := errs.Check(ulv); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ===== Business → App conversions =====

func toAppLifeVision(lv lifevisionbus.LifeVision) LifeVision {
	return LifeVision{
		ID:          lv.ID.String(),
		ValueID:     lv.ValueID.String(),
		Content:     lv.Content.String(),
		DateCreated: lv.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated: lv.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppLifeVisions(lvs []lifevisionbus.LifeVision) []LifeVision {
	app := make([]LifeVision, len(lvs))
	for i, lv := range lvs {
		app[i] = toAppLifeVision(lv)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewLifeVision(nlv NewLifeVision, userID uuid.UUID) (lifevisionbus.NewLifeVision, error) {
	var errors errs.FieldErrors

	// Parse value ID
	valueID, err := uuid.Parse(nlv.ValueID)
	if err != nil {
		errors.Add("valueId", err)
	}

	// Parse content
	content, err := lifevisioncontent.Parse(nlv.Content)
	if err != nil {
		errors.Add("content", err)
	}

	if len(errors) > 0 {
		return lifevisionbus.NewLifeVision{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return lifevisionbus.NewLifeVision{
		UserID:  userID,
		ValueID: valueID,
		Content: content,
	}, nil
}

func toBusUpdateLifeVision(ulv UpdateLifeVision) (lifevisionbus.UpdateLifeVision, error) {
	var errors errs.FieldErrors
	var bus lifevisionbus.UpdateLifeVision

	if ulv.ValueID != nil {
		valueID, err := uuid.Parse(*ulv.ValueID)
		if err != nil {
			errors.Add("valueId", err)
		} else {
			bus.ValueID = &valueID
		}
	}

	if ulv.Content != nil {
		content, err := lifevisioncontent.Parse(*ulv.Content)
		if err != nil {
			errors.Add("content", err)
		} else {
			bus.Content = &content
		}
	}

	if len(errors) > 0 {
		return lifevisionbus.UpdateLifeVision{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}
