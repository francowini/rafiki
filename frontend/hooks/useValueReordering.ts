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
    async (originalValues: Value[], newValues: Value[]): Promise<ReorderResult> => {
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
        const errorMessage = error instanceof Error ? error.message : 'Failed to save value order';

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
