import {
  useQuery,
  useMutation,
  useQueryClient,
  UseQueryResult,
  UseMutationResult,
} from '@tanstack/react-query';
import { api } from '@/lib/api';
import { queryKeys, ObjectiveFilters } from '@/lib/query-keys';
import type {
  Objective,
  NewObjective,
  UpdateObjective,
  ObjectiveListResponse,
  ObjectiveRecord,
  NewObjectiveRecord,
  ObjectiveRecordListResponse,
  ObjectiveActivityData,
  ObjectiveStatus,
} from '@/lib/types';

// ============================================================================
// Query Hooks
// ============================================================================

export function useObjectives(filters?: ObjectiveFilters): UseQueryResult<ObjectiveListResponse> {
  return useQuery({
    queryKey: queryKeys.objectives.list(filters),
    queryFn: () => api.objectives.getAll(filters),
  });
}

export function useObjective(id: string): UseQueryResult<Objective> {
  return useQuery({
    queryKey: queryKeys.objectives.detail(id),
    queryFn: () => api.objectives.getById(id),
    enabled: !!id,
  });
}

export function useObjectiveRecords(
  objectiveId: string,
  params?: { startDate?: string; endDate?: string },
): UseQueryResult<ObjectiveRecordListResponse> {
  return useQuery({
    queryKey: queryKeys.objectives.records(objectiveId, params),
    queryFn: () => api.objectives.getRecords(objectiveId, params),
    enabled: !!objectiveId,
  });
}

export function useObjectiveActivity(
  objectiveId: string,
  year: number,
): UseQueryResult<ObjectiveActivityData> {
  return useQuery({
    queryKey: queryKeys.objectives.activity(objectiveId, year),
    queryFn: () => api.objectives.getActivity(objectiveId, year),
    enabled: !!objectiveId && !!year,
  });
}

// ============================================================================
// Mutation Hooks
// ============================================================================

export function useCreateObjective(): UseMutationResult<Objective, Error, NewObjective> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: NewObjective) => api.objectives.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useUpdateObjective(): UseMutationResult<
  Objective,
  Error,
  { id: string; data: UpdateObjective }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }) => api.objectives.update(id, data),
    onSuccess: (updatedObjective) => {
      queryClient.setQueryData(
        queryKeys.objectives.detail(updatedObjective.id),
        updatedObjective,
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useArchiveObjective(): UseMutationResult<Objective, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.objectives.archive(id),
    onSuccess: (archivedObjective) => {
      queryClient.setQueryData(
        queryKeys.objectives.detail(archivedObjective.id),
        archivedObjective,
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useRestoreObjective(): UseMutationResult<Objective, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.objectives.restore(id),
    onSuccess: (restoredObjective) => {
      queryClient.setQueryData(
        queryKeys.objectives.detail(restoredObjective.id),
        restoredObjective,
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useDeleteObjective(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.objectives.delete(id),
    onSuccess: (_, deletedId) => {
      queryClient.removeQueries({ queryKey: queryKeys.objectives.detail(deletedId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useUpdateObjectiveStatus(): UseMutationResult<
  Objective,
  Error,
  { id: string; status: ObjectiveStatus }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }) => api.objectives.updateStatus(id, status),
    onSuccess: (updatedObjective) => {
      queryClient.setQueryData(
        queryKeys.objectives.detail(updatedObjective.id),
        updatedObjective,
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useIncrementProgress(): UseMutationResult<
  Objective,
  Error,
  { id: string; increment: number }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, increment }) => api.objectives.incrementProgress(id, increment),
    onSuccess: (updatedObjective) => {
      queryClient.setQueryData(
        queryKeys.objectives.detail(updatedObjective.id),
        updatedObjective,
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

export function useUpdateProgress(): UseMutationResult<
  Objective,
  Error,
  { id: string; value: number }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, value }) => api.objectives.updateProgress(id, value),
    onSuccess: (updatedObjective) => {
      queryClient.setQueryData(
        queryKeys.objectives.detail(updatedObjective.id),
        updatedObjective,
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.objectives.lists() });
    },
  });
}

// ============================================================================
// Record Hooks with Optimistic Updates
// ============================================================================

export function useLogRecord(
  objectiveId: string,
): UseMutationResult<ObjectiveRecord, Error, NewObjectiveRecord> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: NewObjectiveRecord) => api.objectives.logRecord(objectiveId, data),

    // OPTIMISTIC UPDATE
    onMutate: async (newRecord) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: queryKeys.objectives.records(objectiveId),
      });

      // Snapshot previous value
      const previousRecords = queryClient.getQueryData<ObjectiveRecordListResponse>(
        queryKeys.objectives.records(objectiveId),
      );

      // Optimistically update cache
      if (previousRecords) {
        const optimisticRecord: ObjectiveRecord = {
          id: crypto.randomUUID(),
          objectiveId,
          recordDate: newRecord.recordDate,
          status: newRecord.status,
          notes: newRecord.notes,
          dateCreated: new Date().toISOString(),
        };

        queryClient.setQueryData<ObjectiveRecordListResponse>(
          queryKeys.objectives.records(objectiveId),
          {
            ...previousRecords,
            items: [optimisticRecord, ...previousRecords.items],
            total: previousRecords.total + 1,
          },
        );
      }

      return { previousRecords };
    },

    // Rollback on error
    onError: (_err, _newRecord, context) => {
      if (context?.previousRecords) {
        queryClient.setQueryData(
          queryKeys.objectives.records(objectiveId),
          context.previousRecords,
        );
      }
    },

    // Refetch on settle
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.records(objectiveId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.activity(objectiveId, new Date().getFullYear()),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.detail(objectiveId),
      });
    },
  });
}

// Dynamic version of useLogRecord for cases where objectiveId varies per call
// (e.g., toast actions in Enfoque page)
export function useLogRecordDynamic(): UseMutationResult<
  ObjectiveRecord,
  Error,
  { objectiveId: string; data: NewObjectiveRecord }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ objectiveId, data }) => api.objectives.logRecord(objectiveId, data),
    onSuccess: (_, { objectiveId }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.records(objectiveId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.activity(objectiveId, new Date().getFullYear()),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.detail(objectiveId),
      });
    },
  });
}

export function useDeleteRecord(objectiveId: string): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (recordId: string) => api.objectives.deleteRecord(recordId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.records(objectiveId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.objectives.activity(objectiveId, new Date().getFullYear()),
      });
    },
  });
}
