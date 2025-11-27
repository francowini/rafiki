package vexportdb

import (
	"bytes"
	"strings"

	"github.com/francowini/rafiki/business/domain/vexportbus"
)

func (s *Store) applyFilter(filter vexportbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	// UserID is always required (security: user data isolation)
	data["user_id"] = filter.UserID
	wc = append(wc, "user_id = :user_id")

	if filter.StartDate != nil {
		data["start_date"] = filter.StartDate.UTC()
		wc = append(wc, "item_date >= :start_date")
	}

	if filter.EndDate != nil {
		data["end_date"] = filter.EndDate.UTC()
		wc = append(wc, "item_date <= :end_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
