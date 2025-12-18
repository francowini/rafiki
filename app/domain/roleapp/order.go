package roleapp

import "github.com/francowini/rafiki/business/domain/rolebus"

var orderByFields = map[string]string{
	"role_id":       rolebus.OrderByRoleID,
	"display_order": rolebus.OrderByDisplayOrder,
	"role_type":     rolebus.OrderByRoleType,
	"facet":         rolebus.OrderByFacet,
	"date_created":  rolebus.OrderByDateCreated,
	"date_updated":  rolebus.OrderByDateUpdated,
}
