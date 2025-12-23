// Package year represents a validated year value in the system.
package year

import (
	"fmt"
	"strconv"
)

// Year represents a validated year value (1900-2100).
type Year struct {
	value int
}

// Value returns the int value of the year.
func (y Year) Value() int {
	return y.value
}

// String returns the string representation of the year.
func (y Year) String() string {
	return strconv.Itoa(y.value)
}

// Equal provides support for the go-cmp package and testing.
func (y Year) Equal(y2 Year) bool {
	return y.value == y2.value
}

// MarshalText provides support for logging and any marshal needs.
func (y Year) MarshalText() ([]byte, error) {
	return []byte(y.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (y *Year) UnmarshalText(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid year value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*y = parsed
	return nil
}

// =============================================================================

// Parse validates the int value and returns a Year if the value complies
// with the rules for year (1900-2100).
func Parse(value int) (Year, error) {
	if value < 1900 || value > 2100 {
		return Year{}, fmt.Errorf("year must be between 1900 and 2100, got %d", value)
	}

	return Year{value}, nil
}

// MustParse parses the int value and returns a Year if the value
// complies with the rules for year. If an error occurs the function panics.
func MustParse(value int) Year {
	year, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return year
}
