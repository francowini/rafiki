package valuedb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	valuebus.OrderByValueID:      "value_id",
	valuebus.OrderByDisplayOrder: "display_order",
	valuebus.OrderByFacet:        "facet",
	valuebus.OrderByDateCreated:  "date_created",
	valuebus.OrderByDateUpdated:  "date_updated",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
