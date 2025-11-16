package momentapp

import "github.com/francowini/rafiki/business/domain/momentbus"

// orderByFields maps API field names to business layer constants
var orderByFields = map[string]string{
	"moment_id":    momentbus.OrderByMomentID,
	"moment_date":  momentbus.OrderByMomentDate,
	"intensity":    momentbus.OrderByIntensity,
	"date_created": momentbus.OrderByDateCreated,
	"date_updated": momentbus.OrderByDateUpdated,
}
