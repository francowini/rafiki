import type {
  Think,
  ThinkListResponse,
  NewThink,
  PaginationParams,
  Moment,
  MomentListResponse,
  MomentStats,
  NewMoment,
  UpdateMoment,
  Value,
  ValueListResponse,
  NewValue,
  UpdateValue,
  Facet,
  LifeVision,
  LifeVisionListResponse,
  NewLifeVision,
  UpdateLifeVision,
  ExportParams,
  ExportResponse,
  Objective,
  ObjectiveListResponse,
  NewObjective,
  UpdateObjective,
  ObjectiveStatus,
  TrackingType,
  ObjectiveRecord,
  ObjectiveRecordListResponse,
  NewObjectiveRecord,
  ObjectiveActivityData,
  Task,
  TaskListResponse,
  NewTask,
  UpdateTask,
  TaskStatus,
  CompleteTaskResponse,
} from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';

class APIError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = 'APIError';
  }
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  // Get token from localStorage
  const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // Add Bearer token if available
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  // Merge with options headers
  if (options?.headers) {
    Object.assign(headers, options.headers);
  }

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    });

    // Handle 401 Unauthorized - clear auth and redirect
    if (response.status === 401) {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
      }
      throw new APIError(401, 'Unauthorized - please login again');
    }

    // Handle 403 Forbidden
    if (response.status === 403) {
      throw new APIError(403, 'Forbidden - insufficient permissions');
    }

    if (!response.ok) {
      const errorText = await response.text();
      throw new APIError(response.status, `API request failed: ${response.status} ${errorText}`);
    }

    // Handle 204 No Content responses (e.g., DELETE operations)
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    throw new Error(`Network error: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export const api = {
  thinks: {
    /**
     * Get all thinks with pagination
     */
    getAll: async (params?: PaginationParams): Promise<ThinkListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.set('page', params.page.toString());
      if (params?.rows) queryParams.set('rows', params.rows.toString());
      if (params?.orderBy) queryParams.set('orderBy', params.orderBy);

      const query = queryParams.toString();
      const endpoint = `/v1/thinks${query ? `?${query}` : ''}`;

      return fetchAPI<ThinkListResponse>(endpoint);
    },

    /**
     * Create a new think
     */
    create: async (data: NewThink): Promise<Think> => {
      return fetchAPI<Think>('/v1/thinks', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },
  },

  moments: {
    /**
     * Get all moments with pagination
     * @param orderBy - Combined field and direction, e.g., 'moment_date,DESC'
     *                  Available fields: moment_date, intensity, date_created, date_updated
     *                  Directions: ASC, DESC
     */
    getAll: async (params?: {
      page?: number;
      rows?: number;
      orderBy?: string;
    }): Promise<MomentListResponse> => {
      const queryParams = new URLSearchParams();

      if (params?.page) queryParams.set('page', params.page.toString());
      if (params?.rows) queryParams.set('rows', params.rows.toString());
      if (params?.orderBy) queryParams.set('orderBy', params.orderBy);

      const query = queryParams.toString();
      const endpoint = `/v1/moments${query ? `?${query}` : ''}`;

      return fetchAPI<MomentListResponse>(endpoint);
    },

    /**
     * Get a single moment by ID
     */
    getById: async (id: string): Promise<Moment> => {
      return fetchAPI<Moment>(`/v1/moments/${id}`);
    },

    /**
     * Create a new moment
     */
    create: async (data: NewMoment): Promise<Moment> => {
      return fetchAPI<Moment>('/v1/moments', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    /**
     * Update an existing moment
     */
    update: async (id: string, data: UpdateMoment): Promise<Moment> => {
      return fetchAPI<Moment>(`/v1/moments/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },

    /**
     * Delete a moment
     */
    delete: async (id: string): Promise<void> => {
      return fetchAPI<void>(`/v1/moments/${id}`, {
        method: 'DELETE',
      });
    },

    /**
     * Get moment statistics
     */
    getStats: async (): Promise<MomentStats> => {
      return fetchAPI<MomentStats>('/v1/moments/stats');
    },
  },

  values: {
    /**
     * Get all values for the current user, sorted by display_order
     */
    getAll: async (params?: {
      facet?: Facet;
      includeArchived?: boolean;
    }): Promise<ValueListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.facet) queryParams.set('facet', params.facet);
      if (params?.includeArchived) queryParams.set('includeArchived', 'true');

      const query = queryParams.toString();
      const endpoint = `/v1/values${query ? `?${query}` : ''}`;

      return fetchAPI<ValueListResponse>(endpoint);
    },

    /**
     * Get a single value by ID
     */
    getById: async (id: string): Promise<Value> => {
      return fetchAPI<Value>(`/v1/values/${id}`);
    },

    /**
     * Create a new value
     */
    create: async (data: NewValue): Promise<Value> => {
      return fetchAPI<Value>('/v1/values', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    /**
     * Update an existing value
     */
    update: async (id: string, data: UpdateValue): Promise<Value> => {
      return fetchAPI<Value>(`/v1/values/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },

    /**
     * Delete a value
     */
    delete: async (id: string): Promise<void> => {
      return fetchAPI<void>(`/v1/values/${id}`, {
        method: 'DELETE',
      });
    },

    /**
     * Reorder multiple values in a single atomic operation.
     * Sends all display order updates in one batch request.
     */
    reorder: async (items: { id: string; displayOrder: number }[]): Promise<void> => {
      return fetchAPI<void>('/v1/values/reorder', {
        method: 'POST',
        body: JSON.stringify({ items }),
      });
    },

    /**
     * Archive a value (soft delete).
     * Returns 409 if value has active life visions.
     */
    archive: async (id: string): Promise<Value> => {
      return fetchAPI<Value>(`/v1/values/${id}/archive`, {
        method: 'PUT',
      });
    },

    /**
     * Restore an archived value.
     */
    restore: async (id: string): Promise<Value> => {
      return fetchAPI<Value>(`/v1/values/${id}/restore`, {
        method: 'PUT',
      });
    },
  },

  health: {
    /**
     * Check API readiness
     */
    readiness: async (): Promise<{ status: string }> => {
      return fetchAPI<{ status: string }>('/v1/readiness');
    },

    /**
     * Check API liveness
     */
    liveness: async (): Promise<{ status: string }> => {
      return fetchAPI<{ status: string }>('/v1/liveness');
    },
  },

  lifeVisions: {
    /**
     * Get all life visions for the current user
     */
    getAll: async (params?: {
      valueId?: string;
      includeArchived?: boolean;
    }): Promise<LifeVisionListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.valueId) queryParams.set('valueId', params.valueId);
      if (params?.includeArchived) queryParams.set('includeArchived', 'true');

      const query = queryParams.toString();
      return fetchAPI<LifeVisionListResponse>(`/v1/lifevisions${query ? `?${query}` : ''}`);
    },

    /**
     * Get all life visions for a specific value
     */
    getByValue: async (valueId: string): Promise<LifeVisionListResponse> => {
      return fetchAPI<LifeVisionListResponse>(`/v1/values/${valueId}/lifevisions`);
    },

    /**
     * Get a single life vision by ID
     */
    getById: async (id: string): Promise<LifeVision> => {
      return fetchAPI<LifeVision>(`/v1/lifevisions/${id}`);
    },

    /**
     * Create a new life vision
     */
    create: async (data: NewLifeVision): Promise<LifeVision> => {
      return fetchAPI<LifeVision>('/v1/lifevisions', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    /**
     * Update an existing life vision
     */
    update: async (id: string, data: UpdateLifeVision): Promise<LifeVision> => {
      return fetchAPI<LifeVision>(`/v1/lifevisions/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },

    /**
     * Delete a life vision
     */
    delete: async (id: string): Promise<void> => {
      return fetchAPI<void>(`/v1/lifevisions/${id}`, {
        method: 'DELETE',
      });
    },

    /**
     * Archive a life vision (soft delete).
     */
    archive: async (id: string): Promise<LifeVision> => {
      return fetchAPI<LifeVision>(`/v1/lifevisions/${id}/archive`, {
        method: 'PUT',
      });
    },

    /**
     * Restore an archived life vision.
     */
    restore: async (id: string): Promise<LifeVision> => {
      return fetchAPI<LifeVision>(`/v1/lifevisions/${id}/restore`, {
        method: 'PUT',
      });
    },

    /**
     * Reassign life vision to a different value.
     */
    reassign: async (id: string, newValueId: string): Promise<LifeVision> => {
      return fetchAPI<LifeVision>(`/v1/lifevisions/${id}/reassign`, {
        method: 'PUT',
        body: JSON.stringify({ valueId: newValueId }),
      });
    },
  },

  export: {
    /**
     * Get export items with date range filtering.
     * Rows are clamped to backend limit (1-100).
     */
    getItems: async (params: ExportParams): Promise<ExportResponse> => {
      // Clamp rows to backend limit: min 1, max 100
      const MAX_ROWS = 100;
      const MIN_ROWS = 1;
      const rawRows = params.rows ?? MAX_ROWS;
      const clampedRows = Math.max(MIN_ROWS, Math.min(MAX_ROWS, rawRows));

      const queryParams = new URLSearchParams({
        start_date: params.startDate,
        end_date: params.endDate,
        page: (params.page || 1).toString(),
        rows: clampedRows.toString(),
      });

      // Add orderBy if specified (default: item_date,DESC on backend)
      if (params.orderBy) {
        queryParams.set('orderBy', params.orderBy);
      }

      return fetchAPI<ExportResponse>(`/v1/export?${queryParams.toString()}`);
    },
  },

  objectives: {
    getAll: async (params?: {
      lifeVisionId?: string;
      status?: ObjectiveStatus;
      trackingType?: TrackingType;
      includeArchived?: boolean;
    }): Promise<ObjectiveListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.lifeVisionId) queryParams.set('lifeVisionId', params.lifeVisionId);
      if (params?.status) queryParams.set('status', params.status);
      if (params?.trackingType) queryParams.set('trackingType', params.trackingType);
      if (params?.includeArchived) queryParams.set('includeArchived', 'true');

      const query = queryParams.toString();
      return fetchAPI<ObjectiveListResponse>(`/v1/objectives${query ? `?${query}` : ''}`);
    },

    getById: async (id: string): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}`);
    },

    create: async (data: NewObjective): Promise<Objective> => {
      return fetchAPI<Objective>('/v1/objectives', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    update: async (id: string, data: UpdateObjective): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },

    delete: async (id: string): Promise<void> => {
      return fetchAPI<void>(`/v1/objectives/${id}`, { method: 'DELETE' });
    },

    archive: async (id: string): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}/archive`, { method: 'PUT' });
    },

    restore: async (id: string): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}/restore`, { method: 'PUT' });
    },

    updateStatus: async (id: string, status: ObjectiveStatus): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}/status`, {
        method: 'PUT',
        body: JSON.stringify({ status }),
      });
    },

    incrementProgress: async (id: string, increment: number): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}/progress`, {
        method: 'PUT',
        body: JSON.stringify({ increment }),
      });
    },

    updateProgress: async (id: string, value: number): Promise<Objective> => {
      return fetchAPI<Objective>(`/v1/objectives/${id}/progress`, {
        method: 'PUT',
        body: JSON.stringify({ value }),
      });
    },

    getRecords: async (
      objectiveId: string,
      params?: { startDate?: string; endDate?: string },
    ): Promise<ObjectiveRecordListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.startDate) queryParams.set('startDate', params.startDate);
      if (params?.endDate) queryParams.set('endDate', params.endDate);

      const query = queryParams.toString();
      return fetchAPI<ObjectiveRecordListResponse>(
        `/v1/objectives/${objectiveId}/records${query ? `?${query}` : ''}`,
      );
    },

    logRecord: async (objectiveId: string, data: NewObjectiveRecord): Promise<ObjectiveRecord> => {
      return fetchAPI<ObjectiveRecord>(`/v1/objectives/${objectiveId}/records`, {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    deleteRecord: async (recordId: string): Promise<void> => {
      return fetchAPI<void>(`/v1/objectives/records/${recordId}`, {
        method: 'DELETE',
      });
    },

    getActivity: async (objectiveId: string, year: number): Promise<ObjectiveActivityData> => {
      return fetchAPI<ObjectiveActivityData>(`/v1/objectives/${objectiveId}/activity?year=${year}`);
    },
  },

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
};

export { APIError };
