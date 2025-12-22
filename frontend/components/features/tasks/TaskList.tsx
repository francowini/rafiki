'use client';

import { useState, useMemo } from 'react';
import type { Task } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { TaskItem } from './TaskItem';
import { Circle, CheckCircle2, Plus } from 'lucide-react';
import { cn } from '@/lib/utils';

interface TaskListProps {
  tasks: Task[];
  isLoading?: boolean;
  onToggleTask: (taskId: string) => void;
  onEditTask: (task: Task) => void;
  onDeleteTask: (task: Task) => void;
  onCreateTask: () => void;
  completingTaskId?: string | null;
}

type TabType = 'pending' | 'completed';

export function TaskList({
  tasks,
  isLoading,
  onToggleTask,
  onEditTask,
  onDeleteTask,
  onCreateTask,
  completingTaskId,
}: TaskListProps) {
  const [activeTab, setActiveTab] = useState<TabType>('pending');

  const { pendingTasks, completedTasks } = useMemo(() => {
    const pending = tasks
      .filter((t) => t.status === 'pending')
      .sort((a, b) => new Date(b.dateCreated).getTime() - new Date(a.dateCreated).getTime());

    const completed = tasks
      .filter((t) => t.status === 'completed')
      .sort((a, b) => {
        const aDate = a.completedAt || a.dateCreated;
        const bDate = b.completedAt || b.dateCreated;
        return new Date(bDate).getTime() - new Date(aDate).getTime();
      });

    return { pendingTasks: pending, completedTasks: completed };
  }, [tasks]);

  const displayedTasks = activeTab === 'pending' ? pendingTasks : completedTasks;

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-7 w-24" />
          <Skeleton className="h-9 w-28" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="font-semibold text-lg">Tareas</h2>
        <Button onClick={onCreateTask} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          Nueva tarea
        </Button>
      </div>

      {/* Tabs */}
      <div role="tablist" className="flex gap-2 border-b" aria-label="Estado de tareas">
        <button
          type="button"
          role="tab"
          id="tab-pending"
          aria-selected={activeTab === 'pending'}
          aria-controls="panel-pending"
          tabIndex={activeTab === 'pending' ? 0 : -1}
          onClick={() => setActiveTab('pending')}
          className={cn(
            'flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
            activeTab === 'pending'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          <Circle className="h-4 w-4" />
          Pendientes
          {pendingTasks.length > 0 && (
            <span
              className={cn(
                'px-2 py-0.5 rounded-full text-xs font-semibold',
                activeTab === 'pending' ? 'bg-primary/10 text-primary' : 'bg-gray-100 text-gray-600',
              )}
            >
              {pendingTasks.length}
            </span>
          )}
        </button>

        <button
          type="button"
          role="tab"
          id="tab-completed"
          aria-selected={activeTab === 'completed'}
          aria-controls="panel-completed"
          tabIndex={activeTab === 'completed' ? 0 : -1}
          onClick={() => setActiveTab('completed')}
          className={cn(
            'flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
            activeTab === 'completed'
              ? 'border-green-500 text-green-600'
              : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          <CheckCircle2 className="h-4 w-4" />
          Completadas
          {completedTasks.length > 0 && (
            <span
              className={cn(
                'px-2 py-0.5 rounded-full text-xs font-semibold',
                activeTab === 'completed'
                  ? 'bg-green-100 text-green-600'
                  : 'bg-gray-100 text-gray-600',
              )}
            >
              {completedTasks.length}
            </span>
          )}
        </button>
      </div>

      {/* Task List or Empty State */}
      <div
        role="tabpanel"
        id={activeTab === 'pending' ? 'panel-pending' : 'panel-completed'}
        aria-labelledby={activeTab === 'pending' ? 'tab-pending' : 'tab-completed'}
      >
        {displayedTasks.length === 0 ? (
          <div className="border border-dashed rounded-lg p-12 text-center">
            {activeTab === 'pending' ? (
            <>
              <Circle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium mb-2">No hay tareas pendientes</h3>
              <p className="text-muted-foreground mb-6">
                Crea tu primera tarea para empezar a organizarte.
              </p>
              <Button onClick={onCreateTask}>
                <Plus className="h-5 w-5 mr-2" />
                Crear tarea
              </Button>
            </>
          ) : (
            <>
              <CheckCircle2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium mb-2">Aún no has completado tareas</h3>
              <p className="text-muted-foreground">
                Las tareas completadas aparecerán aquí. ¡Sigue adelante!
              </p>
            </>
          )}
        </div>
      ) : (
        <div className="space-y-2">
          {displayedTasks.map((task) => (
            <TaskItem
              key={task.id}
              task={task}
              onToggle={() => onToggleTask(task.id)}
              onEdit={() => onEditTask(task)}
              onDelete={() => onDeleteTask(task)}
              isLoading={completingTaskId === task.id}
            />
          ))}
        </div>
      )}
      </div>
    </div>
  );
}
