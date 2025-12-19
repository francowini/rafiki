// Package frecuencian represents a validated frequency N value (must be >= 1).
package frecuencian

import "fmt"

// FrecuenciaN represents a validated frequency count for frecuencia tracking.
// Must be >= 1 (e.g., "5 times per week", "10 times per month").
// Used with n_por_semana and n_por_mes frequency types.
type FrecuenciaN struct {
	value int
}

// Value returns the int value.
func (f FrecuenciaN) Value() int {
	return f.value
}

// String returns the string representation.
func (f FrecuenciaN) String() string {
	return fmt.Sprintf("%d", f.value)
}

// Equal provides support for the go-cmp package and testing.
func (f FrecuenciaN) Equal(f2 FrecuenciaN) bool {
	return f.value == f2.value
}

// MarshalText provides support for logging and any marshal needs.
func (f FrecuenciaN) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// Parse validates and creates a FrecuenciaN (must be >= 1).
func Parse(value int) (FrecuenciaN, error) {
	if value < 1 {
		return FrecuenciaN{}, fmt.Errorf("frecuencia N must be at least 1, got %d", value)
	}
	return FrecuenciaN{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) FrecuenciaN {
	f, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return f
}
