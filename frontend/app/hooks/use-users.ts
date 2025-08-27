import { useState, useEffect, useCallback } from "react";
import {
  type User,
  fetchUsers,
  createUser,
  updateUser,
  deleteUser,
} from "../lib/api";
import { toast } from "sonner";

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
      const data = await fetchUsers();
      setUsers(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "An unknown error occurred",
      );
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    refreshUsers();
  }, [refreshUsers]);

  const addUser = useCallback(async (name: string, email: string) => {
    setLoading(true);
    setError(null);
    try {
      const newUser = await createUser(name, email);
      setUsers((prevUsers) => [...prevUsers, newUser]);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "An unknown error occurred";
      setError(message);
      toast.error(message);
    }
    setLoading(false);
  }, []);

  const editUser = useCallback(async (user: User) => {
    setLoading(true);
    setError(null);
    try {
      const updatedUser = await updateUser(user);
      setUsers((prevUsers) =>
        prevUsers.map((u) => (u.ID === updatedUser.ID ? updatedUser : u)),
      );
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "An unknown error occurred";
      setError(message);
      toast.error(message);
    }
    setLoading(false);
  }, []);

  const removeUser = useCallback(async (id: number) => {
    setLoading(true);
    setError(null);
    try {
      await deleteUser(id);
      setUsers((prevUsers) => prevUsers.filter((user) => user.ID !== id));
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "An unknown error occurred";
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
