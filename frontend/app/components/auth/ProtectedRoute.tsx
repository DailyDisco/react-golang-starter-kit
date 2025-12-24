import React from "react";

import { Navigate, useLocation } from "@tanstack/react-router";

import { useAuth } from "../../hooks/auth/useAuth";
import { AuthLoadingSkeleton } from "../ui/skeletons";

interface ProtectedRouteProps {
  children: React.ReactNode;
  redirectTo?: string;
}

export function ProtectedRoute({ children, redirectTo = "/login" }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return <AuthLoadingSkeleton />;
  }

  if (!isAuthenticated) {
    // Redirect to login page with return url
    return (
      <Navigate
        to={redirectTo}
        search={{}}
        replace
      />
    );
  }

  return <>{children}</>;
}
