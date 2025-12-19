// Package metricaobjetivo represents a validated target metric value (must be > 0).
package metricaobjetivo

import "fmt"

// MetricaObjetivo represents a validated target metric for resultado tracking.
// Must be greater than 0 (e.g., "read 35 books", "run 100km").
type MetricaObjetivo struct {
	value int
}

// Value returns the int value.
func (m MetricaObjetivo) Value() int {
	return m.value
}

// String returns the string representation.
func (m MetricaObjetivo) String() string {
	return fmt.Sprintf("%d", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m MetricaObjetivo) Equal(m2 MetricaObjetivo) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m MetricaObjetivo) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// Parse validates and creates a MetricaObjetivo (must be > 0).
func Parse(value int) (MetricaObjetivo, error) {
	if value <= 0 {
		return MetricaObjetivo{}, fmt.Errorf("metrica objetivo must be greater than 0, got %d", value)
	}
	return MetricaObjetivo{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) MetricaObjetivo {
	m, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return m
}
