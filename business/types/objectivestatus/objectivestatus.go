// Package objectivestatus represents objective status with transition rules.
package objectivestatus

import "fmt"

// The set of statuses for objectives.
var (
	Active    = newStatus("active")
	Completed = newStatus("completed") // TERMINAL - cannot transition out
	Abandoned = newStatus("abandoned") // TERMINAL - cannot transition out
	Paused    = newStatus("paused")
)

// Set of known statuses.
var statuses = make(map[string]ObjectiveStatus)

// ObjectiveStatus represents an objective's status.
type ObjectiveStatus struct {
	value string
}

func newStatus(s string) ObjectiveStatus {
	os := ObjectiveStatus{s}
	statuses[s] = os
	return os
}

// Value returns the underlying string value.
func (os ObjectiveStatus) Value() string {
	return os.value
}

// String returns the string representation.
func (os ObjectiveStatus) String() string {
	return os.value
}

// Equal provides support for the go-cmp package and testing.
func (os ObjectiveStatus) Equal(os2 ObjectiveStatus) bool {
	return os.value == os2.value
}

// MarshalText provides support for logging and any marshal needs.
func (os ObjectiveStatus) MarshalText() ([]byte, error) {
	return []byte(os.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (os *ObjectiveStatus) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*os = parsed
	return nil
}

// IsTerminal returns true if the status is terminal (completed or abandoned).
func (os ObjectiveStatus) IsTerminal() bool {
	return os.value == Completed.value || os.value == Abandoned.value
}

// IsActive returns true if the status is active.
func (os ObjectiveStatus) IsActive() bool {
	return os.value == Active.value
}

// IsPaused returns true if the status is paused.
func (os ObjectiveStatus) IsPaused() bool {
	return os.value == Paused.value
}

// Parse parses the string value and returns a status if one exists.
func Parse(value string) (ObjectiveStatus, error) {
	status, exists := statuses[value]
	if !exists {
		return ObjectiveStatus{}, fmt.Errorf("invalid objective status %q", value)
	}
	return status, nil
}

// MustParse parses the string value and returns a status.
func MustParse(value string) ObjectiveStatus {
	status, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return status
}
