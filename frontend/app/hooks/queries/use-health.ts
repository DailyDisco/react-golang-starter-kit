import { useQuery } from "@tanstack/react-query";

import { CACHE_TIMES } from "../../lib/cache-config";
import { queryKeys } from "../../lib/query-keys";
import { API_BASE_URL } from "../../services";

export const useHealthCheck = () => {
  return useQuery({
    queryKey: queryKeys.health.status,
    queryFn: async () => {
      const response = await fetch(`${API_BASE_URL}/api/v1/health`);
      if (!response.ok) throw new Error("Health check failed");
      return response.json();
    },
    staleTime: CACHE_TIMES.HEALTH,
    refetchInterval: 30000, // Check every 30 seconds
  });
};
