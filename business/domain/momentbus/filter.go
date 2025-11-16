package momentbus

import (
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID      *uuid.UUID
	UserID  *uuid.UUID
	Page    page.Page
	OrderBy order.By
}
