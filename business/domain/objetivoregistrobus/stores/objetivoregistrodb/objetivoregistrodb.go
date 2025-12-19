package objetivoregistrodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/objetivoregistrobus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for objetivo records.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (objetivoregistrobus.Storer, error) {
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

// Create inserts a new objetivo record or updates if one exists for the same date (upsert).
func (s *Store) Create(ctx context.Context, rec objetivoregistrobus.ObjetivoRecord) error {
	const q = `
	INSERT INTO objetivo_records (
		objetivo_record_id, objetivo_id, user_id,
		fecha_registro, status, notes,
		date_created
	) VALUES (
		:objetivo_record_id, :objetivo_id, :user_id,
		:fecha_registro, :status, :notes,
		:date_created
	)
	ON CONFLICT (objetivo_id, fecha_registro)
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

// Update modifies an existing objetivo record in the database.
func (s *Store) Update(ctx context.Context, rec objetivoregistrobus.ObjetivoRecord) error {
	const q = `
	UPDATE objetivo_records SET
		status = :status,
		notes = :notes
	WHERE
		objetivo_record_id = :objetivo_record_id`

	dbRecord, err := toDBRecordEncrypted(rec, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbRecord); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an objetivo record from the database.
func (s *Store) Delete(ctx context.Context, rec objetivoregistrobus.ObjetivoRecord) error {
	const q = `
	DELETE FROM objetivo_records
	WHERE objetivo_record_id = :objetivo_record_id`

	data := struct {
		ID string `db:"objetivo_record_id"`
	}{
		ID: rec.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves objetivo records based on filter criteria.
func (s *Store) Query(ctx context.Context, filter objetivoregistrobus.QueryFilter, orderBy order.By, page page.Page) ([]objetivoregistrobus.ObjetivoRecord, error) {
	// Security guard: Require UserID or ObjetivoID to prevent full-table queries.
	if filter.UserID == nil && filter.ObjetivoID == nil {
		return nil, objetivoregistrobus.ErrMissingUserID
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
		objetivo_record_id, objetivo_id, user_id,
		fecha_registro, status, notes,
		date_created
	FROM objetivo_records
	%s
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)

	var dbRecords []objetivoRecord
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbRecords); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRecordsDecrypted(dbRecords, s.encryptor)
}

// QueryByID finds an objetivo record by its ID.
func (s *Store) QueryByID(ctx context.Context, recordID uuid.UUID) (objetivoregistrobus.ObjetivoRecord, error) {
	const q = `
	SELECT
		objetivo_record_id, objetivo_id, user_id,
		fecha_registro, status, notes,
		date_created
	FROM objetivo_records
	WHERE objetivo_record_id = :objetivo_record_id`

	data := struct {
		ID string `db:"objetivo_record_id"`
	}{
		ID: recordID.String(),
	}

	var dbRecord objetivoRecord
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRecord); err != nil {
		return objetivoregistrobus.ObjetivoRecord{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusRecordDecrypted(dbRecord, s.encryptor)
}

// QueryByObjetivoAndDate finds an objetivo record by objetivo ID and date.
func (s *Store) QueryByObjetivoAndDate(ctx context.Context, objetivoID uuid.UUID, fechaRegistro time.Time) (objetivoregistrobus.ObjetivoRecord, error) {
	const q = `
	SELECT
		objetivo_record_id, objetivo_id, user_id,
		fecha_registro, status, notes,
		date_created
	FROM objetivo_records
	WHERE objetivo_id = :objetivo_id AND fecha_registro = :fecha_registro`

	data := struct {
		ObjetivoID    string    `db:"objetivo_id"`
		FechaRegistro time.Time `db:"fecha_registro"`
	}{
		ObjetivoID:    objetivoID.String(),
		FechaRegistro: time.Date(fechaRegistro.Year(), fechaRegistro.Month(), fechaRegistro.Day(), 0, 0, 0, 0, time.UTC),
	}

	var dbRecord objetivoRecord
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRecord); err != nil {
		return objetivoregistrobus.ObjetivoRecord{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusRecordDecrypted(dbRecord, s.encryptor)
}

// Count returns the total number of objetivo records matching the filter.
func (s *Store) Count(ctx context.Context, filter objetivoregistrobus.QueryFilter) (int, error) {
	// For counting, we allow filtering by ObjetivoID alone (for compliance calculation)
	if filter.UserID == nil && filter.ObjetivoID == nil {
		return 0, objetivoregistrobus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM objetivo_records
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
func buildWhereClause(filter objetivoregistrobus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["objetivo_record_id"] = *filter.ID
		conditions = append(conditions, "objetivo_record_id = :objetivo_record_id")
	}

	if filter.ObjetivoID != nil {
		data["objetivo_id"] = *filter.ObjetivoID
		conditions = append(conditions, "objetivo_id = :objetivo_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.StartDate != nil {
		data["start_date"] = *filter.StartDate
		conditions = append(conditions, "fecha_registro >= :start_date")
	}

	if filter.EndDate != nil {
		data["end_date"] = *filter.EndDate
		conditions = append(conditions, "fecha_registro <= :end_date")
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
