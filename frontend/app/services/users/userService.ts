import {
  API_BASE_URL,
  authenticatedFetch,
  parseErrorResponse,
  authenticatedFetchWithParsing,
} from '../api/client';
import type { User } from '../types';

export class UserService {
  /**
   * Fetch all users
   */
  static async fetchUsers(): Promise<User[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/users`);
    if (!response.ok) {
      const errorMessage = await parseErrorResponse(
        response,
        'Failed to fetch users'
      );
      throw new Error(errorMessage);
    }

    try {
      const responseData = await response.json();
      console.log('Users API response:', responseData); // Debug log

      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        // Backend returns {"users": null, "count": 0} when no users exist
        return responseData.data.users || [];
      }
      // Fallback for old format (if still in use)
      return responseData.users || responseData || [];
    } catch (parseError) {
      console.error('Failed to parse users response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }

  /**
   * Create a new user
   */
  static async createUser(
    name: string,
    email: string,
    password?: string
  ): Promise<User> {
    const userData = password ? { name, email, password } : { name, email };

    const response = await authenticatedFetch(`${API_BASE_URL}/api/users`, {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(
        response,
        'Failed to create user'
      );
      throw new Error(errorMessage);
    }

    try {
      const responseData = await response.json();
      // Handle new success response format
      if (responseData.success === true && responseData.data) {
        return responseData.data;
      }
      // Fallback for old format
      return responseData;
    } catch (parseError) {
      console.error('Failed to parse create user response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }

  /**
   * Update an existing user
   */
  static async updateUser(user: User): Promise<User> {
    return authenticatedFetchWithParsing<User>(
      `${API_BASE_URL}/api/users/${user.id}`,
      {
        method: 'PUT',
        body: JSON.stringify(user),
      }
    );
  }

  /**
   * Delete a user
   */
  static async deleteUser(id: number): Promise<void> {
    const response = await authenticatedFetch(
      `${API_BASE_URL}/api/users/${id}`,
      {
        method: 'DELETE',
      }
    );

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(
        response,
        'Failed to delete user'
      );
      throw new Error(errorMessage);
    }
  }

  /**
   * Get a specific user by ID
   */
  static async getUserById(id: number): Promise<User> {
    return authenticatedFetchWithParsing<User>(
      `${API_BASE_URL}/api/users/${id}`
    );
  }
}
