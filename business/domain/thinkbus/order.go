package thinkbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default ordering
var DefaultOrderBy = order.NewBy(OrderByDateCreated, order.DESC)

// Set of fields for ordering (short constants for indirection)
const (
	OrderByID          = "a"
	OrderByCategory    = "b"
	OrderByDateCreated = "c"
	OrderByDateUpdated = "d"
)
