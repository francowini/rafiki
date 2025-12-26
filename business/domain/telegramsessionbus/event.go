package telegramsessionbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/sdk/delegate"
)

// registerDelegateFunctions registers event handlers for delegate events.
func (b *Business) registerDelegateFunctions(dlg *delegate.Delegate) {
	if dlg != nil {
		dlg.Register(userbus.DomainName, userbus.ActionDeleted, b.actionUserDeleted)
	}
}

// actionUserDeleted handles user deletion by cleaning up their sessions.
// Unmarshal errors are returned to allow job queue retry/alerting.
// Delete failures are logged but do not abort the user deletion flow.
func (b *Business) actionUserDeleted(ctx context.Context, data delegate.Data) error {
	var params userbus.ActionDeletedParms
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		b.log.Error(ctx, "telegramsessionbus.user_deleted.unmarshal",
			"err", err,
		)
		return fmt.Errorf("unmarshal user deleted params: %w", err)
	}

	// Delete all sessions for this user (cascade)
	if err := b.storer.DeleteByUserID(ctx, params.UserID); err != nil {
		b.log.Error(ctx, "telegramsessionbus.user_deleted.delete",
			"user_id", params.UserID,
			"err", err,
		)
		return nil // Log and continue, don't block user deletion
	}

	b.log.Info(ctx, "telegramsessionbus.user_deleted",
		"user_id", params.UserID,
	)

	return nil
}
