import { API_BASE_URL, authenticatedFetch, parseErrorResponse } from "../api/client";

// Types for usage metering
export interface UsageTotals {
  api_calls: number;
  storage_bytes: number;
  compute_ms: number;
  file_uploads: number;
}

export interface UsageLimits {
  api_calls: number;
  storage_bytes: number;
  compute_ms: number;
  file_uploads: number;
}

export interface UsagePercentages {
  api_calls: number;
  storage_bytes: number;
  compute_ms: number;
  file_uploads: number;
}

export interface UsageSummary {
  period_start: string;
  period_end: string;
  totals: UsageTotals;
  limits: UsageLimits;
  percentages: UsagePercentages;
  limits_exceeded: boolean;
}

export interface UsageAlert {
  id: number;
  alert_type: string;
  usage_type: string;
  current_usage: number;
  usage_limit: number;
  percentage_used: number;
  acknowledged: boolean;
  created_at: string;
}

export interface UsageHistoryResponse {
  history: UsageSummary[];
  count: number;
}

export interface UsageAlertsResponse {
  alerts: UsageAlert[];
  count: number;
}

export class UsageService {
  /**
   * Get current billing period usage summary
   */
  static async getCurrentUsage(): Promise<UsageSummary> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/usage`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to get usage");
    }

    return response.json();
  }

  /**
   * Get usage history for past billing periods
   */
  static async getUsageHistory(months: number = 6): Promise<UsageHistoryResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/usage/history?months=${months}`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to get usage history");
    }

    return response.json();
  }

  /**
   * Get unacknowledged usage alerts
   */
  static async getAlerts(): Promise<UsageAlertsResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/usage/alerts`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to get alerts");
    }

    return response.json();
  }

  /**
   * Acknowledge a usage alert
   */
  static async acknowledgeAlert(alertId: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/usage/alerts/${alertId}/acknowledge`, {
      method: "POST",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to acknowledge alert");
    }
  }

  /**
   * Format bytes to human-readable string
   */
  static formatBytes(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  }

  /**
   * Format milliseconds to human-readable time
   */
  static formatComputeTime(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    const seconds = ms / 1000;
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    const minutes = seconds / 60;
    if (minutes < 60) return `${minutes.toFixed(1)}m`;
    const hours = minutes / 60;
    return `${hours.toFixed(1)}h`;
  }

  /**
   * Format large numbers with K/M suffixes
   */
  static formatCount(count: number): string {
    if (count < 1000) return count.toString();
    if (count < 1000000) return (count / 1000).toFixed(1) + "K";
    return (count / 1000000).toFixed(1) + "M";
  }

  /**
   * Get alert type display info
   */
  static getAlertTypeInfo(alertType: string): { label: string; color: string } {
    switch (alertType) {
      case "exceeded":
        return { label: "Exceeded", color: "destructive" };
      case "warning_90":
        return { label: "90% Used", color: "warning" };
      case "warning_80":
        return { label: "80% Used", color: "warning" };
      default:
        return { label: "Alert", color: "default" };
    }
  }

  /**
   * Get usage type display name
   */
  static getUsageTypeLabel(usageType: string): string {
    switch (usageType) {
      case "api_call":
        return "API Calls";
      case "storage":
        return "Storage";
      case "compute":
        return "Compute Time";
      case "file_upload":
        return "File Uploads";
      default:
        return usageType;
    }
  }
}
