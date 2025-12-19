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
  const progress =
    objective.metricaActual && objective.metricaObjetivo
      ? Math.round((objective.metricaActual / objective.metricaObjetivo) * 100)
      : 0;

  return (
    <Card
      className={cn(
        'cursor-pointer transition-all hover:shadow-md bg-gradient-to-br',
        objective.status === 'pausado'
          ? 'opacity-75 border-l-4 border-l-blue-400 hover:border-blue-300 from-blue-50/50 to-white'
          : 'hover:border-blue-300 from-blue-50 to-white',
      )}
      onClick={() => onClick(objective)}
    >
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-lg line-clamp-2">{objective.titulo}</h3>
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
          {objective.status === 'pausado' ? (
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
            <span className="font-medium">
              {objective.metricaActual || 0} / {objective.metricaObjetivo}
            </span>
          </div>
          <Progress value={progress} className="h-2" />
          <p className="text-xs text-muted-foreground text-right">{progress}% completado</p>
        </div>
      </CardContent>

      <CardFooter>
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          disabled={objective.status === 'pausado'}
          onClick={(e) => {
            e.stopPropagation();
            onUpdateProgress?.(objective);
          }}
        >
          <TrendingUp className="h-4 w-4 mr-2" />
          {objective.status === 'pausado' ? 'Objetivo en pausa' : 'Actualizar progreso'}
        </Button>
      </CardFooter>
    </Card>
  );
}
