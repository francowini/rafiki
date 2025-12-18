package roledb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for roles.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (rolebus.Storer, error) {
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

// Create inserts a new role into the database.
func (s *Store) Create(ctx context.Context, role rolebus.Role) error {
	const q = `
	INSERT INTO roles (
		role_id, user_id, name, role_type, facet, description,
		display_order, status, archived_at,
		date_created, date_updated
	) VALUES (
		:role_id, :user_id, :name, :role_type, :facet, :description,
		:display_order, :status, :archived_at,
		:date_created, :date_updated
	)`

	dbRole, err := toDBRoleEncrypted(role, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbRole); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", rolebus.ErrDuplicateOrder)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing role in the database.
func (s *Store) Update(ctx context.Context, role rolebus.Role) error {
	const q = `
	UPDATE roles SET
		name = :name,
		role_type = :role_type,
		facet = :facet,
		description = :description,
		display_order = :display_order,
		status = :status,
		archived_at = :archived_at,
		date_updated = :date_updated
	WHERE
		role_id = :role_id`

	dbRole, err := toDBRoleEncrypted(role, s.encryptor)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbRole); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", rolebus.ErrDuplicateOrder)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a role from the database.
func (s *Store) Delete(ctx context.Context, role rolebus.Role) error {
	const q = `
	DELETE FROM roles
	WHERE role_id = :role_id`

	data := struct {
		ID string `db:"role_id"`
	}{
		ID: role.ID.String(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeleteByUserID removes all roles for a specific user.
func (s *Store) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	const q = `
	DELETE FROM roles
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

// BatchUpdate updates multiple roles atomically.
// All updates happen within the caller's transaction context.
func (s *Store) BatchUpdate(ctx context.Context, roles []rolebus.Role) error {
	if len(roles) == 0 {
		return nil
	}

	const q = `
	UPDATE roles SET
		display_order = :display_order,
		date_updated = :date_updated
	WHERE
		role_id = :role_id`

	for _, role := range roles {
		data := struct {
			RoleID       string    `db:"role_id"`
			DisplayOrder int       `db:"display_order"`
			DateUpdated  time.Time `db:"date_updated"`
		}{
			RoleID:       role.ID.String(),
			DisplayOrder: role.DisplayOrder.Value(),
			DateUpdated:  role.DateUpdated.UTC(),
		}

		if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
			return fmt.Errorf("namedexeccontext: %w", err)
		}
	}

	return nil
}

// Query retrieves roles based on filter criteria.
func (s *Store) Query(ctx context.Context, filter rolebus.QueryFilter, orderBy order.By, page page.Page) ([]rolebus.Role, error) {
	// Security guard: Require UserID to prevent accidental full-table queries across all users.
	// Roles are personal and should always be scoped to a specific user.
	if filter.UserID == nil {
		return nil, rolebus.ErrMissingUserID
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
		role_id, user_id, name, role_type, facet, description,
		display_order, status, archived_at,
		date_created, date_updated
	FROM roles
	%s
	%s
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`, whereClause, orderByClause)

	var dbRoles []role
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbRoles); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRolesDecrypted(dbRoles, s.encryptor)
}

// QueryByID finds a role by its ID.
func (s *Store) QueryByID(ctx context.Context, roleID uuid.UUID) (rolebus.Role, error) {
	const q = `
	SELECT
		role_id, user_id, name, role_type, facet, description,
		display_order, status, archived_at,
		date_created, date_updated
	FROM roles
	WHERE role_id = :role_id`

	data := struct {
		ID string `db:"role_id"`
	}{
		ID: roleID.String(),
	}

	var dbRole role
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRole); err != nil {
		return rolebus.Role{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusRoleDecrypted(dbRole, s.encryptor)
}

// Count returns the total number of roles matching the filter.
func (s *Store) Count(ctx context.Context, filter rolebus.QueryFilter) (int, error) {
	// Security guard: Require UserID to prevent accidental full-table counts across all users.
	// Roles are personal and should always be scoped to a specific user.
	if filter.UserID == nil {
		return 0, rolebus.ErrMissingUserID
	}

	data := map[string]any{}
	whereClause := buildWhereClause(filter, data)

	q := fmt.Sprintf(`
	SELECT COUNT(1) AS count
	FROM roles
	%s`, whereClause)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// =============================================================================
// Junction Table Operations (roles_values)
// =============================================================================

// ConnectValue creates a connection between a role and a value.
func (s *Store) ConnectValue(ctx context.Context, roleID, valueID uuid.UUID) error {
	const q = `
	INSERT INTO roles_values (role_id, value_id, date_created)
	VALUES (:role_id, :value_id, :date_created)`

	data := struct {
		RoleID      uuid.UUID `db:"role_id"`
		ValueID     uuid.UUID `db:"value_id"`
		DateCreated time.Time `db:"date_created"`
	}{
		RoleID:      roleID,
		ValueID:     valueID,
		DateCreated: time.Now().UTC(),
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		// Check for duplicate key (already linked)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return rolebus.ErrDuplicateValueLink
		}
		// Check for foreign key violation (value not found)
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if strings.Contains(pgErr.ConstraintName, "value_id") {
				return rolebus.ErrValueNotFound
			}
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DisconnectValue removes a connection between a role and a value.
func (s *Store) DisconnectValue(ctx context.Context, roleID, valueID uuid.UUID) error {
	const q = `
	DELETE FROM roles_values
	WHERE role_id = :role_id AND value_id = :value_id`

	data := struct {
		RoleID  uuid.UUID `db:"role_id"`
		ValueID uuid.UUID `db:"value_id"`
	}{
		RoleID:  roleID,
		ValueID: valueID,
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// SetValues replaces all value connections for a role (bulk operation).
func (s *Store) SetValues(ctx context.Context, roleID uuid.UUID, valueIDs []uuid.UUID) error {
	// First, delete all existing connections
	const deleteQ = `
	DELETE FROM roles_values
	WHERE role_id = :role_id`

	deleteData := struct {
		RoleID uuid.UUID `db:"role_id"`
	}{
		RoleID: roleID,
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, deleteQ, deleteData); err != nil {
		return fmt.Errorf("delete existing: %w", err)
	}

	// Then insert new connections
	if len(valueIDs) == 0 {
		return nil
	}

	const insertQ = `
	INSERT INTO roles_values (role_id, value_id, date_created)
	VALUES (:role_id, :value_id, :date_created)`

	now := time.Now().UTC()
	for _, valueID := range valueIDs {
		data := struct {
			RoleID      uuid.UUID `db:"role_id"`
			ValueID     uuid.UUID `db:"value_id"`
			DateCreated time.Time `db:"date_created"`
		}{
			RoleID:      roleID,
			ValueID:     valueID,
			DateCreated: now,
		}

		if err := sqldb.NamedExecContext(ctx, s.log, s.db, insertQ, data); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				if strings.Contains(pgErr.ConstraintName, "value_id") {
					return rolebus.ErrValueNotFound
				}
			}
			return fmt.Errorf("namedexeccontext: %w", err)
		}
	}

	return nil
}

// GetValueIDs returns all value IDs connected to a role.
func (s *Store) GetValueIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
	SELECT value_id
	FROM roles_values
	WHERE role_id = :role_id
	ORDER BY date_created`

	data := struct {
		RoleID uuid.UUID `db:"role_id"`
	}{
		RoleID: roleID,
	}

	var results []struct {
		ValueID uuid.UUID `db:"value_id"`
	}

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &results); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	valueIDs := make([]uuid.UUID, len(results))
	for i, r := range results {
		valueIDs[i] = r.ValueID
	}

	return valueIDs, nil
}

// GetRoleIDsForValue returns all role IDs that have a specific value connected.
func (s *Store) GetRoleIDsForValue(ctx context.Context, valueID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
	SELECT role_id
	FROM roles_values
	WHERE value_id = :value_id
	ORDER BY date_created`

	data := struct {
		ValueID uuid.UUID `db:"value_id"`
	}{
		ValueID: valueID,
	}

	var results []struct {
		RoleID uuid.UUID `db:"role_id"`
	}

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &results); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(results))
	for i, r := range results {
		roleIDs[i] = r.RoleID
	}

	return roleIDs, nil
}

// RemoveValueFromAllRoles removes a value from all role connections.
// This is called when a value is deleted (via delegate).
func (s *Store) RemoveValueFromAllRoles(ctx context.Context, valueID uuid.UUID) error {
	const q = `
	DELETE FROM roles_values
	WHERE value_id = :value_id`

	data := struct {
		ValueID uuid.UUID `db:"value_id"`
	}{
		ValueID: valueID,
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildWhereClause constructs the WHERE clause dynamically.
func buildWhereClause(filter rolebus.QueryFilter, data map[string]any) string {
	var conditions []string

	if filter.ID != nil {
		data["role_id"] = *filter.ID
		conditions = append(conditions, "role_id = :role_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		conditions = append(conditions, "user_id = :user_id")
	}

	if filter.RoleType != nil {
		data["role_type"] = filter.RoleType.String()
		conditions = append(conditions, "role_type = :role_type")
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

	return " WHERE " + strings.Join(conditions, " AND ")
}
