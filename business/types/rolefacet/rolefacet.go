// Package rolefacet provides a type for life role facet categorization.
package rolefacet

import (
	"fmt"
)

// Set of known role facets - must be initialized before predefined variables
var roleFacets = make(map[string]RoleFacet)

// Predefined role facet values (15 total)
var (
	FamilyRelationships   = newRoleFacet("family_relationships")
	FriendshipsSocial     = newRoleFacet("friendships_social")
	RomanticRelationships = newRoleFacet("romantic_relationships")
	WorkCareer            = newRoleFacet("work_career")
	EducationGrowth       = newRoleFacet("education_growth")
	RecreationLeisure     = newRoleFacet("recreation_leisure")
	SpiritualityMeaning   = newRoleFacet("spirituality_meaning")
	HealthWellbeing       = newRoleFacet("health_wellbeing")
	CommunityCitizenship  = newRoleFacet("community_citizenship")
	PhysicalHealth        = newRoleFacet("physical_health")
	MentalWellbeing       = newRoleFacet("mental_wellbeing")
	CreativeExpression    = newRoleFacet("creative_expression")
	LeadershipMentorship  = newRoleFacet("leadership_mentorship")
	FinancialStewardship  = newRoleFacet("financial_stewardship")
	EnvironmentalNature   = newRoleFacet("environmental_nature")
)

// RoleFacet represents a validated role facet category.
type RoleFacet struct {
	value string
}

func newRoleFacet(facet string) RoleFacet {
	rf := RoleFacet{value: facet}
	roleFacets[facet] = rf
	return rf
}

// String returns the string value of the role facet.
func (rf RoleFacet) String() string {
	return rf.value
}

// Equal provides support for the go-cmp package and testing.
func (rf RoleFacet) Equal(rf2 RoleFacet) bool {
	return rf.value == rf2.value
}

// Valid returns true if the role facet is a known valid value.
func (rf RoleFacet) Valid() bool {
	_, exists := roleFacets[rf.value]
	return exists
}

// MarshalText provides support for logging and any marshal needs.
func (rf RoleFacet) MarshalText() ([]byte, error) {
	return []byte(rf.value), nil
}

// UnmarshalText provides support for unmarshalling from text.
func (rf *RoleFacet) UnmarshalText(data []byte) error {
	facet, err := Parse(string(data))
	if err != nil {
		return err
	}

	*rf = facet
	return nil
}

// Parse validates and returns a RoleFacet.
func Parse(value string) (RoleFacet, error) {
	facet, exists := roleFacets[value]
	if !exists {
		return RoleFacet{}, fmt.Errorf("invalid role facet %q", value)
	}
	return facet, nil
}

// MustParse parses the role facet string or panics on error. Use in tests only.
func MustParse(value string) RoleFacet {
	facet, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return facet
}

// All returns all valid role facet values.
func All() []RoleFacet {
	return []RoleFacet{
		FamilyRelationships,
		FriendshipsSocial,
		RomanticRelationships,
		WorkCareer,
		EducationGrowth,
		RecreationLeisure,
		SpiritualityMeaning,
		HealthWellbeing,
		CommunityCitizenship,
		PhysicalHealth,
		MentalWellbeing,
		CreativeExpression,
		LeadershipMentorship,
		FinancialStewardship,
		EnvironmentalNature,
	}
}
