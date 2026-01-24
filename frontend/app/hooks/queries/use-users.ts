import { useQuery } from "@tanstack/react-query";

import { queryKeys } from "../../lib/query-keys";
import { UserService, type User, type UserFilters } from "../../services";
import { useUserStore } from "../../stores/user-store";

export const useUsers = () => {
  const filters = useUserStore((state) => state.filters);

  return useQuery({
    // eslint-disable-next-line @tanstack/query/exhaustive-deps -- filters is correctly included via type assertion
    queryKey: queryKeys.users.list(filters as Record<string, unknown>),
    queryFn: () => UserService.fetchUsers(filters as UserFilters),
    select: (data) => data,
  });
};

export const useUser = () => {
  const selectedUserId = useUserStore((state) => state.selectedUserId);

  return useQuery({
    queryKey: queryKeys.users.detail(selectedUserId!),
    queryFn: () => UserService.getUserById(selectedUserId!),
    enabled: !!selectedUserId,
  });
};
