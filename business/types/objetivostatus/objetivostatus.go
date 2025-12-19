// Package objetivostatus represents objective status with transition rules.
package objetivostatus

import "fmt"

// The set of statuses for objectives.
var (
	Activo     = newStatus("activo")
	Completado = newStatus("completado") // TERMINAL - cannot transition out
	Abandonado = newStatus("abandonado") // TERMINAL - cannot transition out
	Pausado    = newStatus("pausado")
)

// Set of known statuses.
var statuses = make(map[string]ObjetivoStatus)

// ObjetivoStatus represents an objective's status.
type ObjetivoStatus struct {
	value string
}

func newStatus(s string) ObjetivoStatus {
	os := ObjetivoStatus{s}
	statuses[s] = os
	return os
}

// Value returns the underlying string value.
func (os ObjetivoStatus) Value() string {
	return os.value
}

// String returns the string representation.
func (os ObjetivoStatus) String() string {
	return os.value
}

// Equal provides support for the go-cmp package and testing.
func (os ObjetivoStatus) Equal(os2 ObjetivoStatus) bool {
	return os.value == os2.value
}

// MarshalText provides support for logging and any marshal needs.
func (os ObjetivoStatus) MarshalText() ([]byte, error) {
	return []byte(os.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (os *ObjetivoStatus) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*os = parsed
	return nil
}

// IsTerminal returns true if the status is terminal (completado or abandonado).
func (os ObjetivoStatus) IsTerminal() bool {
	return os.value == Completado.value || os.value == Abandonado.value
}

// IsActivo returns true if the status is activo.
func (os ObjetivoStatus) IsActivo() bool {
	return os.value == Activo.value
}

// IsPausado returns true if the status is pausado.
func (os ObjetivoStatus) IsPausado() bool {
	return os.value == Pausado.value
}

// Parse parses the string value and returns a status if one exists.
func Parse(value string) (ObjetivoStatus, error) {
	status, exists := statuses[value]
	if !exists {
		return ObjetivoStatus{}, fmt.Errorf("invalid objetivo status %q", value)
	}
	return status, nil
}

// MustParse parses the string value and returns a status.
func MustParse(value string) ObjetivoStatus {
	status, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return status
}
