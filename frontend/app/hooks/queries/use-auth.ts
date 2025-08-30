import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '../../stores/auth-store';
import { AuthService } from '../../services';
import { queryKeys } from '../../lib/query-keys';

export const useCurrentUser = () => {
  const isAuthenticated = useAuthStore(state => state.isAuthenticated);

  return useQuery({
    queryKey: queryKeys.auth.user,
    queryFn: () => AuthService.getCurrentUser(),
    enabled: isAuthenticated,
  });
};
