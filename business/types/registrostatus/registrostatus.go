// Package registrostatus represents objective record status (for tracking skips).
package registrostatus

import "fmt"

// The set of record statuses.
var (
	Completado         = newStatus("completado")          // Completed the habit
	OmitidoIntencional = newStatus("omitido_intencional") // Intentionally skipped (rest day)
	Omitido            = newStatus("omitido")             // Missed (unintentional)
)

// Set of known statuses.
var statuses = make(map[string]RegistroStatus)

// RegistroStatus represents a record's status.
type RegistroStatus struct {
	value string
}

func newStatus(s string) RegistroStatus {
	rs := RegistroStatus{s}
	statuses[s] = rs
	return rs
}

// Value returns the underlying string value.
func (rs RegistroStatus) Value() string {
	return rs.value
}

// String returns the string representation.
func (rs RegistroStatus) String() string {
	return rs.value
}

// Equal provides support for the go-cmp package and testing.
func (rs RegistroStatus) Equal(rs2 RegistroStatus) bool {
	return rs.value == rs2.value
}

// MarshalText provides support for logging and any marshal needs.
func (rs RegistroStatus) MarshalText() ([]byte, error) {
	return []byte(rs.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (rs *RegistroStatus) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*rs = parsed
	return nil
}

// IsCompletado returns true if status is completado.
func (rs RegistroStatus) IsCompletado() bool {
	return rs.value == Completado.value
}

// IsOmitidoIntencional returns true if status is omitido_intentional.
func (rs RegistroStatus) IsOmitidoIntencional() bool {
	return rs.value == OmitidoIntencional.value
}

// IsOmitido returns true if status is omitido.
func (rs RegistroStatus) IsOmitido() bool {
	return rs.value == Omitido.value
}

// Parse parses the string value and returns a status if one exists.
func Parse(value string) (RegistroStatus, error) {
	status, exists := statuses[value]
	if !exists {
		return RegistroStatus{}, fmt.Errorf("invalid registro status %q", value)
	}
	return status, nil
}

// MustParse parses the string value and returns a status.
func MustParse(value string) RegistroStatus {
	status, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return status
}
