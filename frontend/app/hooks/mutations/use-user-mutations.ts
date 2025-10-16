import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { queryKeys } from "../../lib/query-keys";
import { UserService, type User } from "../../services";
import { useUserStore } from "../../stores/user-store";

export const useCreateUser = () => {
  const queryClient = useQueryClient();
  const resetForm = useUserStore((state) => state.resetForm);

  return useMutation({
    mutationFn: ({ name, email, password }: { name: string; email: string; password?: string }) =>
      UserService.createUser(name, email, password),
    onSuccess: (newUser) => {
      queryClient.setQueryData(queryKeys.users.lists(), (old: User[] | undefined) =>
        old ? [...old, newUser] : [newUser]
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      resetForm();
      toast.success("User created successfully");
    },
    onError: (error: Error) => {
      console.error("User creation error:", error);

      // Provide more specific error messages based on common backend responses
      if (error.message.includes("password must contain at least one uppercase")) {
        toast.error("Password validation failed", {
          description: "Password must contain at least one uppercase letter (A-Z)",
        });
      } else if (error.message.includes("password must be at least 8")) {
        toast.error("Password too short", {
          description: "Password must be at least 8 characters long",
        });
      } else if (error.message.includes("Bad Request")) {
        toast.error("Invalid request", {
          description: "Please check your input and try again",
        });
      } else if (error.message.includes("Conflict") || error.message.includes("already exists")) {
        toast.error("User already exists", {
          description: "A user with this email already exists",
        });
      } else {
        toast.error("Failed to create user", {
          description: error.message || "An unexpected error occurred",
        });
      }
    },
  });
};

export const useUpdateUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (user: User) => UserService.updateUser(user),
    onMutate: async (updatedUser) => {
      await queryClient.cancelQueries({
        queryKey: queryKeys.users.detail(updatedUser.id),
      });

      const previousUser = queryClient.getQueryData(queryKeys.users.detail(updatedUser.id));

      queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), updatedUser);

      return { previousUser, updatedUser };
    },
    onError: (err, updatedUser, context) => {
      if (context?.previousUser) {
        queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), context.previousUser);
      }
      toast.error("Failed to update user");
    },
    onSuccess: (updatedUser) => {
      queryClient.setQueryData(queryKeys.users.lists(), (old: User[] | undefined) =>
        old?.map((user) => (user.id === updatedUser.id ? updatedUser : user))
      );
      toast.success("User updated successfully");
    },
  });
};

export const useDeleteUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => UserService.deleteUser(id),
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData(queryKeys.users.lists(), (old: User[] | undefined) =>
        old?.filter((user) => user.id !== deletedId)
      );
      queryClient.removeQueries({
        queryKey: queryKeys.users.detail(deletedId),
      });
      toast.success("User deleted successfully");
    },
    onError: () => {
      toast.error("Failed to delete user");
    },
  });
};
