import { Think, ThinkListResponse, NewThink, PaginationParams, Moment, MomentListResponse, NewMoment, UpdateMoment } from "./types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "APIError";
  }
}

async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  // Get token from localStorage
  const token = typeof window !== 'undefined'
    ? localStorage.getItem('auth_token')
    : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
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
      throw new APIError(
        response.status,
        `API request failed: ${response.status} ${errorText}`
      );
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
    throw new Error(`Network error: ${error instanceof Error ? error.message : "Unknown error"}`);
  }
}

export const api = {
  thinks: {
    /**
     * Get all thinks with pagination
     */
    getAll: async (params?: PaginationParams): Promise<ThinkListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.set("page", params.page.toString());
      if (params?.rows) queryParams.set("rows", params.rows.toString());
      if (params?.orderBy) queryParams.set("orderBy", params.orderBy);

      const query = queryParams.toString();
      const endpoint = `/v1/thinks${query ? `?${query}` : ""}`;

      return fetchAPI<ThinkListResponse>(endpoint);
    },


    /**
     * Create a new think
     */
    create: async (data: NewThink): Promise<Think> => {
      return fetchAPI<Think>("/v1/thinks", {
        method: "POST",
        body: JSON.stringify(data),
      });
    },
  },

  moments: {
    /**
     * Get all moments with pagination
     */
    getAll: async (params?: {
      page?: number;
      rows?: number;
      orderBy?: "moment_date" | "intensity" | "date_created" | "date_updated";
      orderDirection?: "asc" | "desc";
    }): Promise<MomentListResponse> => {
      const queryParams = new URLSearchParams();

      if (params?.page) queryParams.set("page", params.page.toString());
      if (params?.rows) queryParams.set("rows", params.rows.toString());
      if (params?.orderBy) queryParams.set("orderBy", params.orderBy);
      if (params?.orderDirection) queryParams.set("orderDirection", params.orderDirection);

      const query = queryParams.toString();
      const endpoint = `/v1/moments${query ? `?${query}` : ""}`;

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
      return fetchAPI<Moment>("/v1/moments", {
        method: "POST",
        body: JSON.stringify(data),
      });
    },

    /**
     * Update an existing moment
     */
    update: async (id: string, data: UpdateMoment): Promise<Moment> => {
      return fetchAPI<Moment>(`/v1/moments/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      });
    },

    /**
     * Delete a moment
     */
    delete: async (id: string): Promise<void> => {
      return fetchAPI<void>(`/v1/moments/${id}`, {
        method: "DELETE",
      });
    },
  },

  health: {
    /**
     * Check API readiness
     */
    readiness: async (): Promise<{ status: string }> => {
      return fetchAPI<{ status: string }>("/v1/readiness");
    },

    /**
     * Check API liveness
     */
    liveness: async (): Promise<{ status: string }> => {
      return fetchAPI<{ status: string }>("/v1/liveness");
    },
  },
};

export { APIError };