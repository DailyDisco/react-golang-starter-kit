// Create API_BASE_URL safely for SSR
const getApiBaseUrl = () => {
  // In SSR, import.meta.env might not be available or populated
  if (typeof window === 'undefined') {
    return 'http://localhost:8080';
  }
  return import.meta.env.VITE_API_URL || 'http://localhost:8080';
};

export const API_BASE_URL = getApiBaseUrl();

// Debug: Log the API URL being used (only on client side)
if (typeof window !== 'undefined') {
  console.log('🔗 API_BASE_URL:', import.meta.env.VITE_API_URL);
  console.log('🚀 Final API_BASE_URL:', API_BASE_URL);
}

// Get auth token from localStorage
const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('auth_token');
};

// Create headers with optional auth token
export const createHeaders = (
  includeAuth: boolean = true
): Record<string, string> => {
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
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
      // Redirect to login if on a protected page
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
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

/**
 * Safely parse error response from server
 */
export const parseErrorResponse = async (
  response: Response,
  defaultMessage: string
): Promise<string> => {
  try {
    // Get response as text first, then try to parse as JSON
    const responseText = await response.text();

    // Try to parse as JSON if it looks like JSON
    if (
      responseText.trim().startsWith('{') ||
      responseText.trim().startsWith('[')
    ) {
      const errorData = JSON.parse(responseText);
      return errorData.error || errorData.message || defaultMessage;
    } else {
      // Not JSON, use the text directly
      return responseText || `${defaultMessage} with status ${response.status}`;
    }
  } catch (parseError) {
    // If anything fails, use a generic error message
    console.error('Failed to parse error response:', parseError);
    return `${defaultMessage} with status ${response.status}`;
  }
};
