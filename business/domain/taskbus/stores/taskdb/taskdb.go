package taskdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/taskbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for tasks.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (taskbus.Storer, error) {
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

// Create inserts a new task into the database.
func (s *Store) Create(ctx context.Context, tsk taskbus.Task) error {
	const q = `
	INSERT INTO tasks (
		task_id, user_id, objective_id, title, description,
		contribution, status, completed_at,
		date_created, date_updated
	) VALUES (
		:task_id, :user_id, :objective_id, :title, :description,
		:contribution, :status, :completed_at,
		:date_created, :date_updated
	)`

	dbTask, err := toDBTaskEncrypted(tsk, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbTask); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing task in the database.
func (s *Store) Update(ctx context.Context, tsk taskbus.Task) error {
	const q = `
	UPDATE tasks SET
		title = :title,
		description = :description,
		contribution = :contribution,
		status = :status,
		completed_at = :completed_at,
		date_updated = :date_updated
	WHERE
		task_id = :task_id`

	dbTask, err := toDBTaskEncrypted(tsk, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbTask); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a task from the database.
func (s *Store) Delete(ctx context.Context, tsk taskbus.Task) error {
	const q = `
	DELETE FROM tasks
	WHERE task_id = :task_id`

	data := struct {
		ID string `db:"task_id"`
	}{
		ID: tsk.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves tasks based on filter criteria.
func (s *Store) Query(ctx context.Context, filter taskbus.QueryFilter, orderBy order.By, page page.Page) ([]taskbus.Task, error) {
	// Security guard: Require UserID to prevent accidental full-table queries across all users.
	if filter.UserID == nil {
		return nil, taskbus.ErrMissingUserID
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
		task_id, user_id, objective_id, title, description,
		contribution, status, completed_at,
		date_created, date_updated
	FROM tasks
	%s
	%s
	LIMIT :rows_per_page OFFSET :offset`, whereClause, orderByClause)

	var dbTasks []task
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbTasks); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTasksDecrypted(dbTasks, s.encryptor)
}

// QueryByID finds a task by its ID.
func (s *Store) QueryByID(ctx context.Context, taskID uuid.UUID) (taskbus.Task, error) {
	const q = `
	SELECT
		task_id, user_id, objective_id, title, description,
		contribution, status, completed_at,
		date_created, date_updated
	FROM tasks
	WHERE task_id = :task_id`

	data := struct {
		ID string `db:"task_id"`
	}{
		ID: taskID.String(),
	}

	var dbTask task
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbTask); err != nil {
		return taskbus.Task{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusTaskDecrypted(dbTask, s.encryptor)
}

// Count returns the total number of tasks matching the filter.
func (s *Store) Count(ctx context.Context, filter taskbus.QueryFilter) (int, error) {
	// Security guard: Require UserID to prevent accidental full-table counts across all users.
	if filter.UserID == nil {
		return 0, taskbus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM tasks
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
func buildWhereClause(filter taskbus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["task_id"] = *filter.ID
		conditions = append(conditions, "task_id = :task_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	// InboxOnly takes precedence over ObjectiveID to prevent conflicting conditions.
	// When InboxOnly is true, we only look for tasks with no objective.
	if filter.InboxOnly {
		conditions = append(conditions, "objective_id IS NULL")
	} else if filter.ObjectiveID != nil {
		data["objective_id"] = *filter.ObjectiveID
		conditions = append(conditions, "objective_id = :objective_id")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		conditions = append(conditions, "status = :status")
	}

	if len(conditions) == 0 {
		return ""
	}

	return " WHERE " + strings.Join(conditions, " AND ")
}
