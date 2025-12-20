// Package targetmetric represents a validated target metric value (must be > 0).
package targetmetric

import "fmt"

// TargetMetric represents a validated target metric for result tracking.
// Must be greater than 0 (e.g., "read 35 books", "run 100km").
type TargetMetric struct {
	value int
}

// Value returns the int value.
func (m TargetMetric) Value() int {
	return m.value
}

// String returns the string representation.
func (m TargetMetric) String() string {
	return fmt.Sprintf("%d", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m TargetMetric) Equal(m2 TargetMetric) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m TargetMetric) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// Parse validates and creates a TargetMetric (must be > 0).
func Parse(value int) (TargetMetric, error) {
	if value <= 0 {
		return TargetMetric{}, fmt.Errorf("target metric must be greater than 0, got %d", value)
	}
	return TargetMetric{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) TargetMetric {
	m, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return m
}
