'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Value } from '@/lib/types';
import { ValueCard } from './ValueCard';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2 } from 'lucide-react';

interface ValueListProps {
  refresh?: number; // Trigger refetch when this changes
  onValueEdit?: (value: Value) => void;
  onValueDelete?: (value: Value) => void;
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

  useEffect(() => {
    const fetchValues = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await api.values.getAll();
        // Sort by displayOrder (already sorted by backend, but double-check)
        const sortedValues = response.items.sort((a, b) => a.displayOrder - b.displayOrder);
        setValues(sortedValues);

        // Notify parent of values count
        if (onValuesCountChange) {
          onValuesCountChange(sortedValues.length);
        }
      } catch (err: any) {
        setError(err.message || 'Failed to load values');
      } finally {
        setIsLoading(false);
      }
    };

    fetchValues();
  }, [refresh, onValuesCountChange]);

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

  if (values.length === 0) {
    return (
      <Alert className="bg-rose-50 border-rose-200">
        <AlertDescription className="text-rose-900">
          You haven&apos;t created any values yet. Start by defining what matters most to you.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {values.map((value, index) => (
        <ValueCard
          key={value.id}
          value={value}
          rank={index + 1}
          onEdit={() => onValueEdit?.(value)}
          onDelete={() => onValueDelete?.(value)}
        />
      ))}
    </div>
  );
}
