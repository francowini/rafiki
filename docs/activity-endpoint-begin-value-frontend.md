# Activity Endpoint & Begin Value - Frontend Implementation

## Overview

This document specifies the frontend implementation for:
1. **Activity Heatmap**: Binary (activity/no-activity), blue color, sheet drawer on click
2. **Begin Value**: Optional starting point with dual progress visualization

---

## Part 1: Type Updates

### File: `frontend/lib/types.ts`

**Add to Objective interface:**
```typescript
export interface Objective {
  id: string;
  lifeVisionId: string;
  title: string;
  trackingType: TrackingType;
  status: ObjectiveStatus;
  targetMetric?: number;
  currentMetric?: number;
  beginValue?: number;  // NEW
  frequencyType?: FrequencyType;
  frequencyCount?: number;
  complianceTargetPct?: number;
  archivedAt?: string | null;
  dateCreated: string;
  dateUpdated: string;
}
```

**Add to NewObjective and UpdateObjective:**
```typescript
beginValue?: number;  // NEW
```

**Update ObjectiveActivityDay:**
```typescript
export interface ObjectiveActivityDay {
  date: string;
  hasActivity: boolean;  // Changed from level 0-4
  items: ActivityItem[];
}

export interface ActivityItem {
  type: 'task' | 'record';
  id: string;
  title?: string;
  contribution?: number;
  status?: string;
  notes?: string;
}
```

**Update ObjectiveActivityData:**
```typescript
export interface ObjectiveActivityData {
  year: number;
  days: ObjectiveActivityDay[];
  totalCompletions: number;
  streakDays: number;
  longestStreak: number;
  streakUnit: 'days' | 'weeks' | 'months';  // NEW
}
```

---

## Part 2: Schema Updates

### File: `frontend/lib/schemas/objective-schema.ts`

Add beginValue validation:

```typescript
// In the result tracking schema
z.object({
  trackingType: z.literal('result'),
  lifeVisionId: z.string().min(1, 'Visión de vida es requerida'),
  title: z.string().min(5, 'Mínimo 5 caracteres').max(200, 'Máximo 200 caracteres'),
  targetMetric: z.number().min(1, 'La meta debe ser mayor a 0'),
  beginValue: z.number().optional(),
}).refine(
  (data) => {
    if (data.beginValue !== undefined) {
      return data.beginValue !== data.targetMetric;
    }
    return true;
  },
  {
    message: 'El valor inicial debe ser diferente de la meta',
    path: ['beginValue'],
  },
)
```

---

## Part 3: Progress Utility

### New File: `frontend/lib/utils/progress.ts`

```typescript
export interface ProgressResult {
  percentage: number;
  isInverse: boolean;
  isNegativeProgress: boolean;
  displayText: string;
}

export function calculateProgress(
  begin: number | undefined,
  current: number,
  target: number
): ProgressResult {
  const startValue = begin ?? 0;
  const isInverse = startValue > target;

  if (isInverse) {
    // Decrease goal (weight loss: 80 -> 70)
    const totalRange = startValue - target;
    const progress = startValue - current;
    const percentage = Math.min(100, Math.max(0, (progress / totalRange) * 100));
    const isNegativeProgress = current > startValue;

    return {
      percentage: Math.round(percentage),
      isInverse: true,
      isNegativeProgress,
      displayText: isNegativeProgress
        ? `+${current - startValue} desde inicio`
        : `${Math.abs(progress)} de ${totalRange}`,
    };
  } else {
    // Increase goal (books: 0 -> 12)
    const totalRange = target - startValue;
    const progress = current - startValue;
    const percentage = Math.min(100, Math.max(0, (progress / totalRange) * 100));
    const isNegativeProgress = current < startValue;

    return {
      percentage: Math.round(percentage),
      isInverse: false,
      isNegativeProgress,
      displayText: isNegativeProgress
        ? `${startValue - current} por debajo del inicio`
        : `${progress} de ${totalRange}`,
    };
  }
}

export function getProgressColors(isNegativeProgress: boolean) {
  if (isNegativeProgress) {
    return {
      barColor: 'bg-orange-500',
      textColor: 'text-orange-700',
      bgColor: 'bg-orange-50',
    };
  }
  return {
    barColor: 'bg-blue-500',
    textColor: 'text-blue-700',
    bgColor: 'bg-blue-50',
  };
}
```

---

## Part 4: ObjectiveForm Updates

### File: `frontend/components/features/objectives/ObjectiveForm.tsx`

**Add beginValue field after targetMetric (for result type):**

```tsx
{/* Begin Value Field - NEW */}
{trackingType === 'result' && (
  <div className="space-y-2">
    <Label htmlFor="beginValue" className="flex items-center gap-2">
      Valor inicial (opcional)
      <HelpTooltip content="Tu punto de partida. Si está en blanco, se asume 0. Puedes establecer un valor mayor a la meta para objetivos de reducción (ej: perder peso)." />
    </Label>
    <Input
      id="beginValue"
      type="number"
      {...register('beginValue', {
        valueAsNumber: true,
        setValueAs: (v) => v === '' ? undefined : Number(v)
      })}
      placeholder="Ej: 0 (predeterminado)"
    />
    {'beginValue' in errors && errors.beginValue && (
      <p className="text-sm text-red-500">{errors.beginValue.message}</p>
    )}

    {/* Live Preview */}
    {watchedTargetMetric && (
      <div className="bg-gray-50 rounded-lg p-4 space-y-2">
        <p className="text-sm font-medium text-gray-700">Vista previa:</p>
        <div className="flex items-center gap-2">
          <span className="text-2xl">
            {(watchedBeginValue ?? 0) > watchedTargetMetric ? '↓' : '↑'}
          </span>
          <div>
            <p className="text-sm text-gray-600">
              <span className="font-semibold">{watchedBeginValue ?? 0}</span>
              {' → '}
              <span className="font-semibold text-blue-600">{watchedTargetMetric}</span>
            </p>
            <p className="text-xs text-gray-500">
              {(watchedBeginValue ?? 0) > watchedTargetMetric
                ? `Reducir ${(watchedBeginValue ?? 0) - watchedTargetMetric} unidades`
                : `Aumentar ${watchedTargetMetric - (watchedBeginValue ?? 0)} unidades`
              }
            </p>
          </div>
        </div>
      </div>
    )}
  </div>
)}
```

---

## Part 5: ObjectiveHeatmap Updates

### File: `frontend/components/features/objectives/ObjectiveHeatmap.tsx`

**Replace with binary heatmap + sheet drawer:**

```tsx
'use client';

import { useState } from 'react';
import { ActivityCalendar } from 'react-activity-calendar';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Badge } from '@/components/ui/badge';
import { useObjectiveActivity } from '@/lib/hooks/use-objectives';
import { CheckCircle2, ListTodo } from 'lucide-react';
import type { ObjectiveActivityDay } from '@/lib/types';

interface ObjectiveHeatmapProps {
  objectiveId: string;
  year: number;
}

export function ObjectiveHeatmap({ objectiveId, year }: ObjectiveHeatmapProps) {
  const { data, isLoading } = useObjectiveActivity(objectiveId, year);
  const [selectedDay, setSelectedDay] = useState<ObjectiveActivityDay | null>(null);
  const [sheetOpen, setSheetOpen] = useState(false);

  if (isLoading) {
    return <Skeleton className="h-32 w-full" />;
  }

  if (!data) return null;

  // Binary: 0 or 1 (no gradient)
  const calendarData = data.days.map((day) => ({
    date: day.date,
    count: day.hasActivity ? 1 : 0,
    level: day.hasActivity ? 1 : 0,
  }));

  // Blue color scheme (binary)
  const theme = {
    light: ['#e5e7eb', '#3b82f6'],
    dark: ['#1f2937', '#2563eb'],
  };

  const handleDayClick = (event: any) => {
    const clickedDate = event.date;
    const dayData = data.days.find((d) => d.date === clickedDate);
    if (!dayData) {
      console.warn(`[F17] Day data not found for date: ${clickedDate}`);
      return;
    }
    if (dayData.hasActivity) {
      setSelectedDay(dayData);
      setSheetOpen(true);
    }
  };

  const getStreakLabel = (unit: string) => {
    switch (unit) {
      case 'days': return 'días';
      case 'weeks': return 'semanas';
      case 'months': return 'meses';
      default: return 'días';
    }
  };

  return (
    <>
      <div className="space-y-4">
        <ActivityCalendar
          data={calendarData}
          theme={theme}
          blockSize={12}
          blockMargin={4}
          fontSize={14}
          labels={{
            totalCount: `{{count}} ${getStreakLabel(data.streakUnit)} con actividad en {{year}}`,
          }}
          eventHandlers={{
            onClick: () => handleDayClick,
          }}
        />

        <div className="flex gap-6">
          <div>
            <p className="text-sm text-muted-foreground">Total</p>
            <p className="text-2xl font-bold">{data.totalCompletions}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Racha actual</p>
            <p className="text-2xl font-bold">
              {data.streakDays} {getStreakLabel(data.streakUnit)}
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Mejor racha</p>
            <p className="text-2xl font-bold">
              {data.longestStreak} {getStreakLabel(data.streakUnit)}
            </p>
          </div>
        </div>
      </div>

      {/* Activity Detail Sheet */}
      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>
              {selectedDay?.date
                ? (() => {
                    // F7: Validate date before formatting
                    const date = new Date(selectedDay.date);
                    if (isNaN(date.getTime())) {
                      console.warn(`[F7] Invalid date: ${selectedDay.date}`);
                      return 'Fecha inválida';
                    }
                    return date.toLocaleDateString('es-ES', { dateStyle: 'long' });
                  })()
                : ''}
            </SheetTitle>
            <SheetDescription>
              Detalle de actividades
            </SheetDescription>
          </SheetHeader>

          <div className="mt-6 space-y-4">
            {selectedDay?.items?.map((item) => (
              <div key={item.id} className="border rounded-lg p-4 space-y-2">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    {item.type === 'task' ? (
                      <ListTodo className="h-4 w-4 text-blue-600" />
                    ) : (
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                    )}
                    <Badge variant="outline" className="text-xs">
                      {item.type === 'task' ? 'Tarea' : 'Registro'}
                    </Badge>
                  </div>
                  {item.contribution && (
                    <Badge variant="secondary" className="text-xs">
                      +{item.contribution}
                    </Badge>
                  )}
                </div>
                {item.title && <p className="font-medium text-sm">{item.title}</p>}
                {item.status && (
                  <p className="text-xs text-muted-foreground">Estado: {item.status}</p>
                )}
                {item.notes && (
                  <p className="text-xs text-muted-foreground italic">{item.notes}</p>
                )}
              </div>
            ))}
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}
```

---

## Part 6: Progress Display Updates

### File: `frontend/components/features/objectives/ObjectiveCard.tsx`

**Update progress calculation:**

```tsx
import { calculateProgress, getProgressColors } from '@/lib/utils/progress';

// In component:
const progressData = objective.currentMetric && objective.targetMetric
  ? calculateProgress(
      objective.beginValue,
      objective.currentMetric,
      objective.targetMetric
    )
  : { percentage: 0, isInverse: false, isNegativeProgress: false, displayText: '0 de 0' };

const colors = getProgressColors(progressData.isNegativeProgress);

// In JSX:
<div className="space-y-2">
  <div className="flex justify-between text-sm">
    <span className="text-muted-foreground">Progreso</span>
    <span className="font-medium">{progressData.displayText}</span>
  </div>
  <Progress value={progressData.percentage} className={cn("h-2", colors.barColor)} />
  <div className="flex justify-between items-center">
    <p className={cn("text-xs", colors.textColor)}>
      {progressData.percentage}% completado
      {progressData.isNegativeProgress && ' (retroceso)'}
    </p>
    {objective.beginValue !== undefined && (
      <p className="text-xs text-muted-foreground">Inicio: {objective.beginValue}</p>
    )}
  </div>
</div>
```

---

## Part 7: Negative Progress Alert

### Component: NegativeProgressAlert

```tsx
interface NegativeProgressAlertProps {
  isIncreasing: boolean;
  beginValue: number;
  currentValue: number;
}

function NegativeProgressAlert({ isIncreasing, beginValue, currentValue }: NegativeProgressAlertProps) {
  const difference = Math.abs(currentValue - beginValue);

  return (
    <div className="bg-amber-50 border-l-4 border-amber-500 rounded-lg p-4 space-y-2">
      <div className="flex items-start gap-3">
        <AlertCircle className="h-5 w-5 text-amber-600 mt-0.5" />
        <div className="flex-1 space-y-2">
          <p className="text-sm font-medium text-amber-900">
            {isIncreasing
              ? `Has retrocedido ${difference} desde el inicio`
              : `Has aumentado ${difference} desde el inicio`
            }
          </p>
          <p className="text-xs text-amber-700">
            Es normal tener altibajos en el camino. Lo importante es seguir intentando
            y aprender de cada experiencia. Cada día es una nueva oportunidad.
          </p>
        </div>
      </div>
    </div>
  );
}
```

---

## Part 8: Spanish Copy

```typescript
export const OBJECTIVE_COPY = {
  beginValue: {
    label: 'Valor inicial (opcional)',
    placeholder: 'Ej: 0 (predeterminado)',
    helpTooltip: 'Tu punto de partida. Si está en blanco, se asume 0.',

    preview: {
      increasing: {
        title: 'Objetivo de incremento',
        description: 'Aumentar {difference} unidades',
      },
      decreasing: {
        title: 'Objetivo de reducción',
        description: 'Reducir {difference} unidades',
      },
    },

    negativeProgress: {
      increasing: {
        title: 'Has retrocedido {difference} desde el inicio',
        body: 'Es normal tener altibajos en el camino. Lo importante es seguir intentando.',
      },
      decreasing: {
        title: 'Has aumentado {difference} desde el inicio',
        body: 'Los retrocesos son parte del proceso. No te desanimes.',
      },
    },
  },

  heatmap: {
    total: 'Total',
    currentStreak: 'Racha actual',
    longestStreak: 'Mejor racha',
    days: 'días',
    weeks: 'semanas',
    months: 'meses',
  },

  dayDetail: {
    title: 'Detalle de actividades',
    task: 'Tarea',
    record: 'Registro',
    status: 'Estado',
  },
};
```

---

## Implementation Checklist

- [ ] Update `frontend/lib/types.ts` (add beginValue, update activity types)
- [ ] Update `frontend/lib/schemas/objective-schema.ts` (add validation)
- [ ] Create `frontend/lib/utils/progress.ts` (calculation utility)
- [ ] Update `ObjectiveForm.tsx` (add beginValue field + preview)
- [ ] Update `ObjectiveHeatmap.tsx` (binary colors + sheet drawer)
- [ ] Update `ObjectiveCard.tsx` (new progress calculation)
- [ ] Update `ObjectiveDetail.tsx` (dual progress bar)
- [ ] Test inverse goals (begin > target)
- [ ] Test negative progress display
- [ ] Test heatmap click interactions

---

## Errors-to-Avoid Compliance

This documentation has been validated against `/devs/errors-to-avoid-frontend.md`. Here is the compliance status:

### Compliant Patterns

- **F10 (Accessibility: Labels)**: The `beginValue` input field correctly uses `htmlFor="beginValue"` on the label and matching `id="beginValue"` on the Input component
- **F5 (Accessibility: Keyboard Support)**: The ObjectiveHeatmap uses semantic Sheet component with proper event handlers
- **F16 (UI Controls: Terminal States)**: No terminal state patterns in this implementation
- **F21 (API Calls)**: All API calls use centralized hooks (`useObjectiveActivity`, `calculateProgress`)
- **F4 (State Management)**: Uses proper React hooks without state duplication

### Issues Found & Fixed

#### F7 Violation: Unvalidated Date Parsing (FIXED)
**Location**: ObjectiveHeatmap component, SheetTitle display

**Issue**: `new Date(selectedDay.date).toLocaleDateString()` was called without validating the date first. If `selectedDay.date` is malformed, it would display "Invalid Date" to users.

**Fix Applied**: Added date validation with `isNaN(date.getTime())` check before calling `toLocaleDateString()`. Returns fallback message "Fecha inválida" for invalid dates.

#### F17 Violation: Silent Empty Returns (FIXED)
**Location**: ObjectiveHeatmap `handleDayClick` function

**Issue**: The function silently returned without logging when `dayData` was not found, making debugging difficult.

**Fix Applied**: Added explicit logging with `console.warn()` when day data is not found, includes the problematic date for context.

### Patterns to Verify During Implementation

1. **F2 (Stale Responses)**: The `useObjectiveActivity` hook implementation should use either `requestId` or `AbortController` pattern in its internal useEffect to guard against stale responses
2. **F3 (Data Truncation)**: No export functionality in this component, but if added, must check `items.length` vs `total`
3. **F12 (React Query Cache Keys)**: If using React Query for activity data, ensure queryKey includes all filter parameters (objectiveId, year)

### Recommendations

- Extract date formatting to a shared utility function `lib/utils/date-utils.ts` for consistency across the app
- Consider adding a loading skeleton or error boundary around the sheet content
- Ensure the `useObjectiveActivity` hook follows React Query best practices with proper cache invalidation on mutations
