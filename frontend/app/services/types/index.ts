// User-related types
export interface User {
  id: number;
  name: string;
  email: string;
  email_verified: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Authentication types
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

// API Error types
export interface ApiError {
  error: string;
  message?: string;
  statusCode?: number;
}

// Generic API response wrapper
export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
  success: boolean;
}

// Example/demo types (consider removing in production)
export interface ExampleData {
  name: string;
  email: string;
  age: number;
}
