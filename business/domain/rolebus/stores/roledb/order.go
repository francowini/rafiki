package roledb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	rolebus.OrderByRoleID:       "role_id",
	rolebus.OrderByDisplayOrder: "display_order",
	rolebus.OrderByRoleType:     "role_type",
	rolebus.OrderByFacet:        "facet",
	rolebus.OrderByDateCreated:  "date_created",
	rolebus.OrderByDateUpdated:  "date_updated",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
