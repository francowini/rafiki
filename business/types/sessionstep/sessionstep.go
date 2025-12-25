// Package sessionstep represents a validated step in the Telegram ACT Moment flow (1-6).
package sessionstep

import (
	"fmt"
	"strconv"
)

// SessionStep represents a validated step in the Telegram ACT Moment flow (1-6).
type SessionStep struct {
	value int
}

// Step accessor functions provide immutable step values for use in business logic.

// Step1Situacion returns the first step (Situación).
func Step1Situacion() SessionStep { return SessionStep{1} }

// Step2Sintomas returns the second step (Síntomas).
func Step2Sintomas() SessionStep { return SessionStep{2} }

// Step3Conducta returns the third step (Conducta).
func Step3Conducta() SessionStep { return SessionStep{3} }

// Step4Consecuencias returns the fourth step (Consecuencias).
func Step4Consecuencias() SessionStep { return SessionStep{4} }

// Step5Valores returns the fifth step (Valores).
func Step5Valores() SessionStep { return SessionStep{5} }

// Step6Intensidad returns the sixth step (Intensidad).
func Step6Intensidad() SessionStep { return SessionStep{6} }

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

// Parse validates and creates a SessionStep (1-6).
func Parse(value int) (SessionStep, error) {
	if value < 1 || value > 6 {
		return SessionStep{}, fmt.Errorf("session step must be between 1 and 6, got %d", value)
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

// Next returns the next step, or error if already at step 6.
func (s SessionStep) Next() (SessionStep, error) {
	if s.value >= 6 {
		return SessionStep{}, fmt.Errorf("already at final step (6)")
	}

	return SessionStep{s.value + 1}, nil
}

// IsFirst returns true if this is step 1.
func (s SessionStep) IsFirst() bool {
	return s.value == 1
}

// IsFinal returns true if this is step 6 (intensidad).
func (s SessionStep) IsFinal() bool {
	return s.value == 6
}
