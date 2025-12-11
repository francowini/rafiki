package vnotificationdb

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vnotificationbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
)

type valueWithVision struct {
	UserID            uuid.UUID  `db:"user_id"`
	ValueID           uuid.UUID  `db:"value_id"`
	ValueContent      string     `db:"value_content"`
	ValueFacet        string     `db:"value_facet"`
	ValueOrder        int        `db:"value_order"`
	LifeVisionID      *uuid.UUID `db:"life_vision_id"`
	LifeVisionContent *string    `db:"life_vision_content"`
}

func toBusValueWithVision(db valueWithVision, enc encrypt.Encryptor) (vnotificationbus.ValueWithVision, error) {
	// Decrypt value content
	valueContent, err := enc.Decrypt(db.ValueContent)
	if err != nil {
		return vnotificationbus.ValueWithVision{}, fmt.Errorf("decrypt ValueContent: %w", err)
	}

	// Decrypt life vision content if present
	var lifeVisionContent *string
	if db.LifeVisionContent != nil {
		decrypted, err := enc.Decrypt(*db.LifeVisionContent)
		if err != nil {
			return vnotificationbus.ValueWithVision{}, fmt.Errorf("decrypt LifeVisionContent: %w", err)
		}
		lifeVisionContent = &decrypted
	}

	return vnotificationbus.ValueWithVision{
		UserID:            db.UserID,
		ValueID:           db.ValueID,
		ValueContent:      valueContent,
		ValueFacet:        db.ValueFacet,
		ValueOrder:        db.ValueOrder,
		LifeVisionID:      db.LifeVisionID,
		LifeVisionContent: lifeVisionContent,
	}, nil
}

func toBusValuesWithVision(dbs []valueWithVision, enc encrypt.Encryptor) ([]vnotificationbus.ValueWithVision, error) {
	values := make([]vnotificationbus.ValueWithVision, 0, len(dbs))

	for _, db := range dbs {
		v, err := toBusValueWithVision(db, enc)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}
