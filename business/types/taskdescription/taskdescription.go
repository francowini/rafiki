// Package taskdescription represents an optional task description in the system.
package taskdescription

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TaskDescription represents an optional task description (max 2000 characters).
type TaskDescription struct {
	value string
}

// Value returns the string value of the task description.
func (td TaskDescription) Value() string {
	return td.value
}

// String returns the string representation of the task description.
func (td TaskDescription) String() string {
	return td.value
}

// Equal provides support for the go-cmp package and testing.
func (td TaskDescription) Equal(td2 TaskDescription) bool {
	return td.value == td2.value
}

// MarshalText provides support for logging and any marshal needs.
func (td TaskDescription) MarshalText() ([]byte, error) {
	return []byte(td.value), nil
}

// =============================================================================

// Parse validates the string value and returns a TaskDescription if the value complies
// with the rules for task description (max 2000 characters, empty allowed).
func Parse(value string) (TaskDescription, error) {
	value = strings.TrimSpace(value)

	// Empty is allowed (optional field)
	if value == "" {
		return TaskDescription{}, nil
	}

	runeCount := utf8.RuneCountInString(value)
	if runeCount > 2000 {
		return TaskDescription{}, fmt.Errorf("task description must be at most 2000 characters, got %d", runeCount)
	}

	return TaskDescription{value}, nil
}

// MustParse parses the string value and returns a TaskDescription if the value
// complies with the rules for task description. If an error occurs the function panics.
func MustParse(value string) TaskDescription {
	desc, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return desc
}
