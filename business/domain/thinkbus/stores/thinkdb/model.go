package thinkdb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/content"
)

// think represents the database model
type think struct {
	ID          uuid.UUID `db:"think_id"`
	UserID      uuid.UUID `db:"user_id"`
	Category    string    `db:"category"`
	Content     string    `db:"content"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

// toDBThinkEncrypted converts business Think to database model with encrypted content.
func toDBThinkEncrypted(bus thinkbus.Think, enc encrypt.Encryptor) (think, error) {
	encryptedContent, err := enc.Encrypt(bus.Content.String())
	if err != nil {
		return think{}, fmt.Errorf("encrypt content: %w", err)
	}

	return think{
		ID:          bus.ID,
		UserID:      bus.UserID,
		Category:    bus.Category.String(),
		Content:     encryptedContent,
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}, nil
}

// toBusThinkDecrypted converts database model to business Think with decrypted content.
func toBusThinkDecrypted(db think, enc encrypt.Encryptor) (thinkbus.Think, error) {
	decryptedContent, err := enc.Decrypt(db.Content)
	if err != nil {
		return thinkbus.Think{}, fmt.Errorf("decrypt content: %w", err)
	}

	category, err := thinkbus.ParseCategory(db.Category)
	if err != nil {
		return thinkbus.Think{}, fmt.Errorf("parse category: %w", err)
	}

	cnt, err := content.Parse(decryptedContent)
	if err != nil {
		return thinkbus.Think{}, fmt.Errorf("parse content: %w", err)
	}

	return thinkbus.Think{
		ID:          db.ID,
		UserID:      db.UserID,
		Category:    category,
		Content:     cnt,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}, nil
}

// toBusThinkDecryptedSlice converts slice of database models to business models with decryption.
func toBusThinkDecryptedSlice(dbs []think, enc encrypt.Encryptor) ([]thinkbus.Think, error) {
	thinks := make([]thinkbus.Think, len(dbs))

	for i, db := range dbs {
		var err error
		thinks[i], err = toBusThinkDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return thinks, nil
}
