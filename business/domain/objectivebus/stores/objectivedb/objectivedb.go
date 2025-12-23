package objectivedb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/objectivebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for objectives.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (objectivebus.Storer, error) {
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

// Create inserts a new objective into the database.
func (s *Store) Create(ctx context.Context, obj objectivebus.Objective) error {
	const q = `
	INSERT INTO objectives (
		objective_id, user_id, life_vision_id, title,
		tracking_type, status,
		target_metric, current_metric, begin_metric,
		frequency_type, frequency_count, compliance_target_pct,
		archived_at,
		date_created, date_updated
	) VALUES (
		:objective_id, :user_id, :life_vision_id, :title,
		:tracking_type, :status,
		:target_metric, :current_metric, :begin_metric,
		:frequency_type, :frequency_count, :compliance_target_pct,
		:archived_at,
		:date_created, :date_updated
	)`

	dbObjective, err := toDBObjectiveEncrypted(obj, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbObjective); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing objective in the database.
func (s *Store) Update(ctx context.Context, obj objectivebus.Objective) error {
	const q = `
	UPDATE objectives SET
		title = :title,
		status = :status,
		target_metric = :target_metric,
		current_metric = :current_metric,
		begin_metric = :begin_metric,
		frequency_type = :frequency_type,
		frequency_count = :frequency_count,
		compliance_target_pct = :compliance_target_pct,
		archived_at = :archived_at,
		date_updated = :date_updated
	WHERE
		objective_id = :objective_id`

	dbObjective, err := toDBObjectiveEncrypted(obj, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbObjective); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an objective from the database.
func (s *Store) Delete(ctx context.Context, obj objectivebus.Objective) error {
	const q = `
	DELETE FROM objectives
	WHERE objective_id = :objective_id`

	data := struct {
		ID string `db:"objective_id"`
	}{
		ID: obj.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves objectives based on filter criteria.
func (s *Store) Query(ctx context.Context, filter objectivebus.QueryFilter, orderBy order.By, page page.Page) ([]objectivebus.Objective, error) {
	// Security guard: Require UserID to prevent accidental full-table queries across all users.
	if filter.UserID == nil {
		return nil, objectivebus.ErrMissingUserID
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
		objective_id, user_id, life_vision_id, title,
		tracking_type, status,
		target_metric, current_metric, begin_metric,
		frequency_type, frequency_count, compliance_target_pct,
		archived_at,
		date_created, date_updated
	FROM objectives
	%s
	%s
	LIMIT :rows_per_page OFFSET :offset`, whereClause, orderByClause)

	var dbObjectives []objective
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbObjectives); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusObjectivesDecrypted(dbObjectives, s.encryptor)
}

// QueryByID finds an objective by its ID.
func (s *Store) QueryByID(ctx context.Context, objectiveID uuid.UUID) (objectivebus.Objective, error) {
	const q = `
	SELECT
		objective_id, user_id, life_vision_id, title,
		tracking_type, status,
		target_metric, current_metric, begin_metric,
		frequency_type, frequency_count, compliance_target_pct,
		archived_at,
		date_created, date_updated
	FROM objectives
	WHERE objective_id = :objective_id`

	data := struct {
		ID string `db:"objective_id"`
	}{
		ID: objectiveID.String(),
	}

	var dbObjective objective
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbObjective); err != nil {
		return objectivebus.Objective{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusObjectiveDecrypted(dbObjective, s.encryptor)
}

// Count returns the total number of objectives matching the filter.
func (s *Store) Count(ctx context.Context, filter objectivebus.QueryFilter) (int, error) {
	// Security guard: Require UserID to prevent accidental full-table counts across all users.
	if filter.UserID == nil {
		return 0, objectivebus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM objectives
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
func buildWhereClause(filter objectivebus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["objective_id"] = *filter.ID
		conditions = append(conditions, "objective_id = :objective_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.LifeVisionID != nil {
		data["life_vision_id"] = *filter.LifeVisionID
		conditions = append(conditions, "life_vision_id = :life_vision_id")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		conditions = append(conditions, "status = :status")
	}

	if filter.TrackingType != nil {
		data["tracking_type"] = filter.TrackingType.String()
		conditions = append(conditions, "tracking_type = :tracking_type")
	}

	// Handle archived filtering
	if !filter.IncludeArchived {
		conditions = append(conditions, "archived_at IS NULL")
	}

	if len(conditions) == 0 {
		return ""
	}

	return " WHERE " + strings.Join(conditions, " AND ")
}
