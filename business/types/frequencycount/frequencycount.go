// Package frequencycount represents a validated frequency count value (must be >= 1).
package frequencycount

import "fmt"

// FrequencyCount represents a validated frequency count for frequency tracking.
// Must be >= 1 (e.g., "5 times per week", "10 times per month").
// Used with n_per_week and n_per_month frequency types.
type FrequencyCount struct {
	value int
}

// Value returns the int value.
func (f FrequencyCount) Value() int {
	return f.value
}

// String returns the string representation.
func (f FrequencyCount) String() string {
	return fmt.Sprintf("%d", f.value)
}

// Equal provides support for the go-cmp package and testing.
func (f FrequencyCount) Equal(f2 FrequencyCount) bool {
	return f.value == f2.value
}

// MarshalText provides support for logging and any marshal needs.
func (f FrequencyCount) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// Parse validates and creates a FrequencyCount (must be >= 1).
func Parse(value int) (FrequencyCount, error) {
	if value < 1 {
		return FrequencyCount{}, fmt.Errorf("frequency count must be at least 1, got %d", value)
	}
	return FrequencyCount{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) FrequencyCount {
	f, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return f
}
