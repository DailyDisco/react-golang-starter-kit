import { useEffect, useState } from "react";

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { z } from "zod";

import { logger } from "../../lib/logger";
import { AuthService } from "../../services/auth/authService";
import { useAuthStore } from "../../stores/auth-store";

const callbackSearchSchema = z.object({
  success: z.string().optional(),
  error: z.string().optional(),
  new_user: z.string().optional(),
});

export const Route = createFileRoute("/auth/callback")({
  validateSearch: callbackSearchSchema,
  component: OAuthCallback,
});

function OAuthCallback() {
  const navigate = useNavigate();
  const search = Route.useSearch();
  const { login } = useAuthStore();
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [errorMessage, setErrorMessage] = useState<string>("");

  useEffect(() => {
    const handleCallback = async () => {
      // Check for error from OAuth provider
      if (search.error) {
        setStatus("error");
        setErrorMessage(search.error);
        logger.error("OAuth callback error", { error: search.error });
        return;
      }

      // Check for success
      if (search.success === "true") {
        try {
          // Fetch the current user to get their data
          const user = await AuthService.getCurrentUser();

          if (user) {
            // Update auth store - login expects a full User object
            login(user);

            setStatus("success");

            // Redirect after a short delay
            setTimeout(() => {
              const isNewUser = search.new_user === "true";
              if (isNewUser) {
                // Could redirect to onboarding or settings for new users
                navigate({ to: "/settings" });
              } else {
                navigate({ to: "/" });
              }
            }, 1500);
          } else {
            throw new Error("Failed to get user data");
          }
        } catch (error) {
          logger.error("Failed to fetch user after OAuth", error);
          setStatus("error");
          setErrorMessage("Failed to complete authentication. Please try again.");
        }
      } else {
        // No success or error parameter - something went wrong
        setStatus("error");
        setErrorMessage("Invalid callback. Please try logging in again.");
      }
    };

    handleCallback();
  }, [search, navigate, login]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md space-y-8 p-8">
        {status === "loading" && (
          <div className="text-center">
            <div className="mx-auto mb-4 h-12 w-12 animate-spin rounded-full border-b-2 border-blue-600" />
            <h2 className="text-xl font-semibold text-gray-900">Completing sign in...</h2>
            <p className="mt-2 text-gray-600">Please wait while we verify your credentials.</p>
          </div>
        )}

        {status === "success" && (
          <div className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
              <svg
                className="h-6 w-6 text-green-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900">Successfully signed in!</h2>
            <p className="mt-2 text-gray-600">Redirecting you to the app...</p>
          </div>
        )}

        {status === "error" && (
          <div className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
              <svg
                className="h-6 w-6 text-red-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900">Authentication Failed</h2>
            <p className="mt-2 text-gray-600">{errorMessage}</p>
            <button
              onClick={() => navigate({ to: "/login" })}
              className="mt-4 inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
            >
              Return to Login
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
