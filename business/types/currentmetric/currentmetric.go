// Package currentmetric represents a validated current progress value (must be >= 0).
package currentmetric

import (
	"fmt"
	"strconv"
)

// CurrentMetric represents a validated current progress for result tracking.
// Must be non-negative (>= 0). Upper bound validation against TargetMetric
// is handled in the business layer since it depends on another field.
type CurrentMetric struct {
	value int
}

// Value returns the int value.
func (m CurrentMetric) Value() int {
	return m.value
}

// String returns the string representation.
func (m CurrentMetric) String() string {
	return fmt.Sprintf("%d", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m CurrentMetric) Equal(m2 CurrentMetric) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m CurrentMetric) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (m *CurrentMetric) UnmarshalText(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid current metric value %q: %w", string(data), err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*m = parsed
	return nil
}

// Parse validates and creates a CurrentMetric (must be >= 0).
func Parse(value int) (CurrentMetric, error) {
	if value < 0 {
		return CurrentMetric{}, fmt.Errorf("current metric must be non-negative, got %d", value)
	}
	return CurrentMetric{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) CurrentMetric {
	m, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero returns a CurrentMetric with value 0 (initial state).
func Zero() CurrentMetric {
	return CurrentMetric{0}
}
