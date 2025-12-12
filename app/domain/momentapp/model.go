package momentapp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

// Moment represents an API moment.
type Moment struct {
	ID               string `json:"id"`
	MomentDate       string `json:"momentDate"`
	Situation        string `json:"situation"`
	Thoughts         string `json:"thoughts"`
	PhysicalSymptoms string `json:"physicalSymptoms"`
	Behavior         string `json:"behavior"`
	Consequences     string `json:"consequences"`
	ValuesReflection string `json:"valuesReflection"`
	Intensity        int    `json:"intensity"`
	DateCreated      string `json:"dateCreated"`
	DateUpdated      string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (m Moment) Encode() ([]byte, string, error) {
	data, err := json.Marshal(m)
	return data, "application/json", err
}

// NewMoment contains information needed to create a new moment.
type NewMoment struct {
	MomentDate       string `json:"momentDate" validate:"required"`
	Situation        string `json:"situation" validate:"required"`
	Thoughts         string `json:"thoughts" validate:"required"`
	PhysicalSymptoms string `json:"physicalSymptoms" validate:"required"`
	Behavior         string `json:"behavior" validate:"required"`
	Consequences     string `json:"consequences" validate:"required"`
	ValuesReflection string `json:"valuesReflection" validate:"required"`
	Intensity        int    `json:"intensity" validate:"required,min=0,max=10"`
}

// Decode implements the decoder interface.
func (nm *NewMoment) Decode(data []byte) error {
	return json.Unmarshal(data, nm)
}

// Validate checks the data in the model is considered clean.
func (nm NewMoment) Validate() error {
	if err := errs.Check(nm); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateMoment contains information needed to update a moment.
type UpdateMoment struct {
	MomentDate       *string `json:"momentDate"`
	Situation        *string `json:"situation"`
	Thoughts         *string `json:"thoughts"`
	PhysicalSymptoms *string `json:"physicalSymptoms"`
	Behavior         *string `json:"behavior"`
	Consequences     *string `json:"consequences"`
	ValuesReflection *string `json:"valuesReflection"`
	Intensity        *int    `json:"intensity" validate:"omitempty,min=0,max=10"`
}

// Decode implements the decoder interface.
func (um *UpdateMoment) Decode(data []byte) error {
	return json.Unmarshal(data, um)
}

// Validate checks the data in the model is considered clean.
func (um UpdateMoment) Validate() error {
	if err := errs.Check(um); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// AppStats represents moment statistics for the API response.
type AppStats struct {
	ThisWeek   int `json:"thisWeek"`
	ThisMonth  int `json:"thisMonth"`
	Last30Days int `json:"last30Days"`
}

// Encode implements the encoder interface.
func (s AppStats) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}

// =============================================================================

func toAppMoment(moment momentbus.Moment) Moment {
	return Moment{
		ID:               moment.ID.String(),
		MomentDate:       moment.MomentDate.Format(time.RFC3339),
		Situation:        moment.Situation.String(),
		Thoughts:         moment.Thoughts.String(),
		PhysicalSymptoms: moment.PhysicalSymptoms.String(),
		Behavior:         moment.Behavior.String(),
		Consequences:     moment.Consequences.String(),
		ValuesReflection: moment.ValuesReflection.String(),
		Intensity:        moment.Intensity.Value(),
		DateCreated:      moment.DateCreated.Format(time.RFC3339),
		DateUpdated:      moment.DateUpdated.Format(time.RFC3339),
	}
}

func toAppMoments(moments []momentbus.Moment) []Moment {
	app := make([]Moment, len(moments))
	for i, moment := range moments {
		app[i] = toAppMoment(moment)
	}
	return app
}

// =============================================================================

func toBusNewMoment(ctx context.Context, nm NewMoment) (momentbus.NewMoment, error) {
	var errors errs.FieldErrors

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		errors.Add("userID", err)
	}

	momentDate, err := time.Parse(time.RFC3339, nm.MomentDate)
	if err != nil {
		errors.Add("momentDate", fmt.Errorf("must be RFC3339 format"))
	}

	situation, err := content.Parse(nm.Situation)
	if err != nil {
		errors.Add("situation", err)
	}

	thoughts, err := content.Parse(nm.Thoughts)
	if err != nil {
		errors.Add("thoughts", err)
	}

	physicalSymptoms, err := content.Parse(nm.PhysicalSymptoms)
	if err != nil {
		errors.Add("physicalSymptoms", err)
	}

	behavior, err := content.Parse(nm.Behavior)
	if err != nil {
		errors.Add("behavior", err)
	}

	consequences, err := content.Parse(nm.Consequences)
	if err != nil {
		errors.Add("consequences", err)
	}

	valuesReflection, err := content.Parse(nm.ValuesReflection)
	if err != nil {
		errors.Add("valuesReflection", err)
	}

	intensityVal, err := intensity.Parse(nm.Intensity)
	if err != nil {
		errors.Add("intensity", err)
	}

	if len(errors) > 0 {
		return momentbus.NewMoment{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return momentbus.NewMoment{
		UserID:           userID,
		MomentDate:       momentDate,
		Situation:        situation,
		Thoughts:         thoughts,
		PhysicalSymptoms: physicalSymptoms,
		Behavior:         behavior,
		Consequences:     consequences,
		ValuesReflection: valuesReflection,
		Intensity:        intensityVal,
	}, nil
}

//nolint:unparam // ctx kept for API consistency with other toBus* functions
func toBusUpdateMoment(ctx context.Context, um UpdateMoment) (momentbus.UpdateMoment, error) {
	var errors errs.FieldErrors
	var bus momentbus.UpdateMoment

	if um.MomentDate != nil {
		momentDate, err := time.Parse(time.RFC3339, *um.MomentDate)
		if err != nil {
			errors.Add("momentDate", fmt.Errorf("must be RFC3339 format"))
		} else {
			bus.MomentDate = &momentDate
		}
	}

	if um.Situation != nil {
		situation, err := content.Parse(*um.Situation)
		if err != nil {
			errors.Add("situation", err)
		} else {
			bus.Situation = &situation
		}
	}

	if um.Thoughts != nil {
		thoughts, err := content.Parse(*um.Thoughts)
		if err != nil {
			errors.Add("thoughts", err)
		} else {
			bus.Thoughts = &thoughts
		}
	}

	if um.PhysicalSymptoms != nil {
		physicalSymptoms, err := content.Parse(*um.PhysicalSymptoms)
		if err != nil {
			errors.Add("physicalSymptoms", err)
		} else {
			bus.PhysicalSymptoms = &physicalSymptoms
		}
	}

	if um.Behavior != nil {
		behavior, err := content.Parse(*um.Behavior)
		if err != nil {
			errors.Add("behavior", err)
		} else {
			bus.Behavior = &behavior
		}
	}

	if um.Consequences != nil {
		consequences, err := content.Parse(*um.Consequences)
		if err != nil {
			errors.Add("consequences", err)
		} else {
			bus.Consequences = &consequences
		}
	}

	if um.ValuesReflection != nil {
		valuesReflection, err := content.Parse(*um.ValuesReflection)
		if err != nil {
			errors.Add("valuesReflection", err)
		} else {
			bus.ValuesReflection = &valuesReflection
		}
	}

	if um.Intensity != nil {
		intensityVal, err := intensity.Parse(*um.Intensity)
		if err != nil {
			errors.Add("intensity", err)
		} else {
			bus.Intensity = &intensityVal
		}
	}

	if len(errors) > 0 {
		return momentbus.UpdateMoment{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}

func toAppStats(stats momentbus.Stats) AppStats {
	return AppStats{
		ThisWeek:   stats.ThisWeek,
		ThisMonth:  stats.ThisMonth,
		Last30Days: stats.Last30Days,
	}
}
