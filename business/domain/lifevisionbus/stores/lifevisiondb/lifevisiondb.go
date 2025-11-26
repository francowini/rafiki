package lifevisiondb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/lifevisionbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for life visions.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (lifevisionbus.Storer, error) {
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

// Create inserts a new life vision into the database.
func (s *Store) Create(ctx context.Context, lv lifevisionbus.LifeVision) error {
	const q = `
	INSERT INTO life_visions (
		life_vision_id, user_id, value_id, content,
		date_created, date_updated
	) VALUES (
		:life_vision_id, :user_id, :value_id, :content,
		:date_created, :date_updated
	)`

	dbLifeVision, err := toDBLifeVisionEncrypted(lv, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbLifeVision); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing life vision in the database.
func (s *Store) Update(ctx context.Context, lv lifevisionbus.LifeVision) error {
	const q = `
	UPDATE life_visions SET
		content = :content,
		value_id = :value_id,
		date_updated = :date_updated
	WHERE
		life_vision_id = :life_vision_id`

	dbLifeVision, err := toDBLifeVisionEncrypted(lv, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbLifeVision); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a life vision from the database.
func (s *Store) Delete(ctx context.Context, lv lifevisionbus.LifeVision) error {
	const q = `
	DELETE FROM life_visions
	WHERE life_vision_id = :life_vision_id`

	data := struct {
		ID string `db:"life_vision_id"`
	}{
		ID: lv.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves life visions based on filter criteria.
func (s *Store) Query(ctx context.Context, filter lifevisionbus.QueryFilter, orderBy order.By, page page.Page) ([]lifevisionbus.LifeVision, error) {
	// Security guard: Require UserID to prevent accidental full-table queries across all users.
	// Life visions are personal and should always be scoped to a specific user.
	if filter.UserID == nil {
		return nil, lifevisionbus.ErrMissingUserID
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
		life_vision_id, user_id, value_id, content,
		date_created, date_updated
	FROM life_visions
	%s
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)

	var dbLifeVisions []lifeVision
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbLifeVisions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusLifeVisionsDecrypted(dbLifeVisions, s.encryptor)
}

// QueryByID finds a life vision by its ID.
func (s *Store) QueryByID(ctx context.Context, lifeVisionID uuid.UUID) (lifevisionbus.LifeVision, error) {
	const q = `
	SELECT
		life_vision_id, user_id, value_id, content,
		date_created, date_updated
	FROM life_visions
	WHERE life_vision_id = :life_vision_id`

	data := struct {
		ID string `db:"life_vision_id"`
	}{
		ID: lifeVisionID.String(),
	}

	var dbLifeVision lifeVision
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbLifeVision); err != nil {
		return lifevisionbus.LifeVision{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusLifeVisionDecrypted(dbLifeVision, s.encryptor)
}

// Count returns the total number of life visions matching the filter.
func (s *Store) Count(ctx context.Context, filter lifevisionbus.QueryFilter) (int, error) {
	// Security guard: Require UserID to prevent accidental full-table counts across all users.
	// Life visions are personal and should always be scoped to a specific user.
	if filter.UserID == nil {
		return 0, lifevisionbus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM life_visions
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
func buildWhereClause(filter lifevisionbus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["life_vision_id"] = *filter.ID
		conditions = append(conditions, "life_vision_id = :life_vision_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.ValueID != nil {
		data["value_id"] = *filter.ValueID
		conditions = append(conditions, "value_id = :value_id")
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
