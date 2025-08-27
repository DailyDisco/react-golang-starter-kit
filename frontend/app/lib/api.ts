export const API_BASE_URL = "http://localhost:8080/api";

export interface User {
  ID: number;
  name: string;
  email: string;
}

export const fetchUsers = async (): Promise<User[]> => {
  const response = await fetch(`${API_BASE_URL}/users`);
  if (!response.ok) {
    throw new Error(`Failed to fetch users: ${response.statusText}`);
  }
  return response.json();
};

export const createUser = async (
  name: string,
  email: string,
): Promise<User> => {
  const response = await fetch(`${API_BASE_URL}/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, email }),
  });
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(
      errorData.error || `Failed to create user: ${response.statusText}`,
    );
  }
  return response.json();
};

export const updateUser = async (user: User): Promise<User> => {
  const response = await fetch(`${API_BASE_URL}/users/${user.ID}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(user),
  });
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(
      errorData.error || `Failed to update user: ${response.statusText}`,
    );
  }
  return response.json();
};

export const deleteUser = async (id: number): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/users/${id}`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
  });
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(
      errorData.error || `Failed to delete user: ${response.statusText}`,
    );
  }
};
