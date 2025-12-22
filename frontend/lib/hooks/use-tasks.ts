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
  MoveTaskRequest,
} from '@/lib/types';

// ============================================================================
// Query Hooks
// ============================================================================

export function useTasks(filters?: TaskFilters): UseQueryResult<TaskListResponse> {
  return useQuery({
    queryKey: queryKeys.tasks.list(filters),
    queryFn: () => api.tasks.getAll(filters),
  });
}

export function useTasksByObjective(
  objectiveId: string,
  status?: TaskStatus,
): UseQueryResult<TaskListResponse> {
  return useQuery({
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
      // Invalidate ALL task queries (lists + byObjective)
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.all });
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
      // Invalidate ALL task queries (lists + byObjective)
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.all });
    },
  });
}

export function useDeleteTask(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.tasks.delete(id),
    onSuccess: (_, deletedId) => {
      queryClient.removeQueries({ queryKey: queryKeys.tasks.detail(deletedId) });
      // Invalidate ALL task queries (lists + byObjective)
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.all });
    },
  });
}

// ============================================================================
// Complete/Uncomplete with Optimistic Updates
// ============================================================================

interface CompleteTaskContext {
  previousTask?: Task;
  previousLists?: Array<{ key: readonly unknown[]; data: TaskListResponse }>;
}

export function useCompleteTask(): UseMutationResult<CompleteTaskResponse, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskId: string) => api.tasks.complete(taskId),

    // OPTIMISTIC UPDATE
    onMutate: async (taskId): Promise<CompleteTaskContext> => {
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.detail(taskId) });
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.all });

      const previousTask = queryClient.getQueryData<Task>(queryKeys.tasks.detail(taskId));

      if (previousTask) {
        const optimisticTask: Task = {
          ...previousTask,
          status: 'completed',
          completedAt: new Date().toISOString(),
          dateUpdated: new Date().toISOString(),
        };
        queryClient.setQueryData(queryKeys.tasks.detail(taskId), optimisticTask);
      }

      // Snapshot and update all task lists (lists + byObjective)
      const previousLists: Array<{ key: readonly unknown[]; data: TaskListResponse }> = [];
      const listQueries = queryClient.getQueriesData<TaskListResponse>({
        queryKey: queryKeys.tasks.all,
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
      // Invalidate ALL task queries (lists + byObjective)
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.all });

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

interface UncompleteTaskContext {
  previousTask?: Task;
  previousLists?: Array<{ key: readonly unknown[]; data: TaskListResponse }>;
}

export function useUncompleteTask(): UseMutationResult<Task, Error, string, UncompleteTaskContext> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskId: string) => api.tasks.uncomplete(taskId),

    // OPTIMISTIC UPDATE
    onMutate: async (taskId): Promise<UncompleteTaskContext> => {
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.detail(taskId) });
      await queryClient.cancelQueries({ queryKey: queryKeys.tasks.all });

      const previousTask = queryClient.getQueryData<Task>(queryKeys.tasks.detail(taskId));

      if (previousTask) {
        queryClient.setQueryData(queryKeys.tasks.detail(taskId), {
          ...previousTask,
          status: 'pending',
          completedAt: null,
          dateUpdated: new Date().toISOString(),
        });
      }

      // Snapshot and update all task lists (lists + byObjective)
      const previousLists: Array<{ key: readonly unknown[]; data: TaskListResponse }> = [];
      const listQueries = queryClient.getQueriesData<TaskListResponse>({
        queryKey: queryKeys.tasks.all,
      });

      listQueries.forEach(([key, data]) => {
        if (!data) return;
        previousLists.push({ key, data });

        const updatedItems = data.items.map((task) =>
          task.id === taskId
            ? {
                ...task,
                status: 'pending' as TaskStatus,
                completedAt: null,
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
      // Invalidate ALL task queries (lists + byObjective)
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.all });

      if (data?.objectiveId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.objectives.detail(data.objectiveId),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
      }
    },
  });
}

// ============================================================================
// Move Task Hook (Inbox → Objective)
// ============================================================================

export function useMoveTask(): UseMutationResult<
  Task,
  Error,
  { taskId: string; data: MoveTaskRequest }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, data }) => api.tasks.move(taskId, data),
    onSuccess: () => {
      // F20: Invalidate root keys to affect ALL related queries
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.all });
    },
  });
}
