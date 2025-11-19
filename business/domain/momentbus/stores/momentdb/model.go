package momentdb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/momentbus"
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

func toDBMoment(bus momentbus.Moment) moment {
	return moment{
		ID:               bus.ID,
		UserID:           bus.UserID,
		MomentDate:       bus.MomentDate.UTC(),
		Situation:        bus.Situation.String(),
		Thoughts:         bus.Thoughts.String(),
		PhysicalSymptoms: bus.PhysicalSymptoms.String(),
		Behavior:         bus.Behavior.String(),
		Consequences:     bus.Consequences.String(),
		ValuesReflection: bus.ValuesReflection.String(),
		Intensity:        bus.Intensity.Value(),
		DateCreated:      bus.DateCreated.UTC(),
		DateUpdated:      bus.DateUpdated.UTC(),
	}
}

func toBusMoment(db moment) (momentbus.Moment, error) {
	situation, err := content.Parse(db.Situation)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse situation: %w", err)
	}

	thoughts, err := content.Parse(db.Thoughts)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse thoughts: %w", err)
	}

	physicalSymptoms, err := content.Parse(db.PhysicalSymptoms)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse physical_symptoms: %w", err)
	}

	behavior, err := content.Parse(db.Behavior)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse behavior: %w", err)
	}

	consequences, err := content.Parse(db.Consequences)
	if err != nil {
		return momentbus.Moment{}, fmt.Errorf("parse consequences: %w", err)
	}

	valuesReflection, err := content.Parse(db.ValuesReflection)
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

func toBusMoments(dbs []moment) ([]momentbus.Moment, error) {
	moments := make([]momentbus.Moment, len(dbs))

	for i, db := range dbs {
		var err error
		moments[i], err = toBusMoment(db)
		if err != nil {
			return nil, err
		}
	}

	return moments, nil
}
