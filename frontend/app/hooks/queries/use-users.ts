import { useQuery } from "@tanstack/react-query";

import { queryKeys } from "../../lib/query-keys";
import { UserService, type User } from "../../services";
import { useUserStore, type UserFilters } from "../../stores/user-store";

export const useUsers = () => {
  const filters = useUserStore((state) => state.filters);

  return useQuery({
    queryKey: queryKeys.users.list(filters as Record<string, unknown>),
    queryFn: () => UserService.fetchUsers(),
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
