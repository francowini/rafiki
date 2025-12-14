'use client';

import { useState } from 'react';
import type { Moment } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ChevronDown, Edit, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MomentListItemProps {
  moment: Moment;
  onEdit?: () => void;
  onDelete?: () => void;
}

function formatDate(dateValue: string): { dateStr: string; timeStr: string } {
  const date = new Date(dateValue);

  // Guard against invalid dates
  if (isNaN(date.getTime())) {
    return { dateStr: 'Fecha inválida', timeStr: '' };
  }

  const dateStr = date.toLocaleDateString('es-MX', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
  const timeStr = date.toLocaleTimeString('es-MX', {
    hour: '2-digit',
    minute: '2-digit',
  });

  return { dateStr, timeStr };
}

export function MomentListItem({ moment, onEdit, onDelete }: MomentListItemProps) {
  const [expanded, setExpanded] = useState(false);

  const { dateStr, timeStr } = formatDate(moment.momentDate);

  const situationPreview =
    moment.situation.length > 150 ? moment.situation.slice(0, 150) + '...' : moment.situation;

  const thoughtsPreview =
    moment.thoughts.length > 100 ? moment.thoughts.slice(0, 100) + '...' : moment.thoughts;

  const getIntensityColor = (intensity: number) => {
    if (intensity >= 8) return 'bg-red-100 text-red-800 border-red-300';
    if (intensity >= 5) return 'bg-yellow-100 text-yellow-800 border-yellow-300';
    return 'bg-green-100 text-green-800 border-green-300';
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      setExpanded((prev) => !prev);
    }
  };

  return (
    <div className="border rounded-lg bg-white hover:shadow-sm transition-shadow">
      <button
        type="button"
        aria-expanded={expanded}
        className="w-full p-4 cursor-pointer text-left"
        onClick={() => setExpanded((prev) => !prev)}
        onKeyDown={handleKeyDown}
      >
        <div className="flex items-start justify-between mb-2">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-1">
              <h3 className="font-medium text-gray-900">{dateStr}</h3>
              {timeStr && <span className="text-sm text-gray-500">{timeStr}</span>}
              <Badge variant="outline" className={getIntensityColor(moment.intensity)}>
                {moment.intensity}/10
              </Badge>
            </div>
          </div>
          <ChevronDown
            className={cn(
              'h-5 w-5 text-gray-400 transition-transform flex-shrink-0 ml-2',
              expanded && 'rotate-180',
            )}
          />
        </div>

        <div className="space-y-2 text-sm">
          <div>
            <span className="font-medium text-gray-700">Situación: </span>
            <span className="text-gray-600">{situationPreview}</span>
          </div>
          {!expanded && (
            <div>
              <span className="font-medium text-gray-700">Pensamientos: </span>
              <span className="text-gray-600">{thoughtsPreview}</span>
            </div>
          )}
        </div>
      </button>

      {expanded && (
        <div className="px-4 pb-4 space-y-3 border-t pt-3">
          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase mb-1">Pensamientos</p>
            <p className="text-sm text-gray-700">{moment.thoughts}</p>
          </div>

          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase mb-1">Síntomas Físicos</p>
            <p className="text-sm text-gray-700">{moment.physicalSymptoms}</p>
          </div>

          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase mb-1">Comportamiento</p>
            <p className="text-sm text-gray-700">{moment.behavior}</p>
          </div>

          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase mb-1">Consecuencias</p>
            <p className="text-sm text-gray-700">{moment.consequences}</p>
          </div>

          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase mb-1">
              Reflexión de Valores
            </p>
            <p className="text-sm text-gray-700">{moment.valuesReflection}</p>
          </div>

          <div className="flex gap-2 pt-2">
            {onEdit && (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit();
                }}
                className="flex-1"
              >
                <Edit className="h-4 w-4 mr-1" />
                Editar
              </Button>
            )}
            {onDelete && (
              <Button
                type="button"
                size="sm"
                variant="outline"
                aria-label="Eliminar momento"
                title="Eliminar momento"
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete();
                }}
                className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
