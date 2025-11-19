package thinkdb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/thinkbus"
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

// toDBThink converts business Think to database model
func toDBThink(bus thinkbus.Think) think {
	return think{
		ID:          bus.ID,
		UserID:      bus.UserID,
		Category:    bus.Category.String(),
		Content:     bus.Content.String(),
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}
}

// toBusThink converts database model to business Think
func toBusThink(db think) (thinkbus.Think, error) {
	category, err := thinkbus.ParseCategory(db.Category)
	if err != nil {
		return thinkbus.Think{}, fmt.Errorf("parse category: %w", err)
	}

	// Parse content string back to Content type
	cnt, err := content.Parse(db.Content)
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

// toBusThinks converts slice of database models to business models
func toBusThinks(dbs []think) ([]thinkbus.Think, error) {
	thinks := make([]thinkbus.Think, len(dbs))

	for i, db := range dbs {
		var err error
		thinks[i], err = toBusThink(db)
		if err != nil {
			return nil, err
		}
	}

	return thinks, nil
}
