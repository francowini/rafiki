'use client';

import type { Task } from '@/lib/types';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { MoreVertical, Edit, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface TaskItemProps {
  task: Task;
  onToggle: () => void;
  onEdit: () => void;
  onDelete: () => void;
  isLoading?: boolean;
}

export function TaskItem({ task, onToggle, onEdit, onDelete, isLoading }: TaskItemProps) {
  const isCompleted = task.status === 'completed';

  return (
    <div
      className={cn(
        'group flex items-center gap-3 p-3 rounded-lg border transition-all',
        isCompleted
          ? 'bg-gray-50 border-gray-200'
          : 'bg-white border-gray-200 hover:border-blue-300',
        isLoading && 'opacity-60 pointer-events-none',
      )}
    >
      <Checkbox
        checked={isCompleted}
        onCheckedChange={onToggle}
        disabled={isLoading}
        className="h-5 w-5"
        aria-label={isCompleted ? 'Marcar como pendiente' : 'Marcar como completada'}
      />

      <div className="flex-1 min-w-0">
        <p
          className={cn(
            'text-sm font-medium truncate',
            isCompleted && 'line-through text-muted-foreground',
          )}
        >
          {task.title}
        </p>
        {task.description && (
          <p className="text-xs text-muted-foreground line-clamp-1">{task.description}</p>
        )}
      </div>

      {task.contribution && (
        <Badge variant="outline" className="shrink-0 bg-blue-50 text-blue-700 border-blue-200">
          +{task.contribution}
        </Badge>
      )}

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
            aria-label="Opciones de tarea"
          >
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={onEdit}>
            <Edit className="h-4 w-4 mr-2" />
            Editar
          </DropdownMenuItem>
          <DropdownMenuItem onClick={onDelete} className="text-destructive">
            <Trash2 className="h-4 w-4 mr-2" />
            Eliminar
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
