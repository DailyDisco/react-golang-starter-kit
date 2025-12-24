import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import { queryKeys } from "../../lib/query-keys";
import { AuthService, type LoginRequest, type RegisterRequest } from "../../services";
import { useAuthStore } from "../../stores/auth-store";

export const useLogin = () => {
  const { login, setLoading } = useAuthStore();

  return useMutation({
    mutationFn: (credentials: LoginRequest) => AuthService.login(credentials),
    onSuccess: (authData) => {
      login(authData.user);
      AuthService.storeAuthData(authData);
      toast.success("Login successful");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Login failed");
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};

export const useRegister = () => {
  const { login, setLoading } = useAuthStore();

  return useMutation({
    mutationFn: (userData: RegisterRequest) => AuthService.register(userData),
    onSuccess: (authData) => {
      login(authData.user);
      AuthService.storeAuthData(authData);
      toast.success("Registration successful");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Registration failed");
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};
