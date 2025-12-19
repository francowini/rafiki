package objetivobus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// Domain identifiers for the delegate system.
const (
	DomainName    = "objetivo"
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	ObjetivoID   uuid.UUID
	LifeVisionID uuid.UUID
	UserID       uuid.UUID
}

// ActionDeletedData constructs the data payload for the deleted action.
func ActionDeletedData(o Objetivo) delegate.Data {
	params := ActionDeletedParms{
		ObjetivoID:   o.ID,
		LifeVisionID: o.LifeVisionID,
		UserID:       o.UserID,
	}

	rawParams, _ := json.Marshal(params) //nolint:errcheck // ActionDeletedParms is a simple struct that cannot fail to marshal

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}

// registerDelegateFunctions will register action functions with the delegate
// system. If the business was constructed for query only, there won't be a
// delegate provided.
func (b *Business) registerDelegateFunctions() {
	if b.delegate != nil {
		b.delegate.Register(lifevisionbus.DomainName, lifevisionbus.ActionDeleted, b.actionLifeVisionDeleted)
	}
}

// actionLifeVisionDeleted is executed by the life vision domain indirectly when a life vision is deleted.
// Note: The actual deletion is handled by database CASCADE constraint.
// This handler is kept for logging and potential future business logic.
func (b *Business) actionLifeVisionDeleted(ctx context.Context, data delegate.Data) error {
	var params lifevisionbus.ActionDeletedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-lifevisiondeleted", "life_vision_id", params.LifeVisionID, "status", "objetivos deleted via CASCADE")

	return nil
}
