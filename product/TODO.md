# Product TODO

## Tasks Module

### Phase 3 - Future Enhancements

- [ ] **Tasks for Frequency Objectives** - Currently tasks only work with Result-type objectives (contributing to numeric progress). We want to enable tasks for Frequency-type objectives as well:
  - Option A: Tasks as organizational checklist only (no effect on compliance %)
  - Option B: Completing N tasks counts as "Hecho" (Done) for the day
  - Option C: Each task completion can trigger a daily record
  - **Decision needed**: Which approach better fits the habit-tracking workflow?

- [ ] **Standalone Tasks Page** - Create a dedicated `/tasks` page that shows:
  - All tasks across all objectives (unified view)
  - Inbox tasks (tasks not linked to any objective)
  - Filter by: objective, status (pending/completed), date
  - Quick-add task functionality
  - This provides a "task manager" experience separate from the objectives context

---

## Objectives Module

### High Priority

- [ ] **Implement Activity Endpoint** - The frontend calls `GET /v1/objetivos/{objetivo_id}/activity?year={year}` but this endpoint doesn't exist in the backend. The `react-activity-calendar` heatmap component requires:
  - All 365 days of the year with `{ date: "YYYY-MM-DD", count: number, level: 0-4 }`
  - `totalCompletions`: Total completed records
  - `streakDays`: Current streak
  - `longestStreak`: Best streak achieved

  Implementation needed:
  - Add route in `objetivoapp/route.go`
  - Create handler that aggregates records by date
  - Calculate streak statistics


- [ ] **Fix bad login message** - query: email[{ test@rafiki.lat}]: query: email[{ test@rafiki.lat}]: db: user not found


### UX/Design Questions

- [ ] **Completed Objectives Display** - We don't know how to visually handle objectives with `status: completado` to make them look nice. Questions to resolve:
  - Should completed objectives be grayed out or have a special visual treatment?
  - Should they show a celebration/success state?
  - Should they be moved to a separate "Completed" section?
  - Should the heatmap still be interactive or read-only?
  - How to differentiate from archived objectives?


