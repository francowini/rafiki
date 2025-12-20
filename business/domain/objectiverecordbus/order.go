package objectiverecordbus

import "github.com/francowini/rafiki/business/sdk/order"

// DefaultOrderBy represents the default order for queries.
var DefaultOrderBy = order.NewBy(OrderByRecordDate, order.DESC)

// Order field names for objective records.
const (
	OrderByID          = "objective_record_id"
	OrderByRecordDate  = "record_date"
	OrderByDateCreated = "date_created"
)
