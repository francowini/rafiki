package taskbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default order for queries.
var DefaultOrderBy = order.NewBy(OrderByDateCreated, order.DESC)

// Order field names for tasks.
const (
	OrderByID          = "task_id"
	OrderByDateCreated = "date_created"
	OrderByDateUpdated = "date_updated"
	OrderByCompletedAt = "completed_at"
	OrderByTitle       = "title"
)
