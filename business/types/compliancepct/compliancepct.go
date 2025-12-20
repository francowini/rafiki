// Package compliancepct represents compliance target percentage (REQUIRED for frequency).
package compliancepct

import "fmt"

// CompliancePct represents a validated compliance percentage (1-100).
type CompliancePct struct {
	value int
}

// Value returns the int value.
func (cp CompliancePct) Value() int {
	return cp.value
}

// String returns the string representation.
func (cp CompliancePct) String() string {
	return fmt.Sprintf("%d", cp.value)
}

// Equal provides support for the go-cmp package and testing.
func (cp CompliancePct) Equal(cp2 CompliancePct) bool {
	return cp.value == cp2.value
}

// MarshalText provides support for logging and any marshal needs.
func (cp CompliancePct) MarshalText() ([]byte, error) {
	return []byte(cp.String()), nil
}

// Parse validates and creates a CompliancePct (1-100, REQUIRED).
func Parse(value int) (CompliancePct, error) {
	if value < 1 || value > 100 {
		return CompliancePct{}, fmt.Errorf("compliance percentage must be between 1 and 100, got %d", value)
	}
	return CompliancePct{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) CompliancePct {
	pct, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return pct
}
