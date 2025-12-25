// Package retrycount represents a validated retry count (0-2) for AI validation failures.
package retrycount

import (
	"fmt"
)

// RetryCount represents a validated retry count (0-2) for AI validation failures.
type RetryCount struct {
	value int
}

// Value returns the int value of the retry count.
func (r RetryCount) Value() int {
	return r.value
}

// String returns the string representation.
func (r RetryCount) String() string {
	return fmt.Sprintf("%d", r.value)
}

// Equal provides support for the go-cmp package and testing.
func (r RetryCount) Equal(r2 RetryCount) bool {
	return r.value == r2.value
}

// MarshalText provides support for logging and any marshal needs.
func (r RetryCount) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (r *RetryCount) UnmarshalText(data []byte) error {
	var value int
	if _, err := fmt.Sscanf(string(data), "%d", &value); err != nil {
		return fmt.Errorf("invalid retry count value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*r = parsed
	return nil
}

// =============================================================================

// Parse validates and creates a RetryCount (0-2).
func Parse(value int) (RetryCount, error) {
	if value < 0 || value > 2 {
		return RetryCount{}, fmt.Errorf("retry count must be between 0 and 2, got %d", value)
	}

	return RetryCount{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) RetryCount {
	count, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return count
}

// =============================================================================

// Increment returns a new RetryCount incremented by 1.
// Returns error if already at max (2).
func (r RetryCount) Increment() (RetryCount, error) {
	if r.value >= 2 {
		return RetryCount{}, fmt.Errorf("retry count already at maximum (2)")
	}

	return RetryCount{r.value + 1}, nil
}

// IsMaxed returns true if retry count is at maximum (2).
// When maxed, session should auto-approve and advance to next step.
func (r RetryCount) IsMaxed() bool {
	return r.value >= 2
}

// Reset returns a new RetryCount set to 0.
// Use when advancing to a new step.
func (r RetryCount) Reset() RetryCount {
	return RetryCount{0}
}
