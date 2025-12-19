package objetivoregistrobus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// registerDelegateFunctions will register action functions with the delegate
// system. If the business was constructed for query only, there won't be a
// delegate provided.
func (b *Business) registerDelegateFunctions() {
	if b.delegate != nil {
		b.delegate.Register(objetivobus.DomainName, objetivobus.ActionDeleted, b.actionObjetivoDeleted)
	}
}

// actionObjetivoDeleted is executed by the objetivo domain indirectly when an objetivo is deleted.
// Note: The actual deletion is handled by database CASCADE constraint.
// This handler is kept for logging and potential future business logic.
func (b *Business) actionObjetivoDeleted(ctx context.Context, data delegate.Data) error {
	var params objetivobus.ActionDeletedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-objetivodeleted", "objetivo_id", params.ObjetivoID, "status", "objetivo records deleted via CASCADE")

	return nil
}
