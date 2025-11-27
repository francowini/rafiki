// Package lifevisiondb provides database access for life visions.
package lifevisiondb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/lifevisioncontent"
)

// lifeVision represents the database model.
type lifeVision struct {
	ID          uuid.UUID `db:"life_vision_id"`
	UserID      uuid.UUID `db:"user_id"`
	ValueID     uuid.UUID `db:"value_id"`
	Content     string    `db:"content"` // encrypted
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

// toDBLifeVisionEncrypted converts business model to DB model with encryption.
func toDBLifeVisionEncrypted(bus lifevisionbus.LifeVision, enc encrypt.Encryptor) (lifeVision, error) {
	// Encrypt content
	content, err := enc.Encrypt(bus.Content.String())
	if err != nil {
		return lifeVision{}, fmt.Errorf("encrypt content: %w", err)
	}

	return lifeVision{
		ID:          bus.ID,
		UserID:      bus.UserID,
		ValueID:     bus.ValueID,
		Content:     content,
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}, nil
}

// toBusLifeVisionDecrypted converts DB model to business model with decryption.
func toBusLifeVisionDecrypted(db lifeVision, enc encrypt.Encryptor) (lifevisionbus.LifeVision, error) {
	// Decrypt and parse content
	contentStr, err := enc.Decrypt(db.Content)
	if err != nil {
		return lifevisionbus.LifeVision{}, fmt.Errorf("decrypt content: %w", err)
	}

	content, err := lifevisioncontent.Parse(contentStr)
	if err != nil {
		return lifevisionbus.LifeVision{}, fmt.Errorf("parse content: %w", err)
	}

	return lifevisionbus.LifeVision{
		ID:          db.ID,
		UserID:      db.UserID,
		ValueID:     db.ValueID,
		Content:     content,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}, nil
}

// toBusLifeVisionsDecrypted converts a slice of DB models to business models.
func toBusLifeVisionsDecrypted(dbs []lifeVision, enc encrypt.Encryptor) ([]lifevisionbus.LifeVision, error) {
	lifeVisions := make([]lifevisionbus.LifeVision, len(dbs))

	for i, db := range dbs {
		var err error
		lifeVisions[i], err = toBusLifeVisionDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return lifeVisions, nil
}
