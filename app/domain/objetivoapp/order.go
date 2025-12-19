package objetivoapp

import "github.com/francowini/rafiki/business/domain/objetivobus"

var orderByFields = map[string]string{
	"objetivo_id":  objetivobus.OrderByID,
	"date_created": objetivobus.OrderByDateCreated,
	"date_updated": objetivobus.OrderByDateUpdated,
	"titulo":       objetivobus.OrderByTitulo,
}
