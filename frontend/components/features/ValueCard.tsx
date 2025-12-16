'use client';

import { Value } from '@/lib/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Edit, MoreVertical, Archive, RotateCcw, Layers } from 'lucide-react';
import { getFacetConfig } from '@/lib/value-utils';

const gradientMap: Record<string, string> = {
  health: 'bg-gradient-to-br from-emerald-50 to-white',
  relationships: 'bg-gradient-to-br from-blue-50 to-white',
  career: 'bg-gradient-to-br from-amber-50 to-white',
  personal_growth: 'bg-gradient-to-br from-purple-50 to-white',
  family: 'bg-gradient-to-br from-pink-50 to-white',
  creativity: 'bg-gradient-to-br from-orange-50 to-white',
  community: 'bg-gradient-to-br from-green-50 to-white',
  spirituality: 'bg-gradient-to-br from-indigo-50 to-white',
};

// Helper function with date validation
function formatDateSafe(dateValue: string | null, locale = 'es-MX'): string {
  if (!dateValue) return 'Fecha desconocida';

  const date = new Date(dateValue);

  // Validate date is valid before formatting
  if (isNaN(date.getTime())) {
    return 'Fecha invalida';
  }

  return date.toLocaleDateString(locale, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

interface ValueCardProps {
  value: Value;
  rank: number;
  activeLifeVisionCount?: number;
  onEdit?: () => void;
  onArchive?: () => void;
  onRestore?: () => void;
  isDragging?: boolean;
}

export function ValueCard({
  value,
  rank,
  activeLifeVisionCount = 0,
  onEdit,
  onArchive,
  onRestore,
  isDragging = false,
}: ValueCardProps) {
  const facetConfig = getFacetConfig(value.facet);
  const gradientClass = gradientMap[value.facet] || 'bg-white';
  const isArchived = value.status === 'archived';

  return (
    <Card
      className={`transition-all ${gradientClass} ${
        isArchived
          ? 'opacity-60 border-dashed border-2 border-gray-300 bg-gray-50/50'
          : 'hover:border-rose-200'
      } ${isDragging ? 'shadow-lg border-rose-300' : 'hover:shadow-md'}`}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 flex-wrap flex-1">
            {/* Priority Badge - uniform styling for all */}
            <Badge variant="outline" className="bg-gray-100 text-gray-700 border-gray-300">
              #{rank}
            </Badge>

            {/* Facet Badge */}
            <Badge
              variant="outline"
              className={`${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
            >
              <span className="mr-1">{facetConfig.icon}</span>
              {facetConfig.label}
            </Badge>

            {/* Archived Badge */}
            {isArchived && (
              <Badge variant="outline" className="bg-gray-100 text-gray-600 border-gray-400">
                <Archive className="h-3 w-3 mr-1" />
                Retirado
              </Badge>
            )}

            {/* Life Vision Count Badge (only for active values) */}
            {!isArchived && activeLifeVisionCount > 0 && (
              <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-300">
                <Layers className="h-3 w-3 mr-1" />
                {activeLifeVisionCount} vision{activeLifeVisionCount !== 1 ? 'es' : ''}
              </Badge>
            )}
          </div>

          {/* Dropdown menu for actions */}
          {(onEdit || onArchive || onRestore) && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-8 w-8 p-0" aria-label="Actions">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {!isArchived && onEdit && (
                  <DropdownMenuItem onClick={onEdit}>
                    <Edit className="h-4 w-4 mr-2" />
                    Editar
                  </DropdownMenuItem>
                )}
                {!isArchived && onArchive && (
                  <DropdownMenuItem
                    onClick={onArchive}
                    className="text-amber-600 focus:text-amber-600"
                  >
                    <Archive className="h-4 w-4 mr-2" />
                    Retirar
                  </DropdownMenuItem>
                )}
                {isArchived && onRestore && (
                  <DropdownMenuItem
                    onClick={onRestore}
                    className="text-green-600 focus:text-green-600"
                  >
                    <RotateCcw className="h-4 w-4 mr-2" />
                    Restaurar
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <p className="text-sm text-foreground leading-relaxed">{value.content}</p>
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>Actualizado {formatDateSafe(value.dateUpdated)}</span>
          {isArchived && value.archivedAt && (
            <span className="flex items-center gap-1 text-gray-500">
              <Archive className="h-3 w-3" />
              Retirado el {formatDateSafe(value.archivedAt)}
            </span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
