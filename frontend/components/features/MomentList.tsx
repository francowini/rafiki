'use client';

import { useEffect, useState, useCallback } from 'react';
import { api } from '@/lib/api';
import { Moment } from '@/lib/types';
import { MomentCard } from './MomentCard';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';

interface MomentListProps {
  refresh?: number;
  onMomentClick?: (moment: Moment) => void;
  onMomentEdit?: (moment: Moment) => void;
  onMomentDelete?: (moment: Moment) => void;
}

export function MomentList({
  refresh,
  onMomentClick,
  onMomentEdit,
  onMomentDelete,
}: MomentListProps) {
  const [moments, setMoments] = useState<Moment[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const loadMoments = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.moments.getAll({
        page,
        rows: 20,
        orderBy: 'moment_date',
        orderDirection: 'desc',
      });

      setMoments(response.items);
      setTotal(response.total);
    } catch (err: any) {
      setError(err.message || 'Error loading moments');
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadMoments();
  }, [loadMoments, refresh]);

  if (isLoading && moments.length === 0) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[...Array(6)].map((_, i) => (
          <Skeleton key={i} className="h-48 rounded-lg" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
        <Button onClick={loadMoments} variant="outline" className="mt-4">
          Retry
        </Button>
      </Alert>
    );
  }

  if (moments.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="mx-auto max-w-md space-y-4">
          <h3 className="text-lg font-medium text-muted-foreground">No moments recorded</h3>
          <p className="text-sm text-muted-foreground">
            When you're ready, this space is here for you. Record your difficult moments to better
            understand them and find patterns.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {moments.map((moment) => (
          <MomentCard
            key={moment.id}
            moment={moment}
            onClick={() => onMomentClick?.(moment)}
            onEdit={() => onMomentEdit?.(moment)}
            onDelete={() => onMomentDelete?.(moment)}
          />
        ))}
      </div>

      {/* Pagination */}
      {total > 20 && (
        <div className="flex items-center justify-between">
          <Button
            variant="outline"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {Math.ceil(total / 20)}
          </span>
          <Button
            variant="outline"
            onClick={() => setPage((p) => p + 1)}
            disabled={page >= Math.ceil(total / 20)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
