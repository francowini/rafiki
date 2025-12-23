'use client';

import { cloneElement, useState, type SVGAttributes, type HTMLAttributes } from 'react';
import { ActivityCalendar, type Activity, type BlockElement } from 'react-activity-calendar';
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

  // Calculate level based on best status of the day
  // 0 = no activity (white)
  // 1 = skipped (gray)
  // 2 = intentionally_skipped (yellow)
  // 3 = completed or task (green)
  const getLevel = (day: ObjectiveActivityDay): number => {
    if (!day.hasActivity || day.items.length === 0) return 0;

    // Find the "best" status for the day (completed > intentionally_skipped > skipped)
    let hasCompleted = false;
    let hasIntentionallySkipped = false;
    let hasSkipped = false;

    for (const item of day.items) {
      if (item.type === 'task') {
        hasCompleted = true; // Tasks are always completed
      } else if (item.status === 'completed') {
        hasCompleted = true;
      } else if (item.status === 'intentionally_skipped') {
        hasIntentionallySkipped = true;
      } else if (item.status === 'skipped') {
        hasSkipped = true;
      }
    }

    if (hasCompleted) return 3;
    if (hasIntentionallySkipped) return 2;
    if (hasSkipped) return 1;
    return 0;
  };

  const calendarData = data.days.map((day) => ({
    date: day.date,
    count: day.hasActivity ? 1 : 0,
    level: getLevel(day),
  }));

  // Color scheme by status:
  // 0 = white (no activity)
  // 1 = gray (skipped)
  // 2 = yellow (intentionally skipped)
  // 3 = green (completed)
  const theme = {
    light: ['#ffffff', '#d1d5db', '#fbbf24', '#22c55e'],
    dark: ['#1f2937', '#4b5563', '#d97706', '#16a34a'],
  };

  const handleDayClick = (clickedDate: string) => {
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

  const renderBlock = (block: BlockElement, activity: Activity) => {
    return cloneElement(block, {
      onClick: () => handleDayClick(activity.date),
      style: {
        ...(block.props as SVGAttributes<SVGRectElement> & HTMLAttributes<SVGRectElement>).style,
        cursor: activity.level > 0 ? 'pointer' : 'default',
      },
    } as SVGAttributes<SVGRectElement>);
  };

  return (
    <>
      <div className="space-y-3">
        <ActivityCalendar
          data={calendarData}
          theme={theme}
          maxLevel={3}
          blockSize={12}
          blockMargin={4}
          fontSize={14}
          showTotalCount={false}
          showColorLegend={false}
          renderBlock={renderBlock}
        />

        <p className="text-sm text-muted-foreground">
          {data.totalCompletions} registros completados
        </p>
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
            <SheetDescription>Detalle de actividades</SheetDescription>
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
