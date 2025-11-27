'use client';

import { LifeVision, Value } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { getFacetConfig } from '@/lib/value-utils';
import { LifeVisionCard } from './LifeVisionCard';

interface LifeVisionsByValueProps {
  value: Value;
  visions: LifeVision[];
  onEdit: (vision: LifeVision) => void;
  onDelete: (vision: LifeVision) => void;
}

export function LifeVisionsByValue({ value, visions, onEdit, onDelete }: LifeVisionsByValueProps) {
  const facetConfig = getFacetConfig(value.facet);

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 flex-wrap">
        <Badge variant="outline" className="bg-gray-100 text-gray-700 border-gray-300">
          #{value.displayOrder}
        </Badge>
        <span className="text-sm font-medium text-foreground flex-1 min-w-0">
          <span className="line-clamp-1">{value.content}</span>
        </span>
        <Badge
          variant="outline"
          className={`${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor} shrink-0`}
        >
          <span className="mr-1">{facetConfig.icon}</span>
          {facetConfig.label}
        </Badge>
      </div>

      <div className="ml-12 space-y-3">
        {visions.length > 0 ? (
          visions.map((vision) => (
            <LifeVisionCard
              key={vision.id}
              lifeVision={vision}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ))
        ) : (
          <div className="border border-dashed border-gray-300 rounded-lg p-4 text-center">
            <p className="text-sm text-muted-foreground italic">No visions yet for this value.</p>
          </div>
        )}
      </div>
    </div>
  );
}
