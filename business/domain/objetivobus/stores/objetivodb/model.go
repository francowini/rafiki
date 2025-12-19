// Package objetivodb provides database access for objetivos.
package objetivodb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/cumplimientopct"
	"github.com/francowini/rafiki/business/types/frecuencian"
	"github.com/francowini/rafiki/business/types/frecuenciatype"
	"github.com/francowini/rafiki/business/types/metricaactual"
	"github.com/francowini/rafiki/business/types/metricaobjetivo"
	"github.com/francowini/rafiki/business/types/objetivostatus"
	"github.com/francowini/rafiki/business/types/objetivotitle"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// objetivo represents the database model.
type objetivo struct {
	ID                    uuid.UUID  `db:"objetivo_id"`
	UserID                uuid.UUID  `db:"user_id"`
	LifeVisionID          uuid.UUID  `db:"life_vision_id"`
	Titulo                string     `db:"titulo"` // encrypted
	TipoTracking          string     `db:"tipo_tracking"`
	Status                string     `db:"status"`
	MetricaObjetivo       *int       `db:"metrica_objetivo"`
	MetricaActual         *int       `db:"metrica_actual"`
	FrecuenciaTipo        *string    `db:"frecuencia_tipo"`
	FrecuenciaN           *int       `db:"frecuencia_n"`
	CumplimientoTargetPct *int       `db:"cumplimiento_target_pct"`
	ArchivedAt            *time.Time `db:"archived_at"`
	DateCreated           time.Time  `db:"date_created"`
	DateUpdated           time.Time  `db:"date_updated"`
}

// toDBObjetivoEncrypted converts business model to DB model with encryption.
func toDBObjetivoEncrypted(bus objetivobus.Objetivo, enc encrypt.Encryptor) (objetivo, error) {
	// Encrypt titulo
	titulo, err := enc.Encrypt(bus.Titulo.String())
	if err != nil {
		return objetivo{}, fmt.Errorf("encrypt titulo: %w", err)
	}

	var archivedAt *time.Time
	if bus.ArchivedAt != nil {
		t := bus.ArchivedAt.UTC()
		archivedAt = &t
	}

	var frecuenciaTipo *string
	if bus.FrecuenciaTipo != nil {
		s := bus.FrecuenciaTipo.String()
		frecuenciaTipo = &s
	}

	var cumplimientoTargetPct *int
	if bus.CumplimientoTargetPct != nil {
		v := bus.CumplimientoTargetPct.Value()
		cumplimientoTargetPct = &v
	}

	var metricaObjetivo *int
	if bus.MetricaObjetivo != nil {
		v := bus.MetricaObjetivo.Value()
		metricaObjetivo = &v
	}

	var metricaActual *int
	if bus.MetricaActual != nil {
		v := bus.MetricaActual.Value()
		metricaActual = &v
	}

	var frecuenciaN *int
	if bus.FrecuenciaN != nil {
		v := bus.FrecuenciaN.Value()
		frecuenciaN = &v
	}

	return objetivo{
		ID:                    bus.ID,
		UserID:                bus.UserID,
		LifeVisionID:          bus.LifeVisionID,
		Titulo:                titulo,
		TipoTracking:          bus.TipoTracking.String(),
		Status:                bus.Status.String(),
		MetricaObjetivo:       metricaObjetivo,
		MetricaActual:         metricaActual,
		FrecuenciaTipo:        frecuenciaTipo,
		FrecuenciaN:           frecuenciaN,
		CumplimientoTargetPct: cumplimientoTargetPct,
		ArchivedAt:            archivedAt,
		DateCreated:           bus.DateCreated.UTC(),
		DateUpdated:           bus.DateUpdated.UTC(),
	}, nil
}

// toBusObjetivoDecrypted converts DB model to business model with decryption.
func toBusObjetivoDecrypted(db objetivo, enc encrypt.Encryptor) (objetivobus.Objetivo, error) {
	// Decrypt and parse titulo
	tituloStr, err := enc.Decrypt(db.Titulo)
	if err != nil {
		return objetivobus.Objetivo{}, fmt.Errorf("decrypt titulo: %w", err)
	}

	titulo, err := objetivotitle.Parse(tituloStr)
	if err != nil {
		return objetivobus.Objetivo{}, fmt.Errorf("parse titulo: %w", err)
	}

	// Parse tipo_tracking
	tipoTracking, err := trackingtype.Parse(db.TipoTracking)
	if err != nil {
		return objetivobus.Objetivo{}, fmt.Errorf("parse tipo_tracking: %w", err)
	}

	// Parse status
	status, err := objetivostatus.Parse(db.Status)
	if err != nil {
		return objetivobus.Objetivo{}, fmt.Errorf("parse status: %w", err)
	}

	var frecuenciaTipo *frecuenciatype.FrecuenciaType
	if db.FrecuenciaTipo != nil {
		ft, err := frecuenciatype.Parse(*db.FrecuenciaTipo)
		if err != nil {
			return objetivobus.Objetivo{}, fmt.Errorf("parse frecuencia_tipo: %w", err)
		}
		frecuenciaTipo = &ft
	}

	var cumplimientoTargetPct *cumplimientopct.CumplimientoPct
	if db.CumplimientoTargetPct != nil {
		cp, err := cumplimientopct.Parse(*db.CumplimientoTargetPct)
		if err != nil {
			return objetivobus.Objetivo{}, fmt.Errorf("parse cumplimiento_target_pct: %w", err)
		}
		cumplimientoTargetPct = &cp
	}

	var metricaObjetivoPtr *metricaobjetivo.MetricaObjetivo
	if db.MetricaObjetivo != nil {
		mo, err := metricaobjetivo.Parse(*db.MetricaObjetivo)
		if err != nil {
			return objetivobus.Objetivo{}, fmt.Errorf("parse metrica_objetivo: %w", err)
		}
		metricaObjetivoPtr = &mo
	}

	var metricaActualPtr *metricaactual.MetricaActual
	if db.MetricaActual != nil {
		ma, err := metricaactual.Parse(*db.MetricaActual)
		if err != nil {
			return objetivobus.Objetivo{}, fmt.Errorf("parse metrica_actual: %w", err)
		}
		metricaActualPtr = &ma
	}

	var frecuenciaNPtr *frecuencian.FrecuenciaN
	if db.FrecuenciaN != nil {
		fn, err := frecuencian.Parse(*db.FrecuenciaN)
		if err != nil {
			return objetivobus.Objetivo{}, fmt.Errorf("parse frecuencia_n: %w", err)
		}
		frecuenciaNPtr = &fn
	}

	var archivedAt *time.Time
	if db.ArchivedAt != nil {
		t := db.ArchivedAt.UTC()
		archivedAt = &t
	}

	return objetivobus.Objetivo{
		ID:                    db.ID,
		UserID:                db.UserID,
		LifeVisionID:          db.LifeVisionID,
		Titulo:                titulo,
		TipoTracking:          tipoTracking,
		Status:                status,
		MetricaObjetivo:       metricaObjetivoPtr,
		MetricaActual:         metricaActualPtr,
		FrecuenciaTipo:        frecuenciaTipo,
		FrecuenciaN:           frecuenciaNPtr,
		CumplimientoTargetPct: cumplimientoTargetPct,
		ArchivedAt:            archivedAt,
		DateCreated:           db.DateCreated.UTC(),
		DateUpdated:           db.DateUpdated.UTC(),
	}, nil
}

// toBusObjetivosDecrypted converts a slice of DB models to business models.
func toBusObjetivosDecrypted(dbs []objetivo, enc encrypt.Encryptor) ([]objetivobus.Objetivo, error) {
	objetivos := make([]objetivobus.Objetivo, len(dbs))

	for i, db := range dbs {
		var err error
		objetivos[i], err = toBusObjetivoDecrypted(db, enc)
		if err != nil {
			return nil, fmt.Errorf("record %d (id=%s): %w", i, db.ID, err)
		}
	}

	return objetivos, nil
}
