package vexportdb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/vexportbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages the set of APIs for export view database access.
type Store struct {
	log       *logger.Logger
	db        sqlx.ExtContext
	encryptor encrypt.Encryptor
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {
	return &Store{
		log:       log,
		db:        db,
		encryptor: encryptor,
	}
}

// Query retrieves a list of export items from the database view.
func (s *Store) Query(ctx context.Context, filter vexportbus.QueryFilter, orderBy order.By, pg page.Page) ([]vexportbus.ExportItem, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		item_id, user_id, item_type, item_date,
		situation, thoughts, physical_symptoms, behavior,
		consequences, values_reflection, intensity,
		category, content, date_created
	FROM
		view_export_items`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dbItems []exportItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbItems); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusExportItemsDecrypted(dbItems, s.encryptor)
}

// Count returns the total number of export items matching the filter.
func (s *Store) Count(ctx context.Context, filter vexportbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		count(1) AS count
	FROM
		view_export_items`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}
