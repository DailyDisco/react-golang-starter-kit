import type { ActivityItem, ActivityType } from "@/components/dashboard/ActivityFeed";
import { useQuery } from "@tanstack/react-query";

import { CACHE_TIMES } from "../../lib/cache-config";
import { queryKeys } from "../../lib/query-keys";
import { UserService, type ActivityLogItem } from "../../services/users/userService";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Maps audit log action/target_type to ActivityFeed type
 */
function mapToActivityType(action: string, targetType: string): ActivityType {
  // Map based on action
  const actionMap: Record<string, ActivityType> = {
    login: "login",
    logout: "logout",
    password_change: "password_change",
    password_reset: "password_change",
    role_change: "security",
    impersonate: "security",
    stop_impersonate: "security",
  };

  if (actionMap[action]) {
    return actionMap[action];
  }

  // Map based on target type
  const targetMap: Record<string, ActivityType> = {
    user: "profile_update",
    subscription: "subscription",
    file: "file_upload",
    settings: "settings_change",
    feature_flag: "settings_change",
  };

  if (targetMap[targetType]) {
    return targetMap[targetType];
  }

  // Map based on action type
  if (action === "create") return "success";
  if (action === "delete") return "warning";
  if (action === "update") return "profile_update";

  return "success";
}

/**
 * Generates human-readable title from audit log
 */
function generateTitle(action: string, targetType: string): string {
  const titles: Record<string, string> = {
    "login:user": "Signed in successfully",
    "logout:user": "Signed out",
    "create:file": "Uploaded a file",
    "delete:file": "Deleted a file",
    "update:user": "Profile updated",
    "password_change:user": "Password changed",
    "password_reset:user": "Password reset",
    "create:subscription": "Subscription activated",
    "update:subscription": "Subscription updated",
    "update:settings": "Settings updated",
    "role_change:user": "Role changed",
    "impersonate:user": "Impersonation started",
    "stop_impersonate:user": "Impersonation ended",
    "create:api_key": "API key created",
    "delete:api_key": "API key deleted",
  };

  const key = `${action}:${targetType}`;
  if (titles[key]) {
    return titles[key];
  }

  // Fallback: format nicely
  const actionLabel = action.charAt(0).toUpperCase() + action.slice(1).replace(/_/g, " ");
  const targetLabel = targetType.replace(/_/g, " ");
  return `${actionLabel} ${targetLabel}`;
}

/**
 * Transforms backend audit logs to ActivityFeed format
 */
function transformToActivityItems(logs: ActivityLogItem[]): ActivityItem[] {
  return logs.map((log) => ({
    id: String(log.id),
    type: mapToActivityType(log.action, log.target_type),
    title: generateTitle(log.action, log.target_type),
    timestamp: log.created_at,
    metadata: log.changes,
  }));
}

/**
 * Hook for dashboard activity feed
 * Fetches current user's recent activity and transforms it for the ActivityFeed component
 */
export function useMyActivity(limit: number = 10) {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: queryKeys.auditLogs.myActivity(limit),
    queryFn: () => UserService.getMyActivity(limit),
    enabled: isAuthenticated,
    staleTime: CACHE_TIMES.USAGE, // 1 minute - activity can change frequently
    select: (data) => transformToActivityItems(data.activities),
  });
}
