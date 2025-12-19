// Package objetivoregistrodb provides database access for objetivo records.
package objetivoregistrodb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/objetivoregistrobus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/registrostatus"
)

// objetivoRecord represents the database model.
type objetivoRecord struct {
	ID            uuid.UUID `db:"objetivo_record_id"`
	ObjetivoID    uuid.UUID `db:"objetivo_id"`
	UserID        uuid.UUID `db:"user_id"`
	FechaRegistro time.Time `db:"fecha_registro"`
	Status        string    `db:"status"`
	Notes         *string   `db:"notes"` // encrypted
	DateCreated   time.Time `db:"date_created"`
}

// toDBRecordEncrypted converts business model to DB model with encryption.
func toDBRecordEncrypted(bus objetivoregistrobus.ObjetivoRecord, enc encrypt.Encryptor) (objetivoRecord, error) {
	var notesStr *string
	if bus.Notes != nil {
		encrypted, err := enc.Encrypt(bus.Notes.String())
		if err != nil {
			return objetivoRecord{}, fmt.Errorf("encrypt notes: %w", err)
		}
		notesStr = &encrypted
	}

	return objetivoRecord{
		ID:            bus.ID,
		ObjetivoID:    bus.ObjetivoID,
		UserID:        bus.UserID,
		FechaRegistro: bus.FechaRegistro.UTC(),
		Status:        bus.Status.String(),
		Notes:         notesStr,
		DateCreated:   bus.DateCreated.UTC(),
	}, nil
}

// toBusRecordDecrypted converts DB model to business model with decryption.
func toBusRecordDecrypted(db objetivoRecord, enc encrypt.Encryptor) (objetivoregistrobus.ObjetivoRecord, error) {
	// Parse status
	status, err := registrostatus.Parse(db.Status)
	if err != nil {
		return objetivoregistrobus.ObjetivoRecord{}, fmt.Errorf("parse status: %w", err)
	}

	var notesPtr *notes.Notes
	if db.Notes != nil {
		decrypted, err := enc.Decrypt(*db.Notes)
		if err != nil {
			return objetivoregistrobus.ObjetivoRecord{}, fmt.Errorf("decrypt notes: %w", err)
		}
		n, err := notes.Parse(decrypted)
		if err != nil {
			return objetivoregistrobus.ObjetivoRecord{}, fmt.Errorf("parse notes: %w", err)
		}
		notesPtr = &n
	}

	return objetivoregistrobus.ObjetivoRecord{
		ID:            db.ID,
		ObjetivoID:    db.ObjetivoID,
		UserID:        db.UserID,
		FechaRegistro: db.FechaRegistro.UTC(),
		Status:        status,
		Notes:         notesPtr,
		DateCreated:   db.DateCreated.UTC(),
	}, nil
}

// toBusRecordsDecrypted converts a slice of DB models to business models.
func toBusRecordsDecrypted(dbs []objetivoRecord, enc encrypt.Encryptor) ([]objetivoregistrobus.ObjetivoRecord, error) {
	records := make([]objetivoregistrobus.ObjetivoRecord, len(dbs))

	for i, db := range dbs {
		var err error
		records[i], err = toBusRecordDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}
