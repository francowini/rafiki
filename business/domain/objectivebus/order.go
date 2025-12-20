package objectivebus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default order for queries.
var DefaultOrderBy = order.NewBy(OrderByDateCreated, order.DESC)

// Order field names for objectives.
const (
	OrderByID          = "objective_id"
	OrderByDateCreated = "date_created"
	OrderByDateUpdated = "date_updated"
	OrderByTitle       = "title"
)
