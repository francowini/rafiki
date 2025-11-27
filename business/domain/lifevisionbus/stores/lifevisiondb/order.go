package lifevisiondb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	lifevisionbus.OrderByID:          "life_vision_id",
	lifevisionbus.OrderByDateCreated: "date_created",
	lifevisionbus.OrderByDateUpdated: "date_updated",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
