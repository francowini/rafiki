package vobjectiveactivitydb

import (
	"testing"
	"time"

	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
)

func TestCalculateDailyStreaks_EmptySlice(t *testing.T) {
	current, longest := calculateDailyStreaks(nil)
	if current != 0 || longest != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%d, %d)", current, longest)
	}

	current, longest = calculateDailyStreaks([]vobjectiveactivitybus.DayActivity{})
	if current != 0 || longest != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%d, %d)", current, longest)
	}
}

func TestCalculateDailyStreaks_ConsecutiveDays(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today.AddDate(0, 0, -4), HasActivity: true},
		{Date: today.AddDate(0, 0, -3), HasActivity: true},
		{Date: today.AddDate(0, 0, -2), HasActivity: true},
		{Date: today.AddDate(0, 0, -1), HasActivity: true},
		{Date: today, HasActivity: true},
	}

	current, longest := calculateDailyStreaks(days)
	if current != 5 {
		t.Errorf("expected current streak 5, got %d", current)
	}
	if longest != 5 {
		t.Errorf("expected longest streak 5, got %d", longest)
	}
}

func TestCalculateDailyStreaks_GapBreaksStreak(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today.AddDate(0, 0, -5), HasActivity: true},
		{Date: today.AddDate(0, 0, -4), HasActivity: true},
		{Date: today.AddDate(0, 0, -3), HasActivity: false}, // Gap
		{Date: today.AddDate(0, 0, -2), HasActivity: true},
		{Date: today.AddDate(0, 0, -1), HasActivity: true},
		{Date: today, HasActivity: true},
	}

	current, longest := calculateDailyStreaks(days)
	if current != 3 {
		t.Errorf("expected current streak 3, got %d", current)
	}
	if longest != 3 {
		t.Errorf("expected longest streak 3, got %d", longest)
	}
}

func TestCalculateDailyStreaks_LongestIsHistorical(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today.AddDate(0, 0, -9), HasActivity: true},
		{Date: today.AddDate(0, 0, -8), HasActivity: true},
		{Date: today.AddDate(0, 0, -7), HasActivity: true},
		{Date: today.AddDate(0, 0, -6), HasActivity: true},
		{Date: today.AddDate(0, 0, -5), HasActivity: true},
		{Date: today.AddDate(0, 0, -4), HasActivity: false}, // Gap
		{Date: today.AddDate(0, 0, -3), HasActivity: true},
		{Date: today.AddDate(0, 0, -2), HasActivity: true},
		{Date: today.AddDate(0, 0, -1), HasActivity: true},
		{Date: today, HasActivity: true},
	}

	current, longest := calculateDailyStreaks(days)
	if current != 4 {
		t.Errorf("expected current streak 4, got %d", current)
	}
	if longest != 5 {
		t.Errorf("expected longest streak 5, got %d", longest)
	}
}

func TestCalculateDailyStreaks_SkipsFutureDays(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today.AddDate(0, 0, -1), HasActivity: true},
		{Date: today, HasActivity: true},
		{Date: today.AddDate(0, 0, 1), HasActivity: true},  // Future
		{Date: today.AddDate(0, 0, 2), HasActivity: false}, // Future
	}

	current, longest := calculateDailyStreaks(days)
	if current != 2 {
		t.Errorf("expected current streak 2, got %d", current)
	}
	if longest != 2 {
		t.Errorf("expected longest streak 2, got %d", longest)
	}
}

func TestCalculateWeeklyStreaks_EmptySlice(t *testing.T) {
	current, longest := calculateWeeklyStreaks(nil, 3)
	if current != 0 || longest != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%d, %d)", current, longest)
	}

	current, longest = calculateWeeklyStreaks([]vobjectiveactivitybus.DayActivity{}, 3)
	if current != 0 || longest != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%d, %d)", current, longest)
	}
}

func TestCalculateMonthlyStreaks_EmptySlice(t *testing.T) {
	current, longest := calculateMonthlyStreaks(nil, 5)
	if current != 0 || longest != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%d, %d)", current, longest)
	}

	current, longest = calculateMonthlyStreaks([]vobjectiveactivitybus.DayActivity{}, 5)
	if current != 0 || longest != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%d, %d)", current, longest)
	}
}

func TestCalculateStreaks_DailyDefault(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today.AddDate(0, 0, -1), HasActivity: true},
		{Date: today, HasActivity: true},
	}

	// With no frequency type, defaults to daily
	current, longest, unit := calculateStreaks(days, nil, nil)
	if unit != vobjectiveactivitybus.StreakUnitDays {
		t.Errorf("expected unit %s, got %s", vobjectiveactivitybus.StreakUnitDays, unit)
	}
	if current != 2 {
		t.Errorf("expected current streak 2, got %d", current)
	}
	if longest != 2 {
		t.Errorf("expected longest streak 2, got %d", longest)
	}
}

func TestCalculateStreaks_WeeklyType(t *testing.T) {
	// Create days spanning multiple weeks with enough activity
	// Requirement: 3 times per week, all days have activity
	days := make([]vobjectiveactivitybus.DayActivity, 28) // 4 weeks
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range days {
		days[i] = vobjectiveactivitybus.DayActivity{
			Date:        startDate.AddDate(0, 0, i),
			HasActivity: true, // Every day has activity (exceeds 3/week)
		}
	}

	freqType := "n_per_week"
	freqCount := 3
	_, longest, unit := calculateStreaks(days, &freqType, &freqCount)
	if unit != vobjectiveactivitybus.StreakUnitWeeks {
		t.Errorf("expected unit %s, got %s", vobjectiveactivitybus.StreakUnitWeeks, unit)
	}
	// With 28 days of activity and requirement of 3/week, longest should be >= 1
	if longest < 1 {
		t.Errorf("expected longest streak >= 1, got %d", longest)
	}
}

func TestCalculateStreaks_MonthlyType(t *testing.T) {
	// Create days spanning a full year with activity
	days := make([]vobjectiveactivitybus.DayActivity, 365)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range days {
		days[i] = vobjectiveactivitybus.DayActivity{
			Date:        startDate.AddDate(0, 0, i),
			HasActivity: i%3 == 0, // Activity every 3 days (~10 per month)
		}
	}

	freqType := "n_per_month"
	freqCount := 5
	_, longest, unit := calculateStreaks(days, &freqType, &freqCount)
	if unit != vobjectiveactivitybus.StreakUnitMonths {
		t.Errorf("expected unit %s, got %s", vobjectiveactivitybus.StreakUnitMonths, unit)
	}
	// Should have at least some months meeting requirement
	if longest < 1 {
		t.Errorf("expected longest streak >= 1, got %d", longest)
	}
}

func TestCalculateDailyStreaks_NoActivity(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today.AddDate(0, 0, -2), HasActivity: false},
		{Date: today.AddDate(0, 0, -1), HasActivity: false},
		{Date: today, HasActivity: false},
	}

	current, longest := calculateDailyStreaks(days)
	if current != 0 {
		t.Errorf("expected current streak 0, got %d", current)
	}
	if longest != 0 {
		t.Errorf("expected longest streak 0, got %d", longest)
	}
}

func TestCalculateDailyStreaks_SingleDay(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	days := []vobjectiveactivitybus.DayActivity{
		{Date: today, HasActivity: true},
	}

	current, longest := calculateDailyStreaks(days)
	if current != 1 {
		t.Errorf("expected current streak 1, got %d", current)
	}
	if longest != 1 {
		t.Errorf("expected longest streak 1, got %d", longest)
	}
}
