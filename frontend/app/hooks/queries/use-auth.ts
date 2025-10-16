import { useQuery } from "@tanstack/react-query";

import { queryKeys } from "../../lib/query-keys";
import { AuthService } from "../../services";
import { useAuthStore } from "../../stores/auth-store";

export const useCurrentUser = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  return useQuery({
    queryKey: queryKeys.auth.user,
    queryFn: () => AuthService.getCurrentUser(),
    enabled: isAuthenticated,
  });
};
