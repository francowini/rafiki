# Product TODO

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

### UX/Design Questions

- [ ] **Completed Objectives Display** - We don't know how to visually handle objectives with `status: completado` to make them look nice. Questions to resolve:
  - Should completed objectives be grayed out or have a special visual treatment?
  - Should they show a celebration/success state?
  - Should they be moved to a separate "Completed" section?
  - Should the heatmap still be interactive or read-only?
  - How to differentiate from archived objectives?
