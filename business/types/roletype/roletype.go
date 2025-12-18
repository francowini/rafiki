// Package roletype provides a type for life role categorization.
package roletype

import (
	"fmt"
)

// Set of known role types - must be initialized before predefined variables
var roleTypes = make(map[string]RoleType)

// Predefined role type values
var (
	Personal     = newRoleType("personal")
	Relational   = newRoleType("relational")
	Professional = newRoleType("professional")
	HobbyCause   = newRoleType("hobby_cause")
)

// RoleType represents a validated role category.
type RoleType struct {
	value string
}

func newRoleType(roleType string) RoleType {
	rt := RoleType{value: roleType}
	roleTypes[roleType] = rt
	return rt
}

// String returns the string value of the role type.
func (rt RoleType) String() string {
	return rt.value
}

// Equal provides support for the go-cmp package and testing.
func (rt RoleType) Equal(rt2 RoleType) bool {
	return rt.value == rt2.value
}

// Valid returns true if the role type is a known valid value.
func (rt RoleType) Valid() bool {
	_, exists := roleTypes[rt.value]
	return exists
}

// MarshalText provides support for logging and any marshal needs.
func (rt RoleType) MarshalText() ([]byte, error) {
	return []byte(rt.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (rt *RoleType) UnmarshalText(data []byte) error {
	roleType, err := Parse(string(data))
	if err != nil {
		return err
	}

	*rt = roleType
	return nil
}

// Parse validates and returns a RoleType.
func Parse(value string) (RoleType, error) {
	roleType, exists := roleTypes[value]
	if !exists {
		return RoleType{}, fmt.Errorf("invalid role type %q", value)
	}
	return roleType, nil
}

// MustParse parses the role type string or panics on error. Use in tests only.
func MustParse(value string) RoleType {
	roleType, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return roleType
}

// All returns all valid role type values.
func All() []RoleType {
	return []RoleType{
		Personal,
		Relational,
		Professional,
		HobbyCause,
	}
}
