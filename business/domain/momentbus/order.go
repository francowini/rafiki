package momentbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByMomentDate, order.DESC)

// Set of fields that the results can be ordered by.
const (
	OrderByMomentID    = "a"
	OrderByMomentDate  = "b"
	OrderByIntensity   = "c"
	OrderByDateCreated = "d"
	OrderByDateUpdated = "e"
)
