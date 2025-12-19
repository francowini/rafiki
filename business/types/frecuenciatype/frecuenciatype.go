// Package frecuenciatype represents frequency type for frequency-based objectives.
package frecuenciatype

import "fmt"

// The set of frequency types.
var (
	Daily      = newType("daily")        // Every day
	NPorSemana = newType("n_por_semana") // N times per week
	NPorMes    = newType("n_por_mes")    // N times per month
)

// Set of known frequency types.
var types = make(map[string]FrecuenciaType)

// FrecuenciaType represents how often a frequency objective should be completed.
type FrecuenciaType struct {
	value string
}

func newType(t string) FrecuenciaType {
	ft := FrecuenciaType{t}
	types[t] = ft
	return ft
}

// Value returns the underlying string value.
func (ft FrecuenciaType) Value() string {
	return ft.value
}

// String returns the string representation.
func (ft FrecuenciaType) String() string {
	return ft.value
}

// Equal provides support for the go-cmp package and testing.
func (ft FrecuenciaType) Equal(ft2 FrecuenciaType) bool {
	return ft.value == ft2.value
}

// MarshalText provides support for logging and any marshal needs.
func (ft FrecuenciaType) MarshalText() ([]byte, error) {
	return []byte(ft.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (ft *FrecuenciaType) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*ft = parsed
	return nil
}

// IsDaily returns true if frequency type is daily.
func (ft FrecuenciaType) IsDaily() bool {
	return ft.value == Daily.value
}

// IsNPorSemana returns true if frequency type is n_por_semana.
func (ft FrecuenciaType) IsNPorSemana() bool {
	return ft.value == NPorSemana.value
}

// IsNPorMes returns true if frequency type is n_por_mes.
func (ft FrecuenciaType) IsNPorMes() bool {
	return ft.value == NPorMes.value
}

// Parse parses the string value and returns a frequency type if one exists.
func Parse(value string) (FrecuenciaType, error) {
	ft, exists := types[value]
	if !exists {
		return FrecuenciaType{}, fmt.Errorf("invalid frecuencia type %q", value)
	}
	return ft, nil
}

// MustParse parses the string value and returns a frequency type.
func MustParse(value string) FrecuenciaType {
	ft, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return ft
}
