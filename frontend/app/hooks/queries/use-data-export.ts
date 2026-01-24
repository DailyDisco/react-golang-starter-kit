import { useQuery } from "@tanstack/react-query";

import { queryKeys } from "../../lib/query-keys";
import { SettingsService } from "../../services/settings/settingsService";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Hook to fetch data export status with conditional polling
 * Polls every 5 seconds while status is "pending" or "processing"
 */
export function useDataExportStatus() {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: queryKeys.settings.dataExportStatus(),
    queryFn: () => SettingsService.getDataExportStatus(),
    enabled: isAuthenticated,
    refetchInterval: (query) => {
      const data = query.state.data;
      if (data && (data.status === "pending" || data.status === "processing")) {
        return 5000; // Poll every 5 seconds while processing
      }
      return false; // Stop polling when done
    },
  });
}
