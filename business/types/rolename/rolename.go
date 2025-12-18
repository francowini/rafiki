// Package rolename provides validation for role name strings.
package rolename

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// RoleName represents a validated role name (3-100 characters).
type RoleName struct {
	value string
}

// Validation errors
var (
	ErrEmpty    = errors.New("role name cannot be empty")
	ErrTooShort = errors.New("role name must be at least 3 characters")
	ErrTooLong  = errors.New("role name must be at most 100 characters")
)

// Value returns the string value of the role name.
func (rn RoleName) Value() string {
	return rn.value
}

// String returns the string representation.
func (rn RoleName) String() string {
	return rn.value
}

// Equal provides support for the go-cmp package and testing.
func (rn RoleName) Equal(rn2 RoleName) bool {
	return rn.value == rn2.value
}

// MarshalText provides support for logging and any marshal needs.
func (rn RoleName) MarshalText() ([]byte, error) {
	return []byte(rn.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (rn *RoleName) UnmarshalText(data []byte) error {
	name, err := Parse(string(data))
	if err != nil {
		return err
	}

	*rn = name
	return nil
}

// Parse validates and creates a RoleName.
func Parse(value string) (RoleName, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return RoleName{}, ErrEmpty
	}

	// Use rune count for proper UTF-8 character validation
	runeCount := utf8.RuneCountInString(value)

	if runeCount < 3 {
		return RoleName{}, ErrTooShort
	}

	if runeCount > 100 {
		return RoleName{}, ErrTooLong
	}

	return RoleName{value: value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) RoleName {
	name, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return name
}
