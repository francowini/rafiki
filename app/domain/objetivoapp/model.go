// Package objetivoapp provides HTTP handlers for objetivos.
package objetivoapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/types/cumplimientopct"
	"github.com/francowini/rafiki/business/types/frecuencian"
	"github.com/francowini/rafiki/business/types/frecuenciatype"
	"github.com/francowini/rafiki/business/types/metricaobjetivo"
	"github.com/francowini/rafiki/business/types/objetivostatus"
	"github.com/francowini/rafiki/business/types/objetivotitle"
	"github.com/francowini/rafiki/business/types/trackingtype"
)

// Objetivo represents an objetivo for API responses.
type Objetivo struct {
	ID                    string  `json:"id"`
	LifeVisionID          string  `json:"lifeVisionId"`
	Titulo                string  `json:"titulo"`
	TipoTracking          string  `json:"tipoTracking"`
	Status                string  `json:"status"`
	MetricaObjetivo       *int    `json:"metricaObjetivo,omitempty"`
	MetricaActual         *int    `json:"metricaActual,omitempty"`
	FrecuenciaTipo        *string `json:"frecuenciaTipo,omitempty"`
	FrecuenciaN           *int    `json:"frecuenciaN,omitempty"`
	CumplimientoTargetPct *int    `json:"cumplimientoTargetPct,omitempty"`
	ArchivedAt            *string `json:"archivedAt,omitempty"`
	DateCreated           string  `json:"dateCreated"`
	DateUpdated           string  `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (o Objetivo) Encode() ([]byte, string, error) {
	data, err := json.Marshal(o)
	return data, "application/json", err
}

// NewObjetivo represents data for creating a new objetivo.
type NewObjetivo struct {
	LifeVisionID          string  `json:"lifeVisionId" validate:"required,uuid"`
	Titulo                string  `json:"titulo" validate:"required"`
	TipoTracking          string  `json:"tipoTracking" validate:"required,oneof=resultado frecuencia"`
	MetricaObjetivo       *int    `json:"metricaObjetivo"`
	FrecuenciaTipo        *string `json:"frecuenciaTipo" validate:"omitempty,oneof=daily n_por_semana n_por_mes"`
	FrecuenciaN           *int    `json:"frecuenciaN"`
	CumplimientoTargetPct *int    `json:"cumplimientoTargetPct" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (no *NewObjetivo) Decode(data []byte) error {
	return json.Unmarshal(data, no)
}

// Validate checks the data for correctness.
func (no NewObjetivo) Validate() error {
	if err := errs.Check(no); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateObjetivo represents data for updating an objetivo.
type UpdateObjetivo struct {
	Titulo                *string `json:"titulo"`
	MetricaObjetivo       *int    `json:"metricaObjetivo"`
	FrecuenciaTipo        *string `json:"frecuenciaTipo" validate:"omitempty,oneof=daily n_por_semana n_por_mes"`
	FrecuenciaN           *int    `json:"frecuenciaN"`
	CumplimientoTargetPct *int    `json:"cumplimientoTargetPct" validate:"omitempty,min=1,max=100"`
}

// Decode implements the decoder interface.
func (uo *UpdateObjetivo) Decode(data []byte) error {
	return json.Unmarshal(data, uo)
}

// Validate checks the data for correctness.
func (uo UpdateObjetivo) Validate() error {
	if err := errs.Check(uo); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ChangeStatusRequest represents data for changing objetivo status.
type ChangeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=activo completado abandonado pausado"`
}

// Decode implements the decoder interface.
func (csr *ChangeStatusRequest) Decode(data []byte) error {
	return json.Unmarshal(data, csr)
}

// Validate checks the data for correctness.
func (csr ChangeStatusRequest) Validate() error {
	if err := errs.Check(csr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateProgressRequest represents data for updating progress (increment or set value).
type UpdateProgressRequest struct {
	Increment *int `json:"increment"` // Increment/decrement by this value
	Value     *int `json:"value"`     // Set to this exact value
}

// Decode implements the decoder interface.
func (upr *UpdateProgressRequest) Decode(data []byte) error {
	return json.Unmarshal(data, upr)
}

// Validate checks the data for correctness.
func (upr UpdateProgressRequest) Validate() error {
	if upr.Increment == nil && upr.Value == nil {
		return fmt.Errorf("validate: either increment or value must be provided")
	}
	if upr.Increment != nil && upr.Value != nil {
		return fmt.Errorf("validate: cannot provide both increment and value")
	}
	if upr.Value != nil && *upr.Value < 0 {
		return fmt.Errorf("validate: value must be non-negative")
	}
	return nil
}

// ===== Business → App conversions =====

func toAppObjetivo(o objetivobus.Objetivo) Objetivo {
	var archivedAt *string
	if o.ArchivedAt != nil {
		t := o.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		archivedAt = &t
	}

	var frecuenciaTipo *string
	if o.FrecuenciaTipo != nil {
		ft := o.FrecuenciaTipo.String()
		frecuenciaTipo = &ft
	}

	var cumplimientoTargetPct *int
	if o.CumplimientoTargetPct != nil {
		cp := o.CumplimientoTargetPct.Value()
		cumplimientoTargetPct = &cp
	}

	var metricaObjetivo *int
	if o.MetricaObjetivo != nil {
		v := o.MetricaObjetivo.Value()
		metricaObjetivo = &v
	}

	var metricaActual *int
	if o.MetricaActual != nil {
		v := o.MetricaActual.Value()
		metricaActual = &v
	}

	var frecuenciaN *int
	if o.FrecuenciaN != nil {
		v := o.FrecuenciaN.Value()
		frecuenciaN = &v
	}

	return Objetivo{
		ID:                    o.ID.String(),
		LifeVisionID:          o.LifeVisionID.String(),
		Titulo:                o.Titulo.String(),
		TipoTracking:          o.TipoTracking.String(),
		Status:                o.Status.String(),
		MetricaObjetivo:       metricaObjetivo,
		MetricaActual:         metricaActual,
		FrecuenciaTipo:        frecuenciaTipo,
		FrecuenciaN:           frecuenciaN,
		CumplimientoTargetPct: cumplimientoTargetPct,
		ArchivedAt:            archivedAt,
		DateCreated:           o.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated:           o.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppObjetivos(objetivos []objetivobus.Objetivo) []Objetivo {
	app := make([]Objetivo, len(objetivos))
	for i, o := range objetivos {
		app[i] = toAppObjetivo(o)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewObjetivo(no NewObjetivo, userID uuid.UUID) (objetivobus.NewObjetivo, error) {
	var errors errs.FieldErrors

	// Parse life vision ID
	lifeVisionID, err := uuid.Parse(no.LifeVisionID)
	if err != nil {
		errors.Add("lifeVisionId", err)
	}

	// Parse titulo
	titulo, err := objetivotitle.Parse(no.Titulo)
	if err != nil {
		errors.Add("titulo", err)
	}

	// Parse tipo_tracking
	tipoTracking, err := trackingtype.Parse(no.TipoTracking)
	if err != nil {
		errors.Add("tipoTracking", err)
	}

	// Parse frecuencia_tipo if provided
	var frecuenciaTipo *frecuenciatype.FrecuenciaType
	if no.FrecuenciaTipo != nil {
		ft, err := frecuenciatype.Parse(*no.FrecuenciaTipo)
		if err != nil {
			errors.Add("frecuenciaTipo", err)
		} else {
			frecuenciaTipo = &ft
		}
	}

	// Parse cumplimiento_target_pct if provided
	var cumplimientoTargetPct *cumplimientopct.CumplimientoPct
	if no.CumplimientoTargetPct != nil {
		cp, err := cumplimientopct.Parse(*no.CumplimientoTargetPct)
		if err != nil {
			errors.Add("cumplimientoTargetPct", err)
		} else {
			cumplimientoTargetPct = &cp
		}
	}

	// Parse metrica_objetivo if provided
	var metricaObjetivoPtr *metricaobjetivo.MetricaObjetivo
	if no.MetricaObjetivo != nil {
		mo, err := metricaobjetivo.Parse(*no.MetricaObjetivo)
		if err != nil {
			errors.Add("metricaObjetivo", err)
		} else {
			metricaObjetivoPtr = &mo
		}
	}

	// Parse frecuencia_n if provided
	var frecuenciaNPtr *frecuencian.FrecuenciaN
	if no.FrecuenciaN != nil {
		fn, err := frecuencian.Parse(*no.FrecuenciaN)
		if err != nil {
			errors.Add("frecuenciaN", err)
		} else {
			frecuenciaNPtr = &fn
		}
	}

	if len(errors) > 0 {
		return objetivobus.NewObjetivo{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return objetivobus.NewObjetivo{
		UserID:                userID,
		LifeVisionID:          lifeVisionID,
		Titulo:                titulo,
		TipoTracking:          tipoTracking,
		MetricaObjetivo:       metricaObjetivoPtr,
		FrecuenciaTipo:        frecuenciaTipo,
		FrecuenciaN:           frecuenciaNPtr,
		CumplimientoTargetPct: cumplimientoTargetPct,
	}, nil
}

func toBusUpdateObjetivo(uo UpdateObjetivo) (objetivobus.UpdateObjetivo, error) {
	var errors errs.FieldErrors
	var bus objetivobus.UpdateObjetivo

	if uo.Titulo != nil {
		titulo, err := objetivotitle.Parse(*uo.Titulo)
		if err != nil {
			errors.Add("titulo", err)
		} else {
			bus.Titulo = &titulo
		}
	}

	if uo.MetricaObjetivo != nil {
		mo, err := metricaobjetivo.Parse(*uo.MetricaObjetivo)
		if err != nil {
			errors.Add("metricaObjetivo", err)
		} else {
			bus.MetricaObjetivo = &mo
		}
	}

	if uo.FrecuenciaTipo != nil {
		ft, err := frecuenciatype.Parse(*uo.FrecuenciaTipo)
		if err != nil {
			errors.Add("frecuenciaTipo", err)
		} else {
			bus.FrecuenciaTipo = &ft
		}
	}

	if uo.FrecuenciaN != nil {
		fn, err := frecuencian.Parse(*uo.FrecuenciaN)
		if err != nil {
			errors.Add("frecuenciaN", err)
		} else {
			bus.FrecuenciaN = &fn
		}
	}

	if uo.CumplimientoTargetPct != nil {
		cp, err := cumplimientopct.Parse(*uo.CumplimientoTargetPct)
		if err != nil {
			errors.Add("cumplimientoTargetPct", err)
		} else {
			bus.CumplimientoTargetPct = &cp
		}
	}

	if len(errors) > 0 {
		return objetivobus.UpdateObjetivo{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}

func toBusChangeStatusRequest(csr ChangeStatusRequest) (objetivobus.ChangeStatusRequest, error) {
	status, err := objetivostatus.Parse(csr.Status)
	if err != nil {
		return objetivobus.ChangeStatusRequest{}, fmt.Errorf("validate: status: %w", err)
	}

	return objetivobus.ChangeStatusRequest{
		Status: status,
	}, nil
}

func toBusUpdateProgressRequest(upr UpdateProgressRequest) objetivobus.UpdateProgressRequest {
	return objetivobus.UpdateProgressRequest{
		Increment: upr.Increment,
		Value:     upr.Value,
	}
}
