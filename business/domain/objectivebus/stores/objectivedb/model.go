// Package objectivedb provides database access for objectives.
package objectivedb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/compliancepct"
	"github.com/francowini/rafiki/business/types/currentmetric"
	"github.com/francowini/rafiki/business/types/frequencycount"
	"github.com/francowini/rafiki/business/types/frequencytype"
	"github.com/francowini/rafiki/business/types/objectivestatus"
	"github.com/francowini/rafiki/business/types/objectivetitle"
	"github.com/francowini/rafiki/business/types/targetmetric"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// objective represents the database model.
type objective struct {
	ID                  uuid.UUID  `db:"objective_id"`
	UserID              uuid.UUID  `db:"user_id"`
	LifeVisionID        uuid.UUID  `db:"life_vision_id"`
	Title               string     `db:"title"` // encrypted
	TrackingType        string     `db:"tracking_type"`
	Status              string     `db:"status"`
	TargetMetric        *int       `db:"target_metric"`
	CurrentMetric       *int       `db:"current_metric"`
	FrequencyType       *string    `db:"frequency_type"`
	FrequencyCount      *int       `db:"frequency_count"`
	ComplianceTargetPct *int       `db:"compliance_target_pct"`
	ArchivedAt          *time.Time `db:"archived_at"`
	DateCreated         time.Time  `db:"date_created"`
	DateUpdated         time.Time  `db:"date_updated"`
}

// toDBObjectiveEncrypted converts business model to DB model with encryption.
func toDBObjectiveEncrypted(bus objectivebus.Objective, enc encrypt.Encryptor) (objective, error) {
	// Encrypt title
	title, err := enc.Encrypt(bus.Title.String())
	if err != nil {
		return objective{}, fmt.Errorf("encrypt title: %w", err)
	}

	var archivedAt *time.Time
	if bus.ArchivedAt != nil {
		t := bus.ArchivedAt.UTC()
		archivedAt = &t
	}

	var frequencyType *string
	if bus.FrequencyType != nil {
		s := bus.FrequencyType.String()
		frequencyType = &s
	}

	var complianceTargetPct *int
	if bus.ComplianceTargetPct != nil {
		v := bus.ComplianceTargetPct.Value()
		complianceTargetPct = &v
	}

	var targetMetricVal *int
	if bus.TargetMetric != nil {
		v := bus.TargetMetric.Value()
		targetMetricVal = &v
	}

	var currentMetricVal *int
	if bus.CurrentMetric != nil {
		v := bus.CurrentMetric.Value()
		currentMetricVal = &v
	}

	var frequencyCountVal *int
	if bus.FrequencyCount != nil {
		v := bus.FrequencyCount.Value()
		frequencyCountVal = &v
	}

	return objective{
		ID:                  bus.ID,
		UserID:              bus.UserID,
		LifeVisionID:        bus.LifeVisionID,
		Title:               title,
		TrackingType:        bus.TrackingType.String(),
		Status:              bus.Status.String(),
		TargetMetric:        targetMetricVal,
		CurrentMetric:       currentMetricVal,
		FrequencyType:       frequencyType,
		FrequencyCount:      frequencyCountVal,
		ComplianceTargetPct: complianceTargetPct,
		ArchivedAt:          archivedAt,
		DateCreated:         bus.DateCreated.UTC(),
		DateUpdated:         bus.DateUpdated.UTC(),
	}, nil
}

// toBusObjectiveDecrypted converts DB model to business model with decryption.
func toBusObjectiveDecrypted(db objective, enc encrypt.Encryptor) (objectivebus.Objective, error) {
	// Decrypt and parse title
	titleStr, err := enc.Decrypt(db.Title)
	if err != nil {
		return objectivebus.Objective{}, fmt.Errorf("decrypt title: %w", err)
	}

	title, err := objectivetitle.Parse(titleStr)
	if err != nil {
		return objectivebus.Objective{}, fmt.Errorf("parse title: %w", err)
	}

	// Parse tracking_type
	trackingTypeVal, err := trackingtype.Parse(db.TrackingType)
	if err != nil {
		return objectivebus.Objective{}, fmt.Errorf("parse tracking_type: %w", err)
	}

	// Parse status
	status, err := objectivestatus.Parse(db.Status)
	if err != nil {
		return objectivebus.Objective{}, fmt.Errorf("parse status: %w", err)
	}

	var frequencyTypePtr *frequencytype.FrequencyType
	if db.FrequencyType != nil {
		ft, err := frequencytype.Parse(*db.FrequencyType)
		if err != nil {
			return objectivebus.Objective{}, fmt.Errorf("parse frequency_type: %w", err)
		}
		frequencyTypePtr = &ft
	}

	var complianceTargetPctPtr *compliancepct.CompliancePct
	if db.ComplianceTargetPct != nil {
		cp, err := compliancepct.Parse(*db.ComplianceTargetPct)
		if err != nil {
			return objectivebus.Objective{}, fmt.Errorf("parse compliance_target_pct: %w", err)
		}
		complianceTargetPctPtr = &cp
	}

	var targetMetricPtr *targetmetric.TargetMetric
	if db.TargetMetric != nil {
		tm, err := targetmetric.Parse(*db.TargetMetric)
		if err != nil {
			return objectivebus.Objective{}, fmt.Errorf("parse target_metric: %w", err)
		}
		targetMetricPtr = &tm
	}

	var currentMetricPtr *currentmetric.CurrentMetric
	if db.CurrentMetric != nil {
		cm, err := currentmetric.Parse(*db.CurrentMetric)
		if err != nil {
			return objectivebus.Objective{}, fmt.Errorf("parse current_metric: %w", err)
		}
		currentMetricPtr = &cm
	}

	var frequencyCountPtr *frequencycount.FrequencyCount
	if db.FrequencyCount != nil {
		fc, err := frequencycount.Parse(*db.FrequencyCount)
		if err != nil {
			return objectivebus.Objective{}, fmt.Errorf("parse frequency_count: %w", err)
		}
		frequencyCountPtr = &fc
	}

	var archivedAt *time.Time
	if db.ArchivedAt != nil {
		t := db.ArchivedAt.UTC()
		archivedAt = &t
	}

	return objectivebus.Objective{
		ID:                  db.ID,
		UserID:              db.UserID,
		LifeVisionID:        db.LifeVisionID,
		Title:               title,
		TrackingType:        trackingTypeVal,
		Status:              status,
		TargetMetric:        targetMetricPtr,
		CurrentMetric:       currentMetricPtr,
		FrequencyType:       frequencyTypePtr,
		FrequencyCount:      frequencyCountPtr,
		ComplianceTargetPct: complianceTargetPctPtr,
		ArchivedAt:          archivedAt,
		DateCreated:         db.DateCreated.UTC(),
		DateUpdated:         db.DateUpdated.UTC(),
	}, nil
}

// toBusObjectivesDecrypted converts a slice of DB models to business models.
func toBusObjectivesDecrypted(dbs []objective, enc encrypt.Encryptor) ([]objectivebus.Objective, error) {
	objectives := make([]objectivebus.Objective, len(dbs))

	for i, db := range dbs {
		var err error
		objectives[i], err = toBusObjectiveDecrypted(db, enc)
		if err != nil {
			return nil, fmt.Errorf("record %d (id=%s): %w", i, db.ID, err)
		}
	}

	return objectives, nil
}
