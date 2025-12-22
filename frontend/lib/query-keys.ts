import type { TaskFilters, TaskStatus } from './types';

export interface ObjectiveFilters {
  lifeVisionId?: string;
  status?: 'active' | 'completed' | 'abandoned' | 'paused';
  trackingType?: 'result' | 'frequency';
  includeArchived?: boolean;
}

export const queryKeys = {
  objectives: {
    all: ['objectives'] as const,
    lists: () => [...queryKeys.objectives.all, 'list'] as const,
    list: (filters?: ObjectiveFilters) =>
      [...queryKeys.objectives.lists(), filters] as const,
    details: () => [...queryKeys.objectives.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.objectives.details(), id] as const,
    records: (id: string, params?: { startDate?: string; endDate?: string }) =>
      [...queryKeys.objectives.detail(id), 'records', params] as const,
    activity: (id: string, year: number) =>
      [...queryKeys.objectives.detail(id), 'activity', year] as const,
  },

  lifeVisions: {
    all: ['lifeVisions'] as const,
    lists: () => [...queryKeys.lifeVisions.all, 'list'] as const,
    list: (filters?: { valueId?: string; includeArchived?: boolean }) =>
      [...queryKeys.lifeVisions.lists(), filters] as const,
  },

  tasks: {
    all: ['tasks'] as const,
    lists: () => [...queryKeys.tasks.all, 'list'] as const,
    list: (filters?: TaskFilters) => [...queryKeys.tasks.lists(), filters] as const,
    details: () => [...queryKeys.tasks.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.tasks.details(), id] as const,
    byObjective: (objectiveId: string, status?: TaskStatus) =>
      [...queryKeys.tasks.all, 'objective', objectiveId, status] as const,
  },
} as const;
