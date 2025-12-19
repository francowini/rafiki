package objetivodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/objetivobus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for objetivos.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (objetivobus.Storer, error) {
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

// Create inserts a new objetivo into the database.
func (s *Store) Create(ctx context.Context, obj objetivobus.Objetivo) error {
	const q = `
	INSERT INTO objetivos (
		objetivo_id, user_id, life_vision_id, titulo,
		tipo_tracking, status,
		metrica_objetivo, metrica_actual,
		frecuencia_tipo, frecuencia_n, cumplimiento_target_pct,
		archived_at,
		date_created, date_updated
	) VALUES (
		:objetivo_id, :user_id, :life_vision_id, :titulo,
		:tipo_tracking, :status,
		:metrica_objetivo, :metrica_actual,
		:frecuencia_tipo, :frecuencia_n, :cumplimiento_target_pct,
		:archived_at,
		:date_created, :date_updated
	)`

	dbObjetivo, err := toDBObjetivoEncrypted(obj, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbObjetivo); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing objetivo in the database.
func (s *Store) Update(ctx context.Context, obj objetivobus.Objetivo) error {
	const q = `
	UPDATE objetivos SET
		titulo = :titulo,
		status = :status,
		metrica_objetivo = :metrica_objetivo,
		metrica_actual = :metrica_actual,
		frecuencia_tipo = :frecuencia_tipo,
		frecuencia_n = :frecuencia_n,
		cumplimiento_target_pct = :cumplimiento_target_pct,
		archived_at = :archived_at,
		date_updated = :date_updated
	WHERE
		objetivo_id = :objetivo_id`

	dbObjetivo, err := toDBObjetivoEncrypted(obj, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbObjetivo); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an objetivo from the database.
func (s *Store) Delete(ctx context.Context, obj objetivobus.Objetivo) error {
	const q = `
	DELETE FROM objetivos
	WHERE objetivo_id = :objetivo_id`

	data := struct {
		ID string `db:"objetivo_id"`
	}{
		ID: obj.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves objetivos based on filter criteria.
func (s *Store) Query(ctx context.Context, filter objetivobus.QueryFilter, orderBy order.By, page page.Page) ([]objetivobus.Objetivo, error) {
	// Security guard: Require UserID to prevent accidental full-table queries across all users.
	if filter.UserID == nil {
		return nil, objetivobus.ErrMissingUserID
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
		objetivo_id, user_id, life_vision_id, titulo,
		tipo_tracking, status,
		metrica_objetivo, metrica_actual,
		frecuencia_tipo, frecuencia_n, cumplimiento_target_pct,
		archived_at,
		date_created, date_updated
	FROM objetivos
	%s
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)

	var dbObjetivos []objetivo
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbObjetivos); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusObjetivosDecrypted(dbObjetivos, s.encryptor)
}

// QueryByID finds an objetivo by its ID.
func (s *Store) QueryByID(ctx context.Context, objetivoID uuid.UUID) (objetivobus.Objetivo, error) {
	const q = `
	SELECT
		objetivo_id, user_id, life_vision_id, titulo,
		tipo_tracking, status,
		metrica_objetivo, metrica_actual,
		frecuencia_tipo, frecuencia_n, cumplimiento_target_pct,
		archived_at,
		date_created, date_updated
	FROM objetivos
	WHERE objetivo_id = :objetivo_id`

	data := struct {
		ID string `db:"objetivo_id"`
	}{
		ID: objetivoID.String(),
	}

	var dbObjetivo objetivo
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbObjetivo); err != nil {
		return objetivobus.Objetivo{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusObjetivoDecrypted(dbObjetivo, s.encryptor)
}

// Count returns the total number of objetivos matching the filter.
func (s *Store) Count(ctx context.Context, filter objetivobus.QueryFilter) (int, error) {
	// Security guard: Require UserID to prevent accidental full-table counts across all users.
	if filter.UserID == nil {
		return 0, objetivobus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM objetivos
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
func buildWhereClause(filter objetivobus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["objetivo_id"] = *filter.ID
		conditions = append(conditions, "objetivo_id = :objetivo_id")
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

	if filter.TipoTracking != nil {
		data["tipo_tracking"] = filter.TipoTracking.String()
		conditions = append(conditions, "tipo_tracking = :tipo_tracking")
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
