// Package intensity represents a validated intensity value (0-10 scale) in the system.
package intensity

import (
	"fmt"
)

// Intensity represents a validated intensity value on a 0-10 scale.
// This is commonly used for measuring distress, emotion, or pain intensity.
type Intensity struct {
	value int
}

// Value returns the int value of the intensity.
func (i Intensity) Value() int {
	return i.value
}

// String returns the string representation of the intensity.
func (i Intensity) String() string {
	return fmt.Sprintf("%d", i.value)
}

// Equal provides support for the go-cmp package and testing.
func (i Intensity) Equal(i2 Intensity) bool {
	return i.value == i2.value
}

// MarshalText provides support for logging and any marshal needs.
func (i Intensity) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// =============================================================================

// Parse validates the int value and returns an Intensity if the value complies
// with the rules for intensity (0-10 scale).
func Parse(value int) (Intensity, error) {
	if value < 0 || value > 10 {
		return Intensity{}, fmt.Errorf("intensity must be between 0 and 10, got %d", value)
	}

	return Intensity{value}, nil
}

// MustParse parses the int value and returns an Intensity if the value
// complies with the rules for intensity. If an error occurs the function panics.
// Use this only in tests or when you're certain the value is valid.
func MustParse(value int) Intensity {
	intensity, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return intensity
}
