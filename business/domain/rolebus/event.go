package rolebus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// Domain identifiers for the delegate system.
const (
	DomainName    = "role"
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	RoleID uuid.UUID
}

// ActionDeletedData constructs the data payload for the deleted action.
func ActionDeletedData(roleID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		RoleID: roleID,
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
		// Listen for user deletions (roles are deleted via CASCADE, just log)
		b.delegate.Register(userbus.DomainName, userbus.ActionDeleted, b.actionUserDeleted)

		// Listen for value deletions (remove value from all role connections)
		b.delegate.Register(valuebus.DomainName, valuebus.ActionDeleted, b.actionValueDeleted)
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

	b.log.Info(ctx, "rolebus.delegate.user_deleted", "user_id", params.UserID, "status", "roles deleted via CASCADE")

	return nil
}

// actionValueDeleted is executed by the value domain indirectly when a value is deleted.
// This removes the value from all role connections in the junction table.
func (b *Business) actionValueDeleted(ctx context.Context, data delegate.Data) error {
	b.log.Info(ctx, "rolebus.delegate.value_deleted", "rawParams", string(data.RawParams))

	var params valuebus.ActionDeletedParms
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		b.log.Error(ctx, "rolebus.delegate.value_deleted.unmarshal", "err", err)
		return fmt.Errorf("json.unmarshal: %w", err)
	}

	// Remove value from all junction records
	if err := b.storer.RemoveValueFromAllRoles(ctx, params.ValueID); err != nil {
		b.log.Error(ctx, "rolebus.delegate.value_deleted.storer", "err", err, "valueID", params.ValueID)
		return fmt.Errorf("storer.removevaluefromallroles: %w", err)
	}

	b.log.Info(ctx, "rolebus.delegate.value_deleted.success", "valueID", params.ValueID)
	return nil
}
