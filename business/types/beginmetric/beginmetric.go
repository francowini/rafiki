// Package beginmetric represents an optional starting metric value for result objectives.
package beginmetric

import "fmt"

// BeginMetric represents a validated beginning metric for result tracking.
type BeginMetric struct {
	value int
}

// Value returns the int value.
func (m BeginMetric) Value() int {
	return m.value
}

// String returns the string representation.
func (m BeginMetric) String() string {
	return fmt.Sprintf("%d", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m BeginMetric) Equal(m2 BeginMetric) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m BeginMetric) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (m *BeginMetric) UnmarshalText(data []byte) error {
	var value int
	_, err := fmt.Sscanf(string(data), "%d", &value)
	if err != nil {
		return fmt.Errorf("invalid begin metric value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*m = parsed
	return nil
}

// Parse validates and creates a BeginMetric (must be >= 0).
func Parse(value int) (BeginMetric, error) {
	if value < 0 {
		return BeginMetric{}, fmt.Errorf("begin metric must be non-negative, got %d", value)
	}
	return BeginMetric{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) BeginMetric {
	m, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return m
}
