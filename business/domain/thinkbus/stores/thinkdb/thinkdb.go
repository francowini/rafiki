package thinkdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages the set of APIs for think database access
type Store struct {
	log       *logger.Logger
	db        sqlx.ExtContext
	encryptor encrypt.Encryptor
}

// NewStore constructs the api for data access
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {
	return &Store{
		log:       log,
		db:        db,
		encryptor: encryptor,
	}
}

// Create inserts a new think into the database
func (s *Store) Create(ctx context.Context, think thinkbus.Think) error {
	const q = `
	INSERT INTO thinks
		(think_id, user_id, category, content, date_created, date_updated)
	VALUES
		(:think_id, :user_id, :category, :content, :date_created, :date_updated)`

	dbThink, err := toDBThinkEncrypted(think, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbThink); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of all thinks from the database with pagination
func (s *Store) Query(ctx context.Context, userID uuid.UUID, orderBy order.By, page page.Page) ([]thinkbus.Think, error) {
	data := map[string]any{
		"user_id":       userID,
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
	SELECT
		think_id, user_id, category, content, date_created, date_updated
	FROM
		thinks
	WHERE
		user_id = :user_id
	%s
	LIMIT :rows_per_page OFFSET :offset`, orderByClause)

	var dbThinks []think
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbThinks); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusThinkDecryptedSlice(dbThinks, s.encryptor)
}

// Count returns the total number of thinks
func (s *Store) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	data := map[string]any{
		"user_id": userID,
	}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		thinks
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

// QueryByID retrieves a single think by its ID
func (s *Store) QueryByID(ctx context.Context, thinkID, userID uuid.UUID) (thinkbus.Think, error) {
	data := struct {
		ThinkID string `db:"think_id"`
		UserID  string `db:"user_id"`
	}{
		ThinkID: thinkID.String(),
		UserID:  userID.String(),
	}

	const q = `
	SELECT
		think_id, user_id, category, content, date_created, date_updated
	FROM
		thinks
	WHERE
		think_id = :think_id AND user_id = :user_id`

	var dbThink think
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbThink); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return thinkbus.Think{}, fmt.Errorf("db: %w", thinkbus.ErrNotFound)
		}
		return thinkbus.Think{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusThinkDecrypted(dbThink, s.encryptor)
}
