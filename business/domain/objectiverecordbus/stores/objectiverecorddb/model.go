// Package objectiverecorddb provides database access for objective records.
package objectiverecorddb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/objectiverecordbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/notes"
	"github.com/francowini/rafiki/business/types/recordstatus"
)

// objectiveRecord represents the database model.
type objectiveRecord struct {
	ID          uuid.UUID `db:"objective_record_id"`
	ObjectiveID uuid.UUID `db:"objective_id"`
	UserID      uuid.UUID `db:"user_id"`
	RecordDate  time.Time `db:"record_date"`
	Status      string    `db:"status"`
	Notes       *string   `db:"notes"` // encrypted
	DateCreated time.Time `db:"date_created"`
}

// toDBRecordEncrypted converts business model to DB model with encryption.
func toDBRecordEncrypted(bus objectiverecordbus.ObjectiveRecord, enc encrypt.Encryptor) (objectiveRecord, error) {
	var notesStr *string
	if bus.Notes != nil {
		encrypted, err := enc.Encrypt(bus.Notes.String())
		if err != nil {
			return objectiveRecord{}, fmt.Errorf("encrypt notes: %w", err)
		}
		notesStr = &encrypted
	}

	return objectiveRecord{
		ID:          bus.ID,
		ObjectiveID: bus.ObjectiveID,
		UserID:      bus.UserID,
		RecordDate:  bus.RecordDate.UTC(),
		Status:      bus.Status.String(),
		Notes:       notesStr,
		DateCreated: bus.DateCreated.UTC(),
	}, nil
}

// toBusRecordDecrypted converts DB model to business model with decryption.
func toBusRecordDecrypted(db objectiveRecord, enc encrypt.Encryptor) (objectiverecordbus.ObjectiveRecord, error) {
	// Parse status
	status, err := recordstatus.Parse(db.Status)
	if err != nil {
		return objectiverecordbus.ObjectiveRecord{}, fmt.Errorf("parse status: %w", err)
	}

	var notesPtr *notes.Notes
	if db.Notes != nil {
		decrypted, err := enc.Decrypt(*db.Notes)
		if err != nil {
			return objectiverecordbus.ObjectiveRecord{}, fmt.Errorf("decrypt notes: %w", err)
		}
		n, err := notes.Parse(decrypted)
		if err != nil {
			return objectiverecordbus.ObjectiveRecord{}, fmt.Errorf("parse notes: %w", err)
		}
		notesPtr = &n
	}

	return objectiverecordbus.ObjectiveRecord{
		ID:          db.ID,
		ObjectiveID: db.ObjectiveID,
		UserID:      db.UserID,
		RecordDate:  db.RecordDate.UTC(),
		Status:      status,
		Notes:       notesPtr,
		DateCreated: db.DateCreated.UTC(),
	}, nil
}

// toBusRecordsDecrypted converts a slice of DB models to business models.
func toBusRecordsDecrypted(dbs []objectiveRecord, enc encrypt.Encryptor) ([]objectiverecordbus.ObjectiveRecord, error) {
	records := make([]objectiverecordbus.ObjectiveRecord, len(dbs))

	for i, db := range dbs {
		var err error
		records[i], err = toBusRecordDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}
