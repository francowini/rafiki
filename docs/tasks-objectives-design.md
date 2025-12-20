# Tasks-Objectives Integration Design

**Version 2.0 | December 2025**

This document defines the data architecture and phased implementation plan for Tasks in Rafiki.

---

## 1. Design Decisions (Finalized)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Apply Pattern** | Auto-apply on completion | Less friction, immediate dopamine feedback, 10-second undo window |
| **Recurring Tasks** | Defer entirely | ObjectiveRecord handles frequency habits; add Task Templates later if needed |
| **Contribution** | Include in Phase 1 (1-10) | Users set how much each task contributes to RESULT objectives |
| **Inbox** | Include in Phase 1 | FAB quick capture + minimal Inbox page |
| **Initiatives** | Defer to post-MVP | Not needed for core task functionality |

---

## 2. Data Architecture

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           DATA ARCHITECTURE                              │
└─────────────────────────────────────────────────────────────────────────┘

    users (1)
      │
      ├──────────────────────────────────────────┐
      │                                          │
      ▼                                          ▼
    values (N)                               tasks (N)
      │                                     [Inbox: objective_id = NULL]
      │
      ▼
    life_visions (N)
      │
      │
      ▼
    objectives (N)
      │
      ├─────────────────────┐
      │                     │
      ▼                     ▼
    objective_records   tasks (N)
    (for FREQUENCY)     (for RESULT)
                            │
                            │ Auto-apply on complete:
                            │ objective.current_metric += task.contribution
                            ▼
                        [Progress Updated]
```

### Key Relationships

| Parent | Child | Relationship | On Delete |
|--------|-------|--------------|-----------|
| users | tasks | 1:N | CASCADE |
| objectives | tasks | 1:N (optional) | CASCADE |
| objectives | objective_records | 1:N | CASCADE |

### Task States

```
PENDING ──────► COMPLETED ──────► [Auto-Applied]
    │               │
    │               └──► UNCOMPLETED (via Undo)
    │
    └──────────────► CANCELLED
```

---

## 3. Database Schema

### Tasks Table (Migration V16)

```sql
CREATE TABLE tasks (
    task_id              UUID        PRIMARY KEY,
    user_id              UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    objective_id         UUID        NULL REFERENCES objectives(objective_id) ON DELETE CASCADE,
    title                TEXT        NOT NULL,  -- Encrypted, 3-200 chars
    description          TEXT        NULL,      -- Encrypted, max 2000 chars
    status               TEXT        NOT NULL DEFAULT 'pending'
                                     CHECK (status IN ('pending', 'completed', 'cancelled')),
    contribution         INTEGER     NOT NULL DEFAULT 1 CHECK (contribution >= 1 AND contribution <= 10),
    completed_at         TIMESTAMP   NULL,
    date_created         TIMESTAMP   NOT NULL,
    date_updated         TIMESTAMP   NOT NULL
);

-- Indexes
CREATE INDEX tasks_user_id_idx ON tasks(user_id);
CREATE INDEX tasks_objective_id_idx ON tasks(objective_id);
CREATE INDEX tasks_status_idx ON tasks(status);
CREATE INDEX tasks_inbox_idx ON tasks(user_id) WHERE objective_id IS NULL;
```

### Constraints (Business Layer)

- `completed_at` must be set when `status = completed`
- Only RESULT-type objectives can have tasks that update progress
- FREQUENCY objectives: tasks serve as reminders only (no auto-apply)

---

## 4. Domain Architecture

### Domain Hierarchy

```
userbus (Level 1 - Root)
    │
    ▼
valuebus (Level 2)
    │
    ▼
lifevisionbus (Level 3)
    │
    ▼
objectivebus (Level 4)
    │
    ▼
taskbus (Level 5)  ◄── Only imports objectivebus.ExtBusiness
```

### Architecture Rules

| Rule | Implementation |
|------|----------------|
| **Import Direction** | taskbus → objectivebus only (never reverse) |
| **Interface Contract** | taskbus uses `objectivebus.ExtBusiness` interface |
| **Cascade Delete** | Database FK CASCADE + delegate logging |
| **Transaction** | App layer manages cross-domain transaction |
| **Strong Types** | All values use business/types/ (no primitives) |

### Business Types Needed

| Type | Package | Validation |
|------|---------|------------|
| TaskStatus | `business/types/taskstatus` | pending, completed, cancelled |
| TaskTitle | `business/types/tasktitle` | 3-200 chars, encrypted |
| TaskDescription | `business/types/taskdescription` | Optional, max 2000, encrypted |
| Contribution | `business/types/contribution` | 1-10 integer |

---

## 5. Auto-Apply Flow (Critical)

### Complete Task Transaction

```
┌─────────────────────────────────────────────────────────────┐
│                    APP LAYER (taskapp)                       │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                    BEGIN TRANSACTION
                            │
                ┌───────────┴───────────┐
                │                       │
                ▼                       ▼
        taskbus.Complete()      objectivebus.UpdateProgress()
        - status = completed    - current_metric += contribution
        - completed_at = NOW    (only if RESULT type)
                │                       │
                └───────────┬───────────┘
                            │
                            ▼
                    COMMIT TRANSACTION
                            │
                            ▼
            Response: { task, objective, previousMetric, newMetric }
```

### Uncomplete (Undo) Transaction

```
POST /v1/tasks/:id/uncomplete

- Revert task.status to 'pending'
- Revert objective.current_metric -= contribution
- Only allowed within 10 seconds (enforced by frontend)
```

---

## 6. API Endpoints

### Task CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/tasks` | Create task (Inbox or linked to objective) |
| GET | `/v1/tasks` | List tasks (filter: objective_id, status, inbox) |
| GET | `/v1/tasks/:id` | Get single task |
| PUT | `/v1/tasks/:id` | Update task (title, description, contribution, objective_id) |
| DELETE | `/v1/tasks/:id` | Delete task |

### Task Actions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/tasks/:id/complete` | Complete + auto-apply (returns task + objective) |
| POST | `/v1/tasks/:id/uncomplete` | Undo completion (reverts progress) |

### Query Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/objectives/:id/tasks` | All tasks for an objective |
| GET | `/v1/tasks/inbox` | Inbox tasks (objective_id = NULL) |

---

## 7. Frontend Components

### Component Hierarchy

```
ObjectiveDetail
├── TaskList (tabs: Pending | Completed)
│   └── TaskItem (checkbox, title, contribution badge)
├── TaskForm (Sheet - create/edit)
└── QuickTaskSheet (FAB bottom sheet)

InboxPage
├── TaskList (inbox tasks only)
└── AssignToObjectiveDialog
```

### Key UI Patterns

| Pattern | Implementation |
|---------|----------------|
| **Complete Action** | Checkbox click → API call → Toast with Undo |
| **Undo Pattern** | Toast shows 10 seconds, "Deshacer" button calls uncomplete |
| **Contribution** | Slider (1-10) in TaskForm, Badge in TaskItem |
| **Quick Capture** | FAB → Sheet with title-only input → Creates in Inbox |
| **Assign to Objective** | Select dropdown in TaskForm or AssignDialog |

### Toast Format

```
✓ Tarea completada
  +3 → 28/100 Leer 35 libros
  [Deshacer]
```

---

## 8. Phased Implementation Plan

### Overview

| Phase | Name | Scope | Deliverable |
|-------|------|-------|-------------|
| **Phase 1** | Backend Core | Database + Business Types + Domain + API | Backend deployed, API testable |
| **Phase 2** | Frontend Core | Task components + Objective integration | Full feature usable |
| **Phase 3** | Inbox & Polish | FAB + Inbox page + Assign flow | Complete user experience |

---

## Phase 1: Backend Core

### Objective
Deploy backend with full Task API. Frontend not required. API testable via curl/Postman.

### Deliverables

1. **Database Migration V16**
   - Create `tasks` table with all columns
   - Add indexes for performance
   - FK constraints with CASCADE

2. **Business Types** (4 types)
   - `business/types/taskstatus/taskstatus.go`
   - `business/types/tasktitle/tasktitle.go`
   - `business/types/taskdescription/taskdescription.go`
   - `business/types/contribution/contribution.go`

3. **Task Domain** (taskbus)
   - `business/domain/taskbus/model.go` - Task, NewTask, UpdateTask
   - `business/domain/taskbus/taskbus.go` - Business logic
   - `business/domain/taskbus/filter.go` - QueryFilter
   - `business/domain/taskbus/order.go` - OrderBy options
   - `business/domain/taskbus/event.go` - Delegate registration
   - `business/domain/taskbus/stores/taskdb/` - PostgreSQL store

4. **Task App Layer** (taskapp)
   - `app/domain/taskapp/model.go` - API request/response types
   - `app/domain/taskapp/taskapp.go` - HTTP handlers
   - `app/domain/taskapp/route.go` - Route registration

5. **Complete Action** (cross-domain transaction)
   - Complete task + update objective progress atomically
   - Return both task and objective in response

### Key Implementation Details

```go
// taskbus.Complete signature
func (b *Business) Complete(ctx context.Context, task Task) (Task, error)

// taskapp.complete handler - manages transaction
func (a *app) complete(ctx context.Context, r *http.Request) web.Encoder {
    // BEGIN TX
    // taskbus.Complete(task)
    // if task.ObjectiveID != nil && objective.TrackingType.IsResult():
    //     objectivebus.UpdateProgress(objective, +contribution)
    // COMMIT TX
    // Return: { task, objective, previousMetric, newMetric }
}
```

### Acceptance Criteria

- [ ] `POST /v1/tasks` creates task in database
- [ ] `GET /v1/tasks?objective_id=X` returns tasks for objective
- [ ] `GET /v1/tasks/inbox` returns tasks without objective
- [ ] `POST /v1/tasks/:id/complete` updates task AND objective atomically
- [ ] `POST /v1/tasks/:id/uncomplete` reverts both changes
- [ ] Objective deletion cascades to delete tasks
- [ ] All fields encrypted (title, description)

---

## Phase 2: Frontend Core

### Objective
Integrate tasks into Objective Detail page. Users can create, complete, and see tasks for their objectives.

### Deliverables

1. **Type Definitions**
   - Add Task types to `frontend/lib/types.ts`
   - Add task endpoints to `frontend/lib/api.ts`
   - Add query keys to `frontend/lib/query-keys.ts`

2. **React Query Hooks**
   - `frontend/lib/hooks/use-tasks.ts`
   - useTasks, useCreateTask, useCompleteTask, useUncompleteTask

3. **Task Components**
   - `frontend/components/features/tasks/TaskItem.tsx`
   - `frontend/components/features/tasks/TaskList.tsx`
   - `frontend/components/features/tasks/TaskForm.tsx` (Sheet)

4. **Objective Detail Integration**
   - Add TaskList to ObjectiveDetail page
   - Tabs: Pending | Completed
   - "+ Nueva tarea" button opens TaskForm

5. **Complete + Undo Pattern**
   - Checkbox triggers complete mutation
   - Toast with progress update and Undo button
   - Undo button calls uncomplete within 10s window

### Key Implementation Details

```typescript
// Complete mutation with optimistic update
const completeTask = useMutation({
  mutationFn: (taskId) => api.tasks.complete(taskId),
  onMutate: async (taskId) => {
    // Optimistically update task + objective
  },
  onSuccess: (response) => {
    // Show toast with undo
    toast({
      title: "Tarea completada",
      description: `+${response.contribution} → ${response.newMetric}/${response.objective.targetMetric}`,
      action: <Button onClick={() => uncomplete(taskId)}>Deshacer</Button>,
      duration: 10000,
    });
  },
});
```

### Acceptance Criteria

- [ ] Objective detail shows task list with tabs
- [ ] Users can create tasks from objective detail
- [ ] Checkbox completes task and updates progress
- [ ] Toast shows progress change with Undo button
- [ ] Undo reverts both task and objective
- [ ] Contribution slider (1-10) in task form
- [ ] Task form validates title (3-200 chars)

---

## Phase 3: Inbox & Polish

### Objective
Add FAB quick capture and Inbox page for task organization.

### Deliverables

1. **Quick Capture FAB**
   - `frontend/components/features/tasks/QuickTaskFAB.tsx`
   - `frontend/components/features/tasks/QuickTaskSheet.tsx`
   - FAB visible on objectives page
   - Sheet with title-only input, creates in Inbox

2. **Inbox Page**
   - `frontend/app/(dashboard)/inbox/page.tsx`
   - List of tasks without objective
   - Assign to objective action

3. **Assign to Objective Dialog**
   - `frontend/components/features/tasks/AssignObjectiveDialog.tsx`
   - Objective selector dropdown
   - Updates task.objective_id

4. **Navigation Update**
   - Add "Inbox" to sidebar navigation
   - Badge showing inbox count

### Acceptance Criteria

- [ ] FAB appears on objectives page
- [ ] Tapping FAB opens quick capture sheet
- [ ] Quick capture creates task in Inbox
- [ ] Inbox page shows unassigned tasks
- [ ] Users can assign inbox tasks to objectives
- [ ] Assigned tasks can be completed (auto-apply works)
- [ ] Sidebar shows Inbox with count badge

---

## 9. Architecture Compliance Checklist

### Before Implementation

- [ ] taskbus imports only objectivebus.ExtBusiness
- [ ] No objectivebus imports of taskbus (one-directional)
- [ ] All values use strong types from business/types/
- [ ] Transaction managed at app layer, not business layer
- [ ] Delegate registered for objective.deleted event
- [ ] Database FK uses ON DELETE CASCADE

### Code Review Checkpoints

- [ ] No primitive types in domain models (int, string for business values)
- [ ] Error messages use errs package patterns
- [ ] Logging follows foundation/logger patterns
- [ ] API follows existing route patterns in mux.go

---

## 10. Multi-Mind Execution Guide

### For Backend Implementation (Phase 1)

Run: `/multi-mind implement Phase 1 of Tasks - Backend Core`

Expected output:
- Migration SQL file
- 4 business type packages
- taskbus domain package
- taskapp app layer package
- Route registration in mux.go

### For Frontend Implementation (Phase 2)

Run: `/multi-mind implement Phase 2 of Tasks - Frontend Core`

Expected output:
- Type definitions in types.ts
- API client extensions
- React Query hooks
- TaskItem, TaskList, TaskForm components
- ObjectiveDetail integration

### For Inbox Implementation (Phase 3)

Run: `/multi-mind implement Phase 3 of Tasks - Inbox & Polish`

Expected output:
- QuickTaskFAB and QuickTaskSheet components
- Inbox page
- AssignObjectiveDialog component
- Navigation updates

---

## 11. Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | Dec 2025 | Initial design from Multi-Mind analysis |
| 2.0 | Dec 2025 | Simplified to auto-apply, 3-phase plan, architecture compliance |

---

*Updated by Multi-Mind Analysis Team*
*Decisions: Auto-apply, Defer recurring, Include contribution, Include Inbox*
