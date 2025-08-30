import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useUserStore } from '../../stores/user-store';
import { UserService } from '../../services';
import { queryKeys } from '../../lib/query-keys';
import { toast } from 'sonner';
import type { User } from '../../services';

export const useCreateUser = () => {
  const queryClient = useQueryClient();
  const resetForm = useUserStore(state => state.resetForm);

  return useMutation({
    mutationFn: ({
      name,
      email,
      password,
    }: {
      name: string;
      email: string;
      password?: string;
    }) => UserService.createUser(name, email, password),
    onSuccess: newUser => {
      queryClient.setQueryData(
        queryKeys.users.lists(),
        (old: User[] | undefined) => (old ? [...old, newUser] : [newUser])
      );
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      resetForm();
      toast.success('User created successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create user');
    },
  });
};

export const useUpdateUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (user: User) => UserService.updateUser(user),
    onMutate: async updatedUser => {
      await queryClient.cancelQueries({
        queryKey: queryKeys.users.detail(updatedUser.id),
      });

      const previousUser = queryClient.getQueryData(
        queryKeys.users.detail(updatedUser.id)
      );

      queryClient.setQueryData(
        queryKeys.users.detail(updatedUser.id),
        updatedUser
      );

      return { previousUser, updatedUser };
    },
    onError: (err, updatedUser, context) => {
      if (context?.previousUser) {
        queryClient.setQueryData(
          queryKeys.users.detail(updatedUser.id),
          context.previousUser
        );
      }
      toast.error('Failed to update user');
    },
    onSuccess: updatedUser => {
      queryClient.setQueryData(
        queryKeys.users.lists(),
        (old: User[] | undefined) =>
          old?.map(user => (user.id === updatedUser.id ? updatedUser : user))
      );
      toast.success('User updated successfully');
    },
  });
};

export const useDeleteUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => UserService.deleteUser(id),
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData(
        queryKeys.users.lists(),
        (old: User[] | undefined) => old?.filter(user => user.id !== deletedId)
      );
      queryClient.removeQueries({
        queryKey: queryKeys.users.detail(deletedId),
      });
      toast.success('User deleted successfully');
    },
    onError: () => {
      toast.error('Failed to delete user');
    },
  });
};
