package thinkapp

import "github.com/francowini/rafiki/business/domain/thinkbus"

// orderByFields maps API field names to business layer constants
var orderByFields = map[string]string{
	"think_id":     thinkbus.OrderByID,
	"category":     thinkbus.OrderByCategory,
	"date_created": thinkbus.OrderByDateCreated,
	"date_updated": thinkbus.OrderByDateUpdated,
}
