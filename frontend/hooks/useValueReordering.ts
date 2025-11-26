import { useState, useCallback, useRef, useEffect } from 'react';
import { Value } from '@/lib/types';
import { api } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';

export function useValueReordering() {
  const [isUpdating, setIsUpdating] = useState(false);
  const { toast } = useToast();
  const previousStateRef = useRef<Value[] | null>(null);
  const isUpdatingRef = useRef(false);

  // Keep ref in sync with state
  useEffect(() => {
    isUpdatingRef.current = isUpdating;
  }, [isUpdating]);

  const handleValueReorder = useCallback(
    async (
      originalValues: Value[],
      newValues: Value[],
    ): Promise<Value[] | null> => {
      if (isUpdatingRef.current) return null;

      // Save previous state for rollback
      previousStateRef.current = [...originalValues];
      setIsUpdating(true);

      try {
        // Find values that changed displayOrder
        const updates = newValues
          .map((newVal, index) => {
            const newOrder = index + 1;
            const originalVal = originalValues.find((v) => v.id === newVal.id);
            if (originalVal && originalVal.displayOrder !== newOrder) {
              return { id: newVal.id, newOrder };
            }
            return null;
          })
          .filter((u): u is { id: string; newOrder: number } => u !== null);

        if (updates.length === 0) {
          setIsUpdating(false);
          return null;
        }

        // Find an empty slot to use as temporary parking
        const usedOrders = new Set(originalValues.map((v) => v.displayOrder));
        let emptySlot: number | null = null;
        for (let i = 1; i <= 10; i++) {
          if (!usedOrders.has(i)) {
            emptySlot = i;
            break;
          }
        }

        if (emptySlot !== null) {
          // We have an empty slot - use it as parking
          // 1. Move first affected value to empty slot
          // 2. Update others sequentially
          // 3. Move parked value to final position
          const [firstUpdate, ...restUpdates] = updates;

          // Park first value in empty slot
          await api.values.update(firstUpdate.id, { displayOrder: emptySlot });

          // Update others sequentially
          for (const update of restUpdates) {
            await api.values.update(update.id, { displayOrder: update.newOrder });
          }

          // Move parked value to final position
          await api.values.update(firstUpdate.id, {
            displayOrder: firstUpdate.newOrder,
          });
        } else {
          // All 10 slots full - cannot reorder without backend support
          throw new Error(
            'Cannot reorder when all 10 slots are full. Delete a value first.',
          );
        }

        toast({
          title: 'Values reordered',
          description: 'Your value priorities have been updated.',
        });

        return null; // Success, no rollback needed
      } catch (error) {
        toast({
          variant: 'destructive',
          title: 'Error saving order',
          description:
            error instanceof Error
              ? error.message
              : 'Failed to save value order. Your changes have been reverted.',
        });

        // Return previous state for rollback
        return previousStateRef.current;
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
