import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { logger } from "../../lib/logger";
import { showMutationError, showMutationSuccess } from "../../lib/mutation-toast";
import { queryKeys } from "../../lib/query-keys";
import { UserService, type User } from "../../services";
import { useUserStore } from "../../stores/user-store";

// Temporary ID prefix for optimistic items
const TEMP_ID_PREFIX = "temp_";

interface OptimisticUser extends User {
  _optimistic?: boolean;
}

export const useCreateUser = () => {
  const queryClient = useQueryClient();
  const resetForm = useUserStore((state) => state.resetForm);

  const mutation = useMutation({
    mutationFn: ({ name, email, password }: { name: string; email: string; password?: string }) =>
      UserService.createUser(name, email, password),

    // Optimistic add with temporary ID
    onMutate: async (variables) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: queryKeys.users.lists() });

      // Snapshot previous value
      const previousUsers = queryClient.getQueryData<OptimisticUser[]>(queryKeys.users.lists());

      // Create optimistic user
      const optimisticUser: OptimisticUser = {
        id: Date.now(), // Temporary ID
        name: variables.name,
        email: variables.email,
        email_verified: false,
        is_active: true,
        role: "user",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        _optimistic: true,
      };

      // Add optimistic user to cache
      queryClient.setQueryData<OptimisticUser[]>(queryKeys.users.lists(), (old) =>
        old ? [...old, optimisticUser] : [optimisticUser]
      );

      return { previousUsers, optimisticUser };
    },

    onSuccess: (newUser, _, context) => {
      // Replace optimistic user with real user
      queryClient.setQueryData<OptimisticUser[]>(
        queryKeys.users.lists(),
        (old) => old?.map((user) => (user._optimistic ? newUser : user)) ?? [newUser]
      );
      resetForm();
      showMutationSuccess({ message: "User created successfully" });
    },

    onError: (error: Error, variables, context) => {
      logger.error("User creation error", error);

      // Rollback optimistic update
      if (context?.previousUsers) {
        queryClient.setQueryData(queryKeys.users.lists(), context.previousUsers);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(variables),
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
    },
  });

  return mutation;
};

export const useUpdateUser = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (user: User) => UserService.updateUser(user),

    onMutate: async (updatedUser) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: queryKeys.users.detail(updatedUser.id),
      });
      await queryClient.cancelQueries({
        queryKey: queryKeys.users.lists(),
      });

      // Snapshot previous values
      const previousUser = queryClient.getQueryData<User>(queryKeys.users.detail(updatedUser.id));
      const previousUsers = queryClient.getQueryData<User[]>(queryKeys.users.lists());

      // Optimistically update both caches
      queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), updatedUser);
      queryClient.setQueryData<User[]>(queryKeys.users.lists(), (old) =>
        old?.map((user) => (user.id === updatedUser.id ? updatedUser : user))
      );

      return { previousUser, previousUsers, updatedUser };
    },

    onError: (error: Error, updatedUser, context) => {
      logger.error("User update error", error);

      // Rollback optimistic updates
      if (context?.previousUser) {
        queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), context.previousUser);
      }
      if (context?.previousUsers) {
        queryClient.setQueryData(queryKeys.users.lists(), context.previousUsers);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(updatedUser),
      });
    },

    onSuccess: (updatedUser) => {
      // Update caches with server response
      queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), updatedUser);
      queryClient.setQueryData<User[]>(queryKeys.users.lists(), (old) =>
        old?.map((user) => (user.id === updatedUser.id ? updatedUser : user))
      );
      showMutationSuccess({ message: "User updated successfully" });
    },

    onSettled: (_, __, updatedUser) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(updatedUser.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
    },
  });

  return mutation;
};

export const useDeleteUser = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (id: number) => UserService.deleteUser(id),

    // Optimistic delete
    onMutate: async (deletedId) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: queryKeys.users.lists() });

      // Snapshot previous value
      const previousUsers = queryClient.getQueryData<User[]>(queryKeys.users.lists());

      // Optimistically remove from cache
      queryClient.setQueryData<User[]>(queryKeys.users.lists(), (old) => old?.filter((user) => user.id !== deletedId));

      return { previousUsers, deletedId };
    },

    onSuccess: (_, deletedId) => {
      queryClient.removeQueries({
        queryKey: queryKeys.users.detail(deletedId),
      });
      showMutationSuccess({ message: "User deleted successfully" });
    },

    onError: (error: Error, deletedId, context) => {
      logger.error("User deletion error", error);

      // Rollback optimistic update
      if (context?.previousUsers) {
        queryClient.setQueryData(queryKeys.users.lists(), context.previousUsers);
      }

      showMutationError({
        error,
        onRetry: () => mutation.mutate(deletedId),
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
    },
  });

  return mutation;
};
