import { useQuery } from '@tanstack/react-query';
import { useUserStore } from '../../stores/user-store';
import { UserService } from '../../services';
import { queryKeys } from '../../lib/query-keys';
import type { User } from '../../services';
import type { UserFilters } from '../../stores/user-store';

export const useUsers = () => {
  const filters = useUserStore(state => state.filters);

  return useQuery({
    queryKey: queryKeys.users.list(filters as Record<string, unknown>),
    queryFn: () => UserService.fetchUsers(),
    select: data => data,
  });
};

export const useUser = () => {
  const selectedUserId = useUserStore(state => state.selectedUserId);

  return useQuery({
    queryKey: queryKeys.users.detail(selectedUserId!),
    queryFn: () => UserService.getUserById(selectedUserId!),
    enabled: !!selectedUserId,
  });
};
