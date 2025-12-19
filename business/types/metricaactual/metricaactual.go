// Package metricaactual represents a validated current progress value (must be >= 0).
package metricaactual

import "fmt"

// MetricaActual represents a validated current progress for resultado tracking.
// Must be non-negative (>= 0). Upper bound validation against MetricaObjetivo
// is handled in the business layer since it depends on another field.
type MetricaActual struct {
	value int
}

// Value returns the int value.
func (m MetricaActual) Value() int {
	return m.value
}

// String returns the string representation.
func (m MetricaActual) String() string {
	return fmt.Sprintf("%d", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m MetricaActual) Equal(m2 MetricaActual) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m MetricaActual) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// Parse validates and creates a MetricaActual (must be >= 0).
func Parse(value int) (MetricaActual, error) {
	if value < 0 {
		return MetricaActual{}, fmt.Errorf("metrica actual must be non-negative, got %d", value)
	}
	return MetricaActual{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) MetricaActual {
	m, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero returns a MetricaActual with value 0 (initial state).
func Zero() MetricaActual {
	return MetricaActual{0}
}
