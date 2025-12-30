import { useState } from "react";

import { useTranslation } from "react-i18next";

import { logger } from "../../lib/logger";
import { OAuthService, type OAuthProvider } from "../../services/oauth";

interface OAuthButtonsProps {
  onError?: (error: string) => void;
  disabled?: boolean;
  mode?: "login" | "register" | "link";
}

interface ProviderConfig {
  name: string;
  icon: React.ReactNode;
  bgColor: string;
  hoverColor: string;
  textColor: string;
}

const providers: Record<OAuthProvider, ProviderConfig> = {
  google: {
    name: "Google",
    icon: (
      <svg
        className="h-5 w-5"
        viewBox="0 0 24 24"
        fill="currentColor"
      >
        <path
          d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
          fill="#4285F4"
        />
        <path
          d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
          fill="#34A853"
        />
        <path
          d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
          fill="#FBBC05"
        />
        <path
          d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
          fill="#EA4335"
        />
      </svg>
    ),
    bgColor: "bg-white",
    hoverColor: "hover:bg-gray-50",
    textColor: "text-gray-700",
  },
  github: {
    name: "GitHub",
    icon: (
      <svg
        className="h-5 w-5"
        fill="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
        />
      </svg>
    ),
    bgColor: "bg-gray-900",
    hoverColor: "hover:bg-gray-800",
    textColor: "text-white",
  },
};

export function OAuthButtons({ onError, disabled = false, mode = "login" }: OAuthButtonsProps) {
  const { t } = useTranslation("auth");
  const [loading, setLoading] = useState<OAuthProvider | null>(null);

  const handleOAuthClick = async (provider: OAuthProvider) => {
    if (disabled || loading) return;

    setLoading(provider);
    try {
      await OAuthService.initiateOAuth(provider);
    } catch (error) {
      logger.error(`OAuth ${provider} error`, error);
      const message = error instanceof Error ? error.message : `Failed to connect with ${providers[provider].name}`;
      onError?.(message);
      setLoading(null);
    }
  };

  const getButtonText = (provider: OAuthProvider) => {
    if (loading === provider) {
      return t("oauth.connecting");
    }
    switch (mode) {
      case "register":
        return t("oauth.signUpWith", { provider: providers[provider].name });
      case "link":
        return t("oauth.linkWith", { provider: providers[provider].name });
      default:
        return t("oauth.continueWith", { provider: providers[provider].name });
    }
  };

  return (
    <div className="space-y-3">
      {(Object.keys(providers) as OAuthProvider[]).map((provider) => {
        const config = providers[provider];
        const isLoading = loading === provider;

        return (
          <button
            key={provider}
            type="button"
            onClick={() => handleOAuthClick(provider)}
            disabled={disabled || loading !== null}
            className={`flex w-full items-center justify-center gap-3 rounded-lg border border-gray-300 px-4 py-2.5 shadow-sm ${config.bgColor} ${config.hoverColor} ${config.textColor} font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50`}
          >
            {isLoading ? (
              <svg
                className="h-5 w-5 animate-spin"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
            ) : (
              config.icon
            )}
            <span>{getButtonText(provider)}</span>
          </button>
        );
      })}
    </div>
  );
}

interface OAuthDividerProps {
  text?: string;
}

export function OAuthDivider({ text }: OAuthDividerProps) {
  const { t } = useTranslation("auth");
  const displayText = text ?? t("oauth.or");

  return (
    <div className="relative my-6">
      <div className="absolute inset-0 flex items-center">
        <div className="w-full border-t border-gray-300" />
      </div>
      <div className="relative flex justify-center text-sm">
        <span className="bg-white px-2 text-gray-500">{displayText}</span>
      </div>
    </div>
  );
}
