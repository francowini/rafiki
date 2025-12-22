'use client';

import { useState } from 'react';
import { useTasks, useCompleteTask, useUncompleteTask, useDeleteTask } from '@/lib/hooks/use-tasks';
import { useObjectives, useLogRecordDynamic } from '@/lib/hooks/use-objectives';
import { TaskItem } from '@/components/features/tasks/TaskItem';
import { TaskDeleteDialog } from '@/components/features/tasks/TaskDeleteDialog';
import { AssignToObjectiveDialog } from '@/components/features/tasks/AssignToObjectiveDialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Inbox, Target, ChevronDown, Loader2 } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { ToastAction } from '@/components/ui/toast';
import { detectMilestone } from '@/lib/utils/milestone-utils';
import { formatDateForInput } from '@/lib/date-utils';
import { cn } from '@/lib/utils';
import type { Task, TaskStatus } from '@/lib/types';

export default function EnfoquePage() {
  const [inboxOpen, setInboxOpen] = useState(true);
  const [taskToAssign, setTaskToAssign] = useState<Task | null>(null);
  const [taskToDelete, setTaskToDelete] = useState<Task | null>(null);
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'completed'>('pending');
  const { toast } = useToast();

  const { data: tasksData, isLoading: tasksLoading } = useTasks({
    status: statusFilter === 'all' ? undefined : (statusFilter as TaskStatus),
  });
  const { data: objectivesData } = useObjectives({ status: 'active' });
  const completeMutation = useCompleteTask();
  const uncompleteMutation = useUncompleteTask();
  const deleteMutation = useDeleteTask();
  const logRecordMutation = useLogRecordDynamic();

  const tasks = tasksData?.items ?? [];
  const objectives = objectivesData?.items ?? [];

  // Separate tasks
  const inboxTasks = tasks.filter((t) => !t.objectiveId);
  const objectiveTasks = tasks.filter((t) => t.objectiveId);

  // Group by objective
  const tasksByObjective = objectiveTasks.reduce(
    (acc, task) => {
      if (!task.objectiveId) return acc;
      if (!acc[task.objectiveId]) acc[task.objectiveId] = [];
      acc[task.objectiveId].push(task);
      return acc;
    },
    {} as Record<string, Task[]>,
  );

  const handleToggleTask = async (task: Task) => {
    if (task.status === 'completed') {
      uncompleteMutation.mutate(task.id);
      return;
    }

    completeMutation.mutate(task.id, {
      onSuccess: (response) => {
        const objective = objectives.find((o) => o.id === task.objectiveId);

        // Smart prompt for frequency objectives - prominent toast
        if (response.objectiveTracking === 'frequency' && objective) {
          toast({
            title: '🎯 ¡Tarea completada!',
            description: `¿Registrar hoy como completado en "${objective.title}"?`,
            duration: 15000,
            action: (
              <ToastAction
                altText="Marcar hoy"
                className="bg-green-600 hover:bg-green-700 text-white font-semibold px-4"
                onClick={async () => {
                  // Prevent duplicate submissions
                  if (logRecordMutation.isPending) return;

                  try {
                    await logRecordMutation.mutateAsync({
                      objectiveId: objective.id,
                      data: {
                        recordDate: formatDateForInput(new Date()),
                        status: 'completed',
                      },
                    });
                    toast({ title: '✅ Hoy marcado como Hecho' });
                  } catch (err) {
                    console.error('Failed to mark day:', err);
                    toast({
                      variant: 'destructive',
                      title: 'Error',
                      description: 'No se pudo marcar el día. Intenta de nuevo.',
                    });
                  }
                }}
              >
                ✓ Marcar hoy
              </ToastAction>
            ),
          });
          return;
        }

        // Milestone detection for result objectives
        if (response.objectiveProgress && objective?.targetMetric) {
          const previousProgress = response.objectiveProgress - (task.contribution ?? 0);
          const milestone = detectMilestone(
            previousProgress,
            response.objectiveProgress,
            objective.targetMetric,
          );

          if (milestone) {
            toast({
              title: milestone.title,
              description: `${objective.title}: ${response.objectiveProgress}/${objective.targetMetric}`,
              duration: 8000,
            });
            return;
          }
        }

        // Default toast with undo
        toast({
          title: 'Tarea completada',
          description: response.objectiveProgress
            ? `+${task.contribution} -> ${response.objectiveProgress}`
            : 'Tarea marcada como completada',
          duration: 10000,
          action: (
            <ToastAction altText="Deshacer" onClick={() => uncompleteMutation.mutate(task.id)}>
              Deshacer
            </ToastAction>
          ),
        });
      },
    });
  };

  const handleDeleteTask = async () => {
    if (!taskToDelete) return;
    try {
      await deleteMutation.mutateAsync(taskToDelete.id);
      toast({ title: 'Tarea eliminada' });
      setTaskToDelete(null);
    } catch {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: 'No se pudo eliminar la tarea',
      });
    }
  };

  if (tasksLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Enfoque</h1>
          <p className="text-muted-foreground mt-1">Tus tareas organizadas</p>
        </div>

        {/* Filter Pills */}
        <div className="flex gap-2">
          {(['pending', 'completed', 'all'] as const).map((status) => (
            <Button
              key={status}
              variant={statusFilter === status ? 'default' : 'outline'}
              size="sm"
              onClick={() => setStatusFilter(status)}
            >
              {status === 'pending' && 'Pendientes'}
              {status === 'completed' && 'Completadas'}
              {status === 'all' && 'Todas'}
            </Button>
          ))}
        </div>
      </div>

      {/* Inbox Section (Collapsible) - F5: Accessibility */}
      <Collapsible open={inboxOpen} onOpenChange={setInboxOpen}>
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <CollapsibleTrigger asChild>
            <button
              type="button"
              className="w-full flex items-center justify-between"
              aria-expanded={inboxOpen}
            >
              <div className="flex items-center gap-2">
                <Inbox className="h-5 w-5 text-indigo-600" />
                <h2 className="text-xl font-semibold">Inbox</h2>
                <Badge variant="secondary">{inboxTasks.length}</Badge>
              </div>
              <ChevronDown
                className={cn('h-5 w-5 transition-transform', inboxOpen && 'rotate-180')}
              />
            </button>
          </CollapsibleTrigger>

          <CollapsibleContent className="mt-4">
            {inboxTasks.length === 0 ? (
              <p className="text-muted-foreground text-sm">No hay tareas en inbox</p>
            ) : (
              <div className="space-y-2">
                {inboxTasks.map((task) => (
                  <div key={task.id} className="flex items-center gap-2">
                    <div className="flex-1">
                      <TaskItem
                        task={task}
                        onToggle={() => handleToggleTask(task)}
                        onDelete={() => setTaskToDelete(task)}
                      />
                    </div>
                    <Button variant="outline" size="sm" onClick={() => setTaskToAssign(task)}>
                      Asignar
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CollapsibleContent>
        </div>
      </Collapsible>

      {/* By Objective Section (Accordion) */}
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="flex items-center gap-2 mb-4">
          <Target className="h-5 w-5 text-orange-600" />
          <h2 className="text-xl font-semibold">Por objetivo</h2>
        </div>

        {Object.keys(tasksByObjective).length === 0 ? (
          <p className="text-muted-foreground text-sm">No hay tareas asignadas a objetivos</p>
        ) : (
          <Accordion type="single" collapsible className="w-full">
            {Object.entries(tasksByObjective).map(([objectiveId, objTasks]) => {
              const objective = objectives.find((o) => o.id === objectiveId);
              if (!objective) {
                console.warn(`Tasks reference unknown objective: ${objectiveId}`);
              }
              return (
                <AccordionItem key={objectiveId} value={objectiveId}>
                  <AccordionTrigger>
                    <div className="flex items-center gap-2">
                      <span>{objective?.title ?? 'Objetivo'}</span>
                      <Badge variant="outline">{objTasks.length}</Badge>
                      {objective?.trackingType === 'frequency' && (
                        <Badge variant="secondary" className="bg-purple-50 text-purple-700">
                          Frecuencia
                        </Badge>
                      )}
                    </div>
                  </AccordionTrigger>
                  <AccordionContent>
                    <div className="space-y-2 pt-2">
                      {objTasks.map((task) => (
                        <TaskItem
                          key={task.id}
                          task={task}
                          onToggle={() => handleToggleTask(task)}
                          onDelete={() => setTaskToDelete(task)}
                        />
                      ))}
                    </div>
                  </AccordionContent>
                </AccordionItem>
              );
            })}
          </Accordion>
        )}
      </div>

      {/* Assign Dialog */}
      <AssignToObjectiveDialog
        task={taskToAssign}
        open={!!taskToAssign}
        onOpenChange={(open) => !open && setTaskToAssign(null)}
      />

      {/* Delete Dialog */}
      <TaskDeleteDialog
        task={taskToDelete}
        open={!!taskToDelete}
        onOpenChange={(open) => !open && setTaskToDelete(null)}
        onConfirm={handleDeleteTask}
        isDeleting={deleteMutation.isPending}
      />
    </div>
  );
}
