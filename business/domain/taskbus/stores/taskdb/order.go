package taskdb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/taskbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	taskbus.OrderByID:          "task_id",
	taskbus.OrderByDateCreated: "date_created",
	taskbus.OrderByDateUpdated: "date_updated",
	taskbus.OrderByCompletedAt: "completed_at",
	taskbus.OrderByTitle:       "title",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
