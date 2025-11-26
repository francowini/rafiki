package valuebus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/facet"
)

// QueryFilter holds filtering options for querying values.
type QueryFilter struct {
	ID     *uuid.UUID
	UserID *uuid.UUID
	Facet  *facet.Facet
}
