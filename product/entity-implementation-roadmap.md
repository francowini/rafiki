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

### 2.2 Example Business Type Implementation

```go
// business/types/objetivostatus/objetivostatus.go
package objetivostatus

import (
    "fmt"
)

// Set of valid objective statuses
var (
    Activo     = newStatus("activo")
    Completado = newStatus("completado")
    Abandonado = newStatus("abandonado")
    Pausado    = newStatus("pausado")
)

// Set holds all valid statuses for iteration/validation
var Set = map[string]Status{
    Activo.value:     Activo,
    Completado.value: Completado,
    Abandonado.value: Abandonado,
    Pausado.value:    Pausado,
}

// Status represents a validated objective status
type Status struct {
    value string
}

func newStatus(value string) Status {
    return Status{value}
}

// String returns the string representation
func (s Status) String() string {
    return s.value
}

// IsActive returns true if the status is activo
func (s Status) IsActive() bool {
    return s.value == Activo.value
}

// Equal provides support for go-cmp package
func (s Status) Equal(s2 Status) bool {
    return s.value == s2.value
}

// MarshalText implements encoding.TextMarshaler
func (s Status) MarshalText() ([]byte, error) {
    return []byte(s.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (s *Status) UnmarshalText(data []byte) error {
    status, err := Parse(string(data))
    if err != nil {
        return err
    }
    *s = status
    return nil
}

// Parse validates and returns a Status
func Parse(value string) (Status, error) {
    status, exists := Set[value]
    if !exists {
        return Status{}, fmt.Errorf("invalid status %q, valid values: activo, completado, abandonado, pausado", value)
    }
    return status, nil
}

// MustParse panics on invalid status. Use only in tests.
func MustParse(value string) Status {
    status, err := Parse(value)
    if err != nil {
        panic(err)
    }
    return status
}
```

### 2.3 Recurrence Pattern Type

```go
// business/types/recurrencia/recurrencia.go
package recurrencia

import (
    "encoding/json"
    "errors"
    "fmt"
)

// Type constants
const (
    TypeSemanal = "semanal"
    TypeMensual = "mensual"
    TypeDiario  = "diario"
)

// Recurrencia represents a task recurrence pattern
type Recurrencia struct {
    Tipo       string `json:"tipo"`
    DiasSemana []int  `json:"dias_semana,omitempty"` // 1=Monday to 7=Sunday
    DiaMes     *int   `json:"dia_mes,omitempty"`     // 1-31
}

// Validate checks if the recurrence pattern is valid
func (r Recurrencia) Validate() error {
    switch r.Tipo {
    case TypeDiario:
        // No additional validation needed
        return nil

    case TypeSemanal:
        if len(r.DiasSemana) == 0 {
            return errors.New("semanal requires at least one day in dias_semana")
        }
        for _, dia := range r.DiasSemana {
            if dia < 1 || dia > 7 {
                return fmt.Errorf("dias_semana values must be 1-7, got %d", dia)
            }
        }
        return nil

    case TypeMensual:
        if r.DiaMes == nil {
            return errors.New("mensual requires dia_mes")
        }
        if *r.DiaMes < 1 || *r.DiaMes > 31 {
            return fmt.Errorf("dia_mes must be 1-31, got %d", *r.DiaMes)
        }
        return nil

    default:
        return fmt.Errorf("invalid tipo %q, valid values: diario, semanal, mensual", r.Tipo)
    }
}

// MarshalJSON implements json.Marshaler
func (r Recurrencia) MarshalJSON() ([]byte, error) {
    type Alias Recurrencia
    return json.Marshal(Alias(r))
}

// Parse validates and returns a Recurrencia from JSON bytes
func Parse(data []byte) (Recurrencia, error) {
    if len(data) == 0 {
        return Recurrencia{}, errors.New("empty recurrence data")
    }

    var r Recurrencia
    if err := json.Unmarshal(data, &r); err != nil {
        return Recurrencia{}, fmt.Errorf("invalid recurrence JSON: %w", err)
    }

    if err := r.Validate(); err != nil {
        return Recurrencia{}, err
    }

    return r, nil
}

// MustParse panics on error. Use only in tests.
func MustParse(data []byte) Recurrencia {
    r, err := Parse(data)
    if err != nil {
        panic(err)
    }
    return r
}
```

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

### 3.2 Objetivobus Model

```go
// business/domain/objetivobus/model.go
package objetivobus

import (
    "time"

    "github.com/google/uuid"
    "github.com/francowini/rafiki/business/types/frecuenciatype"
    "github.com/francowini/rafiki/business/types/metricatarget"
    "github.com/francowini/rafiki/business/types/metricaunidad"
    "github.com/francowini/rafiki/business/types/objetivostatus"
    "github.com/francowini/rafiki/business/types/objetivotitle"
    "github.com/francowini/rafiki/business/types/trackingtype"
)

// Objetivo represents an annual goal linked to a life vision
type Objetivo struct {
    ID                   uuid.UUID
    UserID               uuid.UUID
    LifeVisionID         uuid.UUID
    Title                objetivotitle.ObjetivoTitle
    Description          *string
    Year                 int
    TrackingType         trackingtype.TrackingType
    MetricaTarget        metricatarget.MetricaTarget
    MetricaActual        int // Auto-calculated, not directly editable
    MetricaUnidad        metricaunidad.MetricaUnidad
    Frecuencia           *frecuenciatype.FrecuenciaType // Only if TrackingType = frecuencia
    FrecuenciaN          *int                           // e.g., 5 for "5 days per week"
    CumplimientoTargetPct *int                          // e.g., 80 for "80% compliance"
    Status               objetivostatus.Status
    ArchivedAt           *time.Time
    DateCreated          time.Time
    DateUpdated          time.Time
}

// NewObjetivo contains fields for creating a new objective
type NewObjetivo struct {
    UserID               uuid.UUID
    LifeVisionID         uuid.UUID
    Title                objetivotitle.ObjetivoTitle
    Description          *string
    Year                 int
    TrackingType         trackingtype.TrackingType
    MetricaTarget        metricatarget.MetricaTarget
    MetricaUnidad        metricaunidad.MetricaUnidad
    Frecuencia           *frecuenciatype.FrecuenciaType
    FrecuenciaN          *int
    CumplimientoTargetPct *int
}

// UpdateObjetivo contains fields for updating an existing objective
type UpdateObjetivo struct {
    Title                *objetivotitle.ObjetivoTitle
    Description          *string
    MetricaTarget        *metricatarget.MetricaTarget
    MetricaUnidad        *metricaunidad.MetricaUnidad
    Frecuencia           *frecuenciatype.FrecuenciaType
    FrecuenciaN          *int
    CumplimientoTargetPct *int
    Status               *objetivostatus.Status
}
```

### 3.3 Objetivobus Business Logic

```go
// business/domain/objetivobus/objetivobus.go
package objetivobus

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/francowini/rafiki/business/domain/lifevisionbus"
    "github.com/francowini/rafiki/business/sdk/delegate"
    "github.com/francowini/rafiki/business/sdk/sqldb"
    "github.com/francowini/rafiki/business/types/objetivostatus"
    "github.com/francowini/rafiki/foundation/logger"
)

// Set of error variables
var (
    ErrNotFound              = errors.New("objetivo not found")
    ErrNotOwner              = errors.New("not owner of objetivo")
    ErrAlreadyArchived       = errors.New("objetivo already archived")
    ErrHasActiveIniciativas  = errors.New("cannot archive objetivo with active iniciativas")
    ErrLifeVisionNotActive   = errors.New("life vision is not active")
)

// ExtBusiness represents the external interface for objetivo operations
type ExtBusiness interface {
    Create(ctx context.Context, no NewObjetivo) (Objetivo, error)
    Update(ctx context.Context, objetivo Objetivo, uo UpdateObjetivo) (Objetivo, error)
    Archive(ctx context.Context, objetivo Objetivo) (Objetivo, error)
    Restore(ctx context.Context, objetivo Objetivo) (Objetivo, error)
    UpdateMetricaActual(ctx context.Context, objetivoID uuid.UUID, value int) error
    QueryByID(ctx context.Context, objetivoID uuid.UUID) (Objetivo, error)
    Query(ctx context.Context, filter QueryFilter, orderBy OrderBy, page sqldb.Page) ([]Objetivo, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for objetivo access
type Business struct {
    log           *logger.Logger
    lifeVisionBus lifevisionbus.ExtBusiness
    delegate      *delegate.Delegate
    storer        Storer
}

// NewBusiness constructs a Business for objetivo operations
func NewBusiness(
    log *logger.Logger,
    lifeVisionBus lifevisionbus.ExtBusiness,
    dlg *delegate.Delegate,
    storer Storer,
) *Business {
    b := Business{
        log:           log,
        lifeVisionBus: lifeVisionBus,
        delegate:      dlg,
        storer:        storer,
    }

    // Register delegate handlers
    b.registerDelegateFunctions()

    return &b
}

// Create adds a new objetivo to the system
func (b *Business) Create(ctx context.Context, no NewObjetivo) (Objetivo, error) {
    // Validate life vision exists and is active
    lv, err := b.lifeVisionBus.QueryByID(ctx, no.LifeVisionID)
    if err != nil {
        return Objetivo{}, fmt.Errorf("lifevision.querybyid: %w", err)
    }

    if lv.Status.IsArchived() {
        return Objetivo{}, ErrLifeVisionNotActive
    }

    // Security: Ensure life vision belongs to requesting user
    if lv.UserID != no.UserID {
        return Objetivo{}, ErrNotOwner
    }

    now := time.Now().UTC()

    objetivo := Objetivo{
        ID:                    uuid.New(),
        UserID:                no.UserID,
        LifeVisionID:          no.LifeVisionID,
        Title:                 no.Title,
        Description:           no.Description,
        Year:                  no.Year,
        TrackingType:          no.TrackingType,
        MetricaTarget:         no.MetricaTarget,
        MetricaActual:         0, // Starts at 0
        MetricaUnidad:         no.MetricaUnidad,
        Frecuencia:            no.Frecuencia,
        FrecuenciaN:           no.FrecuenciaN,
        CumplimientoTargetPct: no.CumplimientoTargetPct,
        Status:                objetivostatus.Activo,
        DateCreated:           now,
        DateUpdated:           now,
    }

    if err := b.storer.Create(ctx, objetivo); err != nil {
        return Objetivo{}, fmt.Errorf("create: %w", err)
    }

    return objetivo, nil
}

// Archive marks an objetivo as archived
func (b *Business) Archive(ctx context.Context, objetivo Objetivo) (Objetivo, error) {
    if objetivo.Status.IsArchived() {
        return Objetivo{}, ErrAlreadyArchived
    }

    now := time.Now().UTC()
    objetivo.Status = objetivostatus.Abandonado // Or a dedicated archived status
    objetivo.ArchivedAt = &now
    objetivo.DateUpdated = now

    if err := b.storer.Update(ctx, objetivo); err != nil {
        // Check if blocked by trigger (has active iniciativas)
        if errors.Is(err, ErrHasActiveIniciativas) {
            return Objetivo{}, ErrHasActiveIniciativas
        }
        return Objetivo{}, fmt.Errorf("update: %w", err)
    }

    return objetivo, nil
}

// UpdateMetricaActual updates the progress metric (called by registro creation)
func (b *Business) UpdateMetricaActual(ctx context.Context, objetivoID uuid.UUID, value int) error {
    objetivo, err := b.QueryByID(ctx, objetivoID)
    if err != nil {
        return fmt.Errorf("querybyid: %w", err)
    }

    objetivo.MetricaActual = value
    objetivo.DateUpdated = time.Now().UTC()

    if err := b.storer.Update(ctx, objetivo); err != nil {
        return fmt.Errorf("update: %w", err)
    }

    return nil
}
```

### 3.4 Tareabus with Dual Parentage

```go
// business/domain/tareabus/tareabus.go (partial)
package tareabus

// Business manages task operations with dual parent validation
type Business struct {
    log           *logger.Logger
    userBus       userbus.ExtBusiness      // For standalone tasks
    iniciativaBus iniciativabus.ExtBusiness // For derived tasks
    delegate      *delegate.Delegate
    storer        Storer
}

// Create adds a new task (derived or standalone)
func (b *Business) Create(ctx context.Context, nt NewTarea) (Tarea, error) {
    // ALWAYS validate user exists
    user, err := b.userBus.QueryByID(ctx, nt.UserID)
    if err != nil {
        return Tarea{}, fmt.Errorf("user.querybyid: %w", err)
    }
    if !user.Enabled {
        return Tarea{}, ErrUserDisabled
    }

    // CONDITIONALLY validate initiative (if provided)
    if nt.IniciativaID != nil {
        iniciativa, err := b.iniciativaBus.QueryByID(ctx, *nt.IniciativaID)
        if err != nil {
            return Tarea{}, fmt.Errorf("iniciativa.querybyid: %w", err)
        }

        // Security: Ensure initiative belongs to same user
        // (initiative's objective's life vision's value's user)
        if iniciativa.UserID != nt.UserID {
            return Tarea{}, ErrNotOwner
        }
    }

    // Create task...
    now := time.Now().UTC()
    tarea := Tarea{
        ID:             uuid.New(),
        UserID:         nt.UserID,
        IniciativaID:   nt.IniciativaID, // Can be nil for standalone
        Title:          nt.Title,
        // ... other fields
        DateCreated:    now,
        DateUpdated:    now,
    }

    if err := b.storer.Create(ctx, tarea); err != nil {
        return Tarea{}, fmt.Errorf("create: %w", err)
    }

    return tarea, nil
}
```

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

```typescript
// frontend/lib/query-client.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10,   // 10 minutes
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

// Query key factory
export const queryKeys = {
  objectives: {
    all: ['objectives'] as const,
    lists: () => [...queryKeys.objectives.all, 'list'] as const,
    list: (filters: string) => [...queryKeys.objectives.lists(), filters] as const,
    detail: (id: string) => [...queryKeys.objectives.all, 'detail', id] as const,
    records: (id: string) => [...queryKeys.objectives.all, 'records', id] as const,
  },
  initiatives: {
    all: ['initiatives'] as const,
    lists: () => [...queryKeys.initiatives.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.initiatives.all, 'detail', id] as const,
  },
  tasks: {
    all: ['tasks'] as const,
    lists: () => [...queryKeys.tasks.all, 'list'] as const,
    list: (filters: object) => [...queryKeys.tasks.lists(), filters] as const,
    detail: (id: string) => [...queryKeys.tasks.all, 'detail', id] as const,
    activity: () => [...queryKeys.tasks.all, 'activity'] as const,
  },
};
```

### 5.2 Optimistic Task Completion

```typescript
// frontend/lib/hooks/use-tasks.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { queryKeys } from '@/lib/query-client';
import { useToast } from '@/hooks/use-toast';

export function useCompleteTask() {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: (taskId: string) => api.tasks.complete(taskId),

    // Optimistic update
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.lists() });

      const previousTasks = queryClient.getQueryData(queryKeys.tasks.lists());

      queryClient.setQueryData(queryKeys.tasks.lists(), (old: any) => {
        if (!old?.items) return old;
        return {
          ...old,
          items: old.items.map((task: Task) =>
            task.id === taskId
              ? { ...task, status: 'completada', completedAt: new Date().toISOString() }
              : task
          ),
        };
      });

      return { previousTasks };
    },

    // Rollback on error
    onError: (error: Error, taskId, context) => {
      if (context?.previousTasks) {
        queryClient.setQueryData(queryKeys.tasks.lists(), context.previousTasks);
      }
      toast({
        variant: 'destructive',
        title: 'Error completing task',
        description: error.message,
      });
    },

    // Refetch to ensure consistency
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.activity() });
      queryClient.invalidateQueries({ queryKey: queryKeys.initiatives.all });
    },

    onSuccess: () => {
      toast({
        title: 'Task completed',
        description: 'Great progress!',
      });
    },
  });
}
```

### 5.3 Calendar Heatmap Component

```typescript
// frontend/components/features/objectives/FrequencyHeatmap.tsx
'use client';

import ActivityCalendar, { Activity } from 'react-activity-calendar';
import { useObjectiveRecords } from '@/lib/hooks/use-objectives';
import { Skeleton } from '@/components/ui/skeleton';

interface FrequencyHeatmapProps {
  objectiveId: string;
  year?: number;
}

export function FrequencyHeatmap({ objectiveId, year }: FrequencyHeatmapProps) {
  const currentYear = year || new Date().getFullYear();
  const { data, isLoading } = useObjectiveRecords(objectiveId, currentYear);

  if (isLoading) {
    return <Skeleton className="h-32 w-full rounded-lg" />;
  }

  // Transform records to activity format
  const activityData: Activity[] = data?.items.map((record) => ({
    date: record.fecha,
    count: record.completado ? 1 : 0,
    level: record.completado ? 2 : 0,
  })) ?? [];

  return (
    <div className="w-full p-4 bg-card rounded-lg border">
      <ActivityCalendar
        data={activityData}
        theme={{
          light: ['#f4f4f5', '#bbf7d0', '#4ade80', '#22c55e', '#16a34a'],
          dark: ['#27272a', '#166534', '#22c55e', '#4ade80', '#86efac'],
        }}
        labels={{
          totalCount: '{{count}} completions in {{year}}',
        }}
        showWeekdayLabels
        blockSize={12}
        blockMargin={3}
        blockRadius={2}
        fontSize={12}
      />
    </div>
  );
}
```

### 5.4 Quick Task Capture FAB

```typescript
// frontend/components/features/tasks/QuickTaskFAB.tsx
'use client';

import { useState, useEffect } from 'react';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useCreateTask } from '@/lib/hooks/use-tasks';

export function QuickTaskFAB() {
  const [isOpen, setIsOpen] = useState(false);
  const [title, setTitle] = useState('');
  const createTask = useCreateTask();

  // Keyboard shortcut: Cmd+K / Ctrl+K
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsOpen(true);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    await createTask.mutateAsync({
      title: title.trim(),
      tipo: 'unica',
      prioridad: 'media',
      status: 'pendiente',
      // No initiative = goes to Inbox
    });

    setTitle('');
    setIsOpen(false);
  };

  return (
    <>
      {/* Floating Action Button */}
      <Button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-6 right-6 h-14 w-14 rounded-full shadow-lg z-50"
        size="icon"
        aria-label="Quick add task"
      >
        <Plus className="h-6 w-6" />
      </Button>

      {/* Quick Add Sheet */}
      <Sheet open={isOpen} onOpenChange={setIsOpen}>
        <SheetContent side="right" className="w-full sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Captura Rápida</SheetTitle>
          </SheetHeader>

          <form onSubmit={handleSubmit} className="mt-6 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="task-title">¿Qué necesitas hacer?</Label>
              <Input
                id="task-title"
                autoFocus
                placeholder="Ej: Llamar al dentista"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
              />
            </div>

            <p className="text-sm text-muted-foreground">
              💡 Presiona <kbd className="px-1 bg-muted rounded">Cmd+K</kbd> para
              capturar tareas rápidamente desde cualquier lugar.
            </p>

            <div className="flex justify-end gap-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsOpen(false)}
              >
                Cancelar
              </Button>
              <Button
                type="submit"
                disabled={!title.trim() || createTask.isPending}
              >
                {createTask.isPending ? 'Agregando...' : 'Agregar'}
              </Button>
            </div>
          </form>
        </SheetContent>
      </Sheet>
    </>
  );
}
```

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
