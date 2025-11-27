'use client';

import { LifeVision } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Pencil, Trash2 } from 'lucide-react';

interface LifeVisionCardProps {
  lifeVision: LifeVision;
  onEdit: (lifeVision: LifeVision) => void;
  onDelete: (lifeVision: LifeVision) => void;
}

export function LifeVisionCard({ lifeVision, onEdit, onDelete }: LifeVisionCardProps) {
  return (
    <div className="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
      <div className="flex items-start justify-between gap-3">
        <p className="text-base font-medium text-foreground leading-relaxed whitespace-pre-wrap flex-1">
          {lifeVision.content}
        </p>
        <div className="flex gap-1 shrink-0">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(lifeVision)}
            className="h-8 px-2 text-muted-foreground hover:text-foreground"
            aria-label="Edit life vision"
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(lifeVision)}
            className="h-8 px-2 text-muted-foreground hover:text-destructive"
            aria-label="Delete life vision"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
