import { useCallback, useEffect, useState } from "react";

import { AlertCircle, Bug, ChevronDown, ChevronUp, Copy, Home, RefreshCw, Wifi, WifiOff } from "lucide-react";
import { useTranslation } from "react-i18next";

import { queryClient } from "../../lib/query-client";
import { captureError } from "../../lib/sentry";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";

export interface ErrorFallbackProps {
  error: Error;
  resetError?: () => void;
  showStack?: boolean;
}

type ErrorType = "network" | "auth" | "notFound" | "server" | "unknown";

function classifyError(error: Error): ErrorType {
  const message = error.message.toLowerCase();
  const name = error.name.toLowerCase();

  if (message.includes("network") || message.includes("fetch") || name.includes("network")) {
    return "network";
  }
  if (
    message.includes("401") ||
    message.includes("403") ||
    message.includes("unauthorized") ||
    message.includes("forbidden")
  ) {
    return "auth";
  }
  if (message.includes("404") || message.includes("not found")) {
    return "notFound";
  }
  if (message.includes("500") || message.includes("server")) {
    return "server";
  }
  return "unknown";
}

function getErrorInfo(errorType: ErrorType) {
  switch (errorType) {
    case "network":
      return {
        icon: WifiOff,
        title: "Connection Problem",
        description: "Unable to connect to the server. Please check your internet connection.",
        suggestions: [
          "Check your internet connection",
          "Try refreshing the page",
          "If the problem persists, the server might be temporarily unavailable",
        ],
      };
    case "auth":
      return {
        icon: AlertCircle,
        title: "Authentication Error",
        description: "Your session may have expired or you don't have permission to access this resource.",
        suggestions: ["Try logging in again", "Contact support if you believe this is a mistake"],
      };
    case "notFound":
      return {
        icon: AlertCircle,
        title: "Not Found",
        description: "The requested resource could not be found.",
        suggestions: [
          "Check if the URL is correct",
          "The resource may have been moved or deleted",
          "Go back to the home page",
        ],
      };
    case "server":
      return {
        icon: AlertCircle,
        title: "Server Error",
        description: "Something went wrong on our end. Our team has been notified.",
        suggestions: ["Try again in a few moments", "If the problem persists, contact support"],
      };
    default:
      return {
        icon: AlertCircle,
        title: "Something Went Wrong",
        description: "An unexpected error occurred.",
        suggestions: [
          "Try refreshing the page",
          "Clear your browser cache",
          "If the problem persists, contact support",
        ],
      };
  }
}

export function ErrorFallback({ error, resetError, showStack = false }: ErrorFallbackProps) {
  const { t } = useTranslation("errors");
  const isDev = typeof import.meta !== "undefined" && import.meta.env?.DEV;
  const [isOnline, setIsOnline] = useState(typeof navigator !== "undefined" ? navigator.onLine : true);
  const [showDetails, setShowDetails] = useState(false);
  const [copied, setCopied] = useState(false);
  const [reportSent, setReportSent] = useState(false);

  const errorType = classifyError(error);
  const errorInfo = getErrorInfo(errorType);
  const IconComponent = errorInfo.icon;

  // Monitor online status
  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  // Report error to Sentry when the component mounts
  useEffect(() => {
    captureError(error, {
      componentStack: "ErrorFallback",
      location: typeof window !== "undefined" ? window.location.href : "unknown",
      errorType,
    });
  }, [error, errorType]);

  // Handle retry: invalidate query cache to ensure fresh data, then reset
  const handleRetry = useCallback(() => {
    void queryClient.invalidateQueries();
    resetError?.();
  }, [resetError]);

  // Copy error details to clipboard
  const handleCopyError = useCallback(() => {
    const errorDetails = `
Error: ${error.name}
Message: ${error.message}
URL: ${typeof window !== "undefined" ? window.location.href : "unknown"}
Time: ${new Date().toISOString()}
User Agent: ${typeof navigator !== "undefined" ? navigator.userAgent : "unknown"}
${error.stack ? `\nStack Trace:\n${error.stack}` : ""}
    `.trim();

    navigator.clipboard.writeText(errorDetails).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }, [error]);

  // Report bug (sends to Sentry with user report flag)
  const handleReportBug = useCallback(() => {
    captureError(error, {
      userReport: true,
      componentStack: "ErrorFallback",
      location: typeof window !== "undefined" ? window.location.href : "unknown",
    });
    setReportSent(true);
    setTimeout(() => setReportSent(false), 3000);
  }, [error]);

  return (
    <div className="bg-background flex min-h-[400px] items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          {/* Offline indicator */}
          {!isOnline && (
            <div className="bg-warning/10 text-warning mb-4 flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm">
              <WifiOff className="h-4 w-4" />
              <span>You're offline</span>
            </div>
          )}

          {/* Error icon */}
          <div className="bg-destructive/10 mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full">
            <IconComponent className="text-destructive h-7 w-7" />
          </div>

          {/* Title and description */}
          <CardTitle className="text-destructive text-xl">{t("generic.title", errorInfo.title)}</CardTitle>
          <CardDescription className="mt-2 text-base">
            {error.message || t("generic.message", errorInfo.description)}
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Suggestions */}
          <div className="bg-muted/50 rounded-lg p-4">
            <h4 className="text-foreground mb-2 text-sm font-medium">Try these steps:</h4>
            <ul className="text-muted-foreground space-y-1 text-sm">
              {errorInfo.suggestions.map((suggestion, index) => (
                <li
                  key={index}
                  className="flex items-start gap-2"
                >
                  <span className="text-primary mt-1">•</span>
                  <span>{suggestion}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Technical details (collapsible) */}
          {(showStack || isDev) && error.stack && (
            <div className="border-border rounded-lg border">
              <button
                onClick={() => setShowDetails(!showDetails)}
                className="text-muted-foreground hover:text-foreground flex w-full items-center justify-between p-3 text-sm transition-colors"
              >
                <span>Technical Details</span>
                {showDetails ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
              </button>
              {showDetails && (
                <div className="border-border border-t">
                  <div className="bg-muted max-h-48 overflow-auto p-4">
                    <pre className="text-muted-foreground text-xs whitespace-pre-wrap">{error.stack}</pre>
                  </div>
                  <div className="border-border flex justify-end border-t p-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleCopyError}
                      className="text-xs"
                    >
                      <Copy className="mr-1 h-3 w-3" />
                      {copied ? "Copied!" : "Copy Error"}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Action buttons */}
          <div className="flex flex-col gap-2 sm:flex-row sm:justify-center">
            {resetError && (
              <Button
                onClick={handleRetry}
                variant="default"
                className="gap-2"
              >
                <RefreshCw className="h-4 w-4" />
                {t("generic.tryAgain", "Try Again")}
              </Button>
            )}
            <Button
              variant="outline"
              asChild
              className="gap-2"
            >
              <a href="/">
                <Home className="h-4 w-4" />
                {t("generic.goHome", "Go Home")}
              </a>
            </Button>
          </div>

          {/* Report bug button */}
          <div className="flex justify-center pt-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleReportBug}
              disabled={reportSent}
              className="text-muted-foreground text-xs"
            >
              <Bug className="mr-1 h-3 w-3" />
              {reportSent ? "Report Sent!" : "Report This Issue"}
            </Button>
          </div>

          {/* Online status restored */}
          {isOnline && errorType === "network" && (
            <div className="text-success flex items-center justify-center gap-1 text-xs">
              <Wifi className="h-3 w-3" />
              <span>Connection restored - try again</span>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * Minimal error fallback for use in smaller components
 */
export function ErrorFallbackInline({ error, resetError }: { error: Error; resetError?: () => void }) {
  return (
    <div className="bg-destructive/10 text-destructive flex items-center gap-3 rounded-lg p-4">
      <AlertCircle className="h-5 w-5 shrink-0" />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium">Error loading content</p>
        <p className="text-destructive/80 truncate text-xs">{error.message}</p>
      </div>
      {resetError && (
        <Button
          variant="ghost"
          size="sm"
          onClick={resetError}
        >
          <RefreshCw className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}
