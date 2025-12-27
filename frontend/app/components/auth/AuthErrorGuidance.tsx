import { Link } from "@tanstack/react-router";
import { HelpCircle, LogIn, UserPlus } from "lucide-react";

import { ApiError } from "../../services/api/client";

interface AuthErrorGuidanceProps {
  error: ApiError | Error | null;
  context: "login" | "register";
}

export function AuthErrorGuidance({ error, context }: AuthErrorGuidanceProps) {
  if (!error) return null;

  const errorCode = error instanceof ApiError ? error.code : null;

  // Login context guidance
  if (context === "login") {
    // For "Invalid credentials" - show help message with link to create account
    // We show generic guidance to avoid revealing if email exists (security)
    if (errorCode === "UNAUTHORIZED") {
      return (
        <div className="mt-3 space-y-2 text-sm">
          <p className="text-muted-foreground flex items-center gap-1">
            <HelpCircle className="h-4 w-4" />
            Having trouble signing in?
          </p>
          <div className="flex flex-col gap-1 pl-5">
            <p className="text-muted-foreground">
              Double-check your email and password, or create a new account if you don&apos;t have one yet.
            </p>
            <Link
              to="/register"
              className="text-primary flex items-center gap-1 hover:underline"
            >
              <UserPlus className="h-3 w-3" />
              Create a new account
            </Link>
          </div>
        </div>
      );
    }

    // Account deactivated
    if (errorCode === "ACCOUNT_INACTIVE") {
      return (
        <div className="text-muted-foreground mt-3 text-sm">
          <p>Your account has been deactivated. Please contact support for assistance.</p>
        </div>
      );
    }
  }

  // Register context guidance
  if (context === "register") {
    // User already exists - suggest login
    if (errorCode === "CONFLICT") {
      return (
        <div className="mt-3 space-y-2 text-sm">
          <p className="text-muted-foreground">An account with this email already exists.</p>
          <Link
            to="/login"
            className="text-primary flex items-center gap-1 hover:underline"
          >
            <LogIn className="h-3 w-3" />
            Sign in to your existing account
          </Link>
        </div>
      );
    }
  }

  return null;
}
