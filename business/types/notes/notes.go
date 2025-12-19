// Package notes represents optional notes in the system.
package notes

import (
	"errors"
	"strings"
)

// Notes represents an optional notes value in the system.
type Notes struct {
	value string
}

// String returns the value of the notes.
func (n Notes) String() string {
	return n.value
}

// Equal provides support for the go-cmp package and testing.
func (n Notes) Equal(n2 Notes) bool {
	return n.value == n2.value
}

// MarshalText provides support for logging and any marshal needs.
func (n Notes) MarshalText() ([]byte, error) {
	return []byte(n.value), nil
}

// =============================================================================

// ErrNotesEmpty is returned when notes is empty.
var ErrNotesEmpty = errors.New("notes cannot be empty")

// Parse parses the string value and returns notes if the value complies
// with the rules for notes (non-empty after trimming).
func Parse(value string) (Notes, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Notes{}, ErrNotesEmpty
	}

	return Notes{value}, nil
}

// MustParse parses the string value and returns notes if the value
// complies with the rules for notes. If an error occurs the function panics.
func MustParse(value string) Notes {
	notes, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return notes
}
