'use client';

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { taskSchema, type TaskFormData } from '@/lib/schemas/task-schema';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Slider } from '@/components/ui/slider';
import { Label } from '@/components/ui/label';
import { HelpTooltip } from '@/components/ui/HelpTooltip';
import { Loader2 } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { useCreateTask, useUpdateTask } from '@/lib/hooks/use-tasks';
import type { Task } from '@/lib/types';

interface TaskFormProps {
  objectiveId: string;
  task?: Task | null;
  isFrequencyObjective?: boolean;
  onSuccess: () => void;
  onCancel: () => void;
}

function getContributionLabel(value: number): string {
  if (value <= 3) return 'Pequeño paso';
  if (value <= 6) return 'Avance sólido';
  if (value <= 8) return 'Gran contribución';
  return 'Contribución máxima';
}

export function TaskForm({
  objectiveId,
  task,
  isFrequencyObjective = false,
  onSuccess,
  onCancel,
}: TaskFormProps) {
  const isEditMode = !!task;
  const { toast } = useToast();
  const createMutation = useCreateTask();
  const updateMutation = useUpdateTask();

  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
    setValue,
    watch,
    reset,
  } = useForm<TaskFormData>({
    resolver: zodResolver(taskSchema),
    defaultValues: {
      objectiveId,
      title: '',
      description: '',
      contribution: 1,
    },
    mode: 'onChange',
  });

  const contribution = watch('contribution') ?? 1;
  const titleLength = watch('title')?.length ?? 0;
  const descriptionLength = watch('description')?.length ?? 0;

  useEffect(() => {
    if (task) {
      reset({
        objectiveId: task.objectiveId ?? objectiveId,
        title: task.title,
        description: task.description ?? '',
        contribution: task.contribution ?? 1,
      });
    } else {
      reset({
        objectiveId,
        title: '',
        description: '',
        contribution: 1,
      });
    }
  }, [task, objectiveId, reset]);

  const onSubmit = async (data: TaskFormData) => {
    try {
      // Frequency objectives don't have contribution
      const contributionValue = isFrequencyObjective ? null : data.contribution;

      if (isEditMode && task) {
        await updateMutation.mutateAsync({
          id: task.id,
          data: {
            title: data.title,
            description: data.description || null,
            contribution: contributionValue,
          },
        });
      } else {
        await createMutation.mutateAsync({
          objectiveId: data.objectiveId ?? null,
          title: data.title,
          description: data.description || null,
          contribution: contributionValue,
        });
      }
      onSuccess();
    } catch (error) {
      console.error('Task form submission failed:', error);
      toast({
        variant: 'destructive',
        title: 'Error',
        description: isEditMode
          ? 'No se pudo actualizar la tarea. Intenta de nuevo.'
          : 'No se pudo crear la tarea. Intenta de nuevo.',
      });
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Title */}
      <div className="space-y-2">
        <Label htmlFor="task-title" className="text-base font-medium">
          Título de la tarea
        </Label>
        <Input
          id="task-title"
          {...register('title')}
          placeholder="Ej: Leer capítulo 3 del libro"
          autoFocus
          aria-invalid={!!errors.title}
          aria-describedby={errors.title ? 'title-error' : undefined}
        />
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>{titleLength}/200 caracteres</span>
        </div>
        {errors.title && (
          <p id="title-error" className="text-sm text-destructive">
            {errors.title.message}
          </p>
        )}
      </div>

      {/* Description */}
      <div className="space-y-2">
        <Label htmlFor="task-description" className="text-base font-medium flex items-center gap-2">
          Descripción (opcional)
          <HelpTooltip content="Agrega detalles adicionales sobre esta tarea" />
        </Label>
        <Textarea
          id="task-description"
          {...register('description')}
          placeholder="Detalles adicionales sobre esta tarea..."
          rows={3}
          aria-invalid={!!errors.description}
        />
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>{descriptionLength}/2000 caracteres</span>
        </div>
        {errors.description && (
          <p className="text-sm text-destructive">{errors.description.message}</p>
        )}
      </div>

      {/* Contribution Slider - Only for result objectives */}
      {!isFrequencyObjective && (
        <div className="space-y-4 bg-blue-50 rounded-lg p-4">
          <Label
            htmlFor="task-contribution"
            className="text-base font-medium flex items-center gap-2"
          >
            Contribucion al objetivo
            <HelpTooltip content="Cuanto aporta esta tarea al progreso de tu objetivo (1-10)" />
          </Label>

          <div className="text-center py-2">
            <span className="text-4xl font-bold text-blue-600">+{contribution}</span>
            <p className="text-sm text-muted-foreground mt-1">{getContributionLabel(contribution)}</p>
          </div>

          <Slider
            id="task-contribution"
            min={1}
            max={10}
            step={1}
            value={[contribution]}
            onValueChange={(value) => setValue('contribution', value[0])}
            className="py-4"
            aria-label="Contribucion al objetivo"
          />

          <div className="flex justify-between text-xs text-muted-foreground">
            <span>Pequeno</span>
            <span>Moderado</span>
            <span>Grande</span>
          </div>

          {errors.contribution && (
            <p className="text-sm text-destructive">{errors.contribution.message}</p>
          )}
        </div>
      )}

      {/* Frequency objective notice */}
      {isFrequencyObjective && (
        <div className="p-4 bg-purple-50 rounded-lg border border-purple-200">
          <p className="text-sm text-purple-700">
            Las tareas de objetivos de frecuencia sirven como checklist organizacional y no tienen
            contribucion numerica.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-3 pt-4">
        <Button type="button" variant="outline" onClick={onCancel} className="flex-1">
          Cancelar
        </Button>
        <Button type="submit" disabled={isPending || !isValid} className="flex-1">
          {isPending ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              {isEditMode ? 'Actualizando...' : 'Creando...'}
            </>
          ) : isEditMode ? (
            'Actualizar tarea'
          ) : (
            'Crear tarea'
          )}
        </Button>
      </div>
    </form>
  );
}
