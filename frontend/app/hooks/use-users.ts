import { useCallback, useEffect, useState } from "react";

import { toast } from "sonner";

import { UserService, type User } from "../services";

interface UseUsersResult {
  users: User[];
  loading: boolean;
  error: string | null;
  addUser: (name: string, email: string) => Promise<void>;
  editUser: (user: User) => Promise<void>;
  removeUser: (id: number) => Promise<void>;
  refreshUsers: () => Promise<void>;
}

export const useUsers = (): UseUsersResult => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refreshUsers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await UserService.fetchUsers();
      setUsers(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An unknown error occurred");
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    void refreshUsers();
  }, [refreshUsers]);

  const addUser = useCallback(async (name: string, email: string) => {
    setLoading(true);
    setError(null);
    try {
      const newUser = await UserService.createUser(name, email);
      setUsers((prevUsers) => [...prevUsers, newUser]);
    } catch (err) {
      const message = err instanceof Error ? err.message : "An unknown error occurred";
      setError(message);
      toast.error(message);
    }
    setLoading(false);
  }, []);

  const editUser = useCallback(async (user: User) => {
    setLoading(true);
    setError(null);
    try {
      const updatedUser = await UserService.updateUser(user);
      setUsers((prevUsers) => prevUsers.map((u) => (u.id === updatedUser.id ? updatedUser : u)));
    } catch (err) {
      const message = err instanceof Error ? err.message : "An unknown error occurred";
      setError(message);
      toast.error(message);
    }
    setLoading(false);
  }, []);

  const removeUser = useCallback(async (id: number) => {
    setLoading(true);
    setError(null);
    try {
      await UserService.deleteUser(id);
      setUsers((prevUsers) => prevUsers.filter((user) => user.id !== id));
    } catch (err) {
      const message = err instanceof Error ? err.message : "An unknown error occurred";
      setError(message);
      toast.error(message);
    }
    setLoading(false);
  }, []);

  return {
    users,
    loading,
    error,
    addUser,
    editUser,
    removeUser,
    refreshUsers,
  };
};
