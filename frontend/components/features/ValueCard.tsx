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
import { Edit, Trash2, MoreVertical } from 'lucide-react';
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

interface ValueCardProps {
  value: Value;
  rank: number;
  onEdit?: () => void;
  onDelete?: () => void;
  isDragging?: boolean;
}

export function ValueCard({
  value,
  rank,
  onEdit,
  onDelete,
  isDragging = false,
}: ValueCardProps) {
  const facetConfig = getFacetConfig(value.facet);
  const gradientClass = gradientMap[value.facet] || 'bg-white';

  return (
    <Card
      className={`transition-all ${gradientClass} hover:border-rose-200 ${
        isDragging ? 'shadow-lg border-rose-300' : 'hover:shadow-md'
      }`}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 flex-wrap flex-1">
            {/* Priority Badge - uniform styling for all */}
            <Badge
              variant="outline"
              className="bg-gray-100 text-gray-700 border-gray-300"
            >
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
          </div>

          {/* Dropdown menu for actions */}
          {(onEdit || onDelete) && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  aria-label="Actions"
                >
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {onEdit && (
                  <DropdownMenuItem onClick={onEdit}>
                    <Edit className="h-4 w-4 mr-2" />
                    Edit
                  </DropdownMenuItem>
                )}
                {onDelete && (
                  <DropdownMenuItem
                    onClick={onDelete}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <p className="text-sm text-foreground leading-relaxed">
          {value.content}
        </p>
        <p className="text-xs text-muted-foreground">
          Updated{' '}
          {(() => {
            try {
              const date = new Date(value.dateUpdated);
              if (isNaN(date.getTime())) return 'recently';
              return date.toLocaleDateString();
            } catch {
              return 'recently';
            }
          })()}
        </p>
      </CardContent>
    </Card>
  );
}
