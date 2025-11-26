package lifevisionapp

import "github.com/francowini/rafiki/business/domain/lifevisionbus"

var orderByFields = map[string]string{
	"life_vision_id": lifevisionbus.OrderByID,
	"date_created":   lifevisionbus.OrderByDateCreated,
	"date_updated":   lifevisionbus.OrderByDateUpdated,
}
