import { useQuery } from '@tanstack/react-query';
import { API_BASE_URL } from '../../services';
import { queryKeys } from '../../lib/query-keys';

export const useHealthCheck = () => {
  return useQuery({
    queryKey: queryKeys.health.status,
    queryFn: async () => {
      const response = await fetch(`${API_BASE_URL}/api/health`);
      if (!response.ok) throw new Error('Health check failed');
      return response.json();
    },
    refetchInterval: 30000, // Check every 30 seconds
  });
};
