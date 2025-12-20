package valuedb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for values.
type Store struct {
	log       *logger.Logger
	db        sqlx.ExtContext
	encryptor encrypt.Encryptor
}

// NewStore constructs a Store for database access.
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {
	return &Store{
		log:       log,
		db:        db,
		encryptor: encryptor,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (valuebus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log:       s.log,
		db:        ec,
		encryptor: s.encryptor,
	}, nil
}

// Create inserts a new value into the database.
func (s *Store) Create(ctx context.Context, value valuebus.Value) error {
	const q = `
	INSERT INTO values (
		value_id, user_id, content, facet, display_order,
		status, archived_at,
		date_created, date_updated
	) VALUES (
		:value_id, :user_id, :content, :facet, :display_order,
		:status, :archived_at,
		:date_created, :date_updated
	)`

	dbValue, err := toDBValueEncrypted(value, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbValue); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", valuebus.ErrDuplicateOrder)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing value in the database.
func (s *Store) Update(ctx context.Context, value valuebus.Value) error {
	const q = `
	UPDATE values SET
		content = :content,
		facet = :facet,
		display_order = :display_order,
		status = :status,
		archived_at = :archived_at,
		date_updated = :date_updated
	WHERE
		value_id = :value_id`

	dbValue, err := toDBValueEncrypted(value, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbValue); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", valuebus.ErrDuplicateOrder)
		}

		// Detect database trigger rejection for archiving with active life visions.
		// The trigger uses ERRCODE '23503' with message containing 'active life visions'.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && strings.Contains(pgErr.Message, "active life visions") {
			return valuebus.ErrHasActiveLifeVisions
		}

		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a value from the database.
func (s *Store) Delete(ctx context.Context, value valuebus.Value) error {
	const q = `
	DELETE FROM values
	WHERE value_id = :value_id`

	data := struct {
		ID string `db:"value_id"`
	}{
		ID: value.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeleteByUserID removes all values for a specific user.
func (s *Store) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	const q = `
	DELETE FROM values
	WHERE user_id = :user_id`

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// BatchUpdate updates multiple values atomically.
// All updates happen within the caller's transaction context.
func (s *Store) BatchUpdate(ctx context.Context, values []valuebus.Value) error {
	if len(values) == 0 {
		return nil
	}

	const q = `
	UPDATE values SET
		display_order = :display_order,
		date_updated = :date_updated
	WHERE
		value_id = :value_id`

	for _, value := range values {
		data := struct {
			ValueID      string    `db:"value_id"`
			DisplayOrder int       `db:"display_order"`
			DateUpdated  time.Time `db:"date_updated"`
		}{
			ValueID:      value.ID.String(),
			DisplayOrder: value.DisplayOrder.Value(),
			DateUpdated:  value.DateUpdated.UTC(),
		}

		if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
			return fmt.Errorf("namedexeccontext: %w", err)
		}
	}

	return nil
}

// Query retrieves values based on filter criteria.
func (s *Store) Query(ctx context.Context, filter valuebus.QueryFilter, orderBy order.By, page page.Page) ([]valuebus.Value, error) {
	// Security guard: Require UserID to prevent accidental full-table queries across all users.
	// Values are personal and should always be scoped to a specific user.
	if filter.UserID == nil {
		return nil, valuebus.ErrMissingUserID
	}

	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	whereClause := buildWhereClause(filter, data)
	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
	SELECT
		value_id, user_id, content, facet, display_order,
		status, archived_at,
		date_created, date_updated
	FROM values
	%s
	%s
	LIMIT :rows_per_page OFFSET :offset`, whereClause, orderByClause)

	var dbValues []value
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbValues); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusValuesDecrypted(dbValues, s.encryptor)
}

// QueryByID finds a value by its ID.
func (s *Store) QueryByID(ctx context.Context, valueID uuid.UUID) (valuebus.Value, error) {
	const q = `
	SELECT
		value_id, user_id, content, facet, display_order,
		status, archived_at,
		date_created, date_updated
	FROM values
	WHERE value_id = :value_id`

	data := struct {
		ID string `db:"value_id"`
	}{
		ID: valueID.String(),
	}

	var dbValue value
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbValue); err != nil {
		return valuebus.Value{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusValueDecrypted(dbValue, s.encryptor)
}

// Count returns the total number of values matching the filter.
func (s *Store) Count(ctx context.Context, filter valuebus.QueryFilter) (int, error) {
	// Security guard: Require UserID to prevent accidental full-table counts across all users.
	// Values are personal and should always be scoped to a specific user.
	if filter.UserID == nil {
		return 0, valuebus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM values
	%s`, whereClause)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// buildWhereClause constructs the WHERE clause dynamically.
func buildWhereClause(filter valuebus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["value_id"] = *filter.ID
		conditions = append(conditions, "value_id = :value_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.Facet != nil {
		data["facet"] = filter.Facet.String()
		conditions = append(conditions, "facet = :facet")
	}

	// Handle status filtering
	if filter.Status != nil {
		data["status"] = filter.Status.String()
		conditions = append(conditions, "status = :status")
	} else if !filter.IncludeArchived {
		// Default: only show active items
		data["status"] = entitystatus.Active.String()
		conditions = append(conditions, "status = :status")
	}

	if len(conditions) == 0 {
		return ""
	}

	whereClause := " WHERE "
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return whereClause
}
