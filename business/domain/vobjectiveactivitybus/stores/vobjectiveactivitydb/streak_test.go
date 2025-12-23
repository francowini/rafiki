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

// TestCalculateWeeklyStreaks_YearBoundaryDec30_31 verifies that ISO week year
// boundaries are handled correctly. Dec 30/31 of some years belong to ISO week 1
// of the next year (e.g., Dec 31, 2024 is in ISO week 2025-01).
// The weeks list should only contain weeks from the data year's ISO calendar.
func TestCalculateWeeklyStreaks_YearBoundaryDec30_31(t *testing.T) {
	// 2024: Dec 30 is Monday of week 2025-01, Dec 31 is Tuesday of week 2025-01
	// Create a full year of 2024 data with activity on Dec 30-31
	days := make([]vobjectiveactivitybus.DayActivity, 366) // 2024 is a leap year
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range days {
		d := startDate.AddDate(0, 0, i)
		days[i] = vobjectiveactivitybus.DayActivity{
			Date:        d,
			HasActivity: true, // All days have activity
		}
	}

	// With activity every day, all 52 weeks of 2024 should meet requirement of 3/week
	// Dec 30-31 fall into ISO week 2025-01 and should NOT create a week "2025-01"
	current, longest := calculateWeeklyStreaks(days, 3)

	// 2024 has 52 ISO weeks (week 52 ends Dec 29)
	// Longest should be 52 (all weeks of 2024 meet requirement)
	if longest != 52 {
		t.Errorf("expected longest streak 52 (all ISO weeks of 2024), got %d", longest)
	}

	// Current streak depends on current date - but shouldn't be affected by 2025-01
	// Since we're testing historical data, current may be 0 or some value
	// The key assertion is that longest is correct and no panic occurs
	_ = current
}

// TestCalculateWeeklyStreaks_YearBoundaryJan1_2 verifies early January dates
// that may belong to the previous year's ISO week are handled correctly.
func TestCalculateWeeklyStreaks_YearBoundaryJan1_2(t *testing.T) {
	// 2025: Jan 1-5 are in ISO week 2025-01
	// Create data for 2025 with activity in first week
	days := make([]vobjectiveactivitybus.DayActivity, 365)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range days {
		d := startDate.AddDate(0, 0, i)
		// Activity only in first 4 weeks (28 days)
		days[i] = vobjectiveactivitybus.DayActivity{
			Date:        d,
			HasActivity: i < 28,
		}
	}

	current, longest := calculateWeeklyStreaks(days, 3)

	// First 4 weeks have 7 days of activity each (meets 3/week)
	// Longest should be 4
	if longest != 4 {
		t.Errorf("expected longest streak 4, got %d", longest)
	}

	// Current streak should be 0 since activity stopped after week 4
	// and we're testing 2025 data (assuming current date is in 2025)
	_ = current
}

// TestCalculateWeeklyStreaks_ConsecutiveWeeksWithGap tests that gaps properly
// break streaks and current vs longest are calculated correctly.
func TestCalculateWeeklyStreaks_ConsecutiveWeeksWithGap(t *testing.T) {
	// Create 12 weeks of data with a gap in week 5
	days := make([]vobjectiveactivitybus.DayActivity, 84) // 12 weeks
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range days {
		d := startDate.AddDate(0, 0, i)
		weekNum := i / 7
		// Activity in weeks 0-3 (4 weeks), gap in week 4, activity in weeks 5-11 (7 weeks)
		hasActivity := weekNum != 4
		days[i] = vobjectiveactivitybus.DayActivity{
			Date:        d,
			HasActivity: hasActivity,
		}
	}

	_, longest := calculateWeeklyStreaks(days, 3)

	// Weeks 5-11 have 7 consecutive weeks meeting requirement
	// Weeks 0-3 have 4 consecutive weeks
	// Longest should be 7
	if longest != 7 {
		t.Errorf("expected longest streak 7, got %d", longest)
	}
}

// TestCalculateMonthlyStreaks_ConsecutiveMonthsWithGap tests monthly streak
// calculation with a gap in the middle.
func TestCalculateMonthlyStreaks_ConsecutiveMonthsWithGap(t *testing.T) {
	// Create a year of data: activity in Jan-Apr (4 months), gap in May,
	// activity in Jun-Dec (7 months)
	days := make([]vobjectiveactivitybus.DayActivity, 365)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range days {
		d := startDate.AddDate(0, 0, i)
		month := d.Month()
		// Activity in all months except May (month 5)
		hasActivity := month != 5
		days[i] = vobjectiveactivitybus.DayActivity{
			Date:        d,
			HasActivity: hasActivity,
		}
	}

	// With ~30 days/month of activity and requirement of 10/month
	_, longest := calculateMonthlyStreaks(days, 10)

	// Jun-Dec = 7 consecutive months meeting requirement
	// Jan-Apr = 4 consecutive months
	// Longest should be 7
	if longest != 7 {
		t.Errorf("expected longest streak 7, got %d", longest)
	}
}
