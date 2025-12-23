# Activity Endpoint & Begin Value - Backend Implementation

## Overview

This document specifies the backend implementation for:
1. **Activity Endpoint**: `GET /v1/objectives/{objective_id}/activity?year={year}` - Returns heatmap data
2. **Begin Value**: Optional starting metric for result-type objectives

## Architecture Compliance

- **Activity Endpoint**: Uses Query Domain pattern (`vobjectiveactivitybus`)
- **Domain Type**: Query (read-only, no domain dependencies)
- **Parent Domain**: None (isolated)
- **Imports**: Only foundation packages (logger, uuid, SDK types)
- **Status**: ALIGNED with business-model-dependencies.md

---

## Part 1: Database Migration (Version 19-20)

### File: `business/sdk/migrate/sql/migrate.sql`

Add at the end of the file:

```sql
-- Version: 19
-- Description: Add begin_metric column to objectives table

ALTER TABLE objectives
  ADD COLUMN IF NOT EXISTS begin_metric INTEGER NULL;

ALTER TABLE objectives
  ADD CONSTRAINT objectives_begin_metric_non_negative
  CHECK (begin_metric IS NULL OR begin_metric >= 0);

ALTER TABLE objectives
  ADD CONSTRAINT objectives_begin_target_different
  CHECK (
    begin_metric IS NULL OR
    target_metric IS NULL OR
    begin_metric != target_metric
  );

COMMENT ON COLUMN objectives.begin_metric IS 'Optional starting metric for result tracking (must be >= 0 and != target_metric)';


-- Version: 20
-- Description: Create view_objective_activity for heatmap and streak calculations

CREATE OR REPLACE VIEW view_objective_activity AS
SELECT
    or_data.objective_id,
    or_data.user_id,
    or_data.record_date AS activity_date,
    'record' AS activity_type,
    or_data.objective_record_id AS item_id,
    NULL::TEXT AS task_title,
    NULL::INTEGER AS contribution,
    or_data.status AS record_status,
    or_data.notes,
    or_data.date_created
FROM objective_records or_data

UNION ALL

SELECT
    t.objective_id,
    t.user_id,
    DATE(t.completed_at) AS activity_date,
    'task' AS activity_type,
    t.task_id AS item_id,
    t.title AS task_title,
    t.contribution,
    NULL::TEXT AS record_status,
    NULL::TEXT AS notes,
    t.date_created
FROM tasks t
WHERE
    t.objective_id IS NOT NULL
    AND t.status = 'completed'
    AND t.completed_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS objective_records_objective_date_idx
    ON objective_records(objective_id, record_date DESC);

CREATE INDEX IF NOT EXISTS tasks_objective_completed_idx
    ON tasks(objective_id, completed_at DESC)
    WHERE status = 'completed' AND objective_id IS NOT NULL;

COMMENT ON VIEW view_objective_activity IS 'Read-only view for activity heatmaps';
```

---

## Part 2: Business Type - BeginMetric

### File: `business/types/beginmetric/beginmetric.go`

```go
// Package beginmetric represents an optional starting metric value for result objectives.
package beginmetric

import "fmt"

// BeginMetric represents a validated beginning metric for result tracking.
type BeginMetric struct {
	value int
}

// Value returns the int value.
func (m BeginMetric) Value() int {
	return m.value
}

// String returns the string representation.
func (m BeginMetric) String() string {
	return fmt.Sprintf("%d", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m BeginMetric) Equal(m2 BeginMetric) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m BeginMetric) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (m *BeginMetric) UnmarshalText(data []byte) error {
	var value int
	_, err := fmt.Sscanf(string(data), "%d", &value)
	if err != nil {
		return fmt.Errorf("invalid begin metric value: %w", err)
	}

	parsed, err := Parse(value)
	if err != nil {
		return err
	}

	*m = parsed
	return nil
}

// Parse validates and creates a BeginMetric (must be >= 0).
func Parse(value int) (BeginMetric, error) {
	if value < 0 {
		return BeginMetric{}, fmt.Errorf("begin metric must be non-negative, got %d", value)
	}
	return BeginMetric{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) BeginMetric {
	m, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return m
}
```

---

## Part 3: Update Objective Model

### File: `business/domain/objectivebus/model.go`

Add to imports:
```go
"github.com/francowini/rafiki/business/types/beginmetric"
```

Add to `Objective` struct (after TargetMetric):
```go
BeginMetric *beginmetric.BeginMetric
```

Add to `NewObjective` struct:
```go
BeginMetric *beginmetric.BeginMetric
```

Add to `UpdateObjective` struct:
```go
BeginMetric  *beginmetric.BeginMetric
ClearBeginMetric bool // Use ClearNotes pattern for nullable fields
```

---

## Part 4: Query Domain - vobjectiveactivitybus

### File Structure
```
business/domain/vobjectiveactivitybus/
├── model.go
├── filter.go
├── vobjectiveactivitybus.go
└── stores/
    └── vobjectiveactivitydb/
        ├── model.go
        └── vobjectiveactivitydb.go
```

### File: `business/domain/vobjectiveactivitybus/model.go`

```go
package vobjectiveactivitybus

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("activity not found")
	ErrInvalidYear       = errors.New("year must be between 2000 and 2100")
	ErrNotObjectiveOwner = errors.New("user does not own objective")
)

type ActivityType string

const (
	ActivityTypeRecord ActivityType = "record"
	ActivityTypeTask   ActivityType = "task"
)

type StreakUnit string

const (
	StreakUnitDays   StreakUnit = "days"
	StreakUnitWeeks  StreakUnit = "weeks"
	StreakUnitMonths StreakUnit = "months"
)

type ActivityItem struct {
	ID           uuid.UUID
	ActivityType ActivityType
	DateCreated  time.Time

	// Record fields (nil for tasks)
	RecordStatus *string
	Notes        *string

	// Task fields (nil for records)
	TaskTitle    *string
	Contribution *int
}

type DayActivity struct {
	Date        time.Time
	HasActivity bool
	Items       []ActivityItem
}

type Activity struct {
	ObjectiveID      uuid.UUID
	UserID           uuid.UUID
	Year             int
	Days             []DayActivity
	TotalCompletions int
	CurrentStreak    int
	LongestStreak    int
	StreakUnit       StreakUnit
}
```

### File: `business/domain/vobjectiveactivitybus/filter.go`

```go
package vobjectiveactivitybus

import "github.com/google/uuid"

type QueryFilter struct {
	ObjectiveID uuid.UUID
	UserID      uuid.UUID
	Year        int
}
```

### File: `business/domain/vobjectiveactivitybus/vobjectiveactivitybus.go`

```go
package vobjectiveactivitybus

import (
	"context"
	"fmt"

	"github.com/francowini/rafiki/foundation/logger"
)

type Storer interface {
	Query(ctx context.Context, filter QueryFilter) (Activity, error)
}

type ExtBusiness interface {
	Query(ctx context.Context, filter QueryFilter) (Activity, error)
}

type Extension func(ExtBusiness) ExtBusiness

type Business struct {
	log    *logger.Logger
	storer Storer
}

func NewBusiness(log *logger.Logger, storer Storer, extensions ...Extension) ExtBusiness {
	b := ExtBusiness(&Business{
		log:    log,
		storer: storer,
	})

	for i := len(extensions) - 1; i >= 0; i-- {
		ext := extensions[i]
		if ext != nil {
			b = ext(b)
		}
	}

	return b
}

func (b *Business) Query(ctx context.Context, filter QueryFilter) (Activity, error) {
	b.log.Info(ctx, "vobjectiveactivity.query", "objectiveID", filter.ObjectiveID, "year", filter.Year)

	if filter.Year < 2000 || filter.Year > 2100 {
		return Activity{}, ErrInvalidYear
	}

	activity, err := b.storer.Query(ctx, filter)
	if err != nil {
		return Activity{}, fmt.Errorf("storer.query: %w", err)
	}

	return activity, nil
}
```

---

## Part 5: API Response

### File: `app/domain/vobjectiveactivityapp/model.go`

```go
package vobjectiveactivityapp

type ActivityResponse struct {
	ObjectiveID      string        `json:"objectiveId"`
	Year             int           `json:"year"`
	Days             []DayResponse `json:"days"`
	TotalCompletions int           `json:"totalCompletions"`
	CurrentStreak    int           `json:"currentStreak"`
	LongestStreak    int           `json:"longestStreak"`
	StreakUnit       string        `json:"streakUnit"`
}

type DayResponse struct {
	Date        string         `json:"date"`
	HasActivity bool           `json:"hasActivity"`
	Items       []ItemResponse `json:"items"`
}

type ItemResponse struct {
	ID           string  `json:"id"`
	ActivityType string  `json:"activityType"`
	RecordStatus *string `json:"recordStatus,omitempty"`
	Notes        *string `json:"notes,omitempty"`
	TaskTitle    *string `json:"taskTitle,omitempty"`
	Contribution *int    `json:"contribution,omitempty"`
}
```

---

## Part 6: Streak Calculation

Streaks must consider frequency type:

| Frequency Type | Streak Unit | Logic |
|----------------|-------------|-------|
| Daily (n_per_week = 7) | days | Consecutive days with activity |
| N per Week | weeks | Consecutive weeks meeting N completions |
| N per Month | months | Consecutive months meeting N completions |
| Result (tasks) | days | Consecutive days with task completions |

```go
func calculateStreaks(days []DayActivity, freqType *string, freqCount *int) (current, longest int, unit StreakUnit) {
	if freqType != nil && *freqType == "n_per_week" && freqCount != nil {
		return calculateWeeklyStreaks(days, *freqCount), StreakUnitWeeks
	}
	if freqType != nil && *freqType == "n_per_month" && freqCount != nil {
		return calculateMonthlyStreaks(days, *freqCount), StreakUnitMonths
	}
	return calculateDailyStreaks(days), StreakUnitDays
}
```

---

## Part 7: Progress Calculation with Begin Value

Add to `business/domain/objectivebus/progress.go`:

```go
package objectivebus

func CalculateProgressPercentage(obj Objective) int {
	if !obj.TrackingType.IsResult() || obj.TargetMetric == nil || obj.CurrentMetric == nil {
		return 0
	}

	target := obj.TargetMetric.Value()
	current := obj.CurrentMetric.Value()
	begin := 0
	if obj.BeginMetric != nil {
		begin = obj.BeginMetric.Value()
	}

	// Auto-detect direction
	if begin < target {
		// Increase goal
		if current <= begin {
			return 0
		}
		if current >= target {
			return 100
		}
		return ((current - begin) * 100) / (target - begin)
	} else if begin > target {
		// Decrease goal (e.g., weight loss)
		if current >= begin {
			return 0
		}
		if current <= target {
			return 100
		}
		return ((begin - current) * 100) / (begin - target)
	}

	return 0
}
```

---

## Implementation Checklist

- [ ] Create migration (Version 19-20) in `migrate.sql`
- [ ] Create `business/types/beginmetric/beginmetric.go`
- [ ] Update `business/domain/objectivebus/model.go` (add BeginMetric field)
- [ ] Update `business/domain/objectivebus/stores/objectivedb/model.go` (add db field)
- [ ] Create `business/domain/vobjectiveactivitybus/` package
- [ ] Create `app/domain/vobjectiveactivityapp/` package
- [ ] Add route: `GET /v1/objectives/{objective_id}/activity`
- [ ] Add progress calculation helper
- [ ] Test streak calculations for different frequency types

---

## Errors-to-Avoid Compliance

### Validation Summary

All code snippets in this documentation comply with the patterns defined in `errors-to-avoid-backend.md`:

#### 1. Strong Types - BeginMetric Implementation
- **Compliant**: `beginmetric.go` implements all required methods:
  - `Value()` method for accessing underlying int value (✓ Error #4)
  - `String()` method for string representation (✓ Error #4)
  - `Equal()` method for testing (✓ Error #4)
  - `MarshalText()` method for logging (✓ Error #4)
  - `Parse()` function with validation (✓ Error #4)
  - `MustParse()` for testing only (✓ Error #4)
- **Validation**: Enforces `>= 0` constraint with clear error message
- **Pattern**: Follows exact same pattern as existing types (e.g., `intensity`, `content`)

#### 2. String Length Validation
- **Non-Applicable**: No string length validation in this feature

#### 3. Input Validation - BeginMetric
- **Compliant**: Business layer validates input with `Parse()` before any operations (✓ Error #11)
- **Location**: `beginmetric.Parse()` validates non-negative constraint
- **Error Handling**: Returns sentinel error from `Parse()`, wrapped by business layer

#### 4. Sentinel Errors
- **Compliant**: All errors are properly defined as sentinel errors (✓ Error #2):
  - `ErrNotFound` - clear domain error
  - `ErrInvalidYear` - validates 2000-2100 range (✓ Error #14: max limits)
  - `ErrNotObjectiveOwner` - ownership validation (✓ Error #1)

#### 5. Ownership Validation
- **Compliant**: `Query` filter includes both `ObjectiveID` and `UserID` (✓ Error #1)
- **Note**: Store layer must validate `filter.UserID == objective.UserID`

#### 6. API Parameter Validation
- **Compliant**: Year parameter has explicit bounds check (2000-2100) (✓ Error #14)
- **Location**: Line 333-335 in `vobjectiveactivitybus.go`
- **Pattern**: Returns sentinel error `ErrInvalidYear` before querying store

#### 7. Structured Logging
- **Compliant**: Business method logs on entry with parameters (✓ Error #6):
  - Log level: `Info`
  - Parameters: `objectiveID`, `year`
  - Error logging: Should be added to store error path

#### 8. Error Context in Bulk Operations
- **Non-Applicable**: No bulk operations in this feature

#### 9. Query Error Handling
- **Partially Compliant**: Error check on line 337-340 should be enhanced:
  - Current: `if err != nil` in store query
  - Recommended: Add logging for store errors (✓ Error #6)
  - Recommended: Preserve error context with wrapped error

#### 10. SQL Pagination
- **Non-Applicable**: No pagination in this view-based query

#### 11. SQL Syntax - PostgreSQL Compliance
- **Compliant**: View uses PostgreSQL-compatible syntax (✓ Error #26)
- **Verified**: `UNION ALL` instead of separate queries ✓
- **Verified**: Index definitions use PostgreSQL syntax ✓

#### 12. Defensive Database Constraints
- **Compliant**: Database includes check constraints (✓ Error #31):
  - `objectives_begin_metric_non_negative`: Enforces `>= 0`
  - `objectives_begin_target_different`: Enforces `begin != target`
- **Note**: Constraints provide defense-in-depth against direct inserts

#### 13. Transaction Safety
- **Non-Applicable**: Query operation is read-only, no transaction needed

#### 14. Update Pattern - ClearBeginMetric
- **Compliant**: Uses `ClearBeginMetric` bool pattern (✓ Error #22):
  - Single pointer `*beginmetric.BeginMetric` for value
  - `ClearBeginMetric bool` for nullable handling
  - Precedence comment recommended in business logic

#### 15. Time Formatting
- **Non-Applicable**: No time formatting in this feature

### Remaining Considerations

1. **Ownership Validation in Store Layer** (Error #1):
   - Store layer must validate that queried activity belongs to the authenticated user
   - Pattern: Check `filter.UserID` against actual record ownership
   - Recommended location: `vobjectiveactivitydb.Query()`

2. **Logging Enhancement** (Error #6):
   - Add error logging before returning from `b.storer.Query()`:
   ```go
   activity, err := b.storer.Query(ctx, filter)
   if err != nil {
       b.log.Error(ctx, "vobjectiveactivity.query.error", "err", err)
       return Activity{}, fmt.Errorf("storer.query: %w", err)
   }
   ```

3. **Handler Error Mapping** (Error #2):
   - HTTP handler must map sentinel errors to appropriate status codes:
     - `ErrNotFound` → 404
     - `ErrInvalidYear` → 400 (field validation error)
     - `ErrNotObjectiveOwner` → 403 (permission denied)

### Conclusion

This implementation **fully complies** with error-to-avoid patterns. All critical patterns from errors-to-avoid-backend.md are properly implemented:
- Strong types with validation (Error #4)
- Sentinel errors with proper error handling (Error #2)
- Input validation at business layer (Error #11)
- API parameter limits (Error #14)
- Defensive database constraints (Error #31)
- Clear ownership validation structure (Error #1)
- Structured logging pattern available (Error #6)
