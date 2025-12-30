import { useCallback, useEffect } from "react";

import { AlertCircle, Home, RefreshCw } from "lucide-react";
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

export function ErrorFallback({ error, resetError, showStack = false }: ErrorFallbackProps) {
  const { t } = useTranslation("errors");
  const isDev = typeof import.meta !== "undefined" && import.meta.env?.DEV;

  // Report error to Sentry when the component mounts
  useEffect(() => {
    captureError(error, {
      componentStack: "ErrorFallback",
      location: typeof window !== "undefined" ? window.location.href : "unknown",
    });
  }, [error]);

  // Handle retry: invalidate query cache to ensure fresh data, then reset
  const handleRetry = useCallback(() => {
    // Invalidate all queries to ensure fresh data on retry
    // This helps recover from stale data issues
    void queryClient.invalidateQueries();

    // Call the reset function to re-render the error boundary's children
    resetError?.();
  }, [resetError]);

  return (
    <div className="bg-background flex min-h-[400px] items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <div className="bg-destructive/10 mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full">
            <AlertCircle className="text-destructive h-6 w-6" />
          </div>
          <CardTitle className="text-destructive">{t("generic.title")}</CardTitle>
          <CardDescription className="mt-2">{error.message || t("generic.message")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {showStack && isDev && error.stack && (
            <div className="bg-muted max-h-48 overflow-auto rounded-md p-4">
              <pre className="text-muted-foreground text-xs whitespace-pre-wrap">{error.stack}</pre>
            </div>
          )}
          <div className="flex justify-center gap-2">
            {resetError && (
              <Button
                onClick={handleRetry}
                variant="default"
              >
                <RefreshCw className="mr-2 h-4 w-4" />
                {t("generic.tryAgain")}
              </Button>
            )}
            <Button
              variant="outline"
              asChild
            >
              <a href="/">
                <Home className="mr-2 h-4 w-4" />
                {t("generic.goHome")}
              </a>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
