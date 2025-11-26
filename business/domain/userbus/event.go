package userbus

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/delegate"
)

// Domain identifiers for the delegate system.
const (
	DomainName    = "user"
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	UserID uuid.UUID
}

// ActionDeletedData constructs the data payload for the deleted action.
func ActionDeletedData(userID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		UserID: userID,
	}

	rawParams, _ := json.Marshal(params)

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}
