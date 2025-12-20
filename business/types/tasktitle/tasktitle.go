// Package tasktitle represents a validated task title in the system.
package tasktitle

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TaskTitle represents a validated task title (3-200 characters).
type TaskTitle struct {
	value string
}

// Value returns the string value of the task title.
func (tt TaskTitle) Value() string {
	return tt.value
}

// String returns the string representation of the task title.
func (tt TaskTitle) String() string {
	return tt.value
}

// Equal provides support for the go-cmp package and testing.
func (tt TaskTitle) Equal(tt2 TaskTitle) bool {
	return tt.value == tt2.value
}

// MarshalText provides support for logging and any marshal needs.
func (tt TaskTitle) MarshalText() ([]byte, error) {
	return []byte(tt.value), nil
}

// =============================================================================

// Parse validates the string value and returns a TaskTitle if the value complies
// with the rules for task title (3-200 characters).
func Parse(value string) (TaskTitle, error) {
	value = strings.TrimSpace(value)

	runeCount := utf8.RuneCountInString(value)
	if runeCount < 3 {
		return TaskTitle{}, fmt.Errorf("task title must be at least 3 characters, got %d", runeCount)
	}

	if runeCount > 200 {
		return TaskTitle{}, fmt.Errorf("task title must be at most 200 characters, got %d", runeCount)
	}

	return TaskTitle{value}, nil
}

// MustParse parses the string value and returns a TaskTitle if the value
// complies with the rules for task title. If an error occurs the function panics.
func MustParse(value string) TaskTitle {
	title, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return title
}
