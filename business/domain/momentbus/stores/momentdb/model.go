package momentdb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

type moment struct {
	ID               uuid.UUID `db:"moment_id"`
	UserID           uuid.UUID `db:"user_id"`
	MomentDate       time.Time `db:"moment_date"`
	Situation        string    `db:"situation"`
	Thoughts         string    `db:"thoughts"`
	PhysicalSymptoms string    `db:"physical_symptoms"`
	Behavior         string    `db:"behavior"`
	Consequences     string    `db:"consequences"`
	ValuesReflection string    `db:"values_reflection"`
	Intensity        int       `db:"intensity"`
	DateCreated      time.Time `db:"date_created"`
	DateUpdated      time.Time `db:"date_updated"`
}

// toDBMomentEncrypted converts business Moment to database model with encrypted fields.
func toDBMomentEncrypted(bus momentbus.Moment, enc encrypt.Encryptor) (moment, error) {
	situation, err := enc.Encrypt(bus.Situation.String())
	if err != nil {
		return moment{}, fmt.Errorf("encrypt situation: %w", err)
	}

	thoughts, err := enc.Encrypt(bus.Thoughts.String())
	if err != nil {
		return moment{}, fmt.Errorf("encrypt thoughts: %w", err)
	}

	physicalSymptoms, err := enc.Encrypt(bus.PhysicalSymptoms.String())
	if err != nil {
		return moment{}, fmt.Errorf("encrypt physical_symptoms: %w", err)
	}

	behavior, err := enc.Encrypt(bus.Behavior.String())
	if err != nil {
		return moment{}, fmt.Errorf("encrypt behavior: %w", err)
	}

	consequences, err := enc.Encrypt(bus.Consequences.String())
	if err != nil {
		return moment{}, fmt.Errorf("encrypt consequences: %w", err)
	}

	valuesReflection, err := enc.Encrypt(bus.ValuesReflection.String())
	if err != nil {
		return moment{}, fmt.Errorf("encrypt values_reflection: %w", err)
	}

	return moment{
		ID:               bus.ID,
		UserID:           bus.UserID,
		MomentDate:       bus.MomentDate.UTC(),
		Situation:        situation,
		Thoughts:         thoughts,
		PhysicalSymptoms: physicalSymptoms,
		Behavior:         behavior,
		Consequences:     consequences,
		ValuesReflection: valuesReflection,
		Intensity:        bus.Intensity.Value(),
		DateCreated:      bus.DateCreated.UTC(),
		DateUpdated:      bus.DateUpdated.UTC(),
	}, nil
}

// toBusMomentDecrypted converts database model to business Moment with decrypted fields.
func toBusMomentDecrypted(db moment, enc encrypt.Encryptor) (momentbus.Moment, error) {
	situationStr, err := enc.Decrypt(db.Situation)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("decrypt situation: %w", err)
	}
	situation, err := content.Parse(situationStr)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse situation: %w", err)
	}

	thoughtsStr, err := enc.Decrypt(db.Thoughts)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("decrypt thoughts: %w", err)
	}
	thoughts, err := content.Parse(thoughtsStr)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse thoughts: %w", err)
	}

	physicalSymptomsStr, err := enc.Decrypt(db.PhysicalSymptoms)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("decrypt physical_symptoms: %w", err)
	}
	physicalSymptoms, err := content.Parse(physicalSymptomsStr)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse physical_symptoms: %w", err)
	}

	behaviorStr, err := enc.Decrypt(db.Behavior)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("decrypt behavior: %w", err)
	}
	behavior, err := content.Parse(behaviorStr)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse behavior: %w", err)
	}

	consequencesStr, err := enc.Decrypt(db.Consequences)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("decrypt consequences: %w", err)
	}
	consequences, err := content.Parse(consequencesStr)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse consequences: %w", err)
	}

	valuesReflectionStr, err := enc.Decrypt(db.ValuesReflection)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("decrypt values_reflection: %w", err)
	}
	valuesReflection, err := content.Parse(valuesReflectionStr)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse values_reflection: %w", err)
	}

	intensityVal, err := intensity.Parse(db.Intensity)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse intensity: %w", err)
	}

	return momentbus.Moment{
		ID:               db.ID,
		UserID:           db.UserID,
		MomentDate:       db.MomentDate.In(time.Local),
		Situation:        situation,
		Thoughts:         thoughts,
		PhysicalSymptoms: physicalSymptoms,
		Behavior:         behavior,
		Consequences:     consequences,
		ValuesReflection: valuesReflection,
		Intensity:        intensityVal,
		DateCreated:      db.DateCreated.In(time.Local),
		DateUpdated:      db.DateUpdated.In(time.Local),
	}, nil
}

// toBusMomentsDecrypted converts slice of database models to business models with decryption.
func toBusMomentsDecrypted(dbs []moment, enc encrypt.Encryptor) ([]momentbus.Moment, error) {
	moments := make([]momentbus.Moment, len(dbs))

	for i, db := range dbs {
		var err error
		moments[i], err = toBusMomentDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return moments, nil
}
