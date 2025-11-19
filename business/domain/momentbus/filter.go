package momentbus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
)

// QueryFilter holds the available fields a query can be filtered on.
// Fields with nil pointer values are ignored in the query.
type QueryFilter struct {
	// ID filters by moment ID. If nil, all moment IDs are included.
	ID *uuid.UUID

	// UserID filters by user ID. If nil, all users are included.
	// WARNING: Always set UserID to enforce user isolation in multi-tenant scenarios.
	UserID *uuid.UUID

	// Page specifies pagination parameters. Use page.Parse to create valid Page instances.
	// Defaults to page 1 with a default page size if not specified.
	Page page.Page

	// OrderBy specifies the sort order. Use order.NewBy or order.Parse to create valid instances.
	// Defaults to DefaultOrderBy (moment_date DESC) if an empty value is provided.
	OrderBy order.By
}
