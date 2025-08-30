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

// API Response types (matching backend)
export interface ApiSuccessResponse<T = any> {
  success: true;
  message: string;
  data?: T;
}

export interface ApiErrorResponse {
  error: string;
  message: string;
  code: number;
}

export type ApiResponse<T = any> = ApiSuccessResponse<T> | ApiErrorResponse;

// Legacy API Error type (keeping for backward compatibility)
export interface ApiError {
  error: string;
  message?: string;
  statusCode?: number;
}

// File-related types
export interface File {
  id: number;
  file_name: string;
  content_type: string;
  file_size: number;
  location: string;
  storage_type: 's3' | 'database';
  created_at: string;
  updated_at: string;
}

export interface FileResponse {
  id: number;
  file_name: string;
  content_type: string;
  file_size: number;
  location: string;
  storage_type: 's3' | 'database';
  created_at: string;
  updated_at: string;
}

export interface StorageStatus {
  storage_type: 's3' | 'database';
  message: string;
}

// Example/demo types (consider removing in production)
export interface ExampleData {
  name: string;
  email: string;
  age: number;
}
