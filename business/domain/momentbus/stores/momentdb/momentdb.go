package momentdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for moment database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create adds a new moment to the database.
func (s *Store) Create(ctx context.Context, moment momentbus.Moment) error {
	const q = `
	INSERT INTO moments (
		moment_id, user_id, moment_date, situation, thoughts,
		physical_symptoms, behavior, consequences, values_reflection,
		intensity, date_created, date_updated
	) VALUES (
		:moment_id, :user_id, :moment_date, :situation, :thoughts,
		:physical_symptoms, :behavior, :consequences, :values_reflection,
		:intensity, :date_created, :date_updated
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMoment(moment)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", momentbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about a moment in the database.
func (s *Store) Update(ctx context.Context, moment momentbus.Moment) error {
	const q = `
	UPDATE moments SET
		moment_date = :moment_date,
		situation = :situation,
		thoughts = :thoughts,
		physical_symptoms = :physical_symptoms,
		behavior = :behavior,
		consequences = :consequences,
		values_reflection = :values_reflection,
		intensity = :intensity,
		date_updated = :date_updated
	WHERE
		moment_id = :moment_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMoment(moment)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", momentbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a moment from the database.
func (s *Store) Delete(ctx context.Context, moment momentbus.Moment) error {
	const q = `
	DELETE FROM moments
	WHERE moment_id = :moment_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMoment(moment)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing moments from the database.
func (s *Store) Query(ctx context.Context, userID uuid.UUID, orderBy order.By, pg page.Page) ([]momentbus.Moment, error) {
	data := map[string]any{
		"user_id":       userID,
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
	SELECT
		moment_id, user_id, moment_date, situation, thoughts,
		physical_symptoms, behavior, consequences, values_reflection,
		intensity, date_created, date_updated
	FROM
		moments
	WHERE
		user_id = :user_id
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, orderByClause)

	var dbMoms []moment
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbMoms); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusMoments(dbMoms)
}

// QueryByID retrieves a single moment by its ID.
func (s *Store) QueryByID(ctx context.Context, momentID uuid.UUID) (momentbus.Moment, error) {
	data := map[string]any{
		"moment_id": momentID,
	}

	const q = `
	SELECT
		moment_id, user_id, moment_date, situation, thoughts,
		physical_symptoms, behavior, consequences, values_reflection,
		intensity, date_created, date_updated
	FROM
		moments
	WHERE
		moment_id = :moment_id`

	var dbMom moment
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbMom); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return momentbus.Moment{}, fmt.Errorf("db: %w", momentbus.ErrNotFound)
		}
		return momentbus.Moment{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusMoment(dbMom)
}

// Count returns the total number of moments in the database.
func (s *Store) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	data := map[string]any{
		"user_id": userID,
	}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		moments
	WHERE
		user_id = :user_id`

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}
