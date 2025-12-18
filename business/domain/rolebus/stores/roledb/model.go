// Package roledb provides database access for roles.
package roledb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/entitystatus"
	"github.com/francowini/rafiki/business/types/roledescription"
	"github.com/francowini/rafiki/business/types/roledisplayorder"
	"github.com/francowini/rafiki/business/types/rolefacet"
	"github.com/francowini/rafiki/business/types/rolename"
	"github.com/francowini/rafiki/business/types/roletype"
)

// role represents the database model.
type role struct {
	ID           uuid.UUID  `db:"role_id"`
	UserID       uuid.UUID  `db:"user_id"`
	Name         string     `db:"name"` // encrypted
	RoleType     string     `db:"role_type"`
	Facet        string     `db:"facet"`
	Description  string     `db:"description"` // encrypted
	DisplayOrder int        `db:"display_order"`
	Status       string     `db:"status"`
	ArchivedAt   *time.Time `db:"archived_at"`
	DateCreated  time.Time  `db:"date_created"`
	DateUpdated  time.Time  `db:"date_updated"`
}

// toDBRoleEncrypted converts business model to DB model with encryption.
func toDBRoleEncrypted(bus rolebus.Role, enc encrypt.Encryptor) (role, error) {
	// Encrypt name
	name, err := enc.Encrypt(bus.Name.Value())
	if err != nil {
		return role{}, fmt.Errorf("encrypt name: %w", err)
	}

	// Encrypt description
	description, err := enc.Encrypt(bus.Description.Value())
	if err != nil {
		return role{}, fmt.Errorf("encrypt description: %w", err)
	}

	var archivedAt *time.Time
	if bus.ArchivedAt != nil {
		t := bus.ArchivedAt.UTC()
		archivedAt = &t
	}

	return role{
		ID:           bus.ID,
		UserID:       bus.UserID,
		Name:         name,
		RoleType:     bus.RoleType.String(),
		Facet:        bus.Facet.String(),
		Description:  description,
		DisplayOrder: bus.DisplayOrder.Value(),
		Status:       bus.Status.String(),
		ArchivedAt:   archivedAt,
		DateCreated:  bus.DateCreated.UTC(),
		DateUpdated:  bus.DateUpdated.UTC(),
	}, nil
}

// toBusRoleDecrypted converts DB model to business model with decryption.
func toBusRoleDecrypted(db role, enc encrypt.Encryptor) (rolebus.Role, error) {
	// Decrypt and parse name
	nameStr, err := enc.Decrypt(db.Name)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("decrypt name: %w", err)
	}

	name, err := rolename.Parse(nameStr)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("parse name: %w", err)
	}

	// Decrypt and parse description
	descStr, err := enc.Decrypt(db.Description)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("decrypt description: %w", err)
	}

	description, err := roledescription.Parse(descStr)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("parse description: %w", err)
	}

	// Parse role type
	roleTypeVal, err := roletype.Parse(db.RoleType)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("parse role type: %w", err)
	}

	// Parse facet
	facetVal, err := rolefacet.Parse(db.Facet)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("parse facet: %w", err)
	}

	// Parse display order
	displayOrder, err := roledisplayorder.Parse(db.DisplayOrder)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("parse display order: %w", err)
	}

	// Parse status
	status, err := entitystatus.Parse(db.Status)
	if err != nil {
		return rolebus.Role{}, fmt.Errorf("parse status: %w", err)
	}

	var archivedAt *time.Time
	if db.ArchivedAt != nil {
		t := db.ArchivedAt.In(time.Local)
		archivedAt = &t
	}

	return rolebus.Role{
		ID:           db.ID,
		UserID:       db.UserID,
		Name:         name,
		RoleType:     roleTypeVal,
		Facet:        facetVal,
		Description:  description,
		DisplayOrder: displayOrder,
		Status:       status,
		ArchivedAt:   archivedAt,
		DateCreated:  db.DateCreated.In(time.Local),
		DateUpdated:  db.DateUpdated.In(time.Local),
	}, nil
}

// toBusRolesDecrypted converts a slice of DB models to business models.
func toBusRolesDecrypted(dbs []role, enc encrypt.Encryptor) ([]rolebus.Role, error) {
	roles := make([]rolebus.Role, len(dbs))

	for i, db := range dbs {
		var err error
		roles[i], err = toBusRoleDecrypted(db, enc)
		if err != nil {
			return nil, err
		}
	}

	return roles, nil
}
