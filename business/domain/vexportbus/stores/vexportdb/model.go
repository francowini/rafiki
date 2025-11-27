package vexportdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

type exportItem struct {
	ID               uuid.UUID      `db:"item_id"`
	UserID           uuid.UUID      `db:"user_id"`
	ItemType         string         `db:"item_type"`
	ItemDate         time.Time      `db:"item_date"`
	Situation        sql.NullString `db:"situation"`
	Thoughts         sql.NullString `db:"thoughts"`
	PhysicalSymptoms sql.NullString `db:"physical_symptoms"`
	Behavior         sql.NullString `db:"behavior"`
	Consequences     sql.NullString `db:"consequences"`
	ValuesReflection sql.NullString `db:"values_reflection"`
	Intensity        sql.NullInt32  `db:"intensity"`
	Category         sql.NullString `db:"category"`
	Content          sql.NullString `db:"content"`
	DateCreated      time.Time      `db:"date_created"`
}

// decryptContent decrypts and parses an encrypted content field.
// Returns nil if the field is not valid (NULL in database).
func decryptContent(enc encrypt.Encryptor, field sql.NullString, fieldName string) (*content.Content, error) {
	if !field.Valid {
		return nil, nil
	}

	decrypted, err := enc.Decrypt(field.String)
	if err != nil {
		return nil, fmt.Errorf("decrypt %s: %w", fieldName, err)
	}

	parsed, err := content.Parse(decrypted)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", fieldName, err)
	}

	return &parsed, nil
}

func toBusExportItemDecrypted(db exportItem, enc encrypt.Encryptor) (vexportbus.ExportItem, error) {
	// Validate ItemType before assignment
	itemType, err := vexportbus.ParseItemType(db.ItemType)
	if err != nil {
		return vexportbus.ExportItem{}, fmt.Errorf("parse item_type: %w", err)
	}

	// Keep timestamps in UTC - timezone conversion happens at the API layer
	item := vexportbus.ExportItem{
		ID:          db.ID,
		UserID:      db.UserID,
		ItemType:    itemType,
		ItemDate:    db.ItemDate.UTC(),
		DateCreated: db.DateCreated.UTC(),
	}

	// Decrypt moment-specific content fields
	if item.Situation, err = decryptContent(enc, db.Situation, "situation"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	if item.Thoughts, err = decryptContent(enc, db.Thoughts, "thoughts"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	if item.PhysicalSymptoms, err = decryptContent(enc, db.PhysicalSymptoms, "physical_symptoms"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	if item.Behavior, err = decryptContent(enc, db.Behavior, "behavior"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	if item.Consequences, err = decryptContent(enc, db.Consequences, "consequences"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	if item.ValuesReflection, err = decryptContent(enc, db.ValuesReflection, "values_reflection"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	// Parse intensity (not encrypted, just needs validation)
	if db.Intensity.Valid {
		parsed, err := intensity.Parse(int(db.Intensity.Int32))
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse intensity: %w", err)
		}
		item.Intensity = &parsed
	}

	// Parse think-specific fields
	if db.Category.Valid {
		cat, err := thinkbus.ParseCategory(db.Category.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse category: %w", err)
		}
		item.Category = &cat
	}

	// Decrypt think content field
	if item.Content, err = decryptContent(enc, db.Content, "content"); err != nil {
		return vexportbus.ExportItem{}, err
	}

	return item, nil
}

func toBusExportItemsDecrypted(dbs []exportItem, enc encrypt.Encryptor) ([]vexportbus.ExportItem, error) {
	items := make([]vexportbus.ExportItem, len(dbs))

	for i, db := range dbs {
		var err error
		items[i], err = toBusExportItemDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}
