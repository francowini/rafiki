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
    <div className="bg-rose-50 border border-rose-200 rounded-lg p-4">
      <div className="flex items-start justify-between gap-3">
        <p className="text-sm text-foreground leading-relaxed whitespace-pre-wrap flex-1">
          {lifeVision.content}
        </p>
        <div className="flex gap-1 shrink-0">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(lifeVision)}
            className="h-8 px-2 text-muted-foreground hover:text-foreground"
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(lifeVision)}
            className="h-8 px-2 text-muted-foreground hover:text-destructive"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
