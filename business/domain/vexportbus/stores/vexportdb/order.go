package vexportdb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	vexportbus.OrderByItemID:      "item_id",
	vexportbus.OrderByItemType:    "item_type",
	vexportbus.OrderByItemDate:    "item_date",
	vexportbus.OrderByDateCreated: "date_created",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
