'use client';

import { ActivityCalendar } from 'react-activity-calendar';
import { Skeleton } from '@/components/ui/skeleton';
import { useObjectiveActivity } from '@/lib/hooks/use-objectives';

interface ObjectiveHeatmapProps {
  objectiveId: string;
  year: number;
}

export function ObjectiveHeatmap({ objectiveId, year }: ObjectiveHeatmapProps) {
  const { data, isLoading } = useObjectiveActivity(objectiveId, year);

  if (isLoading) {
    return <Skeleton className="h-32 w-full" />;
  }

  if (!data) return null;

  return (
    <div className="space-y-4">
      <ActivityCalendar
        data={data.days}
        theme={{
          light: ['#ebedf0', '#9be9a8', '#40c463', '#30a14e', '#216e39'],
          dark: ['#161b22', '#0e4429', '#006d32', '#26a641', '#39d353'],
        }}
        blockSize={12}
        blockMargin={4}
        fontSize={14}
        labels={{
          totalCount: '{{count}} días completados en {{year}}',
        }}
      />

      <div className="flex gap-6">
        <div>
          <p className="text-sm text-muted-foreground">Total</p>
          <p className="text-2xl font-bold">{data.totalCompletions}</p>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">Racha actual</p>
          <p className="text-2xl font-bold">{data.streakDays} días</p>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">Mejor racha</p>
          <p className="text-2xl font-bold">{data.longestStreak} días</p>
        </div>
      </div>
    </div>
  );
}
