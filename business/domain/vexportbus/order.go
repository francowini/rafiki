package vexportbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByItemDate, order.DESC)

// Set of fields that the results can be ordered by.
const (
	OrderByItemID      = "item_id"
	OrderByItemType    = "item_type"
	OrderByItemDate    = "item_date"
	OrderByDateCreated = "date_created"
)
