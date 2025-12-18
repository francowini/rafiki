package rolebus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/business/types/rolefacet"
	"github.com/francowini/rafiki/business/types/roletype"
)

// QueryFilter holds filtering options for querying roles.
type QueryFilter struct {
	ID              *uuid.UUID
	UserID          *uuid.UUID
	RoleType        *roletype.RoleType
	Facet           *rolefacet.RoleFacet
	Status          *entitystatus.Status
	IncludeArchived bool // if false, only active (default)
}
