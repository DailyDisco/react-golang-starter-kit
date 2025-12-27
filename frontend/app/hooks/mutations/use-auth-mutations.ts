import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import { queryKeys } from "../../lib/query-keys";
import { AuthService, type LoginRequest, type RegisterRequest } from "../../services";
import { ApiError } from "../../services/api/client";
import { useAuthStore } from "../../stores/auth-store";

export const useLogin = () => {
  const { login, logout, setLoading } = useAuthStore();

  return useMutation({
    mutationFn: (credentials: LoginRequest) => AuthService.login(credentials),
    onSuccess: (authData) => {
      login(authData.user);
      AuthService.storeAuthData(authData);

      // Start session heartbeat to detect expiration
      AuthService.startSessionHeartbeat(5 * 60 * 1000, () => {
        logout();
        if (typeof window !== "undefined") {
          window.dispatchEvent(new CustomEvent("session-expired"));
        }
      });

      toast.success("Login successful");
    },
    onError: (error: Error) => {
      if (error instanceof ApiError) {
        switch (error.code) {
          case "UNAUTHORIZED":
            toast.error("Unable to sign in", {
              description: "Please check your email and password, or create a new account.",
            });
            break;
          case "ACCOUNT_INACTIVE":
            toast.error("Account deactivated", {
              description: "Please contact support for assistance.",
            });
            break;
          default:
            toast.error(error.message || "Login failed");
        }
      } else {
        toast.error(error.message || "Login failed");
      }
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};

export const useRegister = () => {
  const { login, logout, setLoading } = useAuthStore();

  return useMutation({
    mutationFn: (userData: RegisterRequest) => AuthService.register(userData),
    onSuccess: (authData) => {
      login(authData.user);
      AuthService.storeAuthData(authData);

      // Start session heartbeat to detect expiration
      AuthService.startSessionHeartbeat(5 * 60 * 1000, () => {
        logout();
        if (typeof window !== "undefined") {
          window.dispatchEvent(new CustomEvent("session-expired"));
        }
      });

      toast.success("Registration successful");
    },
    onError: (error: Error) => {
      if (error instanceof ApiError) {
        switch (error.code) {
          case "CONFLICT":
            toast.error("Email already registered", {
              description: "Try signing in instead, or reset your password.",
            });
            break;
          case "BAD_REQUEST":
          case "VALIDATION_ERROR":
            toast.error("Invalid input", {
              description: error.message,
            });
            break;
          default:
            toast.error(error.message || "Registration failed");
        }
      } else {
        toast.error(error.message || "Registration failed");
      }
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};
