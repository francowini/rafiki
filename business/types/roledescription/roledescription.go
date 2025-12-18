// Package roledescription provides validation for role description strings.
package roledescription

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// RoleDescription represents a validated role description (10-500 characters).
type RoleDescription struct {
	value string
}

// Validation errors
var (
	ErrEmpty    = errors.New("role description cannot be empty")
	ErrTooShort = errors.New("role description must be at least 10 characters")
	ErrTooLong  = errors.New("role description must be at most 500 characters")
)

// Value returns the string value of the role description.
func (rd RoleDescription) Value() string {
	return rd.value
}

// String returns the string representation.
func (rd RoleDescription) String() string {
	return rd.value
}

// Equal provides support for the go-cmp package and testing.
func (rd RoleDescription) Equal(rd2 RoleDescription) bool {
	return rd.value == rd2.value
}

// MarshalText provides support for logging and any marshal needs.
func (rd RoleDescription) MarshalText() ([]byte, error) {
	return []byte(rd.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (rd *RoleDescription) UnmarshalText(data []byte) error {
	desc, err := Parse(string(data))
	if err != nil {
		return err
	}

	*rd = desc
	return nil
}

// Parse validates and creates a RoleDescription.
func Parse(value string) (RoleDescription, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return RoleDescription{}, ErrEmpty
	}

	// Use rune count for proper UTF-8 character validation
	runeCount := utf8.RuneCountInString(value)

	if runeCount < 10 {
		return RoleDescription{}, ErrTooShort
	}

	if runeCount > 500 {
		return RoleDescription{}, ErrTooLong
	}

	return RoleDescription{value: value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) RoleDescription {
	desc, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return desc
}
