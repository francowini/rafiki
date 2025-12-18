// Package roleapp provides HTTP handlers for roles.
package roleapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/app/sdk/mid"
	"github.com/francowini/rafiki/business/domain/rolebus"
	"github.com/francowini/rafiki/business/types/roledescription"
	"github.com/francowini/rafiki/business/types/roledisplayorder"
	"github.com/francowini/rafiki/business/types/rolefacet"
	"github.com/francowini/rafiki/business/types/rolename"
	"github.com/francowini/rafiki/business/types/roletype"
)

// =============================================================================
// API Response Models
// =============================================================================

// Role represents a role for API responses.
type Role struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	RoleType     string   `json:"roleType"`
	Facet        string   `json:"facet"`
	Description  string   `json:"description"`
	DisplayOrder int      `json:"displayOrder"`
	Status       string   `json:"status"`
	ArchivedAt   *string  `json:"archivedAt"`
	DateCreated  string   `json:"dateCreated"`
	DateUpdated  string   `json:"dateUpdated"`
	ValueIDs     []string `json:"valueIds,omitempty"`
}

// Encode implements the encoder interface.
func (r Role) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// RoleValue represents a role-value connection for API responses.
type RoleValue struct {
	RoleID      string `json:"roleId"`
	ValueID     string `json:"valueId"`
	DateCreated string `json:"dateCreated"`
}

// Encode implements the encoder interface.
func (rv RoleValue) Encode() ([]byte, string, error) {
	data, err := json.Marshal(rv)
	return data, "application/json", err
}

// RoleFacetInfo represents role facet metadata.
type RoleFacetInfo struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// RoleFacetList represents the list of all role facets.
type RoleFacetList struct {
	Facets []RoleFacetInfo `json:"facets"`
}

// Encode implements the encoder interface.
func (rfl RoleFacetList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(rfl)
	return data, "application/json", err
}

// =============================================================================
// API Request Models
// =============================================================================

// NewRole represents data for creating a new role.
type NewRole struct {
	Name         string   `json:"name" validate:"required"`
	RoleType     string   `json:"roleType" validate:"required"`
	Facet        string   `json:"facet" validate:"required"`
	Description  string   `json:"description" validate:"required"`
	DisplayOrder int      `json:"displayOrder" validate:"required,min=1,max=15"`
	ValueIDs     []string `json:"valueIds"`
}

// Decode implements the decoder interface.
func (nr *NewRole) Decode(data []byte) error {
	return json.Unmarshal(data, nr)
}

// Validate checks the data for correctness.
func (nr NewRole) Validate() error {
	if err := errs.Check(nr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateRole represents data for updating a role.
type UpdateRole struct {
	Name         *string `json:"name"`
	RoleType     *string `json:"roleType"`
	Facet        *string `json:"facet"`
	Description  *string `json:"description"`
	DisplayOrder *int    `json:"displayOrder" validate:"omitempty,min=1,max=15"`
}

// Decode implements the decoder interface.
func (ur *UpdateRole) Decode(data []byte) error {
	return json.Unmarshal(data, ur)
}

// Validate checks the data for correctness.
func (ur UpdateRole) Validate() error {
	if err := errs.Check(ur); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ConnectValueRequest represents a request to connect a value to a role.
type ConnectValueRequest struct {
	ValueID string `json:"value_id" validate:"required,uuid"`
}

// Decode implements the decoder interface.
func (cvr *ConnectValueRequest) Decode(data []byte) error {
	return json.Unmarshal(data, cvr)
}

// Validate checks the data for correctness.
func (cvr ConnectValueRequest) Validate() error {
	if err := errs.Check(cvr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// SetValuesRequest represents a request to set all values for a role.
type SetValuesRequest struct {
	ValueIDs []string `json:"value_ids" validate:"dive,uuid"`
}

// Decode implements the decoder interface.
func (svr *SetValuesRequest) Decode(data []byte) error {
	return json.Unmarshal(data, svr)
}

// Validate checks the data for correctness.
func (svr SetValuesRequest) Validate() error {
	if err := errs.Check(svr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// =============================================================================
// Reorder Types
// =============================================================================

// ReorderRequest represents a batch reorder request.
type ReorderRequest struct {
	Items []ReorderItem `json:"items" validate:"required,min=1,max=15,dive"`
}

// ReorderItem represents a single item to be reordered.
type ReorderItem struct {
	ID           string `json:"id" validate:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" validate:"required,min=1,max=15"`
}

// Decode implements the decoder interface.
func (rr *ReorderRequest) Decode(data []byte) error {
	return json.Unmarshal(data, rr)
}

// Validate checks the data for correctness.
func (rr ReorderRequest) Validate() error {
	if err := errs.Check(rr); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// Check for duplicate displayOrders in request
	seen := make(map[int]bool)
	for _, item := range rr.Items {
		if seen[item.DisplayOrder] {
			return fmt.Errorf("duplicate displayOrder: %d", item.DisplayOrder)
		}
		seen[item.DisplayOrder] = true
	}

	return nil
}

// =============================================================================
// Business -> App Conversions
// =============================================================================

func toAppRole(role rolebus.Role) Role {
	var archivedAt *string
	if role.ArchivedAt != nil {
		t := role.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		archivedAt = &t
	}

	return Role{
		ID:           role.ID.String(),
		Name:         role.Name.Value(),
		RoleType:     role.RoleType.String(),
		Facet:        role.Facet.String(),
		Description:  role.Description.Value(),
		DisplayOrder: role.DisplayOrder.Value(),
		Status:       role.Status.String(),
		ArchivedAt:   archivedAt,
		DateCreated:  role.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated:  role.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppRoleWithValues(role rolebus.Role, valueIDs []uuid.UUID) Role {
	appRole := toAppRole(role)
	appRole.ValueIDs = make([]string, len(valueIDs))
	for i, id := range valueIDs {
		appRole.ValueIDs[i] = id.String()
	}
	return appRole
}

func toAppRoles(roles []rolebus.Role) []Role {
	app := make([]Role, len(roles))
	for i, role := range roles {
		app[i] = toAppRole(role)
	}
	return app
}

// =============================================================================
// App -> Business Conversions
// =============================================================================

func toBusNewRole(ctx context.Context, nr NewRole) (rolebus.NewRole, error) {
	var errors errs.FieldErrors

	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		errors.Add("userID", err)
	}

	// Parse name
	name, err := rolename.Parse(nr.Name)
	if err != nil {
		errors.Add("name", err)
	}

	// Parse role type
	roleTypeVal, err := roletype.Parse(nr.RoleType)
	if err != nil {
		errors.Add("roleType", err)
	}

	// Parse facet
	facetVal, err := rolefacet.Parse(nr.Facet)
	if err != nil {
		errors.Add("facet", err)
	}

	// Parse description
	description, err := roledescription.Parse(nr.Description)
	if err != nil {
		errors.Add("description", err)
	}

	// Parse display order
	displayOrderVal, err := roledisplayorder.Parse(nr.DisplayOrder)
	if err != nil {
		errors.Add("displayOrder", err)
	}

	// Parse value IDs
	valueIDs := make([]uuid.UUID, 0, len(nr.ValueIDs))
	for i, idStr := range nr.ValueIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			errors.Add(fmt.Sprintf("valueIds[%d]", i), err)
		} else {
			valueIDs = append(valueIDs, id)
		}
	}

	if len(errors) > 0 {
		return rolebus.NewRole{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return rolebus.NewRole{
		UserID:       userID,
		Name:         name,
		RoleType:     roleTypeVal,
		Facet:        facetVal,
		Description:  description,
		DisplayOrder: displayOrderVal,
		ValueIDs:     valueIDs,
	}, nil
}

func toBusUpdateRole(ur UpdateRole) (rolebus.UpdateRole, error) {
	var errors errs.FieldErrors
	var bus rolebus.UpdateRole

	if ur.Name != nil {
		name, err := rolename.Parse(*ur.Name)
		if err != nil {
			errors.Add("name", err)
		} else {
			bus.Name = &name
		}
	}

	if ur.RoleType != nil {
		roleTypeVal, err := roletype.Parse(*ur.RoleType)
		if err != nil {
			errors.Add("roleType", err)
		} else {
			bus.RoleType = &roleTypeVal
		}
	}

	if ur.Facet != nil {
		facetVal, err := rolefacet.Parse(*ur.Facet)
		if err != nil {
			errors.Add("facet", err)
		} else {
			bus.Facet = &facetVal
		}
	}

	if ur.Description != nil {
		description, err := roledescription.Parse(*ur.Description)
		if err != nil {
			errors.Add("description", err)
		} else {
			bus.Description = &description
		}
	}

	if ur.DisplayOrder != nil {
		displayOrderVal, err := roledisplayorder.Parse(*ur.DisplayOrder)
		if err != nil {
			errors.Add("displayOrder", err)
		} else {
			bus.DisplayOrder = &displayOrderVal
		}
	}

	if len(errors) > 0 {
		return rolebus.UpdateRole{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}

func toBusReorderRequest(appReorder ReorderRequest) (rolebus.ReorderRequest, error) {
	var fieldErrors errs.FieldErrors
	items := make([]rolebus.ReorderItem, 0, len(appReorder.Items))

	for i, item := range appReorder.Items {
		var roleID uuid.UUID
		var displayOrderVal roledisplayorder.RoleDisplayOrder
		itemValid := true

		id, err := uuid.Parse(item.ID)
		if err != nil {
			fieldErrors.Add(fmt.Sprintf("items[%d].id", i), err)
			itemValid = false
		} else {
			roleID = id
		}

		order, err := roledisplayorder.Parse(item.DisplayOrder)
		if err != nil {
			fieldErrors.Add(fmt.Sprintf("items[%d].displayOrder", i), err)
			itemValid = false
		} else {
			displayOrderVal = order
		}

		if itemValid {
			items = append(items, rolebus.ReorderItem{
				ID:           roleID,
				DisplayOrder: displayOrderVal,
			})
		}
	}

	if len(fieldErrors) > 0 {
		return rolebus.ReorderRequest{}, fmt.Errorf("validate: %w", fieldErrors.ToError())
	}

	return rolebus.ReorderRequest{Items: items}, nil
}

// =============================================================================
// Facet Label Mapping
// =============================================================================

var roleFacetLabels = map[string]string{
	"family_relationships":   "Familia",
	"friendships_social":     "Amistades",
	"romantic_relationships": "Pareja",
	"work_career":            "Carrera",
	"education_growth":       "Aprendizaje",
	"recreation_leisure":     "Recreation",
	"spirituality_meaning":   "Espiritualidad",
	"health_wellbeing":       "Salud General",
	"community_citizenship":  "Comunidad",
	"physical_health":        "Salud Fisica",
	"mental_wellbeing":       "Bienestar Mental",
	"creative_expression":    "Creatividad",
	"leadership_mentorship":  "Liderazgo",
	"financial_stewardship":  "Finanzas",
	"environmental_nature":   "Naturaleza",
}

func getRoleFacetLabel(facet string) string {
	if label, ok := roleFacetLabels[facet]; ok {
		return label
	}
	return facet
}
