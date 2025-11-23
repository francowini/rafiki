package valueapp

import "github.com/francowini/rafiki/business/domain/valuebus"

var orderByFields = map[string]string{
	"value_id":      valuebus.OrderByValueID,
	"display_order": valuebus.OrderByDisplayOrder,
	"facet":         valuebus.OrderByFacet,
	"date_created":  valuebus.OrderByDateCreated,
	"date_updated":  valuebus.OrderByDateUpdated,
}
