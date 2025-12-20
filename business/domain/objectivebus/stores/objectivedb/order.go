package objectivedb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	objectivebus.OrderByID:          "objective_id",
	objectivebus.OrderByDateCreated: "date_created",
	objectivebus.OrderByDateUpdated: "date_updated",
	objectivebus.OrderByTitle:       "title",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
