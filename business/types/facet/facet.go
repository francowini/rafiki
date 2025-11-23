// Package facet provides a type for life facet categorization.
package facet

import (
	"fmt"
)

// Set of known facets - must be initialized before predefined variables
var facets = make(map[string]Facet)

// Predefined facet values
var (
	Health         = newFacet("health")
	Relationships  = newFacet("relationships")
	Career         = newFacet("career")
	PersonalGrowth = newFacet("personal_growth")
	Family         = newFacet("family")
	Creativity     = newFacet("creativity")
	Community      = newFacet("community")
	Spirituality   = newFacet("spirituality")
)

// Facet represents a validated life domain category.
type Facet struct {
	value string
}

func newFacet(facet string) Facet {
	f := Facet{value: facet}
	facets[facet] = f
	return f
}

// String returns the string value of the facet.
func (f Facet) String() string {
	return f.value
}

// Equal provides support for the go-cmp package and testing.
func (f Facet) Equal(f2 Facet) bool {
	return f.value == f2.value
}

// MarshalText provides support for logging and any marshal needs.
func (f Facet) MarshalText() ([]byte, error) {
	return []byte(f.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (f *Facet) UnmarshalText(data []byte) error {
	facet, err := Parse(string(data))
	if err != nil {
		return err
	}

	*f = facet
	return nil
}

// Parse validates and returns a Facet.
func Parse(value string) (Facet, error) {
	facet, exists := facets[value]
	if !exists {
		return Facet{}, fmt.Errorf("invalid facet %q", value)
	}
	return facet, nil
}

// MustParse parses the facet string or panics on error. Use in tests only.
func MustParse(value string) Facet {
	facet, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return facet
}

// All returns all valid facet values.
func All() []Facet {
	return []Facet{
		Health,
		Relationships,
		Career,
		PersonalGrowth,
		Family,
		Creativity,
		Community,
		Spirituality,
	}
}
