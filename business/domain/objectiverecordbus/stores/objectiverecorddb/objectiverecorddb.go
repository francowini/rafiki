package objectiverecorddb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/objectiverecordbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for objective records.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (objectiverecordbus.Storer, error) {
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

// Create inserts a new objective record or updates if one exists for the same date (upsert).
func (s *Store) Create(ctx context.Context, rec objectiverecordbus.ObjectiveRecord) error {
	const q = `
	INSERT INTO objective_records (
		objective_record_id, objective_id, user_id,
		record_date, status, notes,
		date_created
	) VALUES (
		:objective_record_id, :objective_id, :user_id,
		:record_date, :status, :notes,
		:date_created
	)
	ON CONFLICT (objective_id, record_date)
	DO UPDATE SET
		status = EXCLUDED.status,
		notes = EXCLUDED.notes`

	dbRecord, err := toDBRecordEncrypted(rec, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbRecord); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing objective record in the database.
func (s *Store) Update(ctx context.Context, rec objectiverecordbus.ObjectiveRecord) error {
	const q = `
	UPDATE objective_records SET
		status = :status,
		notes = :notes
	WHERE
		objective_record_id = :objective_record_id`

	dbRecord, err := toDBRecordEncrypted(rec, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbRecord); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an objective record from the database.
func (s *Store) Delete(ctx context.Context, rec objectiverecordbus.ObjectiveRecord) error {
	const q = `
	DELETE FROM objective_records
	WHERE objective_record_id = :objective_record_id`

	data := struct {
		ID string `db:"objective_record_id"`
	}{
		ID: rec.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves objective records based on filter criteria.
func (s *Store) Query(ctx context.Context, filter objectiverecordbus.QueryFilter, orderBy order.By, page page.Page) ([]objectiverecordbus.ObjectiveRecord, error) {
	// Security guard: Require UserID or ObjectiveID to prevent full-table queries.
	if filter.UserID == nil && filter.ObjectiveID == nil {
		return nil, objectiverecordbus.ErrMissingUserID
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
		objective_record_id, objective_id, user_id,
		record_date, status, notes,
		date_created
	FROM objective_records
	%s
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)

	var dbRecords []objectiveRecord
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbRecords); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRecordsDecrypted(dbRecords, s.encryptor)
}

// QueryByID finds an objective record by its ID.
func (s *Store) QueryByID(ctx context.Context, recordID uuid.UUID) (objectiverecordbus.ObjectiveRecord, error) {
	const q = `
	SELECT
		objective_record_id, objective_id, user_id,
		record_date, status, notes,
		date_created
	FROM objective_records
	WHERE objective_record_id = :objective_record_id`

	data := struct {
		ID string `db:"objective_record_id"`
	}{
		ID: recordID.String(),
	}

	var dbRecord objectiveRecord
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRecord); err != nil {
		return objectiverecordbus.ObjectiveRecord{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusRecordDecrypted(dbRecord, s.encryptor)
}

// QueryByObjectiveAndDate finds an objective record by objective ID and date.
func (s *Store) QueryByObjectiveAndDate(ctx context.Context, objectiveID uuid.UUID, recordDate time.Time) (objectiverecordbus.ObjectiveRecord, error) {
	const q = `
	SELECT
		objective_record_id, objective_id, user_id,
		record_date, status, notes,
		date_created
	FROM objective_records
	WHERE objective_id = :objective_id AND record_date = :record_date`

	data := struct {
		ObjectiveID string    `db:"objective_id"`
		RecordDate  time.Time `db:"record_date"`
	}{
		ObjectiveID: objectiveID.String(),
		RecordDate:  time.Date(recordDate.Year(), recordDate.Month(), recordDate.Day(), 0, 0, 0, 0, time.UTC),
	}

	var dbRecord objectiveRecord
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRecord); err != nil {
		return objectiverecordbus.ObjectiveRecord{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusRecordDecrypted(dbRecord, s.encryptor)
}

// Count returns the total number of objective records matching the filter.
func (s *Store) Count(ctx context.Context, filter objectiverecordbus.QueryFilter) (int, error) {
	// For counting, we allow filtering by ObjectiveID alone (for compliance calculation)
	if filter.UserID == nil && filter.ObjectiveID == nil {
		return 0, objectiverecordbus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM objective_records
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
func buildWhereClause(filter objectiverecordbus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["objective_record_id"] = *filter.ID
		conditions = append(conditions, "objective_record_id = :objective_record_id")
	}

	if filter.ObjectiveID != nil {
		data["objective_id"] = *filter.ObjectiveID
		conditions = append(conditions, "objective_id = :objective_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.StartDate != nil {
		data["start_date"] = *filter.StartDate
		conditions = append(conditions, "record_date >= :start_date")
	}

	if filter.EndDate != nil {
		data["end_date"] = *filter.EndDate
		conditions = append(conditions, "record_date <= :end_date")
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
