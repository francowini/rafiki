package vobjectiveactivitydb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/vobjectiveactivitybus"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages database operations for objective activity queries.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs a Store for database access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Query retrieves activity data for an objective within a specific year.
func (s *Store) Query(ctx context.Context, filter vobjectiveactivitybus.QueryFilter) (vobjectiveactivitybus.Activity, error) {
	// First verify the objective exists and user owns it
	objectiveQuery := `
	SELECT objective_id, user_id, frequency_type, frequency_count
	FROM objectives
	WHERE objective_id = :objective_id`

	objData := struct {
		ObjectiveID string `db:"objective_id"`
	}{
		ObjectiveID: filter.ObjectiveID.String(),
	}

	var objInfo struct {
		ObjectiveID    uuid.UUID `db:"objective_id"`
		UserID         uuid.UUID `db:"user_id"`
		FrequencyType  *string   `db:"frequency_type"`
		FrequencyCount *int      `db:"frequency_count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, objectiveQuery, objData, &objInfo); err != nil {
		return vobjectiveactivitybus.Activity{}, fmt.Errorf("namedquerystruct objective: %w", err)
	}

	// Verify ownership
	if objInfo.UserID != filter.UserID {
		return vobjectiveactivitybus.Activity{}, vobjectiveactivitybus.ErrNotObjectiveOwner
	}

	// Query activity data from the view for the specified year
	yearVal := filter.Year.Value()
	startDate := time.Date(yearVal, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(yearVal, 12, 31, 23, 59, 59, 0, time.UTC)

	// Query combines objective_records and completed tasks
	activityQuery := `
	SELECT
		or_data.objective_id,
		or_data.user_id,
		or_data.record_date AS activity_date,
		'record' AS activity_type,
		or_data.objective_record_id AS item_id,
		CAST(NULL AS TEXT) AS task_title,
		CAST(NULL AS INTEGER) AS contribution,
		or_data.status AS record_status,
		or_data.notes,
		or_data.date_created
	FROM objective_records or_data
	WHERE or_data.objective_id = :objective_id
	  AND or_data.record_date >= :start_date
	  AND or_data.record_date <= :end_date

	UNION ALL

	SELECT
		t.objective_id,
		t.user_id,
		DATE(t.completed_at) AS activity_date,
		'task' AS activity_type,
		t.task_id AS item_id,
		t.title AS task_title,
		t.contribution,
		CAST(NULL AS TEXT) AS record_status,
		CAST(NULL AS TEXT) AS notes,
		t.date_created
	FROM tasks t
	WHERE t.objective_id = :objective_id
	  AND t.status = 'completed'
	  AND t.completed_at IS NOT NULL
	  AND DATE(t.completed_at) >= :start_date
	  AND DATE(t.completed_at) <= :end_date

	ORDER BY activity_date DESC, date_created DESC`

	data := map[string]any{
		"objective_id": filter.ObjectiveID.String(),
		"start_date":   startDate,
		"end_date":     endDate,
	}

	var rows []activityRow
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, activityQuery, data, &rows); err != nil {
		return vobjectiveactivitybus.Activity{}, fmt.Errorf("namedqueryslice activity: %w", err)
	}

	// Build day activity map
	dayMap := make(map[string][]vobjectiveactivitybus.ActivityItem)
	for _, row := range rows {
		dateKey := row.ActivityDate.Format("2006-01-02")
		dayMap[dateKey] = append(dayMap[dateKey], toBusActivityItem(ctx, s.log, row))
	}

	// Build days slice for all days in the year
	var days []vobjectiveactivitybus.DayActivity
	totalCompletions := 0

	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateKey := currentDate.Format("2006-01-02")
		items := dayMap[dateKey]
		hasActivity := len(items) > 0

		days = append(days, vobjectiveactivitybus.DayActivity{
			Date:        currentDate,
			HasActivity: hasActivity,
			Items:       items,
		})

		if hasActivity {
			// Count completions (records with status="completed" or tasks)
			for _, item := range items {
				if item.ActivityType == vobjectiveactivitybus.ActivityTypeTask {
					totalCompletions++
				} else if item.RecordStatus != nil && item.RecordStatus.IsCompleted() {
					totalCompletions++
				}
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Calculate streaks
	currentStreak, longestStreak, streakUnit := calculateStreaks(days, objInfo.FrequencyType, objInfo.FrequencyCount)

	return vobjectiveactivitybus.Activity{
		ObjectiveID:      filter.ObjectiveID,
		UserID:           filter.UserID,
		Year:             yearVal,
		Days:             days,
		TotalCompletions: totalCompletions,
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
		StreakUnit:       streakUnit,
	}, nil
}

// calculateStreaks computes current and longest streaks based on frequency type.
func calculateStreaks(days []vobjectiveactivitybus.DayActivity, freqType *string, freqCount *int) (current, longest int, unit vobjectiveactivitybus.StreakUnit) {
	if freqType != nil && *freqType == "n_per_week" && freqCount != nil {
		current, longest = calculateWeeklyStreaks(days, *freqCount)
		return current, longest, vobjectiveactivitybus.StreakUnitWeeks
	}
	if freqType != nil && *freqType == "n_per_month" && freqCount != nil {
		current, longest = calculateMonthlyStreaks(days, *freqCount)
		return current, longest, vobjectiveactivitybus.StreakUnitMonths
	}
	current, longest = calculateDailyStreaks(days)
	return current, longest, vobjectiveactivitybus.StreakUnitDays
}

// calculateDailyStreaks computes streaks for daily frequency (consecutive days with activity).
// Iterates backwards from most recent day, counting consecutive activity days.
// Current streak is the most recent unbroken run; longest is the maximum seen.
func calculateDailyStreaks(days []vobjectiveactivitybus.DayActivity) (current, longest int) {
	if len(days) == 0 {
		return 0, 0
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	streak := 0
	foundCurrentStreak := false

	// Iterate backwards from most recent day
	for i := len(days) - 1; i >= 0; i-- {
		day := days[i]

		// Skip future days
		if day.Date.After(today) {
			continue
		}

		if day.HasActivity {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			// First gap encountered - capture current streak and stop
			if !foundCurrentStreak && streak > 0 {
				current = streak
				foundCurrentStreak = true
			}
			streak = 0
		}
	}

	// If we never hit a gap, the entire run is the current streak
	if !foundCurrentStreak && streak > 0 {
		current = streak
	}

	return current, longest
}

// calculateWeeklyStreaks computes streaks for n_per_week frequency.
func calculateWeeklyStreaks(days []vobjectiveactivitybus.DayActivity, requiredPerWeek int) (current, longest int) {
	if len(days) == 0 {
		return 0, 0
	}

	// Group days by ISO week
	weekActivity := make(map[string]int)

	for _, day := range days {
		year, week := day.Date.ISOWeek()
		key := fmt.Sprintf("%d-%02d", year, week)
		if day.HasActivity {
			weekActivity[key]++
		}
	}

	// Sort weeks and count consecutive weeks meeting the requirement
	// For simplicity, we'll iterate through all weeks in the year
	var weeks []string
	currentDate := days[0].Date
	for i := 0; i < 53; i++ {
		year, week := currentDate.ISOWeek()
		key := fmt.Sprintf("%d-%02d", year, week)
		weeks = append(weeks, key)
		currentDate = currentDate.AddDate(0, 0, 7)
	}

	streak := 0
	for _, week := range weeks {
		count := weekActivity[week]
		if count >= requiredPerWeek {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 0
		}
	}

	current = streak
	return current, longest
}

// calculateMonthlyStreaks computes streaks for n_per_month frequency.
func calculateMonthlyStreaks(days []vobjectiveactivitybus.DayActivity, requiredPerMonth int) (current, longest int) {
	if len(days) == 0 {
		return 0, 0
	}

	// Group days by month
	monthActivity := make(map[string]int)

	for _, day := range days {
		key := day.Date.Format("2006-01")
		if day.HasActivity {
			monthActivity[key]++
		}
	}

	// Iterate through months
	streak := 0
	for month := 1; month <= 12; month++ {
		key := fmt.Sprintf("%d-%02d", days[0].Date.Year(), month)
		count := monthActivity[key]
		if count >= requiredPerMonth {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 0
		}
	}

	current = streak
	return current, longest
}
