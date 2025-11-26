// Package content represents content in the system.
package content

import (
	"errors"
	"strings"
)

// Content represents a content value in the system.
type Content struct {
	value string
}

// String returns the value of the content.
func (c Content) String() string {
	return c.value
}

// Equal provides support for the go-cmp package and testing.
func (c Content) Equal(c2 Content) bool {
	return c.value == c2.value
}

// MarshalText provides support for logging and any marshal needs.
func (c Content) MarshalText() ([]byte, error) {
	return []byte(c.value), nil
}

// =============================================================================

// ErrContentEmpty is returned when content is empty.
var ErrContentEmpty = errors.New("content cannot be empty")

// Parse parses the string value and returns a content if the value complies
// with the rules for content (non-empty).
func Parse(value string) (Content, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Content{}, ErrContentEmpty
	}

	return Content{value}, nil
}

// MustParse parses the string value and returns a content if the value
// complies with the rules for content. If an error occurs the function panics.
func MustParse(value string) Content {
	content, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return content
}
