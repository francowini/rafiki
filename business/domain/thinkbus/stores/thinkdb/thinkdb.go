package thinkdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for think database access
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new think into the database
func (s *Store) Create(ctx context.Context, think thinkbus.Think) error {
	const q = `
	INSERT INTO thinks
		(think_id, category, content, date_created, date_updated)
	VALUES
		(:think_id, :category, :content, :date_created, :date_updated)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBThink(think)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of all thinks from the database with pagination
func (s *Store) Query(ctx context.Context, orderBy order.By, page page.Page) ([]thinkbus.Think, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
	SELECT
		think_id, category, content, date_created, date_updated
	FROM
		thinks
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, orderByClause)

	var dbThinks []think
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbThinks); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusThinks(dbThinks)
}

// Count returns the total number of thinks
func (s *Store) Count(ctx context.Context) (int, error) {
	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		thinks`

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.QueryStruct(ctx, s.log, s.db, q, &count); err != nil {
		return 0, fmt.Errorf("querystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single think by its ID
func (s *Store) QueryByID(ctx context.Context, thinkID uuid.UUID) (thinkbus.Think, error) {
	data := struct {
		ID string `db:"think_id"`
	}{
		ID: thinkID.String(),
	}

	const q = `
	SELECT
		think_id, category, content, date_created, date_updated
	FROM
		thinks
	WHERE
		think_id = :think_id`

	var dbThink think
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbThink); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return thinkbus.Think{}, fmt.Errorf("db: %w", thinkbus.ErrNotFound)
		}
		return thinkbus.Think{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusThink(dbThink)
}
