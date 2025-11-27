// Package lifevisioncontent provides validation for life vision content strings.
package lifevisioncontent

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// LifeVisionContent represents validated life vision content (10-500 characters).
type LifeVisionContent struct {
	value string
}

// Validation errors
var (
	ErrEmpty    = errors.New("life vision content cannot be empty")
	ErrTooShort = errors.New("life vision content must be at least 10 characters")
	ErrTooLong  = errors.New("life vision content must be at most 500 characters")
)

// Value returns the string value of the life vision content.
func (c LifeVisionContent) Value() string {
	return c.value
}

// String returns the string representation.
func (c LifeVisionContent) String() string {
	return c.value
}

// Equal provides support for the go-cmp package and testing.
func (c LifeVisionContent) Equal(c2 LifeVisionContent) bool {
	return c.value == c2.value
}

// MarshalText provides support for logging and any marshal needs.
func (c LifeVisionContent) MarshalText() ([]byte, error) {
	return []byte(c.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (c *LifeVisionContent) UnmarshalText(data []byte) error {
	content, err := Parse(string(data))
	if err != nil {
		return err
	}

	*c = content
	return nil
}

// Parse validates and creates a LifeVisionContent.
func Parse(value string) (LifeVisionContent, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return LifeVisionContent{}, ErrEmpty
	}

	// Use rune count for proper UTF-8 character validation
	runeCount := utf8.RuneCountInString(value)

	if runeCount < 10 {
		return LifeVisionContent{}, ErrTooShort
	}

	if runeCount > 500 {
		return LifeVisionContent{}, ErrTooLong
	}

	return LifeVisionContent{value: value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) LifeVisionContent {
	content, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return content
}
