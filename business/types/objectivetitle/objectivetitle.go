// Package objectivetitle represents an objective title in the system.
package objectivetitle

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ObjectiveTitle represents a validated objective title.
type ObjectiveTitle struct {
	value string
}

// Value returns the underlying string value.
func (ot ObjectiveTitle) Value() string {
	return ot.value
}

// String returns the value of the title.
func (ot ObjectiveTitle) String() string {
	return ot.value
}

// Equal provides support for the go-cmp package and testing.
func (ot ObjectiveTitle) Equal(ot2 ObjectiveTitle) bool {
	return ot.value == ot2.value
}

// MarshalText provides support for logging and any marshal needs.
func (ot ObjectiveTitle) MarshalText() ([]byte, error) {
	return []byte(ot.value), nil
}

// Parse validates and creates an ObjectiveTitle (5-200 chars).
func Parse(value string) (ObjectiveTitle, error) {
	value = strings.TrimSpace(value)

	runeCount := utf8.RuneCountInString(value)
	if runeCount < 5 {
		return ObjectiveTitle{}, fmt.Errorf("objective title must be at least 5 characters, got %d", runeCount)
	}
	if runeCount > 200 {
		return ObjectiveTitle{}, fmt.Errorf("objective title must be at most 200 characters, got %d", runeCount)
	}

	return ObjectiveTitle{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) ObjectiveTitle {
	title, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return title
}
