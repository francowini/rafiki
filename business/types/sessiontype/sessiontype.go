// Package sessiontype represents a validated session type for multi-step conversations.
package sessiontype

import (
	"fmt"
	"strings"
)

// Known session types. Add new types here as needed.
const (
	MomentTracking = "moment_tracking"
	HabitCheckIn   = "habit_check_in"
)

// validTypes maps known session types to their total steps.
var validTypes = map[string]int{
	MomentTracking: 6, // 6-step ACT functional analysis
	HabitCheckIn:   3, // Simple check-in flow (placeholder)
}

// SessionType represents a validated session type string.
type SessionType struct {
	value string
}

// Value returns the string value of the session type.
func (s SessionType) Value() string {
	return s.value
}

// String returns the string representation.
func (s SessionType) String() string {
	return s.value
}

// Equal provides support for the go-cmp package and testing.
func (s SessionType) Equal(s2 SessionType) bool {
	return s.value == s2.value
}

// MarshalText provides support for logging and any marshal needs.
func (s SessionType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *SessionType) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

// =============================================================================

// Parse validates and creates a SessionType.
func Parse(value string) (SessionType, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return SessionType{}, fmt.Errorf("session type cannot be empty")
	}

	if _, ok := validTypes[value]; !ok {
		validKeys := make([]string, 0, len(validTypes))
		for k := range validTypes {
			validKeys = append(validKeys, k)
		}
		return SessionType{}, fmt.Errorf("invalid session type %q, valid types: %v", value, validKeys)
	}

	return SessionType{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value string) SessionType {
	st, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return st
}

// =============================================================================

// TotalSteps returns the total number of steps for this session type.
func (s SessionType) TotalSteps() int {
	return validTypes[s.value]
}

// IsMomentTracking returns true if this is the moment tracking session type.
func (s SessionType) IsMomentTracking() bool {
	return s.value == MomentTracking
}

// IsHabitCheckIn returns true if this is the habit check-in session type.
func (s SessionType) IsHabitCheckIn() bool {
	return s.value == HabitCheckIn
}
