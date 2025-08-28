import {
  API_BASE_URL,
  authenticatedFetch,
  parseErrorResponse,
} from '../api/client';
import type { User } from '../types';

export class UserService {
  /**
   * Fetch all users
   */
  static async fetchUsers(): Promise<User[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/users`);
    if (!response.ok) {
      throw new Error(`Failed to fetch users: ${response.statusText}`);
    }

    try {
      return await response.json();
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
      return await response.json();
    } catch (parseError) {
      console.error('Failed to parse create user response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }

  /**
   * Update an existing user
   */
  static async updateUser(user: User): Promise<User> {
    const response = await authenticatedFetch(
      `${API_BASE_URL}/api/users/${user.id}`,
      {
        method: 'PUT',
        body: JSON.stringify(user),
      }
    );

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(
        response,
        'Failed to update user'
      );
      throw new Error(errorMessage);
    }

    try {
      return await response.json();
    } catch (parseError) {
      console.error('Failed to parse update user response:', parseError);
      throw new Error('Invalid response format from server');
    }
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
    const response = await authenticatedFetch(
      `${API_BASE_URL}/api/users/${id}`
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.statusText}`);
    }

    try {
      return await response.json();
    } catch (parseError) {
      console.error('Failed to parse user response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }
}
