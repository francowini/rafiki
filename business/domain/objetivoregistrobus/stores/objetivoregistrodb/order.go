package objetivoregistrodb

import (
	"fmt"

	"github.com/francowini/rafiki/business/domain/objetivoregistrobus"
	"github.com/francowini/rafiki/business/sdk/order"
)

var orderByFields = map[string]string{
	objetivoregistrobus.OrderByID:            "objetivo_record_id",
	objetivoregistrobus.OrderByFechaRegistro: "fecha_registro",
	objetivoregistrobus.OrderByDateCreated:   "date_created",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
