// Package roledisplayorder represents a validated display order value (1-15 scale) for roles.
package roledisplayorder

import (
	"fmt"
	"strconv"
)

// RoleDisplayOrder represents a validated display order value on a 1-15 scale.
// This is used for user-controlled priority ranking where 1 is the highest priority.
type RoleDisplayOrder struct {
	value int
}

// Value returns the int value of the display order.
func (d RoleDisplayOrder) Value() int {
	return d.value
}

// String returns the string representation of the display order.
func (d RoleDisplayOrder) String() string {
	return fmt.Sprintf("%d", d.value)
}

// Equal provides support for the go-cmp package and testing.
func (d RoleDisplayOrder) Equal(d2 RoleDisplayOrder) bool {
	return d.value == d2.value
}

// MarshalText provides support for logging and any marshal needs.
func (d RoleDisplayOrder) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *RoleDisplayOrder) UnmarshalText(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid display order value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// =============================================================================

// Parse validates the int value and returns a RoleDisplayOrder if the value complies
// with the rules for display order (1-15 scale).
func Parse(value int) (RoleDisplayOrder, error) {
	if value < 1 || value > 15 {
		return RoleDisplayOrder{}, fmt.Errorf("display order must be between 1 and 15, got %d", value)
	}

	return RoleDisplayOrder{value}, nil
}

// MustParse parses the int value and returns a RoleDisplayOrder if the value
// complies with the rules for display order. If an error occurs the function panics.
// Use this only in tests or when you're certain the value is valid.
func MustParse(value int) RoleDisplayOrder {
	displayOrder, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return displayOrder
}
