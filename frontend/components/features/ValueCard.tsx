'use client';

import { Value } from '@/lib/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Edit, Trash2 } from 'lucide-react';
import { getFacetConfig } from '@/lib/value-utils';

interface ValueCardProps {
  value: Value;
  rank: number; // 1-based ranking (1 = most important)
  onEdit?: () => void;
  onDelete?: () => void;
}

export function ValueCard({ value, rank, onEdit, onDelete }: ValueCardProps) {
  const facetConfig = getFacetConfig(value.facet);
  const isTopValue = rank === 1;

  return (
    <Card
      className={`transition-all hover:shadow-md ${
        isTopValue
          ? 'ring-2 ring-rose-500 bg-gradient-to-br from-rose-50 to-white'
          : 'hover:border-rose-200'
      }`}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 flex-wrap">
            {/* Priority Badge */}
            <Badge
              variant="outline"
              className={
                isTopValue
                  ? 'bg-rose-100 text-rose-800 border-rose-300 font-semibold'
                  : 'bg-gray-100 text-gray-700 border-gray-300'
              }
            >
              {isTopValue ? '#1 Core Value' : `#${rank}`}
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
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Content */}
        <p
          className={`${isTopValue ? 'text-base font-medium' : 'text-sm'} text-foreground leading-relaxed`}
        >
          {value.content}
        </p>

        {/* Actions */}
        <div className="flex gap-2 pt-2">
          {onEdit && (
            <Button size="sm" variant="outline" onClick={onEdit} className="flex-1">
              <Edit className="h-4 w-4 mr-1" />
              Edit
            </Button>
          )}
          {onDelete && (
            <Button
              size="sm"
              variant="outline"
              onClick={onDelete}
              className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          )}
        </div>

        {/* Metadata (small timestamp) */}
        <p className="text-xs text-muted-foreground">
          Updated {new Date(value.dateUpdated).toLocaleDateString()}
        </p>
      </CardContent>
    </Card>
  );
}
