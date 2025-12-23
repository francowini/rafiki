// Package vobjectiveactivityapp provides HTTP handlers for objective activity queries.
package vobjectiveactivityapp

import (
	"encoding/json"

	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
)

// ActivityResponse represents the API response for objective activity.
type ActivityResponse struct {
	ObjectiveID      string        `json:"objectiveId"`
	Year             int           `json:"year"`
	Days             []DayResponse `json:"days"`
	TotalCompletions int           `json:"totalCompletions"`
	CurrentStreak    int           `json:"currentStreak"`
	LongestStreak    int           `json:"longestStreak"`
	StreakUnit       string        `json:"streakUnit"`
}

// Encode implements the encoder interface.
func (ar ActivityResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(ar)
	return data, "application/json", err
}

// DayResponse represents a single day's activity in the API response.
type DayResponse struct {
	Date        string         `json:"date"`
	HasActivity bool           `json:"hasActivity"`
	Items       []ItemResponse `json:"items"`
}

// ItemResponse represents a single activity item in the API response.
type ItemResponse struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`         // "task" or "record"
	Status       *string `json:"status,omitempty"` // record status
	Notes        *string `json:"notes,omitempty"`
	Title        *string `json:"title,omitempty"` // task title
	Contribution *int    `json:"contribution,omitempty"`
}

// toAppActivityResponse converts business activity to API response.
func toAppActivityResponse(a vobjectiveactivitybus.Activity) ActivityResponse {
	days := make([]DayResponse, len(a.Days))
	for i, day := range a.Days {
		items := make([]ItemResponse, len(day.Items))
		for j, item := range day.Items {
			items[j] = ItemResponse{
				ID:           item.ID.String(),
				Type:         string(item.ActivityType),
				Status:       item.RecordStatus,
				Notes:        item.Notes,
				Title:        item.TaskTitle,
				Contribution: item.Contribution,
			}
		}
		days[i] = DayResponse{
			Date:        day.Date.Format("2006-01-02"),
			HasActivity: day.HasActivity,
			Items:       items,
		}
	}

	return ActivityResponse{
		ObjectiveID:      a.ObjectiveID.String(),
		Year:             a.Year,
		Days:             days,
		TotalCompletions: a.TotalCompletions,
		CurrentStreak:    a.CurrentStreak,
		LongestStreak:    a.LongestStreak,
		StreakUnit:       string(a.StreakUnit),
	}
}
