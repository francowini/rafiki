package thinkbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/content"
)

// Set of error variables for CRUD operations.
var (
	ErrInvalidCategory = errors.New("invalid category, must be one of: personal, work, ideas, learning, reflection")
	ErrNotFound        = errors.New("think not found")
)

// Category represents valid think categories (enum)
type Category string

// Think category constants.
const (
	CategoryPersonal   Category = "personal"
	CategoryWork       Category = "work"
	CategoryIdeas      Category = "ideas"
	CategoryLearning   Category = "learning"
	CategoryReflection Category = "reflection"
)

// String returns the string representation of the category
func (c Category) String() string {
	return string(c)
}

// ParseCategory converts a string to a Category (validation happens here)
func ParseCategory(s string) (Category, error) {
	switch s {
	case "personal":
		return CategoryPersonal, nil
	case "work":
		return CategoryWork, nil
	case "ideas":
		return CategoryIdeas, nil
	case "learning":
		return CategoryLearning, nil
	case "reflection":
		return CategoryReflection, nil
	default:
		return "", ErrInvalidCategory
	}
}

// ValidCategories returns all valid categories
func ValidCategories() []string {
	return []string{
		string(CategoryPersonal),
		string(CategoryWork),
		string(CategoryIdeas),
		string(CategoryLearning),
		string(CategoryReflection),
	}
}

// Think represents a thought/note in the system
type Think struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Category    Category
	Content     content.Content
	DateCreated time.Time
	DateUpdated time.Time
}

// NewThink contains information needed to create a new think
type NewThink struct {
	UserID   uuid.UUID
	Category Category
	Content  content.Content
}
