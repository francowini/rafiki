package objetivoregistrobus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default order for queries.
var DefaultOrderBy = order.NewBy(OrderByFechaRegistro, order.DESC)

// Order field names for objetivo records.
const (
	OrderByID            = "objetivo_record_id"
	OrderByFechaRegistro = "fecha_registro"
	OrderByDateCreated   = "date_created"
)
