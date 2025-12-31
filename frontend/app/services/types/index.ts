// Social links for user profiles
export interface SocialLinks {
  twitter?: string;
  github?: string;
  linkedin?: string;
  website?: string;
}

// User-related types
export interface User {
  id: number;
  name: string;
  email: string;
  email_verified: boolean;
  is_active: boolean;
  role?: string;
  created_at: string;
  updated_at: string;
  avatar_url?: string;
  bio?: string;
  location?: string;
  social_links?: string; // JSON string from backend
  two_factor_enabled?: boolean;
}

// Authentication types
export interface AuthResponse {
  user: User;
  token: string;
  refresh_token?: string;
  expires_in?: number; // Access token expiration in seconds
}

export interface RefreshTokenRequest {
  refresh_token: string;
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
  storage_type: "s3" | "database";
  created_at: string;
  updated_at: string;
}

export interface FileResponse {
  id: number;
  file_name: string;
  content_type: string;
  file_size: number;
  location: string;
  storage_type: "s3" | "database";
  created_at: string;
  updated_at: string;
}

export interface StorageStatus {
  storage_type: "s3" | "database";
  message: string;
}

// Example/demo types (consider removing in production)
export interface ExampleData {
  name: string;
  email: string;
  age: number;
}

// Billing/Subscription types
export interface BillingPlan {
  id: string;
  name: string;
  description: string;
  price_id: string;
  amount: number; // Price in cents
  currency: string;
  interval: "month" | "year";
  features?: string[];
}

export interface Subscription {
  id: number;
  user_id: number;
  status: SubscriptionStatus;
  stripe_price_id: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  canceled_at?: string;
  created_at: string;
  updated_at: string;
}

export type SubscriptionStatus = "active" | "past_due" | "canceled" | "trialing" | "unpaid";

export interface BillingConfig {
  publishable_key: string;
}

export interface CheckoutSessionResponse {
  session_id: string;
  url: string;
}

export interface PortalSessionResponse {
  url: string;
}

export interface CreateCheckoutRequest {
  price_id: string;
}
