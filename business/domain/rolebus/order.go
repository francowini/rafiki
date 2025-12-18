package rolebus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default ordering for roles.
var DefaultOrderBy = order.NewBy(OrderByDisplayOrder, order.ASC)

// Order field constants (short codes for API).
const (
	OrderByRoleID       = "role_id"
	OrderByDisplayOrder = "display_order"
	OrderByRoleType     = "role_type"
	OrderByFacet        = "facet"
	OrderByDateCreated  = "date_created"
	OrderByDateUpdated  = "date_updated"
)
