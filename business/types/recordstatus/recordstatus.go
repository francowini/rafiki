// Package recordstatus represents objective record status (for tracking skips).
package recordstatus

import "fmt"

// The set of record statuses.
var (
	Completed           = newStatus("completed")             // Completed the habit
	IntentionallySkipped = newStatus("intentionally_skipped") // Intentionally skipped (rest day)
	Skipped             = newStatus("skipped")               // Missed (unintentional)
)

// Set of known statuses.
var statuses = make(map[string]RecordStatus)

// RecordStatus represents a record's status.
type RecordStatus struct {
	value string
}

func newStatus(s string) RecordStatus {
	rs := RecordStatus{s}
	statuses[s] = rs
	return rs
}

// Value returns the underlying string value.
func (rs RecordStatus) Value() string {
	return rs.value
}

// String returns the string representation.
func (rs RecordStatus) String() string {
	return rs.value
}

// Equal provides support for the go-cmp package and testing.
func (rs RecordStatus) Equal(rs2 RecordStatus) bool {
	return rs.value == rs2.value
}

// MarshalText provides support for logging and any marshal needs.
func (rs RecordStatus) MarshalText() ([]byte, error) {
	return []byte(rs.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (rs *RecordStatus) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*rs = parsed
	return nil
}

// IsCompleted returns true if status is completed.
func (rs RecordStatus) IsCompleted() bool {
	return rs.value == Completed.value
}

// IsIntentionallySkipped returns true if status is intentionally_skipped.
func (rs RecordStatus) IsIntentionallySkipped() bool {
	return rs.value == IntentionallySkipped.value
}

// IsSkipped returns true if status is skipped.
func (rs RecordStatus) IsSkipped() bool {
	return rs.value == Skipped.value
}

// Parse parses the string value and returns a status if one exists.
func Parse(value string) (RecordStatus, error) {
	status, exists := statuses[value]
	if !exists {
		return RecordStatus{}, fmt.Errorf("invalid record status %q", value)
	}
	return status, nil
}

// MustParse parses the string value and returns a status.
func MustParse(value string) RecordStatus {
	status, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return status
}
