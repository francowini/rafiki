package valuebus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// Domain identifiers for the delegate system.
const (
	DomainName    = "value"
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	ValueID uuid.UUID
}

// ActionDeletedData constructs the data payload for the deleted action.
func ActionDeletedData(valueID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		ValueID: valueID,
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
		b.delegate.Register(userbus.DomainName, userbus.ActionDeleted, b.actionUserDeleted)
	}
}

// actionUserDeleted is executed by the user domain indirectly when a user is deleted.
// Note: The actual deletion is handled by database CASCADE constraint.
// This handler is kept for logging and potential future business logic.
func (b *Business) actionUserDeleted(ctx context.Context, data delegate.Data) error {
	var params userbus.ActionDeletedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-userdeleted", "user_id", params.UserID, "status", "values deleted via CASCADE")

	return nil
}
