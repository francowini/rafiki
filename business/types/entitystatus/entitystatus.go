// Package entitystatus represents entity lifecycle status in the system.
package entitystatus

import "fmt"

// The set of statuses that entities can have.
var (
	Active   = newStatus("active")
	Archived = newStatus("archived")
)

// =============================================================================

// Set of known statuses.
var statuses = make(map[string]Status)

// Status represents an entity's lifecycle status.
type Status struct {
	value string
}

func newStatus(status string) Status {
	s := Status{status}
	statuses[status] = s
	return s
}

// Value returns the underlying string value.
func (s Status) Value() string {
	return s.value
}

// String returns the string representation of the status.
func (s Status) String() string {
	return s.value
}

// Equal provides support for the go-cmp package and testing.
func (s Status) Equal(s2 Status) bool {
	return s.value == s2.value
}

// MarshalText provides support for logging and any marshal needs.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *Status) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// IsActive returns true if the status is active.
func (s Status) IsActive() bool {
	return s.value == Active.value
}

// IsArchived returns true if the status is archived.
func (s Status) IsArchived() bool {
	return s.value == Archived.value
}

// =============================================================================

// Parse parses the string value and returns a status if one exists.
func Parse(value string) (Status, error) {
	status, exists := statuses[value]
	if !exists {
		return Status{}, fmt.Errorf("invalid status %q", value)
	}
	return status, nil
}

// MustParse parses the string value and returns a status if one exists.
// If an error occurs the function panics. Use this only in tests.
func MustParse(value string) Status {
	status, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return status
}
