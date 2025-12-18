// Package rolebus provides business logic for life roles management.
package rolebus

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/business/types/roledescription"
	"github.com/francowini/rafiki/business/types/roledisplayorder"
	"github.com/francowini/rafiki/business/types/rolefacet"
	"github.com/francowini/rafiki/business/types/rolename"
	"github.com/francowini/rafiki/business/types/roletype"
)

// MaxRolesPerUser is the maximum number of roles allowed per user.
const MaxRolesPerUser = 15

// Domain errors
var (
	ErrNotFound             = errors.New("role not found")
	ErrMaxRoles             = fmt.Errorf("maximum %d roles allowed per user", MaxRolesPerUser)
	ErrDuplicateOrder       = errors.New("display order already exists for this user")
	ErrMissingUserID        = errors.New("userID is required for querying roles")
	ErrUserDisabled         = errors.New("user disabled")
	ErrAlreadyArchived      = errors.New("role is already archived")
	ErrNotArchived          = errors.New("role is not archived")
	ErrValueNotFound        = errors.New("value not found")
	ErrValueArchived        = errors.New("cannot link archived value")
	ErrValueNotOwned        = errors.New("value not owned by user")
	ErrDuplicateValueLink   = errors.New("value already linked")
	ErrTargetValueNotActive = errors.New("cannot restore role: target value is not active")
	ErrInvalidRoleType      = errors.New("invalid role type")
	ErrInvalidFacet         = errors.New("invalid facet")
)

// Role represents a life role with facet categorization.
type Role struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         rolename.RoleName
	RoleType     roletype.RoleType
	Facet        rolefacet.RoleFacet
	Description  roledescription.RoleDescription
	DisplayOrder roledisplayorder.RoleDisplayOrder
	Status       entitystatus.Status
	ArchivedAt   *time.Time
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewRole contains information needed to create a new role.
type NewRole struct {
	UserID       uuid.UUID
	Name         rolename.RoleName
	RoleType     roletype.RoleType
	Facet        rolefacet.RoleFacet
	Description  roledescription.RoleDescription
	DisplayOrder roledisplayorder.RoleDisplayOrder
	ValueIDs     []uuid.UUID // Initial values to connect
}

// UpdateRole contains information needed to update a role.
// All fields are optional (use pointers for partial updates).
type UpdateRole struct {
	Name         *rolename.RoleName
	RoleType     *roletype.RoleType
	Facet        *rolefacet.RoleFacet
	Description  *roledescription.RoleDescription
	DisplayOrder *roledisplayorder.RoleDisplayOrder
}

// ReorderItem represents a role ID with its new display order.
type ReorderItem struct {
	ID           uuid.UUID
	DisplayOrder roledisplayorder.RoleDisplayOrder
}

// ReorderRequest contains items to be reordered.
type ReorderRequest struct {
	Items []ReorderItem
}

// RoleValue represents a connection between a role and a value.
type RoleValue struct {
	RoleID      uuid.UUID
	ValueID     uuid.UUID
	DateCreated time.Time
}
