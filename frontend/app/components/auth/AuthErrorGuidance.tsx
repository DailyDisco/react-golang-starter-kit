import { Link } from "@tanstack/react-router";
import { HelpCircle, LogIn, UserPlus } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ApiError } from "../../services/api/client";

interface AuthErrorGuidanceProps {
  error: ApiError | Error | null;
  context: "login" | "register";
}

export function AuthErrorGuidance({ error, context }: AuthErrorGuidanceProps) {
  const { t } = useTranslation("auth");

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
            {t("errors.troubleSigningIn")}
          </p>
          <div className="flex flex-col gap-1 pl-5">
            <p className="text-muted-foreground">{t("errors.checkCredentials")}</p>
            <Link
              to="/register"
              className="text-primary flex items-center gap-1 hover:underline"
            >
              <UserPlus className="h-3 w-3" />
              {t("errors.createAccount")}
            </Link>
          </div>
        </div>
      );
    }

    // Account deactivated
    if (errorCode === "ACCOUNT_INACTIVE") {
      return (
        <div className="text-muted-foreground mt-3 text-sm">
          <p>{t("errors.accountDeactivated")}</p>
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
          <p className="text-muted-foreground">{t("errors.emailExists")}</p>
          <Link
            to="/login"
            className="text-primary flex items-center gap-1 hover:underline"
          >
            <LogIn className="h-3 w-3" />
            {t("errors.signInExisting")}
          </Link>
        </div>
      );
    }
  }

  return null;
}
