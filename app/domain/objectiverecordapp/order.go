package objectiverecordapp

import "github.com/francowini/rafiki/business/domain/objectiverecordbus"

var orderByFields = map[string]string{
	"objective_record_id": objectiverecordbus.OrderByID,
	"record_date":         objectiverecordbus.OrderByRecordDate,
	"date_created":        objectiverecordbus.OrderByDateCreated,
}
