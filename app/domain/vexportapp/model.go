package vexportapp

import (
	"time"

	"github.com/francowini/rafiki/business/domain/vexportbus"
)

// ExportItem represents a unified export item for API responses.
type ExportItem struct {
	ID       string `json:"id"`
	ItemType string `json:"itemType"`
	ItemDate string `json:"itemDate"`

	// Moment-specific fields
	Situation        *string `json:"situation,omitempty"`
	Thoughts         *string `json:"thoughts,omitempty"`
	PhysicalSymptoms *string `json:"physicalSymptoms,omitempty"`
	Behavior         *string `json:"behavior,omitempty"`
	Consequences     *string `json:"consequences,omitempty"`
	ValuesReflection *string `json:"valuesReflection,omitempty"`
	Intensity        *int    `json:"intensity,omitempty"`

	// Think-specific fields
	Category *string `json:"category,omitempty"`
	Content  *string `json:"content,omitempty"`

	DateCreated string `json:"dateCreated"`
}

func toAppExportItem(item vexportbus.ExportItem) ExportItem {
	appItem := ExportItem{
		ID:          item.ID.String(),
		ItemType:    item.ItemType.String(),
		ItemDate:    item.ItemDate.Format(time.RFC3339),
		DateCreated: item.DateCreated.Format(time.RFC3339),
	}

	if item.Situation != nil {
		s := item.Situation.String()
		appItem.Situation = &s
	}
	if item.Thoughts != nil {
		s := item.Thoughts.String()
		appItem.Thoughts = &s
	}
	if item.PhysicalSymptoms != nil {
		s := item.PhysicalSymptoms.String()
		appItem.PhysicalSymptoms = &s
	}
	if item.Behavior != nil {
		s := item.Behavior.String()
		appItem.Behavior = &s
	}
	if item.Consequences != nil {
		s := item.Consequences.String()
		appItem.Consequences = &s
	}
	if item.ValuesReflection != nil {
		s := item.ValuesReflection.String()
		appItem.ValuesReflection = &s
	}
	if item.Intensity != nil {
		i := item.Intensity.Value()
		appItem.Intensity = &i
	}

	if item.Category != nil {
		s := item.Category.String()
		appItem.Category = &s
	}
	if item.Content != nil {
		s := item.Content.String()
		appItem.Content = &s
	}

	return appItem
}

func toAppExportItems(items []vexportbus.ExportItem) []ExportItem {
	app := make([]ExportItem, len(items))
	for i, item := range items {
		app[i] = toAppExportItem(item)
	}
	return app
}
