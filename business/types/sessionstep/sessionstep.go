// Package sessionstep represents a validated step in a multi-step conversation flow.
package sessionstep

import (
	"fmt"
	"strconv"
)

// SessionStep represents a validated step number (1-based positive integer).
// Upper bound validation is done at runtime against the session's total_steps.
type SessionStep struct {
	value int
}

// Value returns the int value of the session step.
func (s SessionStep) Value() int {
	return s.value
}

// String returns the string representation.
func (s SessionStep) String() string {
	return fmt.Sprintf("%d", s.value)
}

// Equal provides support for the go-cmp package and testing.
func (s SessionStep) Equal(s2 SessionStep) bool {
	return s.value == s2.value
}

// MarshalText provides support for logging and any marshal needs.
func (s SessionStep) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *SessionStep) UnmarshalText(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid session step value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

// =============================================================================

// Parse validates and creates a SessionStep (must be >= 1).
// Upper bound validation should be done at the business layer against total_steps.
func Parse(value int) (SessionStep, error) {
	if value < 1 {
		return SessionStep{}, fmt.Errorf("session step must be >= 1, got %d", value)
	}

	return SessionStep{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) SessionStep {
	step, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return step
}

// =============================================================================

// IsFirst returns true if this is step 1.
func (s SessionStep) IsFirst() bool {
	return s.value == 1
}

// IsFinal returns true if this step equals totalSteps.
func (s SessionStep) IsFinal(totalSteps int) bool {
	return s.value == totalSteps
}

// Next returns the next step. Returns error if already at or beyond totalSteps.
func (s SessionStep) Next(totalSteps int) (SessionStep, error) {
	if s.value >= totalSteps {
		return SessionStep{}, fmt.Errorf("already at final step (%d)", totalSteps)
	}

	return SessionStep{s.value + 1}, nil
}

// IsValid checks if the step is within valid range for the given totalSteps.
func (s SessionStep) IsValid(totalSteps int) bool {
	return s.value >= 1 && s.value <= totalSteps
}
