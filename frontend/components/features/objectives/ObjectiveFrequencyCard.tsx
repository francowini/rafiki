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
import { MoreVertical, Edit, Archive, CheckCircle2, MinusCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Objective, RecordStatus } from '@/lib/types';

interface ObjectiveFrequencyCardProps {
  objective: Objective;
  onEdit: (objective: Objective) => void;
  onArchive: (id: string) => void;
  onLogRecord: (status: RecordStatus) => void;
  onClick: (objective: Objective) => void;
}

function getFrecuenciaLabel(objective: Objective): string {
  const n = objective.frecuenciaN ?? '?';

  if (objective.frecuenciaTipo === 'daily') {
    return 'Diario';
  }
  if (objective.frecuenciaTipo === 'n_por_semana') {
    return `${n}x/semana`;
  }
  if (objective.frecuenciaTipo === 'n_por_mes') {
    return `${n}x/mes`;
  }

  // Fallback for unknown or missing frequency type
  console.warn(`Unknown frecuenciaTipo for objective ${objective.id}:`, objective.frecuenciaTipo);
  return 'Tipo desconocido';
}

export function ObjectiveFrequencyCard({
  objective,
  onEdit,
  onArchive,
  onLogRecord,
  onClick,
}: ObjectiveFrequencyCardProps) {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onClick(objective);
    }
  };

  return (
    <Card
      className={cn(
        'cursor-pointer transition-all hover:shadow-md bg-gradient-to-br focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2',
        objective.status === 'pausado'
          ? 'opacity-75 border-l-4 border-l-purple-400 hover:border-purple-300 from-purple-50/50 to-white'
          : 'hover:border-purple-300 from-purple-50 to-white',
      )}
      onClick={() => onClick(objective)}
      onKeyDown={handleKeyDown}
      tabIndex={0}
      role="button"
      aria-label={`Ver detalles de ${objective.titulo}`}
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
          <Badge variant="outline" className="bg-purple-100 text-purple-700">
            Frecuencia
          </Badge>
          <Badge variant="outline">{getFrecuenciaLabel(objective)}</Badge>
          {objective.status === 'pausado' && (
            <Badge variant="outline" className="bg-blue-50 text-blue-600 border-blue-200">
              Pausado
            </Badge>
          )}
        </div>
      </CardHeader>

      <CardContent>
        <div className="text-center py-2">
          <p className="text-3xl font-bold text-purple-600">
            {objective.cumplimientoTargetPct ?? 0}%
          </p>
          <p className="text-sm text-muted-foreground">Meta de cumplimiento</p>
        </div>
      </CardContent>

      <CardFooter className="flex gap-2">
        {objective.status === 'pausado' ? (
          <Button size="sm" variant="outline" className="w-full" disabled>
            Objetivo en pausa
          </Button>
        ) : (
          <>
            <Button
              size="sm"
              className="flex-1 bg-green-600 hover:bg-green-700"
              onClick={(e) => {
                e.stopPropagation();
                onLogRecord('completado');
              }}
            >
              <CheckCircle2 className="h-4 w-4 mr-1" />
              Hecho
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="flex-1"
              onClick={(e) => {
                e.stopPropagation();
                onLogRecord('omitido_intencional');
              }}
            >
              <MinusCircle className="h-4 w-4 mr-1" />
              Descanso
            </Button>
          </>
        )}
      </CardFooter>
    </Card>
  );
}
