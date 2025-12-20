package objectiverecorddb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/objectiverecordbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	objectiverecordbus.OrderByID:          "objective_record_id",
	objectiverecordbus.OrderByRecordDate:  "record_date",
	objectiverecordbus.OrderByDateCreated: "date_created",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
