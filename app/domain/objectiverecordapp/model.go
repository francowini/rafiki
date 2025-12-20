// Package objectiverecordapp provides HTTP handlers for objective records.
package objectiverecordapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/objectiverecordbus"
	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/recordstatus"
)

// ObjectiveRecord represents an objective record for API responses.
type ObjectiveRecord struct {
	ID          string `json:"id"`
	ObjectiveID string `json:"objectiveId"`
	RecordDate  string `json:"recordDate"`
	Status      string `json:"status"`
	Notes       string `json:"notes,omitempty"`
	DateCreated string `json:"dateCreated"`
}

// Encode implements the encoder interface.
func (or ObjectiveRecord) Encode() ([]byte, string, error) {
	data, err := json.Marshal(or)
	return data, "application/json", err
}

// NewObjectiveRecord represents data for creating a new objective record.
type NewObjectiveRecord struct {
	RecordDate string  `json:"recordDate" validate:"required"`
	Status     string  `json:"status" validate:"required,oneof=completed intentionally_skipped skipped"`
	Notes      *string `json:"notes"`
}

// Decode implements the decoder interface.
func (nor *NewObjectiveRecord) Decode(data []byte) error {
	return json.Unmarshal(data, nor)
}

// Validate checks the data for correctness.
func (nor NewObjectiveRecord) Validate() error {
	if err := errs.Check(nor); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateObjectiveRecord represents data for updating an objective record.
type UpdateObjectiveRecord struct {
	Status *string  `json:"status" validate:"omitempty,oneof=completed intentionally_skipped skipped"`
	Notes  **string `json:"notes"`
}

// Decode implements the decoder interface.
func (uor *UpdateObjectiveRecord) Decode(data []byte) error {
	return json.Unmarshal(data, uor)
}

// Validate checks the data for correctness.
func (uor UpdateObjectiveRecord) Validate() error {
	if err := errs.Check(uor); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ===== Business → App conversions =====

func toAppRecord(r objectiverecordbus.ObjectiveRecord) ObjectiveRecord {
	notesStr := ""
	if r.Notes != nil {
		notesStr = r.Notes.String()
	}

	return ObjectiveRecord{
		ID:          r.ID.String(),
		ObjectiveID: r.ObjectiveID.String(),
		RecordDate:  r.RecordDate.Format("2006-01-02"),
		Status:      r.Status.String(),
		Notes:       notesStr,
		DateCreated: r.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppRecords(records []objectiverecordbus.ObjectiveRecord) []ObjectiveRecord {
	app := make([]ObjectiveRecord, len(records))
	for i, r := range records {
		app[i] = toAppRecord(r)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewRecord(nor NewObjectiveRecord, objectiveID, userID uuid.UUID) (objectiverecordbus.NewObjectiveRecord, error) {
	var errors errs.FieldErrors

	// Parse record_date (expect YYYY-MM-DD format)
	recordDate, err := time.Parse("2006-01-02", nor.RecordDate)
	if err != nil {
		errors.Add("recordDate", err)
	}

	// Parse status
	status, err := recordstatus.Parse(nor.Status)
	if err != nil {
		errors.Add("status", err)
	}

	// Parse notes (optional)
	var notesPtr *notes.Notes
	if nor.Notes != nil && *nor.Notes != "" {
		n, err := notes.Parse(*nor.Notes)
		if err != nil {
			errors.Add("notes", err)
		} else {
			notesPtr = &n
		}
	}

	if len(errors) > 0 {
		return objectiverecordbus.NewObjectiveRecord{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return objectiverecordbus.NewObjectiveRecord{
		ObjectiveID: objectiveID,
		UserID:      userID,
		RecordDate:  recordDate,
		Status:      status,
		Notes:       notesPtr,
	}, nil
}
