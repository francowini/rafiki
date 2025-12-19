// Package objetivoregistroapp provides HTTP handlers for objetivo records.
package objetivoregistroapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/objetivoregistrobus"
	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/registrostatus"
)

// ObjetivoRecord represents an objetivo record for API responses.
type ObjetivoRecord struct {
	ID            string `json:"id"`
	ObjetivoID    string `json:"objetivoId"`
	FechaRegistro string `json:"fechaRegistro"`
	Status        string `json:"status"`
	Notes         string `json:"notes,omitempty"`
	DateCreated   string `json:"dateCreated"`
}

// Encode implements the encoder interface.
func (or ObjetivoRecord) Encode() ([]byte, string, error) {
	data, err := json.Marshal(or)
	return data, "application/json", err
}

// NewObjetivoRecord represents data for creating a new objetivo record.
type NewObjetivoRecord struct {
	FechaRegistro string  `json:"fechaRegistro" validate:"required"`
	Status        string  `json:"status" validate:"required,oneof=completado omitido_intencional omitido"`
	Notes         *string `json:"notes"`
}

// Decode implements the decoder interface.
func (nor *NewObjetivoRecord) Decode(data []byte) error {
	return json.Unmarshal(data, nor)
}

// Validate checks the data for correctness.
func (nor NewObjetivoRecord) Validate() error {
	if err := errs.Check(nor); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateObjetivoRecord represents data for updating an objetivo record.
type UpdateObjetivoRecord struct {
	Status *string  `json:"status" validate:"omitempty,oneof=completado omitido_intencional omitido"`
	Notes  **string `json:"notes"`
}

// Decode implements the decoder interface.
func (uor *UpdateObjetivoRecord) Decode(data []byte) error {
	return json.Unmarshal(data, uor)
}

// Validate checks the data for correctness.
func (uor UpdateObjetivoRecord) Validate() error {
	if err := errs.Check(uor); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ===== Business → App conversions =====

func toAppRecord(r objetivoregistrobus.ObjetivoRecord) ObjetivoRecord {
	notesStr := ""
	if r.Notes != nil {
		notesStr = r.Notes.String()
	}

	return ObjetivoRecord{
		ID:            r.ID.String(),
		ObjetivoID:    r.ObjetivoID.String(),
		FechaRegistro: r.FechaRegistro.Format("2006-01-02"),
		Status:        r.Status.String(),
		Notes:         notesStr,
		DateCreated:   r.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppRecords(records []objetivoregistrobus.ObjetivoRecord) []ObjetivoRecord {
	app := make([]ObjetivoRecord, len(records))
	for i, r := range records {
		app[i] = toAppRecord(r)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewRecord(nor NewObjetivoRecord, objetivoID, userID uuid.UUID) (objetivoregistrobus.NewObjetivoRecord, error) {
	var errors errs.FieldErrors

	// Parse fecha_registro (expect YYYY-MM-DD format)
	fechaRegistro, err := time.Parse("2006-01-02", nor.FechaRegistro)
	if err != nil {
		errors.Add("fechaRegistro", err)
	}

	// Parse status
	status, err := registrostatus.Parse(nor.Status)
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
		return objetivoregistrobus.NewObjetivoRecord{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return objetivoregistrobus.NewObjetivoRecord{
		ObjetivoID:    objetivoID,
		UserID:        userID,
		FechaRegistro: fechaRegistro,
		Status:        status,
		Notes:         notesPtr,
	}, nil
}
