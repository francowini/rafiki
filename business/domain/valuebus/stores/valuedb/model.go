// Package valuedb provides database access for values.
package valuedb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/business/types/valuecontent"
)

// value represents the database model.
type value struct {
	ID           uuid.UUID `db:"value_id"`
	UserID       uuid.UUID `db:"user_id"`
	Content      string    `db:"content"` // encrypted
	Facet        string    `db:"facet"`
	DisplayOrder int       `db:"display_order"`
	DateCreated  time.Time `db:"date_created"`
	DateUpdated  time.Time `db:"date_updated"`
}

// toDBValueEncrypted converts business model to DB model with encryption.
func toDBValueEncrypted(bus valuebus.Value, enc encrypt.Encryptor) (value, error) {
	// Encrypt content
	content, err := enc.Encrypt(bus.Content.String())
	if err != nil {
		return value{}, fmt.Errorf("encrypt content: %w", err)
	}

	return value{
		ID:           bus.ID,
		UserID:       bus.UserID,
		Content:      content,
		Facet:        bus.Facet.String(),
		DisplayOrder: bus.DisplayOrder,
		DateCreated:  bus.DateCreated.UTC(),
		DateUpdated:  bus.DateUpdated.UTC(),
	}, nil
}

// toBusValueDecrypted converts DB model to business model with decryption.
func toBusValueDecrypted(db value, enc encrypt.Encryptor) (valuebus.Value, error) {
	// Decrypt and parse content
	contentStr, err := enc.Decrypt(db.Content)
	if err != nil {
		return valuebus.Value{}, fmt.Errorf("decrypt content: %w", err)
	}

	content, err := valuecontent.Parse(contentStr)
	if err != nil {
		return valuebus.Value{}, fmt.Errorf("parse content: %w", err)
	}

	// Parse facet
	facetVal, err := facet.Parse(db.Facet)
	if err != nil {
		return valuebus.Value{}, fmt.Errorf("parse facet: %w", err)
	}

	return valuebus.Value{
		ID:           db.ID,
		UserID:       db.UserID,
		Content:      content,
		Facet:        facetVal,
		DisplayOrder: db.DisplayOrder,
		DateCreated:  db.DateCreated.In(time.Local),
		DateUpdated:  db.DateUpdated.In(time.Local),
	}, nil
}

// toBusValuesDecrypted converts a slice of DB models to business models.
func toBusValuesDecrypted(dbs []value, enc encrypt.Encryptor) ([]valuebus.Value, error) {
	values := make([]valuebus.Value, len(dbs))

	for i, db := range dbs {
		var err error
		values[i], err = toBusValueDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}
