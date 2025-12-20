// Package taskstatus represents a validated task status in the system.
package taskstatus

import "fmt"

// TaskStatus represents a validated task status.
type TaskStatus struct {
	value string
}

// Status constants.
const (
	Pending   = "pending"
	Completed = "completed"
	Canceled  = "cancelled" //nolint:misspell // DB uses British spelling
)

// Pre-parsed status values for use in production code.
// Use these instead of MustParse to avoid runtime panics.
var (
	StatusPending   = TaskStatus{Pending}
	StatusCompleted = TaskStatus{Completed}
	StatusCanceled  = TaskStatus{Canceled}
)

// Value returns the string value of the status.
func (ts TaskStatus) Value() string {
	return ts.value
}

// String returns the string representation of the status.
func (ts TaskStatus) String() string {
	return ts.value
}

// Equal provides support for the go-cmp package and testing.
func (ts TaskStatus) Equal(ts2 TaskStatus) bool {
	return ts.value == ts2.value
}

// MarshalText provides support for logging and any marshal needs.
func (ts TaskStatus) MarshalText() ([]byte, error) {
	return []byte(ts.value), nil
}

// IsPending returns true if the status is pending.
func (ts TaskStatus) IsPending() bool {
	return ts.value == Pending
}

// IsCompleted returns true if the status is completed.
func (ts TaskStatus) IsCompleted() bool {
	return ts.value == Completed
}

// IsCanceled returns true if the status is canceled.
func (ts TaskStatus) IsCanceled() bool {
	return ts.value == Canceled
}

// =============================================================================

// Parse validates the string value and returns a TaskStatus if the value complies
// with the rules for task status.
func Parse(value string) (TaskStatus, error) {
	switch value {
	case Pending, Completed, Canceled:
		return TaskStatus{value}, nil
	default:
		return TaskStatus{}, fmt.Errorf("invalid task status %q (must be: pending, completed, cancelled)", value) //nolint:misspell // Match DB value "cancelled"
	}
}

// MustParse parses the string value and returns a TaskStatus if the value
// complies with the rules for task status. If an error occurs the function panics.
func MustParse(value string) TaskStatus {
	status, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return status
}
