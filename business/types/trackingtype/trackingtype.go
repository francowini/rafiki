// Package trackingtype represents objective tracking type (immutable after creation).
package trackingtype

import "fmt"

// The set of tracking types for objectives.
var (
	Resultado  = newType("resultado")  // Outcome-based (metric progress)
	Frecuencia = newType("frecuencia") // Frequency-based (calendar compliance)
)

// Set of known tracking types.
var types = make(map[string]TrackingType)

// TrackingType represents how an objective is tracked (IMMUTABLE after creation).
type TrackingType struct {
	value string
}

func newType(t string) TrackingType {
	tt := TrackingType{t}
	types[t] = tt
	return tt
}

// Value returns the underlying string value.
func (tt TrackingType) Value() string {
	return tt.value
}

// String returns the string representation.
func (tt TrackingType) String() string {
	return tt.value
}

// Equal provides support for the go-cmp package and testing.
func (tt TrackingType) Equal(tt2 TrackingType) bool {
	return tt.value == tt2.value
}

// MarshalText provides support for logging and any marshal needs.
func (tt TrackingType) MarshalText() ([]byte, error) {
	return []byte(tt.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (tt *TrackingType) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*tt = parsed
	return nil
}

// IsResultado returns true if tracking type is resultado.
func (tt TrackingType) IsResultado() bool {
	return tt.value == Resultado.value
}

// IsFrecuencia returns true if tracking type is frecuencia.
func (tt TrackingType) IsFrecuencia() bool {
	return tt.value == Frecuencia.value
}

// Parse parses the string value and returns a tracking type if one exists.
func Parse(value string) (TrackingType, error) {
	tt, exists := types[value]
	if !exists {
		return TrackingType{}, fmt.Errorf("invalid tracking type %q (must be 'resultado' or 'frecuencia')", value)
	}
	return tt, nil
}

// MustParse parses the string value and returns a tracking type.
func MustParse(value string) TrackingType {
	tt, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return tt
}
