# Rafiki Entity Implementation Roadmap

**Version 1.0 | December 2025**

This document provides the technical implementation roadmap for the new entity system, including architecture details, challenges, and solutions.

---

## 1. Architecture Summary

### 1.1 Domain Dependency Graph

```
═══════════════════════════════════════════════════════════════════
                    SIMPLIFIED MVP ARCHITECTURE
                    (No Roles - Tags Only)
═══════════════════════════════════════════════════════════════════

ROOT DOMAIN:
┌─────────────┐
│   userbus   │ ← No imports (root)
└─────────────┘
       │
       │ imports userbus.ExtBusiness
       ▼
┌─────────────┐
│  valuebus   │ (Level 2) ✅ IMPLEMENTED
└─────────────┘
       │
       │ imports valuebus.ExtBusiness
       ▼
┌──────────────┐
│lifevisionbus │ (Level 3) ✅ IMPLEMENTED
└──────────────┘
       │
       │ imports lifevisionbus.ExtBusiness
       ▼
┌──────────────┐
│ objetivobus  │ (Level 4) ⏳ TO IMPLEMENT
└──────────────┘
       │
       ├──────────────────────────────┐
       │                              │
       │ imports objetivobus          │ 1:N relationship
       ▼                              ▼
┌──────────────┐            ┌──────────────────┐
│iniciativabus │ (Level 5)  │objetivoregistrobus│
└──────────────┘            └──────────────────┘
       │
       │ imports iniciativabus.ExtBusiness
       ▼
┌─────────────┐
│  tareabus   │ (Level 6) - DUAL PARENT
└─────────────┘
       │       ↖
       │        └── also imports userbus.ExtBusiness
       │            (for standalone tasks)
       │
       │ 1:N relationship
       ▼
┌──────────────────┐
│ tarearegistrobus │
└──────────────────┘

═══════════════════════════════════════════════════════════════════
```

### 1.2 Key Architectural Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Roles | **Removed for MVP** | Simplifies N:M complexity, use tags instead |
| Standalone Tasks | **Allowed** | Quick capture is essential for UX |
| Archive Behavior | **Block Pattern** | Prevents orphaned data, matches existing pattern |
| Progress Updates | **Optimistic** | Better UX, requires React Query |
| Tracking Storage | **Separate Records Tables** | Cleaner queries, better indexing |
| Recurrence | **JSONB Pattern** | Flexible, validated in business layer |

---

## 2. Business Types Required

### 2.1 New Types to Create

Location: `/Users/francowini/Documents/rafiki/business/types/`

```
business/types/
├── objetivotitle/
│   └── objetivotitle.go      # 5-200 chars validation
├── trackingtype/
│   └── trackingtype.go       # Enum: resultado, frecuencia
├── metricatarget/
│   └── metricatarget.go      # Integer > 0
├── metricaunidad/
│   └── metricaunidad.go      # String unit (libros, sesiones, km)
├── objetivostatus/
│   └── objetivostatus.go     # Enum: activo, completado, abandonado, pausado
├── frecuenciatype/
│   └── frecuenciatype.go     # Enum: daily, n_por_semana, n_por_mes
├── iniciativatitle/
│   └── iniciativatitle.go    # 5-200 chars validation
├── trimestre/
│   └── trimestre.go          # Enum: Q1, Q2, Q3, Q4
├── progresopct/
│   └── progresopct.go        # 0-100 validation
├── iniciativastatus/
│   └── iniciativastatus.go   # Enum: pendiente, en_progreso, completada, cancelada
├── tareatitle/
│   └── tareatitle.go         # 3-200 chars validation
├── tareatipo/
│   └── tareatipo.go          # Enum: unica, recurrente
├── prioridad/
│   └── prioridad.go          # Enum: alta, media, baja
├── tareastatus/
│   └── tareastatus.go        # Enum: pendiente, en_progreso, completada, cancelada
└── recurrencia/
    └── recurrencia.go        # JSONB struct with validation
```

### 2.2 Implementation Pattern Reference

All new business types must follow the established pattern documented in `CLAUDE.md` under "Business Types with Validation". Key requirements:

**For enum types** (objetivostatus, trackingtype, etc.):
- Reference: `business/types/entitystatus/entitystatus.go`
- Required methods: `Value()`, `String()`, `Equal()`, `MarshalText()`, `Parse()`, `MustParse()`
- Export a `Set` map for validation

**For validated primitives** (objetivotitle, progresopct, etc.):
- Reference: `business/types/content/content.go`, `business/types/name/name.go`
- Required methods: `Value()`, `String()`, `Equal()`, `MarshalText()`, `Parse()`, `MustParse()`

**For JSONB structs** (recurrencia):
- Reference: Implement as struct with `Validate()` method
- Store as JSONB in PostgreSQL
- Parse from JSON bytes with validation

---

## 3. Domain Implementation

### 3.1 Objetivobus Structure

```
business/domain/objetivobus/
├── model.go              # Objetivo, NewObjetivo, UpdateObjetivo structs
├── objetivobus.go        # Business logic with ExtBusiness interface
├── filter.go             # QueryFilter struct
├── order.go              # Ordering options
├── event.go              # Delegate events
└── stores/
    └── objetivodb/
        ├── objetivodb.go # PostgreSQL store implementation
        └── model.go      # Database model (dbObjetivo)
```

### 3.2 Objetivobus Model Fields

**Objetivo** (main entity):
- `ID`, `UserID`, `LifeVisionID` (uuid.UUID)
- `Title` (objetivotitle type, 5-200 chars)
- `Description` (*string, optional)
- `Year` (int)
- `TrackingType` (enum: resultado, frecuencia)
- `MetricaTarget`, `MetricaActual` (int)
- `MetricaUnidad` (string: libros, sesiones, km, etc.)
- `Frecuencia` (*enum, only if TrackingType = frecuencia)
- `FrecuenciaN` (*int, e.g., 5 for "5 days per week")
- `CumplimientoTargetPct` (*int, e.g., 80 for "80% compliance")
- `Status` (objetivostatus enum)
- `ArchivedAt`, `DateCreated`, `DateUpdated` (time.Time)

**Reference pattern**: `business/domain/lifevisionbus/model.go`

### 3.3 Objetivobus Business Logic

**Key behaviors**:
- Imports `lifevisionbus.ExtBusiness` for parent validation
- Create: Validates life vision exists, is active, and belongs to user
- Archive: Checks for active children (iniciativas) - block pattern
- UpdateMetricaActual: Called by registro creation to update progress

**Reference pattern**: `business/domain/lifevisionbus/lifevisionbus.go`

### 3.4 Tareabus with Dual Parentage

**Special pattern**: Tasks can be standalone OR derived from initiatives.

**Key behaviors**:
- Imports BOTH `userbus.ExtBusiness` AND `iniciativabus.ExtBusiness`
- Create always validates user exists and is enabled
- If `IniciativaID` is provided, validates initiative exists and belongs to same user
- If `IniciativaID` is nil, task goes to "Inbox" (standalone)

**Reference pattern**: `business/domain/valuebus/valuebus.go` (for user validation)

---

## 4. API Endpoints

### 4.1 Objectives API

```
POST   /v1/objetivos                    Create objective
GET    /v1/objetivos                    List objectives (filtered)
GET    /v1/objetivos/{id}               Get objective by ID
PUT    /v1/objetivos/{id}               Update objective
POST   /v1/objetivos/{id}/archive       Archive objective
POST   /v1/objetivos/{id}/restore       Restore objective
PUT    /v1/objetivos/{id}/status        Change status (complete, pause, etc.)

# Objective Records (frequency tracking)
POST   /v1/objetivos/{id}/registros     Log completion for a day
GET    /v1/objetivos/{id}/registros     Get all records (for heatmap)
PUT    /v1/objetivos/{id}/registros/{fecha}  Update record for date
```

### 4.2 Initiatives API

```
POST   /v1/iniciativas                  Create initiative
GET    /v1/iniciativas                  List initiatives (filtered)
GET    /v1/iniciativas/{id}             Get initiative by ID
PUT    /v1/iniciativas/{id}             Update initiative
PUT    /v1/iniciativas/{id}/status      Change status
DELETE /v1/iniciativas/{id}             Delete initiative (soft)
```

### 4.3 Tasks API

```
POST   /v1/tareas                       Create task (derived or standalone)
GET    /v1/tareas                       List tasks (filtered by date, status, tags)
GET    /v1/tareas/{id}                  Get task by ID
PUT    /v1/tareas/{id}                  Update task
POST   /v1/tareas/{id}/complete         Mark as completed
POST   /v1/tareas/{id}/reschedule       Change scheduled date
DELETE /v1/tareas/{id}                  Delete task (soft)

# Task Records (recurring task tracking)
GET    /v1/tareas/{id}/registros        Get all records for task
POST   /v1/tareas/{id}/registros        Complete/skip a scheduled instance
PUT    /v1/tareas/{id}/registros/{fecha}  Update record for date

# Activity aggregation (for heatmap)
GET    /v1/tareas/activity              Get daily completion counts
```

---

## 5. Frontend Implementation

### 5.1 React Query Setup

**Requirements**:
- Install `@tanstack/react-query` for data fetching and caching
- Configure QueryClient with appropriate stale/cache times
- Create query key factory for consistent cache invalidation

**Key patterns**:
- Query keys: hierarchical structure (e.g., `['objectives', 'list', filters]`)
- Mutations: optimistic updates with rollback on error
- Cache invalidation: invalidate related queries on mutations

**Reference**: TanStack Query documentation

### 5.2 Optimistic Updates Pattern

**For task completion** (and similar instant-feedback actions):
1. `onMutate`: Cancel in-flight queries, snapshot current state, update cache optimistically
2. `onError`: Rollback to previous state, show error toast
3. `onSettled`: Invalidate related queries to ensure consistency
4. `onSuccess`: Show success feedback

### 5.3 Calendar Heatmap

**Requirements**:
- Library: `react-activity-calendar`
- Data format: `{ date: string, count: number, level: 0-4 }`
- Theme: Match application light/dark mode

**Used for**: Frequency objective tracking visualization

### 5.4 Quick Task Capture FAB

**Requirements**:
- Floating action button (fixed position, bottom-right)
- Keyboard shortcut: Cmd+K / Ctrl+K
- Sheet component for quick input form
- Minimal required fields: title only
- Tasks without initiative go to "Inbox"

**Reference components**: `frontend/components/ui/sheet.tsx`, `frontend/components/ui/button.tsx`

---

## 6. DevOps Requirements

### 6.1 Pre-Deployment Checklist

Before deploying new entities:

- [ ] Backup script created and tested (`devops/backup-db.sh`)
- [ ] Rollback script created and tested (`devops/rollback-migration.sh`)
- [ ] Deploy script updated with backup step
- [ ] Manual backup created before deployment
- [ ] Migration SQL reviewed and tested locally
- [ ] All business types implemented
- [ ] Integration tests passing

### 6.2 Deployment Phases

**Phase 1: Objectives (Migration v13)**
- Tables: `objetivos`, `objetivo_registros`
- Triggers: Archive block for life visions
- Indexes: All foreign keys and common queries

**Phase 2: Initiatives (Migration v14)**
- Tables: `iniciativas`
- Triggers: Archive block for objectives
- Indexes: Quarter and year filtering

**Phase 3: Tasks (Migration v15)**
- Tables: `tareas`, `tarea_registros`
- Triggers: Archive block for initiatives
- Indexes: Date, status, and tags filtering

### 6.3 Monitoring

After each deployment:
- Check migration applied successfully
- Verify new tables exist
- Create test entity via API
- Monitor error logs for 30 minutes
- Check query performance baselines

---

## 7. Timeline Summary

| Week | Focus | Deliverables |
|------|-------|--------------|
| 1 | Business Types | All 15+ types implemented |
| 2 | Objectives Backend | objetivobus, objetivoregistrobus, API |
| 3 | Objectives Frontend | Components, pages, React Query setup |
| 4 | Initiatives | Full stack implementation |
| 5 | Tasks Backend | tareabus, tarearegistrobus, recurrence |
| 6 | Tasks Frontend | Quick capture, calendar, heatmap |

**Total: 6 weeks** for complete entity system implementation.

---

*Document Version: 1.0*
*Last Updated: December 2025*
