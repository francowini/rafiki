package vexportapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/app/sdk/query"
	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/foundation/web"
)

// Ordering fields mapping for API
var orderByFields = map[string]string{
	"item_date":    vexportbus.OrderByItemDate,
	"item_type":    vexportbus.OrderByItemType,
	"date_created": vexportbus.OrderByDateCreated,
}

type app struct {
	vexportBus vexportbus.ExtBusiness
}

func newApp(vexportBus vexportbus.ExtBusiness) *app {
	return &app{
		vexportBus: vexportBus,
	}
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Parse pagination (default: page 1, 1000 rows for exports)
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page/rows", err)
	}

	// Parse ordering (default: item_date DESC)
	orderBy, err := order.Parse(orderByFields, qp.OrderBy, vexportbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("orderBy", err)
	}

	// Parse date filters
	var startDate, endDate *time.Time

	if qp.StartDate != "" {
		parsed, err := time.Parse(time.RFC3339, qp.StartDate)
		if err != nil {
			return errs.NewFieldErrors("start_date", fmt.Errorf("must be RFC3339 format"))
		}
		startDate = &parsed
	}

	if qp.EndDate != "" {
		parsed, err := time.Parse(time.RFC3339, qp.EndDate)
		if err != nil {
			return errs.NewFieldErrors("end_date", fmt.Errorf("must be RFC3339 format"))
		}
		endDate = &parsed
	}

	filter := vexportbus.QueryFilter{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Execute query
	items, err := a.vexportBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		if errors.Is(err, vexportbus.ErrInvalidDateRange) {
			return errs.NewFieldErrors("date_range", err)
		}
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	// Get total count
	total, err := a.vexportBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppExportItems(items), total, pg)
}

type queryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	StartDate string
	EndDate   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	return queryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}
