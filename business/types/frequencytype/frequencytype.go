// Package frequencytype represents frequency type for frequency-based objectives.
package frequencytype

import "fmt"

// The set of frequency types.
var (
	Daily     = newType("daily")       // Every day
	NPerWeek  = newType("n_per_week")  // N times per week
	NPerMonth = newType("n_per_month") // N times per month
)

// Set of known frequency types.
var types = make(map[string]FrequencyType)

// FrequencyType represents how often a frequency objective should be completed.
type FrequencyType struct {
	value string
}

func newType(t string) FrequencyType {
	ft := FrequencyType{t}
	types[t] = ft
	return ft
}

// Value returns the underlying string value.
func (ft FrequencyType) Value() string {
	return ft.value
}

// String returns the string representation.
func (ft FrequencyType) String() string {
	return ft.value
}

// Equal provides support for the go-cmp package and testing.
func (ft FrequencyType) Equal(ft2 FrequencyType) bool {
	return ft.value == ft2.value
}

// MarshalText provides support for logging and any marshal needs.
func (ft FrequencyType) MarshalText() ([]byte, error) {
	return []byte(ft.value), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (ft *FrequencyType) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*ft = parsed
	return nil
}

// IsDaily returns true if frequency type is daily.
func (ft FrequencyType) IsDaily() bool {
	return ft.value == Daily.value
}

// IsNPerWeek returns true if frequency type is n_per_week.
func (ft FrequencyType) IsNPerWeek() bool {
	return ft.value == NPerWeek.value
}

// IsNPerMonth returns true if frequency type is n_per_month.
func (ft FrequencyType) IsNPerMonth() bool {
	return ft.value == NPerMonth.value
}

// Parse parses the string value and returns a frequency type if one exists.
func Parse(value string) (FrequencyType, error) {
	ft, exists := types[value]
	if !exists {
		return FrequencyType{}, fmt.Errorf("invalid frequency type %q", value)
	}
	return ft, nil
}

// MustParse parses the string value and returns a frequency type.
func MustParse(value string) FrequencyType {
	ft, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return ft
}
