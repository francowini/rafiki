package thinkapp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/types/content"
)

// Think represents an API think
type Think struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Content     string `json:"content"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (t Think) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
}

// NewThink contains information needed to create a new think
type NewThink struct {
	Category string `json:"category" validate:"required"`
	Content  string `json:"content" validate:"required"`
}

// Decode implements the decoder interface.
func (nt *NewThink) Decode(data []byte) error {
	return json.Unmarshal(data, nt)
}

// Validate checks the data in the model is considered clean.
func (nt NewThink) Validate() error {
	if err := errs.Check(nt); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

// =============================================================================

func toAppThink(think thinkbus.Think) Think {
	return Think{
		ID:          think.ID.String(),
		Category:    think.Category.String(),
		Content:     think.Content.String(),
		DateCreated: think.DateCreated.Format(time.RFC3339),
		DateUpdated: think.DateUpdated.Format(time.RFC3339),
	}
}

func toAppThinks(thinks []thinkbus.Think) []Think {
	app := make([]Think, len(thinks))
	for i, think := range thinks {
		app[i] = toAppThink(think)
	}
	return app
}

// =============================================================================

func toBusNewThink(ctx context.Context, nt NewThink) (thinkbus.NewThink, error) {
	var errors errs.FieldErrors

	// Extract userID from JWT claims via context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		errors.Add("userID", err)
	}

	category, err := thinkbus.ParseCategory(nt.Category)
	if err != nil {
		errors.Add("category", err)
	}

	cnt, err := content.Parse(nt.Content)
	if err != nil {
		errors.Add("content", err)
	}

	if len(errors) > 0 {
		return thinkbus.NewThink{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return thinkbus.NewThink{
		UserID:   userID,
		Category: category,
		Content:  cnt,
	}, nil
}
