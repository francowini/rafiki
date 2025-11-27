package vexportbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// NOTE: Unlike vproductbus, UserID is REQUIRED (not pointer) for security.
// Export data must always be scoped to the authenticated user.
type QueryFilter struct {
	UserID    uuid.UUID  // REQUIRED - always set for user isolation
	StartDate *time.Time // Filter items >= this date (inclusive)
	EndDate   *time.Time // Filter items <= this date (inclusive)
}
