import { apiClient } from "../api/client";
import type { User } from "../types";

// Admin Stats
export interface AdminStats {
  total_users: number;
  active_users: number;
  verified_users: number;
  new_users_today: number;
  new_users_this_week: number;
  new_users_this_month: number;
  total_subscriptions: number;
  active_subscriptions: number;
  canceled_subscriptions: number;
  total_files: number;
  total_file_size: number;
  users_by_role: Record<string, number>;
}

// Audit Logs
export interface AuditLog {
  id: number;
  user_id?: number;
  user_name?: string;
  user_email?: string;
  target_type: string;
  target_id?: number;
  action: string;
  changes?: Record<string, unknown>;
  ip_address?: string;
  user_agent?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AuditLogsResponse {
  logs: AuditLog[];
  count: number;
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface AuditLogFilter {
  user_id?: number;
  target_type?: string;
  target_id?: number;
  action?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  limit?: number;
}

// Feature Flags
export interface FeatureFlag {
  id: number;
  key: string;
  name: string;
  description?: string;
  enabled: boolean;
  rollout_percentage: number;
  allowed_roles?: string[];
  created_at: string;
  updated_at: string;
}

export interface FeatureFlagsResponse {
  flags: FeatureFlag[];
  count: number;
}

export interface CreateFeatureFlagRequest {
  key: string;
  name: string;
  description?: string;
  enabled?: boolean;
  rollout_percentage?: number;
  allowed_roles?: string[];
}

export interface UpdateFeatureFlagRequest {
  name?: string;
  description?: string;
  enabled?: boolean;
  rollout_percentage?: number;
  allowed_roles?: string[];
}

// Impersonation
export interface ImpersonateRequest {
  user_id: number;
  reason?: string;
}

export interface ImpersonateResponse {
  user: User;
  token: string;
  original_user_id: number;
}

export const AdminService = {
  // Stats
  async getStats(): Promise<AdminStats> {
    const response = await apiClient.get<AdminStats>("/admin/stats");
    return response;
  },

  // Audit Logs
  async getAuditLogs(filter?: AuditLogFilter): Promise<AuditLogsResponse> {
    const params = new URLSearchParams();
    if (filter?.user_id) params.set("user_id", filter.user_id.toString());
    if (filter?.target_type) params.set("target_type", filter.target_type);
    if (filter?.target_id) params.set("target_id", filter.target_id.toString());
    if (filter?.action) params.set("action", filter.action);
    if (filter?.start_date) params.set("start_date", filter.start_date);
    if (filter?.end_date) params.set("end_date", filter.end_date);
    if (filter?.page) params.set("page", filter.page.toString());
    if (filter?.limit) params.set("limit", filter.limit.toString());

    const queryString = params.toString();
    const url = queryString ? `/admin/audit-logs?${queryString}` : "/admin/audit-logs";
    return apiClient.get<AuditLogsResponse>(url);
  },

  // User Management
  async updateUserRole(userId: number, role: string): Promise<User> {
    return apiClient.put<User>(`/admin/users/${userId}/role`, { role });
  },

  async deactivateUser(userId: number): Promise<void> {
    await apiClient.post(`/admin/users/${userId}/deactivate`, {});
  },

  async reactivateUser(userId: number): Promise<void> {
    await apiClient.post(`/admin/users/${userId}/reactivate`, {});
  },

  // Impersonation
  async impersonateUser(request: ImpersonateRequest): Promise<ImpersonateResponse> {
    return apiClient.post<ImpersonateResponse>("/admin/impersonate", request);
  },

  async stopImpersonation(): Promise<{ user: User; token: string }> {
    return apiClient.post("/admin/stop-impersonate", {});
  },

  // Feature Flags (Admin)
  async getFeatureFlags(): Promise<FeatureFlagsResponse> {
    return apiClient.get<FeatureFlagsResponse>("/admin/feature-flags");
  },

  async createFeatureFlag(request: CreateFeatureFlagRequest): Promise<FeatureFlag> {
    return apiClient.post<FeatureFlag>("/admin/feature-flags", request);
  },

  async updateFeatureFlag(key: string, request: UpdateFeatureFlagRequest): Promise<FeatureFlag> {
    return apiClient.put<FeatureFlag>(`/admin/feature-flags/${key}`, request);
  },

  async deleteFeatureFlag(key: string): Promise<void> {
    await apiClient.delete(`/admin/feature-flags/${key}`);
  },

  async setUserFeatureFlagOverride(userId: number, key: string, enabled: boolean): Promise<void> {
    await apiClient.put(`/admin/users/${userId}/feature-flags/${key}`, { enabled });
  },

  async deleteUserFeatureFlagOverride(userId: number, key: string): Promise<void> {
    await apiClient.delete(`/admin/users/${userId}/feature-flags/${key}`);
  },
};

// Feature Flags (User-facing)
export const FeatureFlagService = {
  async getFlags(): Promise<Record<string, boolean>> {
    return apiClient.get<Record<string, boolean>>("/feature-flags");
  },
};
