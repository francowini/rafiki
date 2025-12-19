// Package objetivotitle represents an objective title in the system.
package objetivotitle

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ObjetivoTitle represents a validated objective title.
type ObjetivoTitle struct {
	value string
}

// Value returns the underlying string value.
func (ot ObjetivoTitle) Value() string {
	return ot.value
}

// String returns the value of the title.
func (ot ObjetivoTitle) String() string {
	return ot.value
}

// Equal provides support for the go-cmp package and testing.
func (ot ObjetivoTitle) Equal(ot2 ObjetivoTitle) bool {
	return ot.value == ot2.value
}

// MarshalText provides support for logging and any marshal needs.
func (ot ObjetivoTitle) MarshalText() ([]byte, error) {
	return []byte(ot.value), nil
}

// Parse validates and creates an ObjetivoTitle (5-200 chars).
func Parse(value string) (ObjetivoTitle, error) {
	value = strings.TrimSpace(value)

	runeCount := utf8.RuneCountInString(value)
	if runeCount < 5 {
		return ObjetivoTitle{}, fmt.Errorf("objetivo title must be at least 5 characters, got %d", runeCount)
	}
	if runeCount > 200 {
		return ObjetivoTitle{}, fmt.Errorf("objetivo title must be at most 200 characters, got %d", runeCount)
	}

	return ObjetivoTitle{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) ObjetivoTitle {
	title, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return title
}
