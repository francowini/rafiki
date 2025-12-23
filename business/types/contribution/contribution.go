// Package contribution represents a validated contribution value (0-10 scale) in the system.
package contribution

import (
	"fmt"
	"strconv"
)

// Contribution represents a validated contribution value on a 0-10 scale.
// 0 means the task is tracked but doesn't contribute numerically to the objective.
type Contribution struct {
	value int
}

// Value returns the int value of the contribution.
func (c Contribution) Value() int {
	return c.value
}

// String returns the string representation of the contribution.
func (c Contribution) String() string {
	return fmt.Sprintf("%d", c.value)
}

// Equal provides support for the go-cmp package and testing.
func (c Contribution) Equal(c2 Contribution) bool {
	return c.value == c2.value
}

// MarshalText provides support for logging and any marshal needs.
func (c Contribution) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (c *Contribution) UnmarshalText(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid contribution value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*c = parsed
	return nil
}

// =============================================================================

// Parse validates the int value and returns a Contribution if the value complies
// with the rules for contribution (0-10 scale).
func Parse(value int) (Contribution, error) {
	if value < 0 || value > 10 {
		return Contribution{}, fmt.Errorf("contribution must be between 0 and 10, got %d", value)
	}

	return Contribution{value}, nil
}

// MustParse parses the int value and returns a Contribution if the value
// complies with the rules for contribution. If an error occurs the function panics.
func MustParse(value int) Contribution {
	contribution, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return contribution
}
