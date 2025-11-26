package lifevisionbus

import "github.com/google/uuid"

// QueryFilter holds the available fields to filter life visions.
type QueryFilter struct {
	ID      *uuid.UUID
	UserID  *uuid.UUID
	ValueID *uuid.UUID
}
