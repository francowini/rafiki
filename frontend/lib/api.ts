import { Think, ThinkListResponse, NewThink, PaginationParams } from "./types";

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

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new APIError(
        response.status,
        `API request failed: ${response.status} ${errorText}`
      );
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