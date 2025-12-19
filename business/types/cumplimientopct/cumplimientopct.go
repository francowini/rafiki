// Package cumplimientopct represents compliance target percentage (REQUIRED for frequency).
package cumplimientopct

import "fmt"

// CumplimientoPct represents a validated compliance percentage (1-100).
type CumplimientoPct struct {
	value int
}

// Value returns the int value.
func (cp CumplimientoPct) Value() int {
	return cp.value
}

// String returns the string representation.
func (cp CumplimientoPct) String() string {
	return fmt.Sprintf("%d", cp.value)
}

// Equal provides support for the go-cmp package and testing.
func (cp CumplimientoPct) Equal(cp2 CumplimientoPct) bool {
	return cp.value == cp2.value
}

// MarshalText provides support for logging and any marshal needs.
func (cp CumplimientoPct) MarshalText() ([]byte, error) {
	return []byte(cp.String()), nil
}

// Parse validates and creates a CumplimientoPct (1-100, REQUIRED).
func Parse(value int) (CumplimientoPct, error) {
	if value < 1 || value > 100 {
		return CumplimientoPct{}, fmt.Errorf("compliance percentage must be between 1 and 100, got %d", value)
	}
	return CumplimientoPct{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) CumplimientoPct {
	pct, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return pct
}
