package lifevisionbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// Domain identifiers for the delegate system.
const (
	DomainName    = "lifevision"
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	LifeVisionID uuid.UUID
	ValueID      uuid.UUID
	UserID       uuid.UUID
}

// ActionDeletedData constructs the data payload for the deleted action.
func ActionDeletedData(lv LifeVision) delegate.Data {
	params := ActionDeletedParms{
		LifeVisionID: lv.ID,
		ValueID:      lv.ValueID,
		UserID:       lv.UserID,
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
		b.delegate.Register(valuebus.DomainName, valuebus.ActionDeleted, b.actionValueDeleted)
	}
}

// actionValueDeleted is executed by the value domain indirectly when a value is deleted.
// Note: The actual deletion is handled by database CASCADE constraint.
// This handler is kept for logging and potential future business logic.
func (b *Business) actionValueDeleted(ctx context.Context, data delegate.Data) error {
	var params valuebus.ActionDeletedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-valuedeleted", "value_id", params.ValueID, "status", "life visions deleted via CASCADE")

	return nil
}
