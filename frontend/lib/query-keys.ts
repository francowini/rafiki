export interface ObjectiveFilters {
  lifeVisionId?: string;
  status?: 'activo' | 'completado' | 'abandonado' | 'pausado';
  trackingType?: 'resultado' | 'frecuencia';
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
    records: (id: string) =>
      [...queryKeys.objectives.detail(id), 'records'] as const,
    activity: (id: string, year: number) =>
      [...queryKeys.objectives.detail(id), 'activity', year] as const,
  },

  lifeVisions: {
    all: ['lifeVisions'] as const,
    lists: () => [...queryKeys.lifeVisions.all, 'list'] as const,
    list: (filters?: { valueId?: string; includeArchived?: boolean }) =>
      [...queryKeys.lifeVisions.lists(), filters] as const,
  },
} as const;
