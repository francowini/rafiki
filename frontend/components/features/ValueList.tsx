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
      const sortedValues = [...response.items].sort((a, b) => a.displayOrder - b.displayOrder);
      setValues(sortedValues);
      onValuesCountChange?.(sortedValues.length);
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load values';
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
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <SortableContext items={values.map((v) => v.id)} strategy={verticalListSortingStrategy}>
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
