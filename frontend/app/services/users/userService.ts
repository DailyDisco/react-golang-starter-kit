import { API_BASE_URL, authenticatedFetch } from '../api/client';
import type { User } from '../types';

export class UserService {
  /**
   * Fetch all users
   */
  static async fetchUsers(): Promise<User[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/users`);
    if (!response.ok) {
      throw new Error(`Failed to fetch users: ${response.statusText}`);
    }
    return response.json();
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

    const response = await authenticatedFetch(`${API_BASE_URL}/users`, {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(
        errorData.error || `Failed to create user: ${response.statusText}`
      );
    }

    return response.json();
  }

  /**
   * Update an existing user
   */
  static async updateUser(user: User): Promise<User> {
    const response = await authenticatedFetch(
      `${API_BASE_URL}/users/${user.id}`,
      {
        method: 'PUT',
        body: JSON.stringify(user),
      }
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(
        errorData.error || `Failed to update user: ${response.statusText}`
      );
    }

    return response.json();
  }

  /**
   * Delete a user
   */
  static async deleteUser(id: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/users/${id}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(
        errorData.error || `Failed to delete user: ${response.statusText}`
      );
    }
  }

  /**
   * Get a specific user by ID
   */
  static async getUserById(id: number): Promise<User> {
    const response = await authenticatedFetch(`${API_BASE_URL}/users/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.statusText}`);
    }
    return response.json();
  }
}
