package valuebus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default ordering for values.
var DefaultOrderBy = order.NewBy(OrderByDisplayOrder, order.ASC)

// Order field constants (short codes for API).
const (
	OrderByValueID      = "value_id"
	OrderByDisplayOrder = "display_order"
	OrderByFacet        = "facet"
	OrderByDateCreated  = "date_created"
	OrderByDateUpdated  = "date_updated"
)
