import { API_BASE_URL, authenticatedFetch, authenticatedFetchWithParsing, parseErrorResponse } from "../api/client";
import type { User } from "../types";

// Activity types matching backend models
export interface ActivityLogItem {
  id: number;
  target_type: string;
  action: string;
  changes?: Record<string, unknown>;
  created_at: string;
}

export interface MyActivityResponse {
  activities: ActivityLogItem[];
  count: number;
  total: number;
}

export interface UserFilters {
  search?: string;
  role?: string;
  isActive?: boolean;
}

export class UserService {
  /**
   * Fetch all users with optional filters
   */
  static async fetchUsers(filters?: UserFilters): Promise<User[]> {
    const params = new URLSearchParams();
    if (filters?.search) params.set("search", filters.search);
    if (filters?.role) params.set("role", filters.role);
    if (filters?.isActive !== undefined) params.set("is_active", String(filters.isActive));

    const queryString = params.toString();
    const url = `${API_BASE_URL}/api/v1/users${queryString ? `?${queryString}` : ""}`;
    const response = await authenticatedFetch(url);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch users");
      throw apiError;
    }

    try {
      const responseData = await response.json();

      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        // Backend returns {"users": null, "count": 0} when no users exist
        return responseData.data.users || [];
      }
      // Fallback for old format (if still in use)
      return responseData.users || responseData || [];
    } catch {
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Create a new user
   */
  static async createUser(name: string, email: string, password?: string): Promise<User> {
    const userData = password ? { name, email, password } : { name, email };

    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users`, {
      method: "POST",
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to create user");
      throw apiError;
    }

    try {
      const responseData = await response.json();
      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data;
      }
      // Fallback for old format
      return responseData;
    } catch {
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Update an existing user
   */
  static async updateUser(user: User): Promise<User> {
    return authenticatedFetchWithParsing<User>(`${API_BASE_URL}/api/v1/users/${user.id}`, {
      method: "PUT",
      body: JSON.stringify(user),
    });
  }

  /**
   * Delete a user
   */
  static async deleteUser(id: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/${id}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to delete user");
      throw apiError;
    }
  }

  /**
   * Get a specific user by ID
   */
  static async getUserById(id: number): Promise<User> {
    return authenticatedFetchWithParsing<User>(`${API_BASE_URL}/api/v1/users/${id}`);
  }

  /**
   * Get current user's activity feed
   * Returns recent audit log entries for the authenticated user
   */
  static async getMyActivity(limit: number = 10): Promise<MyActivityResponse> {
    return authenticatedFetchWithParsing<MyActivityResponse>(`${API_BASE_URL}/api/v1/users/me/activity?limit=${limit}`);
  }

  /**
   * Update the current user's profile
   */
  static async updateCurrentUser(data: { name?: string; bio?: string }): Promise<User> {
    return authenticatedFetchWithParsing<User>(`${API_BASE_URL}/api/v1/users/me`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }
}
