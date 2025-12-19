package objetivoregistroapp

import "github.com/francowini/rafiki/business/domain/objetivoregistrobus"

var orderByFields = map[string]string{
	"objetivo_record_id": objetivoregistrobus.OrderByID,
	"fecha_registro":     objetivoregistrobus.OrderByFechaRegistro,
	"date_created":       objetivoregistrobus.OrderByDateCreated,
}
