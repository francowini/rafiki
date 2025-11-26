# Value Reorder - Frontend Implementation

## Overview

This document specifies the frontend implementation for atomic value reordering, addressing:
- **Issue #14**: Support drag-and-drop reordering when all 10 value slots are filled
- **Issue #13**: Rollback inconsistency where client state reverts but backend may remain partially reordered

## Solution Summary

Replace sequential PUT calls with a single `POST /v1/values/reorder` API call. On error, auto-refresh from server to ensure client-server consistency.

## Design Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| State management | Keep useState | Simple for MVP, no extra dependencies |
| Error handling | Auto-refresh from server | Server is source of truth |
| User feedback | Toast notification | Non-intrusive |
| Loading state | Disable drag handles | Prevents concurrent reorders |
| Success feedback | Toast only | No animation needed |

## Files to Modify

1. `frontend/hooks/useValueReordering.ts` - Core reorder logic
2. `frontend/lib/api.ts` - Add reorder API method
3. `frontend/components/features/ValueList.tsx` - Disable drag during save
4. `frontend/components/features/ValueDragItem.tsx` - Handle disabled state

## Implementation

### 1. Update `api.ts`

Add the new reorder method to the values namespace:

```typescript
// Add this interface near the top with other types
interface ReorderItem {
  id: string;
  displayOrder: number;
}

// Add to the values object in the api export
export const api = {
  // ... existing code ...

  values: {
    // ... existing methods (getAll, getById, create, update, delete) ...

    /**
     * Reorder multiple values in a single atomic operation.
     * Sends all display order updates in one batch request.
     */
    reorder: async (items: ReorderItem[]): Promise<void> => {
      return fetchAPI<void>('/v1/values/reorder', {
        method: 'POST',
        body: JSON.stringify({ items }),
      });
    },
  },

  // ... rest of api ...
};
```

### 2. Update `useValueReordering.ts`

Replace the entire file with this simplified version:

```typescript
import { useState, useCallback, useRef, useEffect } from 'react';
import { Value } from '@/lib/types';
import { api } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';

interface ReorderItem {
  id: string;
  displayOrder: number;
}

// Return type: null = success, 'refresh' = needs refresh from server
type ReorderResult = null | 'refresh';

export function useValueReordering() {
  const [isUpdating, setIsUpdating] = useState(false);
  const { toast } = useToast();
  const isUpdatingRef = useRef(false);

  // Keep ref in sync with state
  useEffect(() => {
    isUpdatingRef.current = isUpdating;
  }, [isUpdating]);

  const handleValueReorder = useCallback(
    async (
      originalValues: Value[],
      newValues: Value[],
    ): Promise<ReorderResult> => {
      // Prevent concurrent reorders
      if (isUpdatingRef.current) return null;

      setIsUpdating(true);

      try {
        // Build reorder items from new values
        const reorderItems: ReorderItem[] = newValues.map((value, index) => ({
          id: value.id,
          displayOrder: index + 1,
        }));

        // Check if anything actually changed
        const hasChanges = reorderItems.some((item) => {
          const originalValue = originalValues.find((v) => v.id === item.id);
          return originalValue && originalValue.displayOrder !== item.displayOrder;
        });

        if (!hasChanges) {
          return null; // No changes needed
        }

        // Single atomic API call
        await api.values.reorder(reorderItems);

        toast({
          title: 'Values reordered',
          description: 'Your value priorities have been updated.',
        });

        return null; // Success
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Failed to save value order';

        toast({
          variant: 'destructive',
          title: 'Error saving order',
          description: errorMessage,
        });

        // Signal that caller should refresh from server
        return 'refresh';
      } finally {
        setIsUpdating(false);
      }
    },
    [toast],
  );

  return {
    handleValueReorder,
    isUpdating,
  };
}
```

**Key changes:**
- Removed `previousStateRef` (no client-side rollback)
- Removed parking slot logic (no longer needed)
- Removed "all 10 slots full" error
- Single `api.values.reorder()` call
- Returns `'refresh'` signal on error for caller to handle

### 3. Update `ValueList.tsx`

Update the handleDragEnd function and pass disabled state:

```typescript
'use client';

import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';
import { Value } from '@/lib/types';
import { ValueDragItem } from './ValueDragItem';
import { EmptySlot } from './EmptySlot';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2 } from 'lucide-react';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { useValueReordering } from '@/hooks/useValueReordering';

interface ValueListProps {
  refresh?: number;
  onValueEdit: (value: Value) => void;
  onValueDelete: (value: Value) => void;
  onValuesCountChange?: (count: number) => void;
}

export function ValueList({
  refresh,
  onValueEdit,
  onValueDelete,
  onValuesCountChange,
}: ValueListProps) {
  const [values, setValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { handleValueReorder, isUpdating } = useValueReordering();

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  // Fetch values from server
  const fetchValues = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.values.getAll();
      const sortedValues = [...response.items].sort(
        (a, b) => a.displayOrder - b.displayOrder,
      );
      setValues(sortedValues);
      onValuesCountChange?.(sortedValues.length);
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to load values';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  }, [onValuesCountChange]);

  // Initial fetch + refetch on refresh trigger
  useEffect(() => {
    fetchValues();
  }, [refresh, fetchValues]);

  // Handle drag end - reorder values
  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      const { active, over } = event;

      if (!over || active.id === over.id) return;

      const oldIndex = values.findIndex((v) => v.id === active.id);
      const newIndex = values.findIndex((v) => v.id === over.id);

      if (oldIndex === -1 || newIndex === -1) return;

      // Keep original for comparison
      const originalValues = [...values];

      // Optimistic update - show new order immediately
      const newValues = arrayMove(values, oldIndex, newIndex);
      const updatedValues = newValues.map((value, index) => ({
        ...value,
        displayOrder: index + 1,
      }));
      setValues(updatedValues);

      // Save to backend
      const result = await handleValueReorder(originalValues, updatedValues);

      // If result is 'refresh', auto-refresh from server
      if (result === 'refresh') {
        await fetchValues();
      }
      // If result is null, save was successful - keep optimistic update
    },
    [values, handleValueReorder, fetchValues],
  );

  if (isLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-rose-600" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  // Generate 10 slots
  const slots = Array.from({ length: 10 }, (_, i) => {
    const slotNumber = i + 1;
    const value = values.find((v) => v.displayOrder === slotNumber);
    return { slotNumber, value };
  });

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext
        items={values.map((v) => v.id)}
        strategy={verticalListSortingStrategy}
      >
        <div className="space-y-3">
          {slots.map(({ slotNumber, value }) => {
            if (!value) {
              return (
                <EmptySlot key={`empty-${slotNumber}`} slotNumber={slotNumber} />
              );
            }

            return (
              <ValueDragItem
                key={value.id}
                value={value}
                rank={slotNumber}
                onEdit={onValueEdit}
                onDelete={onValueDelete}
                isDisabled={isUpdating}
              />
            );
          })}
        </div>
      </SortableContext>
    </DndContext>
  );
}
```

**Key changes:**
- Added `fetchValues` callback for auto-refresh
- Pass `isDisabled={isUpdating}` to ValueDragItem
- On error result, call `fetchValues()` to sync with server

### 4. Update `ValueDragItem.tsx`

Add disabled state handling:

```typescript
'use client';

import { Value } from '@/lib/types';
import { ValueCard } from './ValueCard';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ValueDragItemProps {
  value: Value;
  rank: number;
  onEdit: (value: Value) => void;
  onDelete: (value: Value) => void;
  isDisabled?: boolean;
}

export function ValueDragItem({
  value,
  rank,
  onEdit,
  onDelete,
  isDisabled = false,
}: ValueDragItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: value.id,
    disabled: isDisabled,  // Disable sorting when updating
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div ref={setNodeRef} style={style} className="flex gap-2 items-stretch">
      {/* Drag handle - disabled during save */}
      <Button
        variant="ghost"
        size="sm"
        disabled={isDisabled}
        className="flex-shrink-0 cursor-grab active:cursor-grabbing h-auto disabled:cursor-not-allowed disabled:opacity-50"
        aria-label={isDisabled ? 'Saving order...' : 'Drag to reorder'}
        {...(isDisabled ? {} : { ...attributes, ...listeners })}
      >
        <GripVertical className="h-5 w-5 text-muted-foreground" />
      </Button>

      {/* Card wrapper */}
      <div className="flex-1">
        <ValueCard
          value={value}
          rank={rank}
          isDragging={isDragging}
          onEdit={() => onEdit(value)}
          onDelete={() => onDelete(value)}
        />
      </div>
    </div>
  );
}
```

**Key changes:**
- Added `isDisabled` prop with default `false`
- Pass `disabled: isDisabled` to `useSortable` hook
- Conditionally spread drag listeners only when not disabled
- Visual feedback: `disabled:cursor-not-allowed disabled:opacity-50`
- Updated aria-label for accessibility

## Testing Checklist

- [ ] Drag works when not saving
- [ ] Drag is disabled during save (cursor changes, opacity reduced)
- [ ] Reordering 3 values among 10 shows correct toast
- [ ] Reordering all 10 filled slots works (no "slots full" error)
- [ ] Network error shows error toast and auto-refreshes
- [ ] After refresh, UI shows server state
- [ ] Rapid drag attempts are prevented during save

## User Experience Flow

```
1. User drags value from position 1 to position 5
   ↓
2. UI updates immediately (optimistic)
   ↓
3. Drag handles become disabled (visual: opacity 50%)
   ↓
4. Single POST /v1/values/reorder sent
   ↓
5a. SUCCESS:
    - Toast: "Values reordered"
    - Drag handles re-enabled
    - UI keeps new order

5b. ERROR:
    - Toast: "Error saving order"
    - Auto-fetch from server
    - UI shows server state
    - Drag handles re-enabled
```

## Deployment Notes

1. Backend must be deployed first with `/v1/values/reorder` endpoint
2. Frontend can be deployed after backend is live
3. No breaking changes - existing functionality preserved
4. Vercel deployment: standard process, no special config needed
