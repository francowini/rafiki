// Package valuecontent provides validation for value content strings.
package valuecontent

import (
	"errors"
	"strings"
)

// ValueContent represents validated value content (3-200 characters).
type ValueContent struct {
	value string
}

// Validation errors
var (
	ErrEmpty    = errors.New("value content cannot be empty")
	ErrTooShort = errors.New("value content must be at least 3 characters")
	ErrTooLong  = errors.New("value content must be at most 200 characters")
)

// String returns the string value of the content.
func (vc ValueContent) String() string {
	return vc.value
}

// Equal provides support for the go-cmp package and testing.
func (vc ValueContent) Equal(vc2 ValueContent) bool {
	return vc.value == vc2.value
}

// MarshalText provides support for logging and any marshal needs.
func (vc ValueContent) MarshalText() ([]byte, error) {
	return []byte(vc.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (vc *ValueContent) UnmarshalText(data []byte) error {
	content, err := Parse(string(data))
	if err != nil {
		return err
	}

	*vc = content
	return nil
}

// Parse validates and creates a ValueContent.
func Parse(value string) (ValueContent, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return ValueContent{}, ErrEmpty
	}

	if len(value) < 3 {
		return ValueContent{}, ErrTooShort
	}

	if len(value) > 200 {
		return ValueContent{}, ErrTooLong
	}

	return ValueContent{value: value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) ValueContent {
	content, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return content
}
