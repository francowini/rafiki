package objetivodb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	objetivobus.OrderByID:          "objetivo_id",
	objetivobus.OrderByDateCreated: "date_created",
	objetivobus.OrderByDateUpdated: "date_updated",
	objetivobus.OrderByTitulo:      "titulo",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
