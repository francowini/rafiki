package lifevisionbus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/entitystatus"
)

// QueryFilter holds the available fields to filter life visions.
type QueryFilter struct {
	ID              *uuid.UUID
	UserID          *uuid.UUID
	ValueID         *uuid.UUID
	Status          *entitystatus.Status
	IncludeArchived bool // if false, only active (default)
}
