package taskapp

import "github.com/francowini/rafiki/business/domain/taskbus"

var orderByFields = map[string]string{
	"dateCreated": taskbus.OrderByDateCreated,
	"dateUpdated": taskbus.OrderByDateUpdated,
	"completedAt": taskbus.OrderByCompletedAt,
	"title":       taskbus.OrderByTitle,
}
