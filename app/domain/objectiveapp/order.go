package objectiveapp

import "github.com/francowini/rafiki/business/domain/objectivebus"

var orderByFields = map[string]string{
	"objective_id": objectivebus.OrderByID,
	"date_created": objectivebus.OrderByDateCreated,
	"date_updated": objectivebus.OrderByDateUpdated,
	"title":        objectivebus.OrderByTitle,
}
