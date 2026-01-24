import { apiClient } from "../api/client";
// Feature Flags (User-facing)
import type { UserFeatureFlagsResponse } from "../feature-flags/types";
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
  async searchUsers(query: string, limit: number = 10): Promise<{ users: User[]; count: number }> {
    const params = new URLSearchParams();
    params.set("query", query);
    params.set("limit", limit.toString());
    return apiClient.get<{ users: User[]; count: number }>(`/admin/users?${params.toString()}`);
  },

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

export const FeatureFlagService = {
  async getFlags(): Promise<UserFeatureFlagsResponse> {
    return apiClient.get<UserFeatureFlagsResponse>("/feature-flags");
  },
};

// ============ Admin Settings Types ============

// System Settings
export interface EmailSettings {
  smtp_host: string;
  smtp_port: number;
  smtp_username: string;
  smtp_password: string;
  smtp_from_email: string;
  smtp_from_name: string;
  smtp_enabled: boolean;
}

export interface SecuritySettings {
  password_min_length: number;
  password_require_uppercase: boolean;
  password_require_lowercase: boolean;
  password_require_number: boolean;
  password_require_special: boolean;
  session_timeout_minutes: number;
  max_login_attempts: number;
  lockout_duration_minutes: number;
  require_2fa_for_admin: boolean;
  allow_registration: boolean;
}

export interface SiteSettings {
  site_name: string;
  site_description: string;
  site_logo_url: string;
  maintenance_mode: boolean;
  maintenance_message: string;
  contact_email: string;
  support_url: string;
}

// IP Blocklist
export interface IPBlock {
  id: number;
  ip_address: string;
  ip_range?: string;
  reason: string;
  block_type: string;
  hit_count: number;
  is_active: boolean;
  expires_at?: string;
  created_at: string;
}

export interface CreateIPBlockRequest {
  ip_address: string;
  ip_range?: string;
  reason: string;
  block_type?: string;
  expires_in_hours?: number;
}

// Announcements
export type AnnouncementDisplayType = "banner" | "modal";
export type AnnouncementCategory = "update" | "feature" | "bugfix";
export type AnnouncementType = "info" | "warning" | "success" | "error";

export interface Announcement {
  id: number;
  title: string;
  message: string;
  type: AnnouncementType;
  display_type: AnnouncementDisplayType;
  category: AnnouncementCategory;
  link_url?: string;
  link_text?: string;
  is_dismissible: boolean;
  priority: number;
  is_active: boolean;
  target_roles?: string[];
  starts_at?: string;
  ends_at?: string;
  published_at?: string;
  email_sent?: boolean;
  email_sent_at?: string;
}

export interface CreateAnnouncementRequest {
  title: string;
  message: string;
  type?: AnnouncementType;
  display_type?: AnnouncementDisplayType;
  category?: AnnouncementCategory;
  link_url?: string;
  link_text?: string;
  is_dismissible?: boolean;
  priority?: number;
  is_active?: boolean;
  target_roles?: string[];
  starts_at?: string;
  ends_at?: string;
  send_email?: boolean;
}

export interface UpdateAnnouncementRequest {
  title?: string;
  message?: string;
  type?: AnnouncementType;
  display_type?: AnnouncementDisplayType;
  category?: AnnouncementCategory;
  link_url?: string;
  link_text?: string;
  is_dismissible?: boolean;
  priority?: number;
  is_active?: boolean;
  target_roles?: string[];
  starts_at?: string;
  ends_at?: string;
}

// Changelog
export interface ChangelogEntry {
  id: number;
  title: string;
  message: string;
  category: AnnouncementCategory;
  link_url?: string;
  link_text?: string;
  published_at: string;
}

export interface ChangelogMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface ChangelogResponse {
  data: ChangelogEntry[];
  meta: ChangelogMeta;
}

// Email Templates
export interface EmailTemplate {
  id: number;
  key: string;
  name: string;
  description?: string;
  subject: string;
  body_html: string;
  body_text?: string;
  available_variables: TemplateVariable[];
  is_active: boolean;
  is_system: boolean;
  send_count: number;
  last_sent_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TemplateVariable {
  name: string;
  description: string;
}

export interface UpdateEmailTemplateRequest {
  name?: string;
  description?: string;
  subject?: string;
  body_html?: string;
  body_text?: string;
  is_active?: boolean;
}

export interface PreviewEmailTemplateResponse {
  subject: string;
  body_html: string;
  body_text: string;
}

// System Health
export interface SystemHealth {
  status: "healthy" | "degraded" | "unhealthy";
  timestamp: string;
  components: HealthComponent[];
  metrics?: SystemMetrics;
}

export interface HealthComponent {
  name: string;
  status: "healthy" | "degraded" | "unhealthy" | "unavailable";
  message?: string;
  latency?: string;
  last_check?: string;
  details?: Record<string, unknown>;
}

export interface SystemMetrics {
  database?: DatabaseMetrics;
  cache?: CacheMetrics;
  storage?: StorageMetrics;
  api?: APIMetrics;
}

export interface DatabaseMetrics {
  status: string;
  connections_active: number;
  connections_idle: number;
  connections_max: number;
  avg_query_time: string;
  slow_queries: number;
  uptime: string;
}

export interface CacheMetrics {
  status: string;
  memory_used: string;
  memory_max: string;
  hit_rate: number;
  keys: number;
  connections: number;
}

export interface StorageMetrics {
  status: string;
  used: string;
  available: string;
  total: string;
  used_pct: number;
  file_count: number;
}

export interface APIMetrics {
  requests_per_minute: number;
  avg_response_time: string;
  p50_response_time: string;
  p95_response_time: string;
  p99_response_time: string;
  error_rate: number;
}

// Admin Settings Service
export const AdminSettingsService = {
  // Email Settings
  async getEmailSettings(): Promise<EmailSettings> {
    return apiClient.get<EmailSettings>("/admin/settings/email");
  },

  async updateEmailSettings(settings: Partial<EmailSettings>): Promise<void> {
    await apiClient.put("/admin/settings/email", settings);
  },

  async testEmailSettings(): Promise<{ success: boolean; message: string }> {
    return apiClient.post("/admin/settings/email/test", {});
  },

  // Security Settings
  async getSecuritySettings(): Promise<SecuritySettings> {
    return apiClient.get<SecuritySettings>("/admin/settings/security");
  },

  async updateSecuritySettings(settings: Partial<SecuritySettings>): Promise<void> {
    await apiClient.put("/admin/settings/security", settings);
  },

  // Site Settings
  async getSiteSettings(): Promise<SiteSettings> {
    return apiClient.get<SiteSettings>("/admin/settings/site");
  },

  async updateSiteSettings(settings: Partial<SiteSettings>): Promise<void> {
    await apiClient.put("/admin/settings/site", settings);
  },

  // IP Blocklist
  async getIPBlocklist(): Promise<IPBlock[]> {
    return apiClient.get<IPBlock[]>("/admin/ip-blocklist");
  },

  async blockIP(request: CreateIPBlockRequest): Promise<IPBlock> {
    return apiClient.post<IPBlock>("/admin/ip-blocklist", request);
  },

  async unblockIP(id: number): Promise<void> {
    await apiClient.delete(`/admin/ip-blocklist/${id}`);
  },

  // Announcements
  async getAnnouncements(): Promise<Announcement[]> {
    return apiClient.get<Announcement[]>("/admin/announcements");
  },

  async createAnnouncement(request: CreateAnnouncementRequest): Promise<Announcement> {
    return apiClient.post<Announcement>("/admin/announcements", request);
  },

  async updateAnnouncement(id: number, request: UpdateAnnouncementRequest): Promise<Announcement> {
    return apiClient.put<Announcement>(`/admin/announcements/${id}`, request);
  },

  async deleteAnnouncement(id: number): Promise<void> {
    await apiClient.delete(`/admin/announcements/${id}`);
  },

  // Email Templates
  async getEmailTemplates(): Promise<EmailTemplate[]> {
    return apiClient.get<EmailTemplate[]>("/admin/email-templates");
  },

  async getEmailTemplate(id: number): Promise<EmailTemplate> {
    return apiClient.get<EmailTemplate>(`/admin/email-templates/${id}`);
  },

  async updateEmailTemplate(id: number, request: UpdateEmailTemplateRequest): Promise<EmailTemplate> {
    return apiClient.put<EmailTemplate>(`/admin/email-templates/${id}`, request);
  },

  async previewEmailTemplate(id: number, variables: Record<string, string>): Promise<PreviewEmailTemplateResponse> {
    return apiClient.post<PreviewEmailTemplateResponse>(`/admin/email-templates/${id}/preview`, { variables });
  },

  // System Health
  async getSystemHealth(): Promise<SystemHealth> {
    return apiClient.get<SystemHealth>("/admin/health");
  },

  async getDatabaseHealth(): Promise<Record<string, unknown>> {
    return apiClient.get<Record<string, unknown>>("/admin/health/database");
  },

  async getCacheHealth(): Promise<Record<string, unknown>> {
    return apiClient.get<Record<string, unknown>>("/admin/health/cache");
  },
};

// Public Announcements Service
export const AnnouncementService = {
  async getActiveAnnouncements(): Promise<Announcement[]> {
    return apiClient.get<Announcement[]>("/announcements");
  },

  async dismissAnnouncement(id: number): Promise<void> {
    await apiClient.post(`/announcements/${id}/dismiss`, {});
  },

  async getUnreadModalAnnouncements(): Promise<Announcement[]> {
    return apiClient.get<Announcement[]>("/announcements/unread-modals");
  },

  async markAnnouncementRead(id: number): Promise<void> {
    await apiClient.post(`/announcements/${id}/read`, {});
  },
};

// Public Changelog Service
export const ChangelogService = {
  async getChangelog(
    page: number = 1,
    limit: number = 10,
    category?: AnnouncementCategory
  ): Promise<ChangelogResponse> {
    const params = new URLSearchParams();
    params.set("page", String(page));
    params.set("limit", String(limit));
    if (category) {
      params.set("category", category);
    }
    return apiClient.get<ChangelogResponse>(`/changelog?${params.toString()}`);
  },
};
