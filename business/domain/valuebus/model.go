// Package valuebus provides business logic for personal values management.
package valuebus

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/displayorder"
	"github.com/francowini/rafiki/business/types/facet"
	"github.com/francowini/rafiki/business/types/valuecontent"
)

// MaxValuesPerUser is the maximum number of values allowed per user.
const MaxValuesPerUser = 10

// Domain errors
var (
	ErrNotFound       = errors.New("value not found")
	ErrMaxValues      = fmt.Errorf("maximum %d values allowed per user", MaxValuesPerUser)
	ErrDuplicateOrder = errors.New("display order already exists for this user")
	ErrMissingUserID  = errors.New("userID is required for querying values")
	ErrUserDisabled   = errors.New("user disabled")
)

// Value represents a personal value with facet categorization.
type Value struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Content      valuecontent.ValueContent
	Facet        facet.Facet
	DisplayOrder displayorder.DisplayOrder
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewValue contains information needed to create a new value.
type NewValue struct {
	UserID       uuid.UUID
	Content      valuecontent.ValueContent
	Facet        facet.Facet
	DisplayOrder displayorder.DisplayOrder
}

// UpdateValue contains information needed to update a value.
// All fields are optional (use pointers for partial updates).
type UpdateValue struct {
	Content      *valuecontent.ValueContent
	Facet        *facet.Facet
	DisplayOrder *displayorder.DisplayOrder
}

// ReorderItem represents a value ID with its new display order.
type ReorderItem struct {
	ID           uuid.UUID
	DisplayOrder displayorder.DisplayOrder
}

// ReorderRequest contains items to be reordered.
type ReorderRequest struct {
	Items []ReorderItem
}
