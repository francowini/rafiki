'use client';

import { Value } from '@/lib/types';
import { ValueCard } from './ValueCard';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ValueDragItemProps {
  value: Value;
  rank: number;
  onEdit: (value: Value) => void;
  onDelete: (value: Value) => void;
  isDisabled?: boolean;
}

export function ValueDragItem({
  value,
  rank,
  onEdit,
  onDelete,
  isDisabled = false,
}: ValueDragItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: value.id,
    disabled: isDisabled, // Disable sorting when updating
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div ref={setNodeRef} style={style} className="flex gap-2 items-stretch">
      {/* Drag handle - disabled during save */}
      <Button
        variant="ghost"
        size="sm"
        disabled={isDisabled}
        className="flex-shrink-0 cursor-grab active:cursor-grabbing h-auto disabled:cursor-not-allowed disabled:opacity-50"
        aria-label={isDisabled ? 'Saving order...' : 'Drag to reorder'}
        {...(isDisabled ? {} : { ...attributes, ...listeners })}
      >
        <GripVertical className="h-5 w-5 text-muted-foreground" />
      </Button>

      {/* Card wrapper */}
      <div className="flex-1">
        <ValueCard
          value={value}
          rank={rank}
          isDragging={isDragging}
          onEdit={() => onEdit(value)}
          onDelete={() => onDelete(value)}
        />
      </div>
    </div>
  );
}
