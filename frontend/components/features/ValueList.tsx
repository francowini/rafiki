'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '@/lib/api';
import { Value, ValueWithLifeVisionCount, STATUS_ACTIVE, STATUS_ARCHIVED } from '@/lib/types';
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
  showArchived: boolean;
  onValueEdit: (value: Value) => void;
  onValueArchive: (value: Value) => void;
  onValueRestore: (value: Value) => void;
  onValuesCountChange?: (count: number) => void;
}

export function ValueList({
  refresh,
  showArchived,
  onValueEdit,
  onValueArchive,
  onValueRestore,
  onValuesCountChange,
}: ValueListProps) {
  const [values, setValues] = useState<ValueWithLifeVisionCount[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { handleValueReorder, isUpdating } = useValueReordering();

  // Request ID ref for freshness guard (prevents stale responses)
  const requestIdRef = useRef(0);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  // Fetch values from server
  const fetchValues = useCallback(async () => {
    // Increment request ID for freshness guard
    const currentRequestId = ++requestIdRef.current;

    setIsLoading(true);
    setError(null);

    try {
      const [response, visionsRes] = await Promise.all([
        api.values.getAll({ includeArchived: showArchived }),
        api.lifeVisions.getAll(),
      ]);

      // Check if this request is still current (prevents stale responses)
      if (currentRequestId !== requestIdRef.current) {
        return; // A newer request was made, discard this result
      }

      // Add life vision counts (backend already filters to active-only when includeArchived is not set)
      const valuesWithCounts: ValueWithLifeVisionCount[] = response.items.map((value) => ({
        ...value,
        activeLifeVisionCount: visionsRes.items.filter((v) => v.valueId === value.id).length,
      }));

      const sortedValues = valuesWithCounts.sort((a, b) => a.displayOrder - b.displayOrder);
      setValues(sortedValues);

      // Count only active values for limit
      const activeCount = sortedValues.filter((v) => v.status === STATUS_ACTIVE).length;
      onValuesCountChange?.(activeCount);
    } catch (err: unknown) {
      // Only update error state if this request is still current
      if (currentRequestId === requestIdRef.current) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to load values';
        setError(errorMessage);
      }
    } finally {
      // Only update loading state if this request is still current
      if (currentRequestId === requestIdRef.current) {
        setIsLoading(false);
      }
    }
  }, [showArchived, onValuesCountChange]);

  // Initial fetch + refetch on refresh trigger
  useEffect(() => {
    fetchValues();
  }, [refresh, fetchValues]);

  // Handle drag end - reorder values
  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      const { active, over } = event;

      if (!over || active.id === over.id) return;

      // Only allow reordering active values
      const activeValues = values.filter((v) => v.status === STATUS_ACTIVE);
      const oldIndex = activeValues.findIndex((v) => v.id === active.id);
      const newIndex = activeValues.findIndex((v) => v.id === over.id);

      if (oldIndex === -1 || newIndex === -1) return;

      // Keep original for comparison
      const originalValues = [...values];

      // Optimistic update - show new order immediately (only for active values)
      const newActiveValues = arrayMove(activeValues, oldIndex, newIndex);
      const updatedActiveValues = newActiveValues.map((value, index) => ({
        ...value,
        displayOrder: index + 1,
      }));

      // Merge back with archived values
      const archivedValues = values.filter((v) => v.status === STATUS_ARCHIVED);
      const updatedValues = [...updatedActiveValues, ...archivedValues].sort(
        (a, b) => a.displayOrder - b.displayOrder,
      );
      setValues(updatedValues);

      // Save to backend (only active values)
      const result = await handleValueReorder(
        originalValues.filter((v) => v.status === STATUS_ACTIVE),
        updatedActiveValues,
      );

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

  // Filter values based on showArchived
  const displayValues = showArchived ? values : values.filter((v) => v.status === STATUS_ACTIVE);

  // Generate 10 slots for active values
  const activeValues = displayValues.filter((v) => v.status === STATUS_ACTIVE);
  const archivedValues = displayValues.filter((v) => v.status === STATUS_ARCHIVED);

  const slots = Array.from({ length: 10 }, (_, i) => {
    const slotNumber = i + 1;
    const value = activeValues.find((v) => v.displayOrder === slotNumber);
    return { slotNumber, value };
  });

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <SortableContext
        items={activeValues.map((v) => v.id)}
        strategy={verticalListSortingStrategy}
      >
        <div className="space-y-3">
          {slots.map(({ slotNumber, value }) => {
            if (!value) {
              return <EmptySlot key={`empty-${slotNumber}`} slotNumber={slotNumber} />;
            }

            return (
              <ValueDragItem
                key={value.id}
                value={value}
                rank={slotNumber}
                activeLifeVisionCount={value.activeLifeVisionCount}
                onEdit={onValueEdit}
                onArchive={onValueArchive}
                onRestore={onValueRestore}
                isDisabled={isUpdating}
              />
            );
          })}

          {/* Archived values section */}
          {showArchived && archivedValues.length > 0 && (
            <>
              <div className="pt-4 pb-2">
                <h3 className="text-sm font-medium text-muted-foreground">
                  Valores retirados ({archivedValues.length})
                </h3>
              </div>
              {archivedValues.map((value) => (
                <ValueDragItem
                  key={value.id}
                  value={value}
                  activeLifeVisionCount={value.activeLifeVisionCount}
                  onEdit={onValueEdit}
                  onArchive={onValueArchive}
                  onRestore={onValueRestore}
                  isDisabled
                />
              ))}
            </>
          )}
        </div>
      </SortableContext>
    </DndContext>
  );
}
