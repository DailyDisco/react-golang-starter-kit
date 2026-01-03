import { API_BASE_URL, authenticatedFetch, createHeaders, parseErrorResponse } from "../api/client";

// User Preferences Types
export interface UserPreferences {
  id: number;
  user_id: number;
  theme: "light" | "dark" | "system";
  timezone: string;
  language: string;
  date_format: string;
  time_format: "12h" | "24h";
  email_notifications: EmailNotificationSettings;
  created_at: string;
  updated_at: string;
}

export interface EmailNotificationSettings {
  marketing: boolean;
  security: boolean;
  updates: boolean;
  weekly_digest: boolean;
}

export interface UpdatePreferencesRequest {
  theme?: "light" | "dark" | "system";
  timezone?: string;
  language?: string;
  date_format?: string;
  time_format?: "12h" | "24h";
  email_notifications?: EmailNotificationSettings;
}

// Session Types
export interface UserSession {
  id: number;
  user_id: number;
  device_info: string;
  ip_address: string;
  location: string;
  is_current: boolean;
  last_active_at: string;
  created_at: string;
  expires_at: string;
}

// Login History Types
export interface LoginHistoryEntry {
  id: number;
  user_id: number;
  ip_address: string;
  user_agent: string;
  device_type: string;
  browser: string;
  os: string;
  location: string;
  success: boolean;
  failure_reason?: string;
  created_at: string;
}

// 2FA Types
export interface TwoFactorSetupResponse {
  secret: string;
  qr_code: string;
  backup_codes: string[];
}

export interface TwoFactorVerifyRequest {
  code: string;
}

export interface BackupCodesResponse {
  backup_codes: string[];
}

// Password Change Types
export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

// Account Deletion Types
export interface AccountDeletionRequest {
  password: string;
  reason?: string;
}

// Data Export Types
export interface DataExportStatus {
  id: number;
  status: "pending" | "processing" | "completed" | "failed";
  download_url?: string;
  expires_at?: string;
  created_at: string;
}

// Connected Account Types
export interface ConnectedAccount {
  provider: "google" | "github";
  provider_user_id: string;
  email: string;
  connected_at: string;
}

// API Key Types
export interface UserAPIKey {
  id: number;
  provider: "gemini" | "openai" | "anthropic";
  name: string;
  key_preview: string;
  is_active: boolean;
  last_used_at?: string;
  usage_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateAPIKeyRequest {
  provider: "gemini" | "openai" | "anthropic";
  name: string;
  api_key: string;
}

export interface UpdateAPIKeyRequest {
  name?: string;
  api_key?: string;
  is_active?: boolean;
}

export interface UserAPIKeysResponse {
  keys: UserAPIKey[];
  count: number;
}

export class SettingsService {
  // ==================== API Keys ====================

  static async getAPIKeys(): Promise<UserAPIKey[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/api-keys`);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch API keys");
      throw apiError;
    }
    const data = await response.json();
    return data.keys || [];
  }

  static async getAPIKey(id: number): Promise<UserAPIKey> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/api-keys/${id}`);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch API key");
      throw apiError;
    }
    return response.json();
  }

  static async createAPIKey(req: CreateAPIKeyRequest): Promise<UserAPIKey> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/api-keys`, {
      method: "POST",
      body: JSON.stringify(req),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to create API key");
      throw apiError;
    }
    return response.json();
  }

  static async updateAPIKey(id: number, req: UpdateAPIKeyRequest): Promise<UserAPIKey> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/api-keys/${id}`, {
      method: "PUT",
      body: JSON.stringify(req),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to update API key");
      throw apiError;
    }
    return response.json();
  }

  static async deleteAPIKey(id: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/api-keys/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to delete API key");
      throw apiError;
    }
  }

  static async testAPIKey(id: number): Promise<{ success: boolean; message: string }> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/api-keys/${id}/test`, {
      method: "POST",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to test API key");
      throw apiError;
    }
    return response.json();
  }

  // ==================== User Preferences ====================

  static async getPreferences(): Promise<UserPreferences> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/preferences`);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch preferences");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async updatePreferences(req: UpdatePreferencesRequest): Promise<UserPreferences> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/preferences`, {
      method: "PUT",
      body: JSON.stringify(req),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to update preferences");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  // ==================== Sessions ====================

  static async getSessions(): Promise<UserSession[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/sessions`);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch sessions");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async revokeSession(sessionId: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/sessions/${sessionId}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to revoke session");
      throw apiError;
    }
  }

  static async revokeAllSessions(): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/sessions`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to revoke all sessions");
      throw apiError;
    }
  }

  // ==================== Login History ====================

  static async getLoginHistory(limit: number = 20): Promise<LoginHistoryEntry[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/login-history?limit=${limit}`);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch login history");
      throw apiError;
    }
    const data = await response.json();
    return data.history || [];
  }

  // ==================== Password ====================

  static async changePassword(req: ChangePasswordRequest): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/password`, {
      method: "PUT",
      body: JSON.stringify(req),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to change password");
      throw apiError;
    }
  }

  // ==================== Two-Factor Authentication ====================

  static async setup2FA(): Promise<TwoFactorSetupResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/2fa/setup`, {
      method: "POST",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to setup 2FA");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async verify2FA(code: string): Promise<BackupCodesResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/2fa/verify`, {
      method: "POST",
      body: JSON.stringify({ code }),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to verify 2FA");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async disable2FA(code: string): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/2fa/disable`, {
      method: "POST",
      body: JSON.stringify({ code }),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to disable 2FA");
      throw apiError;
    }
  }

  static async regenerateBackupCodes(code: string): Promise<BackupCodesResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/2fa/backup-codes`, {
      method: "POST",
      body: JSON.stringify({ code }),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to regenerate backup codes");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  // ==================== Profile / Avatar ====================

  static async updateProfile(data: {
    name?: string;
    email?: string;
    bio?: string;
    location?: string;
    social_links?: string;
  }): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to update profile");
      throw apiError;
    }
  }

  static async uploadAvatar(file: File): Promise<{ avatar_url: string }> {
    const formData = new FormData();
    formData.append("avatar", file);

    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/avatar`, {
      method: "POST",
      headers: {}, // Let browser set content-type for FormData
      body: formData,
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to upload avatar");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async deleteAvatar(): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/avatar`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to delete avatar");
      throw apiError;
    }
  }

  // ==================== Privacy / Account ====================

  static async requestDataExport(): Promise<DataExportStatus> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/export`, {
      method: "POST",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to request data export");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async getDataExportStatus(): Promise<DataExportStatus | null> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/export`);
    if (response.status === 404) {
      return null;
    }
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to get export status");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async requestAccountDeletion(password: string, reason?: string): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/delete`, {
      method: "POST",
      body: JSON.stringify({ password, reason }),
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to request account deletion");
      throw apiError;
    }
  }

  static async cancelAccountDeletion(): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/delete`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to cancel account deletion");
      throw apiError;
    }
  }

  // ==================== Connected Accounts ====================

  static async getConnectedAccounts(): Promise<ConnectedAccount[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/connected-accounts`);
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to fetch connected accounts");
      throw apiError;
    }
    const data = await response.json();
    return data.success ? data.data : data;
  }

  static async disconnectAccount(provider: string): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/users/me/connected-accounts/${provider}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Failed to disconnect account");
      throw apiError;
    }
  }
}

// Static data for settings forms
export const TIMEZONES = [
  "UTC",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Anchorage",
  "Pacific/Honolulu",
  "America/Phoenix",
  "America/Toronto",
  "America/Vancouver",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Europe/Madrid",
  "Europe/Rome",
  "Europe/Amsterdam",
  "Europe/Stockholm",
  "Europe/Moscow",
  "Asia/Dubai",
  "Asia/Kolkata",
  "Asia/Singapore",
  "Asia/Hong_Kong",
  "Asia/Tokyo",
  "Asia/Seoul",
  "Asia/Shanghai",
  "Australia/Sydney",
  "Australia/Melbourne",
  "Pacific/Auckland",
];

export const LANGUAGES = [
  { code: "en", name: "English" },
  { code: "es", name: "Spanish" },
  { code: "fr", name: "French" },
  { code: "de", name: "German" },
  { code: "it", name: "Italian" },
  { code: "pt", name: "Portuguese" },
  { code: "ja", name: "Japanese" },
  { code: "ko", name: "Korean" },
  { code: "zh", name: "Chinese" },
];

export const DATE_FORMATS = [
  { value: "MM/DD/YYYY", label: "MM/DD/YYYY (12/31/2024)" },
  { value: "DD/MM/YYYY", label: "DD/MM/YYYY (31/12/2024)" },
  { value: "YYYY-MM-DD", label: "YYYY-MM-DD (2024-12-31)" },
  { value: "DD.MM.YYYY", label: "DD.MM.YYYY (31.12.2024)" },
  { value: "YYYY/MM/DD", label: "YYYY/MM/DD (2024/12/31)" },
];
