package momentdb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	momentbus.OrderByMomentID:    "moment_id",
	momentbus.OrderByMomentDate:  "moment_date",
	momentbus.OrderByIntensity:   "intensity",
	momentbus.OrderByDateCreated: "date_created",
	momentbus.OrderByDateUpdated: "date_updated",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
