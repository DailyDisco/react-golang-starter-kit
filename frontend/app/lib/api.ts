// Debug: Log the API URL being used
console.log('🔗 API_BASE_URL:', import.meta.env.VITE_API_URL);

export const API_BASE_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Debug: Log the final API URL
console.log('🚀 Final API_BASE_URL:', API_BASE_URL);

// Get auth token from localStorage
const getAuthToken = (): string | null => {
  return localStorage.getItem('auth_token');
};

// Create headers with optional auth token
const createHeaders = (includeAuth: boolean = true): Record<string, string> => {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (includeAuth) {
    const token = getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  return headers;
};

// Enhanced fetch with auth token and error handling
const authenticatedFetch = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const headers = createHeaders(options.method !== 'GET');

  const response = await fetch(url, {
    ...options,
    headers: {
      ...headers,
      ...options.headers,
    },
  });

  // Handle 401 Unauthorized - token might be expired
  if (response.status === 401) {
    // Clear invalid token
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
    // Redirect to login if on a protected page
    if (window.location.pathname !== '/login') {
      window.location.href = '/login';
    }
  }

  return response;
};

export interface User {
  id: number;
  name: string;
  email: string;
  email_verified: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

// Authentication API functions
export const loginUser = async (
  credentials: LoginRequest
): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Login failed');
  }

  return response.json();
};

export const registerUser = async (
  userData: RegisterRequest
): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(userData),
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Registration failed');
  }

  return response.json();
};

export const getCurrentUser = async (): Promise<User> => {
  const response = await authenticatedFetch(`${API_BASE_URL}/auth/me`);
  if (!response.ok) {
    throw new Error('Failed to fetch current user');
  }
  return response.json();
};

// User management API functions (with authentication)
export const fetchUsers = async (): Promise<User[]> => {
  const response = await authenticatedFetch(`${API_BASE_URL}/users`);
  if (!response.ok) {
    throw new Error(`Failed to fetch users: ${response.statusText}`);
  }
  return response.json();
};

export const createUser = async (
  name: string,
  email: string,
  password: string
): Promise<User> => {
  const response = await authenticatedFetch(`${API_BASE_URL}/users`, {
    method: 'POST',
    body: JSON.stringify({ name, email, password }),
  });
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(
      errorData.error || `Failed to create user: ${response.statusText}`
    );
  }
  return response.json();
};

export const updateUser = async (
  id: number,
  userData: Partial<User>
): Promise<User> => {
  const response = await authenticatedFetch(`${API_BASE_URL}/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(userData),
  });
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(
      errorData.error || `Failed to update user: ${response.statusText}`
    );
  }
  return response.json();
};

export const deleteUser = async (id: number): Promise<void> => {
  const response = await authenticatedFetch(`${API_BASE_URL}/users/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(
      errorData.error || `Failed to delete user: ${response.statusText}`
    );
  }
};
