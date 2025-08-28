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
export const authenticatedFetch = async (
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

// Simple fetch without auth (for login/register)
export const apiFetch = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const headers = createHeaders(false);

  return fetch(url, {
    ...options,
    headers: {
      ...headers,
      ...options.headers,
    },
  });
};
