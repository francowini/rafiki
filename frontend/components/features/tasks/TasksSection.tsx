'use client';

import { useState, useRef, useCallback } from 'react';
import { TaskList } from './TaskList';
import { TaskForm } from './TaskForm';
import { TaskDeleteDialog } from './TaskDeleteDialog';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet';
import { useToast } from '@/hooks/use-toast';
import { ToastAction } from '@/components/ui/toast';
import {
  useTasksByObjective,
  useCompleteTask,
  useUncompleteTask,
  useDeleteTask,
} from '@/lib/hooks/use-tasks';
import { useLogRecordDynamic } from '@/lib/hooks/use-objectives';
import type { Task, CompleteTaskResponse } from '@/lib/types';

interface TasksSectionProps {
  objectiveId: string;
  objectiveName: string;
  targetMetric?: number;
  trackingType: 'result' | 'frequency';
}

export function TasksSection({
  objectiveId,
  objectiveName,
  targetMetric,
  trackingType,
}: TasksSectionProps) {
  const isFrequencyObjective = trackingType === 'frequency';
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [taskToEdit, setTaskToEdit] = useState<Task | null>(null);
  const [taskToDelete, setTaskToDelete] = useState<Task | null>(null);
  const [completingTaskId, setCompletingTaskId] = useState<string | null>(null);

  const { toast, dismiss } = useToast();
  const toastIdRef = useRef<string | null>(null);

  const { data: tasksData, isLoading } = useTasksByObjective(objectiveId);
  const completeMutation = useCompleteTask();
  const uncompleteMutation = useUncompleteTask();
  const deleteMutation = useDeleteTask();
  const logRecordMutation = useLogRecordDynamic();

  const handleFormClose = useCallback(() => {
    setIsFormOpen(false);
    setTaskToEdit(null);
  }, []);

  const handleToggleTask = useCallback(
    (taskId: string) => {
      const task = tasksData?.items.find((t) => t.id === taskId);
      if (!task) return;

      setCompletingTaskId(taskId);

      // Dismiss previous toast
      if (toastIdRef.current) {
        dismiss(toastIdRef.current);
      }

      if (task.status === 'completed') {
        // Uncomplete the task
        uncompleteMutation.mutate(taskId, {
          onSuccess: () => {
            setCompletingTaskId(null);
            toast({
              title: 'Tarea reabierta',
              description: 'La tarea ha vuelto a pendiente',
            });
          },
          onError: () => {
            setCompletingTaskId(null);
            toast({
              variant: 'destructive',
              title: 'Error',
              description: 'No se pudo reabrir la tarea',
            });
          },
        });
      } else {
        // Complete the task
        completeMutation.mutate(taskId, {
          onSuccess: (response: CompleteTaskResponse) => {
            setCompletingTaskId(null);

            // For frequency objectives, show toast with "Marcar hoy" action
            if (isFrequencyObjective) {
              const { id } = toast({
                title: 'Tarea completada',
                description: `¿Marcar hoy como Hecho en "${objectiveName}"?`,
                duration: 10000,
                action: (
                  <ToastAction
                    altText="Marcar hoy"
                    onClick={async () => {
                      try {
                        await logRecordMutation.mutateAsync({
                          objectiveId,
                          data: {
                            recordDate: new Date().toISOString().split('T')[0],
                            status: 'completed',
                          },
                        });
                        dismiss(id);
                        toast({ title: 'Hoy marcado como Hecho' });
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
                    Marcar hoy
                  </ToastAction>
                ),
              });
              toastIdRef.current = id;
              return;
            }

            // Build description for result objectives
            let description = 'Tarea marcada como completada';
            if (response.objectiveProgress && targetMetric) {
              description = `+${response.task.contribution} -> ${response.objectiveProgress}/${targetMetric} ${objectiveName}`;
            }

            // Show toast with undo - 10 second duration
            const { id } = toast({
              title: 'Tarea completada',
              description,
              duration: 10000,
              action: (
                <ToastAction
                  altText="Deshacer"
                  onClick={() => {
                    uncompleteMutation.mutate(taskId, {
                      onSuccess: () => {
                        dismiss(id);
                      },
                      onError: () => {
                        dismiss(id);
                        toast({
                          variant: 'destructive',
                          title: 'Error',
                          description: 'No se pudo deshacer la acción',
                        });
                      },
                    });
                  }}
                >
                  Deshacer
                </ToastAction>
              ),
            });

            toastIdRef.current = id;
          },
          onError: () => {
            setCompletingTaskId(null);
            toast({
              variant: 'destructive',
              title: 'Error',
              description: 'No se pudo completar la tarea',
            });
          },
        });
      }
    },
    [tasksData, completeMutation, uncompleteMutation, logRecordMutation, toast, dismiss, objectiveId, objectiveName, targetMetric, isFrequencyObjective],
  );

  const handleDeleteTask = useCallback(() => {
    if (!taskToDelete) return;

    deleteMutation.mutate(taskToDelete.id, {
      onSuccess: () => {
        setTaskToDelete(null);
        toast({
          title: 'Tarea eliminada',
          description: 'La tarea ha sido eliminada',
        });
      },
      onError: () => {
        toast({
          variant: 'destructive',
          title: 'Error',
          description: 'No se pudo eliminar la tarea',
        });
      },
    });
  }, [taskToDelete, deleteMutation, toast]);

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6">
      <TaskList
        tasks={tasksData?.items ?? []}
        isLoading={isLoading}
        onToggleTask={handleToggleTask}
        onEditTask={(task) => {
          setTaskToEdit(task);
          setIsFormOpen(true);
        }}
        onDeleteTask={setTaskToDelete}
        onCreateTask={() => {
          setTaskToEdit(null);
          setIsFormOpen(true);
        }}
        completingTaskId={completingTaskId}
      />

      {/* Task Form Sheet */}
      <Sheet open={isFormOpen} onOpenChange={(open) => !open && handleFormClose()}>
        <SheetContent className="w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{taskToEdit ? 'Editar tarea' : 'Nueva tarea'}</SheetTitle>
            <SheetDescription>
              {taskToEdit
                ? 'Modifica los detalles de la tarea'
                : 'Crea una nueva tarea para este objetivo'}
            </SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <TaskForm
              objectiveId={objectiveId}
              task={taskToEdit}
              isFrequencyObjective={isFrequencyObjective}
              onSuccess={handleFormClose}
              onCancel={handleFormClose}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Delete Confirmation */}
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
