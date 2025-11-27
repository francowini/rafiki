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

func toBusExportItemDecrypted(db exportItem, enc encrypt.Encryptor) (vexportbus.ExportItem, error) {
	item := vexportbus.ExportItem{
		ID:          db.ID,
		UserID:      db.UserID,
		ItemType:    vexportbus.ItemType(db.ItemType),
		ItemDate:    db.ItemDate.In(time.Local),
		DateCreated: db.DateCreated.In(time.Local),
	}

	// Decrypt moment-specific fields
	if db.Situation.Valid {
		decrypted, err := enc.Decrypt(db.Situation.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt situation: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse situation: %w", err)
		}
		item.Situation = &parsed
	}

	if db.Thoughts.Valid {
		decrypted, err := enc.Decrypt(db.Thoughts.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt thoughts: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse thoughts: %w", err)
		}
		item.Thoughts = &parsed
	}

	if db.PhysicalSymptoms.Valid {
		decrypted, err := enc.Decrypt(db.PhysicalSymptoms.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt physical_symptoms: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse physical_symptoms: %w", err)
		}
		item.PhysicalSymptoms = &parsed
	}

	if db.Behavior.Valid {
		decrypted, err := enc.Decrypt(db.Behavior.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt behavior: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse behavior: %w", err)
		}
		item.Behavior = &parsed
	}

	if db.Consequences.Valid {
		decrypted, err := enc.Decrypt(db.Consequences.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt consequences: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse consequences: %w", err)
		}
		item.Consequences = &parsed
	}

	if db.ValuesReflection.Valid {
		decrypted, err := enc.Decrypt(db.ValuesReflection.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt values_reflection: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse values_reflection: %w", err)
		}
		item.ValuesReflection = &parsed
	}

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

	if db.Content.Valid {
		decrypted, err := enc.Decrypt(db.Content.String)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("decrypt content: %w", err)
		}
		parsed, err := content.Parse(decrypted)
		if err != nil {
			return vexportbus.ExportItem{}, fmt.Errorf("parse content: %w", err)
		}
		item.Content = &parsed
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
