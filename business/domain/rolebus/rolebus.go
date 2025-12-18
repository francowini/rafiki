package rolebus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/business/types/roledisplayorder"
	"github.com/francowini/rafiki/foundation/logger"
)

// Storer interface defines required database operations.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, role Role) error
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, role Role) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	BatchUpdate(ctx context.Context, roles []Role) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Role, error)
	QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)

	// Junction table operations
	ConnectValue(ctx context.Context, roleID, valueID uuid.UUID) error
	DisconnectValue(ctx context.Context, roleID, valueID uuid.UUID) error
	SetValues(ctx context.Context, roleID uuid.UUID, valueIDs []uuid.UUID) error
	GetValueIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	GetRoleIDsForValue(ctx context.Context, valueID uuid.UUID) ([]uuid.UUID, error)
	RemoveValueFromAllRoles(ctx context.Context, valueID uuid.UUID) error
}

// ExtBusiness interface provides support for role business operations.
type ExtBusiness interface {
	NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error)
	Create(ctx context.Context, nr NewRole) (Role, error)
	Update(ctx context.Context, role Role, ur UpdateRole) (Role, error)
	Delete(ctx context.Context, role Role) error
	Archive(ctx context.Context, role Role) (Role, error)
	Restore(ctx context.Context, role Role) (Role, error)
	Reorder(ctx context.Context, userID uuid.UUID, rr ReorderRequest) ([]Role, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Role, error)
	QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)

	// Value connection operations
	ConnectValue(ctx context.Context, roleID, valueID, authenticatedUserID uuid.UUID) error
	DisconnectValue(ctx context.Context, roleID, valueID, authenticatedUserID uuid.UUID) error
	SetValues(ctx context.Context, roleID uuid.UUID, valueIDs []uuid.UUID, authenticatedUserID uuid.UUID) error
	GetValueIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	GetRoleIDsForValue(ctx context.Context, valueID, authenticatedUserID uuid.UUID) ([]uuid.UUID, error)
}

// Business manages role operations.
type Business struct {
	log      *logger.Logger
	userBus  userbus.ExtBusiness
	valueBus valuebus.ExtBusiness
	delegate *delegate.Delegate
	storer   Storer
	beginner sqldb.Beginner
}

// NewBusiness constructs a Business for role domain.
func NewBusiness(
	log *logger.Logger,
	userBus userbus.ExtBusiness,
	valueBus valuebus.ExtBusiness,
	dlg *delegate.Delegate,
	storer Storer,
	beginner sqldb.Beginner,
) ExtBusiness {
	b := &Business{
		log:      log,
		userBus:  userBus,
		valueBus: valueBus,
		delegate: dlg,
		storer:   storer,
		beginner: beginner,
	}

	b.registerDelegateFunctions()

	return b
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	userBus, err := b.userBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	valueBus, err := b.valueBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	// Create business without re-registering delegate functions
	// to avoid duplicate handler registration on the shared delegate.
	return &Business{
		log:      b.log,
		userBus:  userBus,
		valueBus: valueBus,
		delegate: b.delegate,
		storer:   storer,
		beginner: b.beginner,
	}, nil
}

// Create adds a new role to the system.
func (b *Business) Create(ctx context.Context, nr NewRole) (Role, error) {
	b.log.Info(ctx, "rolebus.create", "userID", nr.UserID, "name", nr.Name)

	// Validate parent user exists and is enabled
	usr, err := b.userBus.QueryByID(ctx, nr.UserID)
	if err != nil {
		b.log.Error(ctx, "rolebus.create.user.querybyid", "err", err)
		return Role{}, fmt.Errorf("user.querybyid: %s: %w", nr.UserID, err)
	}

	if !usr.Enabled {
		b.log.Error(ctx, "rolebus.create.user.disabled", "userID", nr.UserID)
		return Role{}, ErrUserDisabled
	}

	// Validate enum types
	if !nr.RoleType.Valid() {
		b.log.Error(ctx, "rolebus.create.invalid_role_type", "roleType", nr.RoleType)
		return Role{}, ErrInvalidRoleType
	}
	if !nr.Facet.Valid() {
		b.log.Error(ctx, "rolebus.create.invalid_facet", "facet", nr.Facet)
		return Role{}, ErrInvalidFacet
	}

	// Get all roles for this user to check count
	filter := QueryFilter{UserID: &nr.UserID}
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		b.log.Error(ctx, "rolebus.create.count", "err", err)
		return Role{}, fmt.Errorf("storer.count: %w", err)
	}

	if count >= MaxRolesPerUser {
		b.log.Error(ctx, "rolebus.create.max_roles", "count", count)
		return Role{}, ErrMaxRoles
	}

	// Validate initial values ownership and status
	for _, valueID := range nr.ValueIDs {
		value, err := b.valueBus.QueryByID(ctx, valueID)
		if err != nil {
			b.log.Error(ctx, "rolebus.create.value.querybyid", "err", err, "valueID", valueID)
			return Role{}, fmt.Errorf("value.querybyid: %w", err)
		}
		if value.UserID != nr.UserID {
			b.log.Error(ctx, "rolebus.create.value.not_owned", "valueID", valueID, "valueUserID", value.UserID, "roleUserID", nr.UserID)
			return Role{}, ErrValueNotOwned
		}
		if !value.Status.IsActive() {
			b.log.Error(ctx, "rolebus.create.value.archived", "valueID", valueID)
			return Role{}, ErrValueArchived
		}
	}

	now := time.Now().UTC()

	role := Role{
		ID:           uuid.New(),
		UserID:       nr.UserID,
		Name:         nr.Name,
		RoleType:     nr.RoleType,
		Facet:        nr.Facet,
		Description:  nr.Description,
		DisplayOrder: nr.DisplayOrder,
		Status:       entitystatus.Active,
		ArchivedAt:   nil,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := b.storer.Create(ctx, role); err != nil {
		b.log.Error(ctx, "rolebus.create.storer.create", "err", err)
		return Role{}, fmt.Errorf("storer.create: %w", err)
	}

	// Connect initial values
	for _, valueID := range nr.ValueIDs {
		if err := b.storer.ConnectValue(ctx, role.ID, valueID); err != nil {
			b.log.Error(ctx, "rolebus.create.storer.connectvalue", "err", err, "valueID", valueID)
			return Role{}, fmt.Errorf("storer.connectvalue: %w", err)
		}
	}

	b.log.Info(ctx, "rolebus.create.success", "roleID", role.ID)
	return role, nil
}

// Update modifies an existing role.
func (b *Business) Update(ctx context.Context, role Role, ur UpdateRole) (Role, error) {
	b.log.Info(ctx, "rolebus.update", "roleID", role.ID)

	if ur.Name != nil {
		role.Name = *ur.Name
	}

	if ur.RoleType != nil {
		if !ur.RoleType.Valid() {
			b.log.Error(ctx, "rolebus.update.invalid_role_type", "roleType", *ur.RoleType)
			return Role{}, ErrInvalidRoleType
		}
		role.RoleType = *ur.RoleType
	}

	if ur.Facet != nil {
		if !ur.Facet.Valid() {
			b.log.Error(ctx, "rolebus.update.invalid_facet", "facet", *ur.Facet)
			return Role{}, ErrInvalidFacet
		}
		role.Facet = *ur.Facet
	}

	if ur.Description != nil {
		role.Description = *ur.Description
	}

	if ur.DisplayOrder != nil {
		role.DisplayOrder = *ur.DisplayOrder
	}

	role.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, role); err != nil {
		b.log.Error(ctx, "rolebus.update.storer.update", "err", err)
		return Role{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "rolebus.update.success", "roleID", role.ID)
	return role, nil
}

// Delete removes a role from the system.
func (b *Business) Delete(ctx context.Context, role Role) error {
	b.log.Info(ctx, "rolebus.delete", "roleID", role.ID)

	if err := b.storer.Delete(ctx, role); err != nil {
		b.log.Error(ctx, "rolebus.delete.storer.delete", "err", err)
		return fmt.Errorf("delete: %w", err)
	}

	// Publish event for any domains that might care about role deletion
	if b.delegate != nil {
		if err := b.delegate.Call(ctx, ActionDeletedData(role.ID)); err != nil {
			b.log.Error(ctx, "rolebus.delete.delegate.call", "err", err)
			return fmt.Errorf("delegate call: %w", err)
		}
	}

	b.log.Info(ctx, "rolebus.delete.success", "roleID", role.ID)
	return nil
}

// Archive sets a role's status to archived.
func (b *Business) Archive(ctx context.Context, role Role) (Role, error) {
	b.log.Info(ctx, "rolebus.archive", "roleID", role.ID)

	if role.Status.IsArchived() {
		b.log.Error(ctx, "rolebus.archive.already_archived", "roleID", role.ID)
		return Role{}, ErrAlreadyArchived
	}

	now := time.Now().UTC()
	role.Status = entitystatus.Archived
	role.ArchivedAt = &now
	role.DateUpdated = now

	if err := b.storer.Update(ctx, role); err != nil {
		b.log.Error(ctx, "rolebus.archive.storer.update", "err", err)
		return Role{}, fmt.Errorf("update: %w", err)
	}

	b.log.Info(ctx, "rolebus.archive.success", "roleID", role.ID)
	return role, nil
}

// Restore sets an archived role's status back to active.
// Returns ErrTargetValueNotActive if any connected value is archived.
func (b *Business) Restore(ctx context.Context, role Role) (Role, error) {
	b.log.Info(ctx, "rolebus.restore", "roleID", role.ID)

	if !role.Status.IsArchived() {
		b.log.Error(ctx, "rolebus.restore.not_archived", "roleID", role.ID, "status", role.Status)
		return Role{}, ErrNotArchived
	}

	// Get all connected values and validate they are still active
	valueIDs, err := b.storer.GetValueIDs(ctx, role.ID)
	if err != nil {
		b.log.Error(ctx, "rolebus.restore.storer.getvalueids", "err", err)
		return Role{}, fmt.Errorf("storer.getvalueids: %w", err)
	}

	for _, valueID := range valueIDs {
		value, err := b.valueBus.QueryByID(ctx, valueID)
		if err != nil {
			b.log.Error(ctx, "rolebus.restore.value.querybyid", "err", err, "valueID", valueID)
			return Role{}, fmt.Errorf("value.querybyid: valueID[%s]: %w", valueID, err)
		}
		if !value.Status.IsActive() {
			b.log.Error(ctx, "rolebus.restore.value.not_active", "valueID", valueID, "status", value.Status)
			return Role{}, ErrTargetValueNotActive
		}
	}

	role.Status = entitystatus.Active
	role.ArchivedAt = nil
	role.DateUpdated = time.Now().UTC()

	if err := b.storer.Update(ctx, role); err != nil {
		b.log.Error(ctx, "rolebus.restore.storer.update", "err", err)
		return Role{}, fmt.Errorf("storer.update: %w", err)
	}

	b.log.Info(ctx, "rolebus.restore.success", "roleID", role.ID)
	return role, nil
}

// Query retrieves a list of roles based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Role, error) {
	roles, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return roles, nil
}

// QueryByID finds a role by its ID.
func (b *Business) QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error) {
	role, err := b.storer.QueryByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Role{}, ErrNotFound
		}
		return Role{}, fmt.Errorf("query: %w", err)
	}

	return role, nil
}

// Count returns the total number of roles matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// Reorder atomically updates display order for multiple roles within a transaction.
func (b *Business) Reorder(ctx context.Context, userID uuid.UUID, rr ReorderRequest) ([]Role, error) {
	b.log.Info(ctx, "rolebus.reorder", "userID", userID, "items", len(rr.Items))

	// Validate parent user exists and is enabled
	usr, err := b.userBus.QueryByID(ctx, userID)
	if err != nil {
		b.log.Error(ctx, "rolebus.reorder.user.querybyid", "err", err)
		return nil, fmt.Errorf("user.querybyid: %s: %w", userID, err)
	}

	if !usr.Enabled {
		b.log.Error(ctx, "rolebus.reorder.user.disabled", "userID", userID)
		return nil, ErrUserDisabled
	}

	// Get current roles for the user
	filter := QueryFilter{UserID: &userID}
	orderBy := order.By{Field: OrderByDisplayOrder, Direction: order.ASC}
	pg := page.MustParse("1", fmt.Sprintf("%d", MaxRolesPerUser))

	currentRoles, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		b.log.Error(ctx, "rolebus.reorder.query", "err", err)
		return nil, fmt.Errorf("query current roles: %w", err)
	}

	// Build update map from request and validate no duplicate display_orders in request
	updateMap := make(map[string]roledisplayorder.RoleDisplayOrder)
	seenOrders := make(map[int]string)
	for _, item := range rr.Items {
		orderVal := item.DisplayOrder.Value()
		if existingID, exists := seenOrders[orderVal]; exists && existingID != item.ID.String() {
			b.log.Error(ctx, "rolebus.reorder.duplicate_order", "order", orderVal)
			return nil, ErrDuplicateOrder
		}
		seenOrders[orderVal] = item.ID.String()
		updateMap[item.ID.String()] = item.DisplayOrder
	}

	// Prepare roles to update
	now := time.Now().UTC()
	rolesToUpdate := make([]Role, 0, len(currentRoles))
	for _, current := range currentRoles {
		if newOrder, found := updateMap[current.ID.String()]; found {
			current.DisplayOrder = newOrder
			current.DateUpdated = now
			rolesToUpdate = append(rolesToUpdate, current)
		}
	}

	if len(rolesToUpdate) == 0 {
		return currentRoles, nil
	}

	// Begin transaction for atomic batch update
	tx, err := b.beginner.Begin()
	if err != nil {
		b.log.Error(ctx, "rolebus.reorder.begin", "err", err)
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			//nolint:errcheck // Rollback error is irrelevant when we're already in an error path
			tx.Rollback()
		}
	}()

	txStorer, err := b.storer.NewWithTx(tx)
	if err != nil {
		b.log.Error(ctx, "rolebus.reorder.newwithtx", "err", err)
		return nil, fmt.Errorf("new tx storer: %w", err)
	}

	if err := txStorer.BatchUpdate(ctx, rolesToUpdate); err != nil {
		b.log.Error(ctx, "rolebus.reorder.batchupdate", "err", err)
		return nil, fmt.Errorf("batch update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		b.log.Error(ctx, "rolebus.reorder.commit", "err", err)
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	committed = true

	// Return updated roles
	updatedRoles, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		b.log.Error(ctx, "rolebus.reorder.query_updated", "err", err)
		return nil, fmt.Errorf("query updated roles: %w", err)
	}

	b.log.Info(ctx, "rolebus.reorder.success", "updated", len(rolesToUpdate))
	return updatedRoles, nil
}

// =============================================================================
// Value Connection Operations
// =============================================================================

// ConnectValue creates a connection between a role and a value.
func (b *Business) ConnectValue(ctx context.Context, roleID, valueID, authenticatedUserID uuid.UUID) error {
	b.log.Info(ctx, "rolebus.connectvalue", "roleID", roleID, "valueID", valueID)

	// Validate role exists and is owned by authenticated user
	role, err := b.QueryByID(ctx, roleID)
	if err != nil {
		b.log.Error(ctx, "rolebus.connectvalue.role.querybyid", "err", err)
		return fmt.Errorf("role.querybyid: %w", err)
	}
	if role.UserID != authenticatedUserID {
		b.log.Error(ctx, "rolebus.connectvalue.role.not_owned", "roleUserID", role.UserID, "authUserID", authenticatedUserID)
		return ErrNotFound // Don't leak existence of other users' roles
	}

	// Validate value exists, is active, and is owned by same user
	value, err := b.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		b.log.Error(ctx, "rolebus.connectvalue.value.querybyid", "err", err)
		return fmt.Errorf("value.querybyid: %w", err)
	}
	if value.UserID != authenticatedUserID {
		b.log.Error(ctx, "rolebus.connectvalue.value.not_owned", "valueUserID", value.UserID, "authUserID", authenticatedUserID)
		return ErrValueNotOwned
	}
	if !value.Status.IsActive() {
		b.log.Error(ctx, "rolebus.connectvalue.value.archived", "valueID", valueID)
		return ErrValueArchived
	}

	if err := b.storer.ConnectValue(ctx, roleID, valueID); err != nil {
		if errors.Is(err, ErrDuplicateValueLink) {
			b.log.Info(ctx, "rolebus.connectvalue.duplicate", "roleID", roleID, "valueID", valueID)
			return ErrDuplicateValueLink
		}
		b.log.Error(ctx, "rolebus.connectvalue.storer.connectvalue", "err", err)
		return fmt.Errorf("storer.connectvalue: %w", err)
	}

	b.log.Info(ctx, "rolebus.connectvalue.success", "roleID", roleID, "valueID", valueID)
	return nil
}

// DisconnectValue removes a connection between a role and a value.
func (b *Business) DisconnectValue(ctx context.Context, roleID, valueID, authenticatedUserID uuid.UUID) error {
	b.log.Info(ctx, "rolebus.disconnectvalue", "roleID", roleID, "valueID", valueID)

	// Validate role exists and is owned by authenticated user
	role, err := b.QueryByID(ctx, roleID)
	if err != nil {
		b.log.Error(ctx, "rolebus.disconnectvalue.role.querybyid", "err", err)
		return fmt.Errorf("role.querybyid: %w", err)
	}
	if role.UserID != authenticatedUserID {
		b.log.Error(ctx, "rolebus.disconnectvalue.role.not_owned", "roleUserID", role.UserID, "authUserID", authenticatedUserID)
		return ErrNotFound
	}

	if err := b.storer.DisconnectValue(ctx, roleID, valueID); err != nil {
		b.log.Error(ctx, "rolebus.disconnectvalue.storer.disconnectvalue", "err", err)
		return fmt.Errorf("storer.disconnectvalue: %w", err)
	}

	b.log.Info(ctx, "rolebus.disconnectvalue.success", "roleID", roleID, "valueID", valueID)
	return nil
}

// SetValues replaces all value connections for a role (bulk operation).
func (b *Business) SetValues(ctx context.Context, roleID uuid.UUID, valueIDs []uuid.UUID, authenticatedUserID uuid.UUID) error {
	b.log.Info(ctx, "rolebus.setvalues", "roleID", roleID, "values", len(valueIDs))

	// Validate role exists and is owned by authenticated user
	role, err := b.QueryByID(ctx, roleID)
	if err != nil {
		b.log.Error(ctx, "rolebus.setvalues.role.querybyid", "err", err)
		return fmt.Errorf("role.querybyid: %w", err)
	}
	if role.UserID != authenticatedUserID {
		b.log.Error(ctx, "rolebus.setvalues.role.not_owned", "roleUserID", role.UserID, "authUserID", authenticatedUserID)
		return ErrNotFound
	}

	// Validate all values exist, are active, and are owned by same user
	for _, valueID := range valueIDs {
		value, err := b.valueBus.QueryByID(ctx, valueID)
		if err != nil {
			b.log.Error(ctx, "rolebus.setvalues.value.querybyid", "err", err, "valueID", valueID)
			return fmt.Errorf("value.querybyid: %w", err)
		}
		if value.UserID != authenticatedUserID {
			b.log.Error(ctx, "rolebus.setvalues.value.not_owned", "valueID", valueID, "valueUserID", value.UserID, "authUserID", authenticatedUserID)
			return ErrValueNotOwned
		}
		if !value.Status.IsActive() {
			b.log.Error(ctx, "rolebus.setvalues.value.archived", "valueID", valueID)
			return ErrValueArchived
		}
	}

	if err := b.storer.SetValues(ctx, roleID, valueIDs); err != nil {
		b.log.Error(ctx, "rolebus.setvalues.storer.setvalues", "err", err)
		return fmt.Errorf("storer.setvalues: %w", err)
	}

	b.log.Info(ctx, "rolebus.setvalues.success", "roleID", roleID, "values", len(valueIDs))
	return nil
}

// GetValueIDs returns all value IDs connected to a role.
func (b *Business) GetValueIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	valueIDs, err := b.storer.GetValueIDs(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("storer.getvalueids: %w", err)
	}

	return valueIDs, nil
}

// GetRoleIDsForValue returns all role IDs that have a specific value connected.
func (b *Business) GetRoleIDsForValue(ctx context.Context, valueID, authenticatedUserID uuid.UUID) ([]uuid.UUID, error) {
	// First validate the value is owned by the authenticated user
	value, err := b.valueBus.QueryByID(ctx, valueID)
	if err != nil {
		return nil, fmt.Errorf("value.querybyid: %w", err)
	}
	if value.UserID != authenticatedUserID {
		return nil, ErrValueNotOwned
	}

	roleIDs, err := b.storer.GetRoleIDsForValue(ctx, valueID)
	if err != nil {
		return nil, fmt.Errorf("storer.getroleidsforvalue: %w", err)
	}

	return roleIDs, nil
}
