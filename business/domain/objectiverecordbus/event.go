package objectiverecordbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// registerDelegateFunctions will register action functions with the delegate
// system. If the business was constructed for query only, there won't be a
// delegate provided.
func (b *Business) registerDelegateFunctions() {
	if b.delegate != nil {
		b.delegate.Register(objectivebus.DomainName, objectivebus.ActionDeleted, b.actionObjectiveDeleted)
	}
}

// actionObjectiveDeleted is executed by the objective domain indirectly when an objective is deleted.
// Note: The actual deletion is handled by database CASCADE constraint.
// This handler is kept for logging and potential future business logic.
func (b *Business) actionObjectiveDeleted(ctx context.Context, data delegate.Data) error {
	var params objectivebus.ActionDeletedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-objectivedeleted", "objective_id", params.ObjectiveID, "status", "objective records deleted via CASCADE")

	return nil
}
