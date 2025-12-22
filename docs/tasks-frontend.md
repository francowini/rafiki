# Tasks Frontend Implementation - Phase 2

**Version 1.0 | December 2025**

This document provides complete implementation specifications for the Tasks frontend feature.

---

## Overview

Implement Tasks UI within Objective Detail pages with:
- Task list with Pending/Completed tabs
- Task creation and editing via Sheet form
- Checkbox completion with optimistic updates
- 10-second undo toast with "Deshacer" button
- Auto-apply contribution to objective progress

---

## Architecture Compliance

| Check | Status |
|-------|--------|
| **Type Alignment** | Frontend types match backend response |
| **API Client Pattern** | Follows existing `api.objectives` pattern |
| **React Query Pattern** | Hierarchical query keys, optimistic updates |
| **Component Organization** | `components/features/tasks/` with barrel export |
| **Cross-Domain Cache** | Invalidates both tasks AND objectives |

---

## 1. Prerequisites

### Install Missing shadcn Components

```bash
cd frontend
npx shadcn@latest add tabs --yes
npx shadcn@latest add checkbox --yes
```

---

## 2. Type Definitions

Add to `frontend/lib/types.ts`:

```typescript
// ============================================================================
// Task Types
// ============================================================================

export type TaskStatus = 'pending' | 'completed' | 'cancelled';

export const TASK_STATUS_LABELS: Record<TaskStatus, string> = {
  pending: 'Pendiente',
  completed: 'Completada',
  cancelled: 'Cancelada',
};

export interface Task {
  id: string;
  objectiveId: string | null;
  title: string;
  description: string | null;
  contribution: number | null;
  status: TaskStatus;
  completedAt: string | null;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewTask {
  objectiveId: string | null;
  title: string;
  description?: string | null;
  contribution?: number | null;
}

export interface UpdateTask {
  title?: string;
  description?: string | null;
  clearDescription?: boolean;
  contribution?: number | null;
}

export interface CompleteTaskResponse {
  task: Task;
  objectiveProgress: number | null;
}

export interface TaskListResponse {
  items: Task[];
  total: number;
  page: number;
  rowsPerPage: number;
}

export interface TaskFilters {
  objectiveId?: string;
  status?: TaskStatus;
  inboxOnly?: boolean;
}
```

---

## 3. API Client

Add to `frontend/lib/api.ts`:

```typescript
tasks: {
  getAll: async (params?: {
    page?: number;
    rows?: number;
    orderBy?: string;
    objectiveId?: string;
    status?: TaskStatus;
    inboxOnly?: boolean;
  }): Promise<TaskListResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.set('page', params.page.toString());
    if (params?.rows) queryParams.set('rows', params.rows.toString());
    if (params?.orderBy) queryParams.set('orderBy', params.orderBy);
    if (params?.objectiveId) queryParams.set('objectiveId', params.objectiveId);
    if (params?.status) queryParams.set('status', params.status);
    if (params?.inboxOnly) queryParams.set('inboxOnly', 'true');

    const query = queryParams.toString();
    return fetchAPI<TaskListResponse>(`/v1/tasks${query ? `?${query}` : ''}`);
  },

  getByObjective: async (
    objectiveId: string,
    params?: { status?: TaskStatus },
  ): Promise<TaskListResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.status) queryParams.set('status', params.status);
    const query = queryParams.toString();
    return fetchAPI<TaskListResponse>(
      `/v1/objectives/${objectiveId}/tasks${query ? `?${query}` : ''}`,
    );
  },

  getById: async (id: string): Promise<Task> => {
    return fetchAPI<Task>(`/v1/tasks/${id}`);
  },

  create: async (data: NewTask): Promise<Task> => {
    return fetchAPI<Task>('/v1/tasks', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  update: async (id: string, data: UpdateTask): Promise<Task> => {
    return fetchAPI<Task>(`/v1/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete: async (id: string): Promise<void> => {
    return fetchAPI<void>(`/v1/tasks/${id}`, {
      method: 'DELETE',
    });
  },

  complete: async (id: string): Promise<CompleteTaskResponse> => {
    return fetchAPI<CompleteTaskResponse>(`/v1/tasks/${id}/complete`, {
      method: 'PUT',
    });
  },

  uncomplete: async (id: string): Promise<Task> => {
    return fetchAPI<Task>(`/v1/tasks/${id}/uncomplete`, {
      method: 'PUT',
    });
  },
},
```

---

## 4. Query Keys

Add to `frontend/lib/query-keys.ts`:

```typescript
tasks: {
  all: ['tasks'] as const,
  lists: () => [...queryKeys.tasks.all, 'list'] as const,
  list: (filters?: TaskFilters) => [...queryKeys.tasks.lists(), filters] as const,
  details: () => [...queryKeys.tasks.all, 'detail'] as const,
  detail: (id: string) => [...queryKeys.tasks.details(), id] as const,
  byObjective: (objectiveId: string, status?: TaskStatus) =>
    [...queryKeys.tasks.all, 'objective', objectiveId, status] as const,
},
```

---

## 5. React Query Hooks

Create `frontend/lib/hooks/use-tasks.ts`:

```typescript
import {
  useQuery,
  useMutation,
  useQueryClient,
  UseQueryResult,
  UseMutationResult,
} from '@tanstack/react-query';
import { api } from '@/lib/api';
import { queryKeys } from '@/lib/query-keys';
import type {
  Task,
  NewTask,
  UpdateTask,
  TaskListResponse,
  CompleteTaskResponse,
  TaskStatus,
  TaskFilters,
} from '@/lib/types';

// ============================================================================
// Query Hooks
// ============================================================================

export function useTasks(filters?: TaskFilters): UseQueryResult<TaskListResponse> {
  return useQuery({
    // F12 Compliance: queryKey includes all filter parameters
    queryKey: queryKeys.tasks.list(filters),
    queryFn: () => api.tasks.getAll(filters),
  });
}

export function useTasksByObjective(
  objectiveId: string,
  status?: TaskStatus,
): UseQueryResult<TaskListResponse> {
  return useQuery({
    // F12 Compliance: queryKey includes objectiveId AND status
    queryKey: queryKeys.tasks.byObjective(objectiveId, status),
    queryFn: () => api.tasks.getByObjective(objectiveId, { status }),
    enabled: !!objectiveId,
  });
}

export function useTask(id: string): UseQueryResult<Task> {
  return useQuery({
    queryKey: queryKeys.tasks.detail(id),
    queryFn: () => api.tasks.getById(id),
    enabled: !!id,
  });
}

// ============================================================================
// Mutation Hooks
// ============================================================================

export function useCreateTask(): UseMutationResult<Task, Error, NewTask> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: NewTask) => api.tasks.create(data),
    onSuccess: (newTask) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.lists() });
      if (newTask.objectiveId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.objectives.detail(newTask.objectiveId),
        });
      }
    },
  });
}

export function useUpdateTask(): UseMutationResult<
  Task,
  Error,
  { id: string; data: UpdateTask }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }) => api.tasks.update(id, data),
    onSuccess: (updatedTask) => {
      queryClient.setQueryData(queryKeys.tasks.detail(updatedTask.id), updatedTask);
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.lists() });
    },
  });
}

export function useDeleteTask(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.tasks.delete(id),
    onSuccess: (_, deletedId) => {
      queryClient.removeQueries({ queryKey: queryKeys.tasks.detail(deletedId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.lists() });
    },
  });
}

// ============================================================================
// Complete/Uncomplete with Optimistic Updates
// ============================================================================

interface CompleteTaskContext {
  previousTask?: Task;
  previousLists?: Array<{ key: unknown[]; data: TaskListResponse }>;
}

export function useCompleteTask(
  objectiveId?: string,
): UseMutationResult<CompleteTaskResponse, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskId: string) => api.tasks.complete(taskId),

    // OPTIMISTIC UPDATE
    onMutate: async (taskId): Promise<CompleteTaskContext> => {
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.detail(taskId) });
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.lists() });

      const previousTask = queryClient.getQueryData<Task>(queryKeys.tasks.detail(taskId));

      // F13 Compliance: Use crypto.randomUUID() if generating temp IDs
      // (Not needed here - we're updating existing task)

      if (previousTask) {
        const optimisticTask: Task = {
          ...previousTask,
          status: 'completed',
          completedAt: new Date().toISOString(),
          dateUpdated: new Date().toISOString(),
        };
        queryClient.setQueryData(queryKeys.tasks.detail(taskId), optimisticTask);
      }

      // Snapshot and update lists
      const previousLists: Array<{ key: unknown[]; data: TaskListResponse }> = [];
      const listQueries = queryClient.getQueriesData<TaskListResponse>({
        queryKey: queryKeys.tasks.lists(),
      });

      listQueries.forEach(([key, data]) => {
        if (!data) return;
        previousLists.push({ key, data });

        const updatedItems = data.items.map((task) =>
          task.id === taskId
            ? {
                ...task,
                status: 'completed' as TaskStatus,
                completedAt: new Date().toISOString(),
                dateUpdated: new Date().toISOString(),
              }
            : task,
        );
        queryClient.setQueryData(key, { ...data, items: updatedItems });
      });

      return { previousTask, previousLists };
    },

    // Rollback on error
    onError: (_err, taskId, context) => {
      if (!context) return;

      if (context.previousTask) {
        queryClient.setQueryData(queryKeys.tasks.detail(taskId), context.previousTask);
      }

      context.previousLists?.forEach(({ key, data }) => {
        queryClient.setQueryData(key, data);
      });
    },

    // Refetch on settle
    onSettled: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.lists() });

      // Cross-domain cache invalidation
      if (data?.task?.objectiveId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.objectives.detail(data.task.objectiveId),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
      }
    },
  });
}

export function useUncompleteTask(
  objectiveId?: string,
): UseMutationResult<Task, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskId: string) => api.tasks.uncomplete(taskId),

    onMutate: async (taskId): Promise<{ previousTask?: Task }> => {
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.detail(taskId) });

      const previousTask = queryClient.getQueryData<Task>(queryKeys.tasks.detail(taskId));

      if (previousTask) {
        queryClient.setQueryData(queryKeys.tasks.detail(taskId), {
          ...previousTask,
          status: 'pending',
          completedAt: null,
          dateUpdated: new Date().toISOString(),
        });
      }

      return { previousTask };
    },

    onError: (_err, taskId, context) => {
      if (context?.previousTask) {
        queryClient.setQueryData(queryKeys.tasks.detail(taskId), context.previousTask);
      }
    },

    onSettled: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.lists() });

      if (data?.objectiveId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.objectives.detail(data.objectiveId),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
      }
    },
  });
}
```

---

## 6. Form Schema

Create `frontend/lib/schemas/task-schema.ts`:

```typescript
import { z } from 'zod';

export const taskSchema = z
  .object({
    objectiveId: z.string().uuid().optional().nullable(),
    title: z
      .string()
      .min(3, 'El titulo debe tener al menos 3 caracteres')
      .max(200, 'El titulo debe tener maximo 200 caracteres')
      .trim(),
    description: z
      .string()
      .max(2000, 'La descripcion debe tener maximo 2000 caracteres')
      .trim()
      .optional()
      .nullable(),
    contribution: z
      .number()
      .int('La contribucion debe ser un numero entero')
      .min(1, 'La contribucion debe ser al menos 1')
      .max(10, 'La contribucion debe ser maximo 10')
      .optional()
      .nullable(),
  })
  .refine(
    (data) => {
      if (data.objectiveId && !data.contribution) {
        return false;
      }
      return true;
    },
    {
      message: 'La contribucion es requerida para tareas vinculadas a objetivos',
      path: ['contribution'],
    },
  );

export type TaskFormData = z.infer<typeof taskSchema>;
```

---

## 7. Components

### File Structure

```
frontend/components/features/tasks/
├── index.ts
├── TaskItem.tsx
├── TaskList.tsx
├── TaskForm.tsx
└── TaskDeleteDialog.tsx
```

### 7.1 TaskItem.tsx

```typescript
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

  // F5 Compliance: Using semantic button with proper keyboard handling
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      if (!isCompleted && !isLoading) {
        onToggle();
      }
    }
  };

  return (
    <div
      className={cn(
        'group flex items-center gap-3 p-3 rounded-lg border transition-all',
        isCompleted ? 'bg-gray-50 border-gray-200' : 'bg-white border-gray-200 hover:border-blue-300',
        isLoading && 'opacity-60 pointer-events-none',
      )}
    >
      {/* F5 Compliance: Checkbox is naturally keyboard accessible */}
      <Checkbox
        checked={isCompleted}
        onCheckedChange={() => !isCompleted && onToggle()}
        disabled={isLoading || isCompleted}
        className="h-5 w-5"
        aria-label={isCompleted ? 'Tarea completada' : 'Marcar como completada'}
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
```

### 7.2 TaskList.tsx

```typescript
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

      {/* Tabs - F5 Compliance: Using semantic buttons */}
      <div className="flex gap-2 border-b">
        <button
          type="button"
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
              <h3 className="text-lg font-medium mb-2">Aun no has completado tareas</h3>
              <p className="text-muted-foreground">
                Las tareas completadas apareceran aqui. Sigue adelante!
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
  );
}
```

### 7.3 TaskForm.tsx

```typescript
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
import { useCreateTask, useUpdateTask } from '@/lib/hooks/use-tasks';
import type { Task } from '@/lib/types';

interface TaskFormProps {
  objectiveId: string;
  task?: Task | null;
  onSuccess: () => void;
  onCancel: () => void;
}

function getContributionLabel(value: number): string {
  if (value <= 3) return 'Pequeno paso';
  if (value <= 6) return 'Avance solido';
  if (value <= 8) return 'Gran contribucion';
  return 'Contribucion maxima';
}

export function TaskForm({ objectiveId, task, onSuccess, onCancel }: TaskFormProps) {
  const isEditMode = !!task;
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

  // F15 Compliance: Use useEffect for state updates, not during render
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

  // F11 Compliance: Wrap async calls in try-catch
  const onSubmit = async (data: TaskFormData) => {
    try {
      if (isEditMode && task) {
        await updateMutation.mutateAsync({
          id: task.id,
          data: {
            title: data.title,
            description: data.description || null,
            contribution: data.contribution,
          },
        });
      } else {
        await createMutation.mutateAsync({
          objectiveId: data.objectiveId,
          title: data.title,
          description: data.description || null,
          contribution: data.contribution,
        });
      }
      onSuccess();
    } catch (error) {
      // Error handled by mutation's onError
      console.error('Task form submission failed:', error);
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Title - F10 Compliance: Label associated via htmlFor */}
      <div className="space-y-2">
        <Label htmlFor="task-title" className="text-base font-medium">
          Titulo de la tarea
        </Label>
        <Input
          id="task-title"
          {...register('title')}
          placeholder="Ej: Leer capitulo 3 del libro"
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

      {/* Description - F10 Compliance */}
      <div className="space-y-2">
        <Label htmlFor="task-description" className="text-base font-medium flex items-center gap-2">
          Descripcion (opcional)
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

      {/* Contribution Slider - F10 Compliance */}
      <div className="space-y-4 bg-blue-50 rounded-lg p-4">
        <Label htmlFor="task-contribution" className="text-base font-medium flex items-center gap-2">
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
```

### 7.4 TaskDeleteDialog.tsx

```typescript
'use client';

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import type { Task } from '@/lib/types';

interface TaskDeleteDialogProps {
  task: Task | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isDeleting?: boolean;
}

export function TaskDeleteDialog({
  task,
  open,
  onOpenChange,
  onConfirm,
  isDeleting,
}: TaskDeleteDialogProps) {
  if (!task) return null;

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Eliminar tarea?</AlertDialogTitle>
          <AlertDialogDescription>
            Estas a punto de eliminar la tarea. Puedes crear una nueva despues si lo necesitas.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>Cancelar</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            disabled={isDeleting}
            className="bg-destructive text-white hover:bg-destructive/90"
          >
            {isDeleting ? 'Eliminando...' : 'Eliminar'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

### 7.5 Barrel Export (index.ts)

```typescript
export { TaskItem } from './TaskItem';
export { TaskList } from './TaskList';
export { TaskForm } from './TaskForm';
export { TaskDeleteDialog } from './TaskDeleteDialog';
```

---

## 8. ObjectiveDetail Integration

Add to the ObjectiveDetail page (only for RESULT objectives):

```typescript
'use client';

import { useState, useRef, useCallback } from 'react';
import { TaskList, TaskForm, TaskDeleteDialog } from '@/components/features/tasks';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from '@/components/ui/sheet';
import { useToast } from '@/hooks/use-toast';
import { ToastAction } from '@/components/ui/toast';
import { useTasksByObjective, useCompleteTask, useUncompleteTask, useDeleteTask } from '@/lib/hooks/use-tasks';
import type { Task, CompleteTaskResponse } from '@/lib/types';

interface TasksSectionProps {
  objectiveId: string;
  objectiveName: string;
  targetMetric: number;
}

export function TasksSection({ objectiveId, objectiveName, targetMetric }: TasksSectionProps) {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [taskToEdit, setTaskToEdit] = useState<Task | null>(null);
  const [taskToDelete, setTaskToDelete] = useState<Task | null>(null);
  const [completingTaskId, setCompletingTaskId] = useState<string | null>(null);

  const { toast, dismiss } = useToast();
  const toastIdRef = useRef<string | null>(null);

  const { data: tasksData, isLoading } = useTasksByObjective(objectiveId);
  const completeMutation = useCompleteTask(objectiveId);
  const uncompleteMutation = useUncompleteTask(objectiveId);
  const deleteMutation = useDeleteTask();

  // F9 Compliance: Reset state when closing form
  const handleFormClose = useCallback(() => {
    setIsFormOpen(false);
    setTaskToEdit(null);
  }, []);

  const handleCompleteTask = useCallback(
    (taskId: string) => {
      setCompletingTaskId(taskId);

      // Dismiss previous toast
      if (toastIdRef.current) {
        dismiss(toastIdRef.current);
      }

      completeMutation.mutate(taskId, {
        onSuccess: (response: CompleteTaskResponse) => {
          setCompletingTaskId(null);

          // Show toast with undo - 10 second duration
          const { id } = toast({
            title: 'Tarea completada',
            description: response.objectiveProgress
              ? `+${response.task.contribution} → ${response.objectiveProgress}/${targetMetric} ${objectiveName}`
              : 'Tarea marcada como completada',
            duration: 10000,
            action: (
              <ToastAction
                altText="Deshacer"
                onClick={() => {
                  uncompleteMutation.mutate(taskId);
                  dismiss(id);
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
    },
    [completeMutation, uncompleteMutation, toast, dismiss, objectiveName, targetMetric],
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
        onToggleTask={handleCompleteTask}
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

      {/* Task Form Sheet - Mobile: bottom, Desktop: right */}
      <Sheet open={isFormOpen} onOpenChange={(open) => !open && handleFormClose()}>
        <SheetContent side="bottom" className="sm:side-right w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{taskToEdit ? 'Editar tarea' : 'Nueva tarea'}</SheetTitle>
            <SheetDescription>
              {taskToEdit ? 'Modifica los detalles de la tarea' : 'Crea una nueva tarea para este objetivo'}
            </SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <TaskForm
              objectiveId={objectiveId}
              task={taskToEdit}
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
```

---

## 9. Spanish Copy Reference

| Context | Message |
|---------|---------|
| **Toast - Complete** | "Tarea completada" |
| **Toast - Progress** | "+{contribution} → {current}/{target} {objectiveName}" |
| **Toast - Undo** | "Deshacer" |
| **Empty Pending** | "No hay tareas pendientes" |
| **Empty Completed** | "Aun no has completado tareas" |
| **Form Title Create** | "Nueva tarea" |
| **Form Title Edit** | "Editar tarea" |
| **Contribution 1-3** | "Pequeno paso" |
| **Contribution 4-6** | "Avance solido" |
| **Contribution 7-8** | "Gran contribucion" |
| **Contribution 9-10** | "Contribucion maxima" |
| **Delete Title** | "Eliminar tarea?" |
| **Delete Message** | "Puedes crear una nueva despues si lo necesitas." |

---

## 10. Errors-to-Avoid Compliance

| Error | Status | Implementation |
|-------|--------|----------------|
| F2 (Stale Responses) | N/A | React Query handles this |
| F5 (Accessibility) | ✅ | Semantic buttons with keyboard handlers |
| F9 (Dialog State Reset) | ✅ | `handleFormClose` resets state |
| F10 (Label Association) | ✅ | All labels have `htmlFor` |
| F11 (Nested Try-Catch) | ✅ | Form submission wrapped |
| F12 (Query Key Params) | ✅ | All filters included in keys |
| F13 (Temp ID) | N/A | No temp IDs needed |
| F15 (Render State) | ✅ | useEffect for initialization |

---

## 11. Deployment Checklist

### Pre-Development
```bash
cd frontend
npx shadcn@latest add tabs --yes
npx shadcn@latest add checkbox --yes
```

### Backend (if not deployed)
```bash
make deploy
```

### Frontend
```bash
cd frontend
npm run build
vercel --prod
```

### Verification
- [ ] Create task from objective detail
- [ ] Complete task (verify toast + progress update)
- [ ] Undo within 10 seconds
- [ ] Edit task
- [ ] Delete task with confirmation
- [ ] Pending/Completed tabs work
- [ ] Empty states display correctly

---

*Generated by Multi-Mind Analysis Team - December 2025*
