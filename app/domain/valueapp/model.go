// Package valueapp provides HTTP handlers for values.
package valueapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/types/displayorder"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/business/types/valuecontent"
)

// Value represents a value for API responses.
type Value struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	Facet        string `json:"facet"`
	DisplayOrder int    `json:"displayOrder"`
	DateCreated  string `json:"dateCreated"`
	DateUpdated  string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (v Value) Encode() ([]byte, string, error) {
	data, err := json.Marshal(v)
	return data, "application/json", err
}

// NewValue represents data for creating a new value.
type NewValue struct {
	Content      string `json:"content" validate:"required"`
	Facet        string `json:"facet" validate:"required"`
	DisplayOrder int    `json:"displayOrder" validate:"required,min=1,max=10"`
}

// Decode implements the decoder interface.
func (nv *NewValue) Decode(data []byte) error {
	return json.Unmarshal(data, nv)
}

// Validate checks the data for correctness.
func (nv NewValue) Validate() error {
	if err := errs.Check(nv); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateValue represents data for updating a value.
type UpdateValue struct {
	Content      *string `json:"content"`
	Facet        *string `json:"facet"`
	DisplayOrder *int    `json:"displayOrder" validate:"omitempty,min=1,max=10"`
}

// Decode implements the decoder interface.
func (uv *UpdateValue) Decode(data []byte) error {
	return json.Unmarshal(data, uv)
}

// Validate checks the data for correctness.
func (uv UpdateValue) Validate() error {
	if err := errs.Check(uv); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ===== Business → App conversions =====

func toAppValue(value valuebus.Value) Value {
	return Value{
		ID:           value.ID.String(),
		Content:      value.Content.String(),
		Facet:        value.Facet.String(),
		DisplayOrder: value.DisplayOrder.Value(),
		DateCreated:  value.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated:  value.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppValues(values []valuebus.Value) []Value {
	app := make([]Value, len(values))
	for i, value := range values {
		app[i] = toAppValue(value)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewValue(ctx context.Context, nv NewValue) (valuebus.NewValue, error) {
	var errors errs.FieldErrors

	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		errors.Add("userID", err)
	}

	// Parse content
	content, err := valuecontent.Parse(nv.Content)
	if err != nil {
		errors.Add("content", err)
	}

	// Parse facet
	facetVal, err := facet.Parse(nv.Facet)
	if err != nil {
		errors.Add("facet", err)
	}

	// Parse display order
	displayOrderVal, err := displayorder.Parse(nv.DisplayOrder)
	if err != nil {
		errors.Add("displayOrder", err)
	}

	if len(errors) > 0 {
		return valuebus.NewValue{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return valuebus.NewValue{
		UserID:       userID,
		Content:      content,
		Facet:        facetVal,
		DisplayOrder: displayOrderVal,
	}, nil
}

func toBusUpdateValue(uv UpdateValue) (valuebus.UpdateValue, error) {
	var errors errs.FieldErrors
	var bus valuebus.UpdateValue

	if uv.Content != nil {
		content, err := valuecontent.Parse(*uv.Content)
		if err != nil {
			errors.Add("content", err)
		} else {
			bus.Content = &content
		}
	}

	if uv.Facet != nil {
		facetVal, err := facet.Parse(*uv.Facet)
		if err != nil {
			errors.Add("facet", err)
		} else {
			bus.Facet = &facetVal
		}
	}

	if uv.DisplayOrder != nil {
		displayOrderVal, err := displayorder.Parse(*uv.DisplayOrder)
		if err != nil {
			errors.Add("displayOrder", err)
		} else {
			bus.DisplayOrder = &displayOrderVal
		}
	}

	if len(errors) > 0 {
		return valuebus.UpdateValue{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}

// ===== Reorder types =====

// ReorderRequest represents a batch reorder request.
type ReorderRequest struct {
	Items []ReorderItem `json:"items" validate:"required,min=1,max=10,dive"`
}

// ReorderItem represents a single item to be reordered.
type ReorderItem struct {
	ID           string `json:"id" validate:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" validate:"required,min=1,max=10"`
}

// Decode implements the decoder interface.
func (rr *ReorderRequest) Decode(data []byte) error {
	return json.Unmarshal(data, rr)
}

// Validate checks the data for correctness.
func (rr ReorderRequest) Validate() error {
	if err := errs.Check(rr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// Check for duplicate displayOrders in request
	seen := make(map[int]bool)
	for _, item := range rr.Items {
		if seen[item.DisplayOrder] {
			return fmt.Errorf("duplicate displayOrder: %d", item.DisplayOrder)
		}
		seen[item.DisplayOrder] = true
	}

	return nil
}

// toBusReorderRequest converts app layer to business domain type.
// Collects all validation errors instead of failing on first error.
func toBusReorderRequest(appReorder ReorderRequest) (valuebus.ReorderRequest, error) {
	var fieldErrors errs.FieldErrors
	items := make([]valuebus.ReorderItem, 0, len(appReorder.Items))

	for i, item := range appReorder.Items {
		var valueID uuid.UUID
		var displayOrderVal displayorder.DisplayOrder
		itemValid := true

		id, err := uuid.Parse(item.ID)
		if err != nil {
			fieldErrors.Add(fmt.Sprintf("items[%d].id", i), err)
			itemValid = false
		} else {
			valueID = id
		}

		order, err := displayorder.Parse(item.DisplayOrder)
		if err != nil {
			fieldErrors.Add(fmt.Sprintf("items[%d].displayOrder", i), err)
			itemValid = false
		} else {
			displayOrderVal = order
		}

		if itemValid {
			items = append(items, valuebus.ReorderItem{
				ID:           valueID,
				DisplayOrder: displayOrderVal,
			})
		}
	}

	if len(fieldErrors) > 0 {
		return valuebus.ReorderRequest{}, fmt.Errorf("validate: %w", fieldErrors.ToError())
	}

	return valuebus.ReorderRequest{Items: items}, nil
}
