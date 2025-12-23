'use client';

import { Card, CardHeader, CardContent, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Progress } from '@/components/ui/progress';
import { MoreVertical, Edit, Archive, TrendingUp } from 'lucide-react';
import { cn } from '@/lib/utils';
import { calculateProgress, getProgressColors } from '@/lib/utils/progress';
import type { Objective } from '@/lib/types';

interface ObjectiveCardProps {
  objective: Objective;
  onEdit: (objective: Objective) => void;
  onArchive: (id: string) => void;
  onClick: (objective: Objective) => void;
  onUpdateProgress?: (objective: Objective) => void;
}

export function ObjectiveCard({
  objective,
  onEdit,
  onArchive,
  onClick,
  onUpdateProgress,
}: ObjectiveCardProps) {
  const progressData =
    objective.currentMetric !== undefined && objective.targetMetric
      ? calculateProgress(objective.beginMetric, objective.currentMetric, objective.targetMetric)
      : { percentage: 0, isInverse: false, isNegativeProgress: false, displayText: '0 de 0' };

  const colors = getProgressColors(progressData.isNegativeProgress);

  return (
    <Card
      className={cn(
        'cursor-pointer transition-all hover:shadow-md bg-gradient-to-br',
        objective.status === 'paused'
          ? 'opacity-75 border-l-4 border-l-blue-400 hover:border-blue-300 from-blue-50/50 to-white'
          : 'hover:border-blue-300 from-blue-50 to-white',
      )}
      onClick={() => onClick(objective)}
    >
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-lg line-clamp-2">{objective.title}</h3>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit(objective);
                }}
              >
                <Edit className="h-4 w-4 mr-2" />
                Editar
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onArchive(objective.id);
                }}
              >
                <Archive className="h-4 w-4 mr-2" />
                Archivar
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="flex gap-2 mt-2">
          <Badge variant="outline" className="bg-blue-100 text-blue-700">
            Resultado
          </Badge>
          {objective.status === 'paused' ? (
            <Badge variant="outline" className="bg-blue-50 text-blue-600 border-blue-200">
              Pausado
            </Badge>
          ) : (
            <Badge variant="outline">{objective.status}</Badge>
          )}
        </div>
      </CardHeader>

      <CardContent>
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Progreso</span>
            <span className="font-medium">{progressData.displayText}</span>
          </div>
          <Progress value={progressData.percentage} className={cn('h-2', colors.barColor)} />
          <div className="flex justify-between items-center">
            <p className={cn('text-xs', colors.textColor)}>
              {progressData.percentage}% completado
              {progressData.isNegativeProgress && ' (retroceso)'}
            </p>
            {objective.beginMetric !== undefined && (
              <p className="text-xs text-muted-foreground">Inicio: {objective.beginMetric}</p>
            )}
          </div>
        </div>
      </CardContent>

      <CardFooter>
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          disabled={objective.status === 'paused'}
          onClick={(e) => {
            e.stopPropagation();
            onUpdateProgress?.(objective);
          }}
        >
          <TrendingUp className="h-4 w-4 mr-2" />
          {objective.status === 'paused' ? 'Objetivo en pausa' : 'Actualizar progreso'}
        </Button>
      </CardFooter>
    </Card>
  );
}
