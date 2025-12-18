# Rafiki Entity System Architecture

**Version 2.0 | December 2025**

This document describes the complete entity system for Rafiki, a personal development application based on ACT (Acceptance and Commitment Therapy). It covers the data model, data flows, use cases, architecture alignment, and implementation challenges.

---

## Executive Summary

Rafiki implements a **top-down planning system** that helps users align daily actions with long-term values. The hierarchy flows from abstract (Values) to concrete (Tasks):

```
VALUES → LIFE VISIONS → OBJECTIVES → INITIATIVES → TASKS
```

**Key Decisions:**
- **No Roles for MVP**: Simplified architecture using tags and optional value links instead
- **Standalone Tasks**: Quick capture without full hierarchy (inbox pattern)
- **Frequency Tracking**: GitHub-style heatmap for habit objectives
- **Immediate Feedback**: Optimistic updates via React Query
- **Archive Block Pattern**: Cannot archive parent if active children exist

---

## 1. Entity Overview

### 1.1 Entity Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│                         USERS                                   │
│                    (Authentication & Profile)                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ owns (1:N)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        VALUES (≤10)                             │
│              "Who do I want to be?" - Permanent directions      │
│                                                                 │
│  Examples:                                                      │
│  • "Be a present and loving father"                            │
│  • "Pursue continuous growth and learning"                     │
│  • "Maintain physical and mental vitality"                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ inspires (1:N)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      LIFE VISIONS                               │
│           "What does my ideal future look like?" (5-20 years)   │
│                                                                 │
│  Examples:                                                      │
│  • "Deep, trusting relationship with my adult children"        │
│  • "Recognized expert in my field with published work"         │
│  • "Running marathons at 60 with boundless energy"             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ generates (1:N)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       OBJECTIVES                                │
│              "What will I achieve this year?" (Annual)          │
│                                                                 │
│  Two tracking types:                                            │
│  • OUTCOME: "Read 35 technical books" (count to target)        │
│  • FREQUENCY: "Meditate 5 days/week" (habit tracking)          │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┴───────────────────┐
          │                                       │
          │ decomposes (1:N)                      │ tracks (1:N)
          ▼                                       ▼
┌─────────────────────────┐         ┌─────────────────────────┐
│      INITIATIVES        │         │   OBJECTIVE RECORDS     │
│   (Quarterly projects)  │         │  (Daily frequency logs) │
│                         │         │                         │
│  Example:               │         │  Example:               │
│  "Q1: Establish morning │         │  • Dec 15: Completed ✓  │
│   meditation ritual"    │         │  • Dec 14: Completed ✓  │
└─────────────────────────┘         │  • Dec 13: Skipped ✗    │
          │                         └─────────────────────────┘
          │ generates (1:N)
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                          TASKS                                  │
│              "What do I do today?" (Daily actions)              │
│                                                                 │
│  Two types:                                                     │
│  • DERIVED: From initiative ("Read chapter 3 of book X")       │
│  • STANDALONE: Quick capture ("Call dentist", "Pay bills")     │
│                                                                 │
│  Two execution patterns:                                        │
│  • ONE-TIME: Do once and complete                              │
│  • RECURRING: Repeats on schedule (Mon/Wed/Fri)                │
└─────────────────────────────────────────────────────────────────┘
          │
          │ tracks (1:N, only for recurring)
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      TASK RECORDS                               │
│            (Execution history for recurring tasks)              │
│                                                                 │
│  Example for "Review emails every Mon/Wed/Fri":                │
│  • Dec 16 (Mon): Completed at 9:15am                           │
│  • Dec 18 (Wed): Skipped (holiday)                             │
│  • Dec 20 (Fri): Pending                                       │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Entity Count Summary

| Entity | Relationship | Limit | Status |
|--------|-------------|-------|--------|
| Users | Root | - | ✅ Implemented |
| Values | 1 User : N Values | Max 10 total (incl. archived) | ✅ Implemented |
| Life Visions | 1 Value : N Visions | Unlimited | ✅ Implemented |
| Objectives | 1 Vision : N Objectives | Unlimited | ⏳ To Implement |
| Objective Records | 1 Objective : N Records | 1 per day | ⏳ To Implement |
| Initiatives | 1 Objective : N Initiatives | ~4 per year | ⏳ To Implement |
| Tasks | N:1 Initiative (optional) | Unlimited | ⏳ To Implement |
| Task Records | 1 Task : N Records | 1 per scheduled day | ⏳ To Implement |

---

## 2. Entity Descriptions

### 2.1 Users

**Purpose**: Root entity representing authenticated individuals.

**Key Characteristics**:
- Authentication via email/password
- Optional Telegram integration for notifications
- All personal data is private and encrypted
- Cascade delete removes all child entities

**Relationships**:
- Parent of: Values, Life Visions, Objectives, Tasks (standalone)

---

### 2.2 Values

**Purpose**: Permanent life directions based on ACT therapy. Values are not goals to achieve but qualities of being to manifest continuously.

**Key Characteristics**:
- Maximum 10 total values per user, including archived (cognitive limit; current implementation counts all values, not just active)
- Ordered by priority (1 = most important)
- Categorized by life facet (personal, relationships, work, leisure, growth)
- Soft delete via archive (preserves history)
- Cannot archive if has active Life Visions (block pattern)

**Relationships**:
- Belongs to: User (N:1)
- Inspires: Life Visions (1:N)

**Example Values**:
- "Be fully present in every interaction"
- "Pursue mastery and continuous improvement"
- "Nurture deep, authentic connections"

---

### 2.3 Life Visions

**Purpose**: Long-term aspirational states (5-20 year horizon). Describe what life looks like when fully living a value.

**Key Characteristics**:
- Linked to exactly one Value (the inspiration source)
- Can be reassigned to different Value
- Soft delete via archive
- No limit on quantity

**Relationships**:
- Inspired by: Value (N:1)
- Generates: Objectives (1:N)

**Example Life Visions**:
- Value "Present Father" → Vision "My adult children call me their best friend and confidant"
- Value "Continuous Growth" → Vision "Recognized thought leader with 3 published books"

---

### 2.4 Objectives

**Purpose**: Concrete, measurable annual goals that move toward Life Visions.

**Key Characteristics**:
- Linked to exactly one Life Vision
- Year-scoped (e.g., 2025 objectives)
- Two tracking types:
  - **Outcome**: Count toward target (e.g., "35 books")
  - **Frequency**: Habit consistency (e.g., "5 days/week, 80% compliance")
- Progress auto-calculated from records or manual updates
- Multiple statuses: active, completed, abandoned, paused

**Relationships**:
- Belongs to: Life Vision (N:1)
- Generates: Objective Records (1:N, frequency type only)
- Decomposes into: Initiatives (1:N)

**Example Objectives**:
- Outcome type: "Read 35 technical books in 2025" (target: 35, unit: books)
- Frequency type: "Meditate 5 days per week" (frequency: 5/week, compliance: 80%)

---

### 2.5 Objective Records

**Purpose**: Daily tracking entries for frequency-based objectives.

**Key Characteristics**:
- One record per objective per day (unique constraint)
- Captures: completed (boolean), optional value, notes
- Auto-updates objective's progress metric
- Cannot be future-dated

**Relationships**:
- Belongs to: Objective (N:1)

**Example Record**:
- Objective: "Meditate 5 days/week"
- Date: 2025-12-15
- Completed: true
- Value: 20 (minutes)
- Notes: "Morning session, felt calm afterward"

---

### 2.6 Initiatives

**Purpose**: Quarterly projects that decompose annual objectives into actionable chunks.

**Key Characteristics**:
- Linked to exactly one Objective
- Scoped to quarter (Q1, Q2, Q3, Q4) and year
- Has expected outcome description
- Progress percentage (0-100%)
- Statuses: pending, in_progress, completed, cancelled

**Relationships**:
- Belongs to: Objective (N:1)
- Generates: Tasks (1:N)

**Example Initiatives**:
- Objective: "Read 35 books"
  - Q1 Initiative: "Complete 'Designing Data-Intensive Applications'"
  - Q2 Initiative: "Read 3 books on distributed systems"

---

### 2.7 Tasks

**Purpose**: Daily executable actions. The lowest abstraction level where actual work happens.

**Key Characteristics**:
- Can belong to Initiative (derived) OR be standalone (quick capture)
- Two execution types:
  - **One-time**: Complete once and done
  - **Recurring**: Repeats on pattern (daily, weekly days, monthly)
- Priority levels: high, medium, low
- Scheduled date (when to do) vs due date (deadline)
- Optional tags for organization (freeform text array)
- Optional direct value link (for standalone tasks)

**Relationships**:
- Optionally from: Initiative (N:1, nullable)
- Belongs to: User (N:1, always - for ownership)
- Generates: Task Records (1:N, recurring only)

**Example Tasks**:
- Derived: "Read chapter 5 of DDIA" (from Initiative "Complete DDIA book")
- Standalone: "Call dentist for appointment" (no initiative, goes to Inbox)
- Recurring: "Review and process email inbox" (Mon/Wed/Fri pattern)

---

### 2.8 Task Records

**Purpose**: Execution history for recurring tasks. Each scheduled instance gets a record.

**Key Characteristics**:
- System-generated based on recurrence pattern
- One record per task per scheduled date
- Can be: completed (with timestamp), skipped (intentionally), or pending
- Optional notes for context

**Relationships**:
- Belongs to: Task (N:1)

**Example Records**:
- Task: "Review emails" (Mon/Wed/Fri)
  - Monday Dec 16: Completed at 9:15am
  - Wednesday Dec 18: Skipped (noted: "Holiday")
  - Friday Dec 20: Pending

---

## 3. Data Flows

### 3.1 Top-Down Planning Flow

The primary workflow for strategic planning:

```
USER JOURNEY: Annual Planning Session

1. REFLECT ON VALUES
   └─► Review existing values, archive outdated ones, add new ones
   └─► Reorder by current life priorities
   └─► Result: 5-10 prioritized values

2. ENVISION THE FUTURE
   └─► For each value, ask "What does success look like in 10 years?"
   └─► Create/update Life Visions
   └─► Result: 2-3 visions per value

3. SET ANNUAL OBJECTIVES
   └─► For each vision, ask "What can I achieve THIS YEAR?"
   └─► Define measurable targets (outcome or frequency)
   └─► Result: 3-5 key objectives for the year

4. PLAN THE QUARTER
   └─► For each objective, create Q1 initiative
   └─► Define expected outcomes for the quarter
   └─► Result: 1 initiative per objective per quarter

5. BREAK INTO TASKS
   └─► For each initiative, identify first actions
   └─► Schedule tasks for the week
   └─► Result: Daily task list aligned with values
```

### 3.2 Bottom-Up Execution Flow

The daily workflow for task execution:

```
USER JOURNEY: Daily Execution

1. MORNING REVIEW
   └─► Dashboard shows today's tasks
   └─► Frequency objectives show "X/5 days this week"
   └─► Quick stats: tasks pending, objectives on track

2. CAPTURE NEW TASKS
   └─► FAB button → Quick Add Sheet
   └─► Enter task title (required)
   └─► Optional: link to initiative, set due date, add tags
   └─► Unlinked tasks go to "Inbox" for later organization

3. EXECUTE TASKS
   └─► Check off completed tasks (optimistic update)
   └─► Progress updates IMMEDIATELY in UI
   └─► Task completion → Initiative progress recalculated

4. LOG FREQUENCY HABITS
   └─► Dashboard widget shows frequency objectives
   └─► Click to log today's completion
   └─► Heatmap updates instantly
   └─► Weekly compliance percentage updates

5. EVENING REFLECTION
   └─► Review completed vs planned
   └─► Reschedule incomplete tasks
   └─► Optional: add Momento (emotional reflection)
```

### 3.3 Progress Calculation Flow

How metrics propagate up the hierarchy:

```
TASK COMPLETED
    │
    ▼
TASK RECORD created (for recurring tasks)
    │
    ▼
INITIATIVE progress_pct recalculated
    │  Formula: (completed_tasks / total_tasks) * 100
    │
    ▼
OBJECTIVE metrica_actual updated
    │  • Outcome type: Manual or from initiative completion
    │  • Frequency type: COUNT(completed records this period)
    │
    ▼
DASHBOARD reflects new progress
    │  • Progress bars update
    │  • Heatmap shows new completion
    │  • Weekly/monthly stats refresh
```

### 3.4 Archive Cascade Flow

What happens when archiving entities:

```
ARCHIVE VALUE
    │
    ├─► CHECK: Has active Life Visions?
    │     ├─► YES → BLOCK (return error)
    │     └─► NO  → PROCEED
    │
    ▼
VALUE status = 'archived'
VALUE archived_at = NOW()

─────────────────────────────────────

ARCHIVE LIFE VISION
    │
    ├─► CHECK: Has active Objectives?
    │     ├─► YES → BLOCK (return error)
    │     └─► NO  → PROCEED
    │
    ▼
LIFE_VISION status = 'archived'

─────────────────────────────────────

ARCHIVE OBJECTIVE
    │
    ├─► CHECK: Has active Initiatives?
    │     ├─► YES → BLOCK (return error)
    │     └─► NO  → PROCEED
    │
    ▼
OBJECTIVE status = 'archived'
(Records are preserved for history)

─────────────────────────────────────

And so on for Initiatives → Tasks...
```

---

## 4. Use Cases

### 4.1 Strategic Planning Use Cases

#### UC-1: Annual Values Review
**Actor**: User
**Trigger**: New year or major life change
**Flow**:
1. User opens Values page
2. Reviews each value: still relevant?
3. Archives values no longer aligned
4. Adds new values (up to 10 limit)
5. Reorders by drag-and-drop to reflect priorities
6. System updates display_order for all values

**Outcome**: Updated, prioritized value set for the year

---

#### UC-2: Life Vision Creation
**Actor**: User
**Trigger**: New value created or annual review
**Flow**:
1. User selects a Value
2. Clicks "Add Life Vision"
3. Writes aspirational description (5-20 year horizon)
4. System validates content length (10-500 chars)
5. Vision linked to selected value

**Outcome**: New aspirational target connected to core value

---

#### UC-3: Objective Setting
**Actor**: User
**Trigger**: Annual planning or quarterly review
**Flow**:
1. User selects a Life Vision
2. Clicks "Create Objective"
3. Enters title and description
4. Selects tracking type:
   - Outcome: Sets target number and unit
   - Frequency: Sets days per week and compliance target
5. System creates objective for current year

**Outcome**: Measurable annual goal linked to life vision

---

#### UC-4: Quarterly Initiative Planning
**Actor**: User
**Trigger**: Start of quarter
**Flow**:
1. User views active objectives
2. For each objective, creates Q[N] initiative
3. Defines expected outcome for the quarter
4. System sets quarter and year automatically
5. User can now create tasks under initiative

**Outcome**: Quarterly project decomposing annual goal

---

### 4.2 Daily Execution Use Cases

#### UC-5: Quick Task Capture
**Actor**: User
**Trigger**: Thought of something to do
**Flow**:
1. User clicks FAB (floating action button)
2. Quick Add sheet opens
3. Enters task title (required)
4. Optionally: links to initiative, sets due date, adds tags
5. Clicks "Add Task"
6. Task appears in list immediately (optimistic update)

**Outcome**: Task captured with minimal friction

---

#### UC-6: Task Completion
**Actor**: User
**Trigger**: Finished a task
**Flow**:
1. User checks the checkbox on task card
2. UI immediately shows task as completed (optimistic)
3. Backend updates task status and completion timestamp
4. If task belongs to initiative: progress recalculated
5. Dashboard stats update

**Outcome**: Task marked done with instant feedback

---

#### UC-7: Frequency Habit Logging
**Actor**: User
**Trigger**: Daily habit check-in
**Flow**:
1. Dashboard shows frequency objectives widget
2. User sees "Meditate: 4/5 days this week"
3. Clicks "Log Today"
4. System creates objective record for today
5. Heatmap updates with new green square
6. Progress metric updates (metrica_actual)

**Outcome**: Daily habit tracked with visual feedback

---

#### UC-8: Recurring Task Execution
**Actor**: User
**Trigger**: Scheduled task day arrives
**Flow**:
1. System pre-generates task records for recurring tasks
2. User sees "Review emails" in today's tasks
3. Completes the task
4. Task record updated with completion timestamp
5. Next occurrence appears on next scheduled day

**Outcome**: Recurring routine tracked consistently

---

### 4.3 Review Use Cases

#### UC-9: Weekly Review
**Actor**: User
**Trigger**: End of week (e.g., Sunday)
**Flow**:
1. User opens review section
2. System shows:
   - Tasks completed this week: 12/18 (67%)
   - Frequency objectives compliance: 4/5 days
   - Incomplete tasks from this week
3. User drags incomplete tasks to next week
4. Optionally adds reflection notes
5. Plans priorities for next week

**Outcome**: Week closed out, next week planned

---

#### UC-10: Quarterly Review
**Actor**: User
**Trigger**: End of quarter
**Flow**:
1. User opens quarterly review wizard
2. System shows objective progress:
   - "Read 35 books": 12/35 (34%) ⚠️
   - "Meditate 5 days/week": 78% compliance ✅
3. User reflects: What worked? What didn't?
4. Creates initiatives for next quarter
5. Adjusts objectives if needed (change target, pause, abandon)

**Outcome**: Quarter reviewed, next quarter planned

---

### 4.4 Data Query Use Cases

#### UC-11: Filter Tasks by Tags
**Actor**: User
**Trigger**: Want to see specific context
**Flow**:
1. User opens Tasks page
2. Clicks tag filter dropdown
3. Selects "Work" tag
4. System filters to show only tasks tagged "Work"
5. Can combine with status filter (pending only)

**Outcome**: Focused view of relevant tasks

---

#### UC-12: View Value Alignment
**Actor**: User
**Trigger**: Checking if actions align with values
**Flow**:
1. User selects a Value
2. System shows:
   - Life Visions inspired by this value
   - Objectives derived from those visions
   - Active initiatives under those objectives
   - Tasks in progress
3. User sees full chain from value to daily action

**Outcome**: Visibility into value-action alignment

---

## 5. Architecture Alignment

### 5.1 Domain Structure

The entity system maps to the existing Rafiki architecture:

```
DOMAIN HIERARCHY (6 Levels):

Level 1: userbus (ROOT)
           │
           ├─► imports: NONE
           │
Level 2: valuebus (CHILD)
           │
           ├─► imports: userbus.ExtBusiness
           │
Level 3: lifevisionbus (CHILD)
           │
           ├─► imports: valuebus.ExtBusiness
           │
Level 4: objetivobus (CHILD)
           │
           ├─► imports: lifevisionbus.ExtBusiness
           │
Level 5: iniciativabus (CHILD)
           │
           ├─► imports: objetivobus.ExtBusiness
           │
Level 6: tareabus (CHILD, dual parent)
           │
           └─► imports: userbus.ExtBusiness (for standalone tasks)
                       + iniciativabus.ExtBusiness (for derived tasks)
```

### 5.2 Architectural Patterns Used

| Pattern | Description | Where Applied |
|---------|-------------|---------------|
| **One-Directional Imports** | Child imports parent interface only | All domain relationships |
| **Interface Contracts** | ExtBusiness interfaces for dependency injection | All domain constructors |
| **Soft Delete** | status + archived_at instead of DELETE | Values, Life Visions, Objectives |
| **Archive Block** | Database triggers prevent orphaning | All parent-child relationships |
| **Delegate Events** | Async cascade operations | Delete propagation |
| **Business Types** | Validated domain types | All entity fields |
| **Query Domains** | Read-optimized views | Dashboard aggregations |

### 5.3 Alignment Status

**ALIGNED ✅** - The entity system follows all architectural rules:

1. ✅ **One-directional imports**: Each domain only imports its parent
2. ✅ **No circular dependencies**: Hierarchy flows top-down
3. ✅ **Dependency injection**: Parents injected via constructor
4. ✅ **Business types**: All domain values use validated types
5. ✅ **Soft delete pattern**: Consistent with existing entities
6. ✅ **Transaction support**: NewWithTx() pattern for atomic operations

**Simplified by MVP Decision**:
- ❌ No N:M relationships (Roles removed)
- ❌ No Bridge Domains needed
- ✅ Simpler implementation path

---

## 6. Implementation Challenges

### 6.1 Technical Challenges

#### Challenge 1: Deep Domain Nesting (6 Levels)

**Problem**: Constructor chain becomes complex at depth 6.

**Solution**:
- Follow existing pattern (already working for 3 levels)
- Use factory function in main.go to construct in order
- Each domain only knows immediate parent

```go
// Construction order in main.go
userBus := userbus.NewBusiness(...)
valueBus := valuebus.NewBusiness(log, userBus, ...)
lifevisionBus := lifevisionbus.NewBusiness(log, valueBus, ...)
objetivoBus := objetivobus.NewBusiness(log, lifevisionBus, ...)
iniciativaBus := iniciativabus.NewBusiness(log, objetivoBus, ...)
tareaBus := tareabus.NewBusiness(log, userBus, iniciativaBus, ...)
```

---

#### Challenge 2: Dual Parentage for Tasks

**Problem**: Tasks can belong to Initiative OR be standalone (owned by User directly).

**Solution**:
- `iniciativa_id` is nullable
- Validation logic checks both paths:
  - If `iniciativa_id` is set: validate through initiative chain
  - If `iniciativa_id` is NULL: validate directly against user

```go
func (b *Business) Create(ctx context.Context, nt NewTarea) (Tarea, error) {
    // Always validate user
    user, _ := b.userBus.QueryByID(ctx, nt.UserID)

    // Conditionally validate initiative
    if nt.IniciativaID != nil {
        iniciativa, _ := b.iniciativaBus.QueryByID(ctx, *nt.IniciativaID)
        if iniciativa.UserID != nt.UserID {
            return Tarea{}, ErrNotOwner
        }
    }

    // Create task...
}
```

---

#### Challenge 3: Optimistic Updates in Frontend

**Problem**: User expects immediate feedback when checking off tasks.

**Solution**:
- Migrate to React Query (TanStack Query)
- Implement optimistic updates pattern
- Rollback on server error

```typescript
const completeTask = useMutation({
    mutationFn: (taskId) => api.tasks.complete(taskId),
    onMutate: async (taskId) => {
        // Cancel outgoing refetches
        await queryClient.cancelQueries({ queryKey: ['tasks'] });

        // Snapshot current state
        const previous = queryClient.getQueryData(['tasks']);

        // Optimistically update
        queryClient.setQueryData(['tasks'], (old) => ({
            ...old,
            items: old.items.map(t =>
                t.id === taskId ? { ...t, status: 'completed' } : t
            )
        }));

        return { previous };
    },
    onError: (err, taskId, context) => {
        // Rollback on error
        queryClient.setQueryData(['tasks'], context.previous);
    }
});
```

---

#### Challenge 4: Frequency Tracking Visualization

**Problem**: Users need GitHub-style heatmap for habit tracking.

**Solution**:
- Use `react-activity-calendar` library
- Backend provides aggregated activity data
- Frontend renders with appropriate theming

```typescript
// API response structure
interface ActivityData {
    date: string;      // "2025-12-15"
    count: number;     // completions that day
    level: 0|1|2|3|4;  // intensity for color
}

// Component
<ActivityCalendar
    data={activityData}
    theme={{
        dark: ['#161b22', '#0e4429', '#006d32', '#26a641', '#39d353']
    }}
/>
```

---

#### Challenge 5: Recurrence Pattern Storage

**Problem**: Tasks can recur on complex patterns (Mon/Wed/Fri, 10th of month).

**Solution**:
- Store as JSONB in database
- Business type with validation
- System generates task records ahead of time

```go
type RecurrencePattern struct {
    Type      string   `json:"type"`       // "weekly" | "monthly"
    DaysWeek  []int    `json:"days_week"`  // [1,3,5] = Mon/Wed/Fri
    DayMonth  *int     `json:"day_month"`  // 10 = 10th of month
}

func (r RecurrencePattern) Validate() error {
    switch r.Type {
    case "weekly":
        if len(r.DaysWeek) == 0 {
            return errors.New("weekly requires days_week")
        }
    case "monthly":
        if r.DayMonth == nil {
            return errors.New("monthly requires day_month")
        }
    }
    return nil
}
```

---

### 6.2 UX Challenges

#### Challenge 6: Deep Hierarchy Without Overwhelm

**Problem**: 6-level hierarchy could overwhelm users.

**Solution**:
- Progressive disclosure: show 2 levels at a time
- Breadcrumbs for context awareness
- Collapsible sections
- Quick capture bypasses hierarchy

---

#### Challenge 7: Quick Capture vs. Organization

**Problem**: Users want to capture tasks fast BUT also want organization.

**Solution**:
- FAB for instant capture (just title required)
- "Inbox" for unorganized tasks
- Later: drag tasks from Inbox to initiatives
- Optional tags for lightweight organization

---

### 6.3 DevOps Challenges

#### Challenge 8: Safe Database Migrations

**Problem**: New tables with foreign keys could break if migration fails.

**Solution**:
- Pre-deployment backup script (MANDATORY)
- Phased deployment (3 separate deployments)
- Rollback script ready
- Health checks after each migration

**Deployment Order**:
1. Phase 1: Objectives + Objective Records
2. Phase 2: Initiatives
3. Phase 3: Tasks + Task Records

---

## 7. Implementation Phases

### Phase 1: Objectives Layer (Week 1-2)

**Backend**:
- Business types: `objetivotitle`, `trackingtype`, `metricatarget`, `objetivostatus`
- Domain: `objetivobus` with full CRUD
- Domain: `objetivoregistrobus` for frequency tracking
- API endpoints: 8-10 new endpoints
- Database migration: 2 new tables + triggers

**Frontend**:
- Types: Objective, ObjectiveRecord interfaces
- Components: ObjectiveCard, ObjectiveForm, ProgressBar
- Page: /objectives with filtering
- Integration: React Query setup

---

### Phase 2: Initiatives Layer (Week 3)

**Backend**:
- Business types: `iniciativatitle`, `trimestre`, `progresopct`, `iniciativastatus`
- Domain: `iniciativabus` with full CRUD
- API endpoints: 6-8 new endpoints
- Database migration: 1 new table + triggers

**Frontend**:
- Types: Initiative interface
- Components: InitiativeCard, InitiativeForm
- Page: /initiatives or embedded in Objective detail
- Quarterly timeline view

---

### Phase 3: Tasks Layer (Week 4-5)

**Backend**:
- Business types: `tareatitle`, `tareatipo`, `prioridad`, `tareastatus`, `recurrencia`
- Domain: `tareabus` with dual parentage
- Domain: `tarearegistrobus` for recurring tasks
- API endpoints: 10-12 new endpoints
- Database migration: 2 new tables + triggers + JSONB

**Frontend**:
- Types: Task, TaskRecord, RecurrencePattern interfaces
- Components: TaskCard, TaskForm, QuickAddFAB, RecurrenceEditor
- Page: /tasks with multiple views (list, calendar)
- Inbox for standalone tasks

---

### Phase 4: Integration & Polish (Week 6)

**Backend**:
- Query domains for dashboard aggregations
- Performance optimization (indexes, query tuning)
- End-to-end testing

**Frontend**:
- Dashboard widgets (frequency tracking, progress overview)
- Calendar heatmap integration
- Mobile responsiveness
- Error states and loading skeletons

---

## 8. Future Considerations (Post-MVP)

### 8.1 Roles System
If user feedback shows demand for cross-cutting organization:
- Add `roles` table with optional value link
- N:M with objectives and tasks
- Bridge domains for relationships

### 8.2 Review Entities
If users want to persist reflections:
- Add `reviews` table (weekly, quarterly, annual)
- Store reflection notes and insights
- Track review completion streaks

### 8.3 Collaboration
If multi-user support needed:
- Shared objectives within teams
- Task delegation
- Progress sharing

### 8.4 Analytics
For deeper insights:
- Time tracking on tasks
- Productivity patterns
- Value alignment scores
- AI-generated insights

---

## 9. Conclusion

The Rafiki entity system provides a complete framework for ACT-based personal development:

1. **Values** anchor the system in what matters most
2. **Life Visions** paint the aspirational future
3. **Objectives** make the year concrete and measurable
4. **Initiatives** break the year into manageable quarters
5. **Tasks** bring it all down to daily action
6. **Records** track progress and build habits

The architecture is:
- ✅ Aligned with existing patterns
- ✅ Simplified for MVP (no Roles complexity)
- ✅ Extensible for future needs
- ✅ Psychologically sound (ACT-based)

**Total Implementation Estimate**: 6 weeks (1 developer)

---

*Document Version: 2.0*
*Last Updated: December 2025*
*Authors: Multi-Mind Analysis Team*
