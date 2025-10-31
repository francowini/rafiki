package habitbus

import (
	"fmt"
	"time"




	"github.com/francowini/rafiki/business/types/name"
	"github.com/google/uuid"
)

// ============================================================================
// Habit Type

type Habit struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        name.Name
	Description name.Null
	Category    Category
	Frequency   Frequency
	Streak      int
	LastDone    time.Time
	Enabled     bool
	DateCreated time.Time
	DateUpdated time.Time
}

type NewHabit struct {
	Name        name.Name
	Description name.Null
	Category    Category
	Frequency   Frequency
}

// ============================================================================
// CATEGORY TYPE

type Category struct {
	value string
}

var categoryRegistry = map[string]Category{
	"health":       {value: "health"},
	"productivity": {value: "productivity"},
	"learning":     {value: "learning"},
	"social":       {value: "social"},
	"finance":      {value: "finance"},
	"hobby":        {value: "hobby"},
}

func ParseCategory(value string) (Category, error) {
	cat, exists := categoryRegistry[value]
	if !exists {
		return Category{}, fmt.Errorf("invalid category %q", value)
	}
	return cat, nil
}

func (c Category) String() string {
	return c.value
}

// ============================================================================
// FREQUENCY TYPE

type Frequency struct {
	value string
}

var frequencyRegistry = map[string]Frequency{
	"daily":   {value: "daily"},
	"weekly":  {value: "weekly"},
	"monthly": {value: "monthly"},
}

func ParseFrequency(value string) (Frequency, error) {
	freq, exists := frequencyRegistry[value]
	if !exists {
		return Frequency{}, fmt.Errorf("invalid frequency %q", value)
	}
	return freq, nil
}

func (f Frequency) String() string {
	return f.value
}
