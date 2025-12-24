import { AlertTriangle, RefreshCw, ServerCrash, WifiOff } from "lucide-react";

import { cn } from "../../lib/utils";
import { Button } from "../ui/button";

interface QueryErrorProps {
  /** The error that occurred */
  error: Error | null;
  /** Function to call when retry is clicked */
  onRetry?: () => void;
  /** Whether a retry is currently in progress */
  isRetrying?: boolean;
  /** Custom title for the error */
  title?: string;
  /** Whether to show the error message */
  showMessage?: boolean;
  /** Size variant */
  variant?: "compact" | "default" | "full";
  /** Additional className */
  className?: string;
}

/**
 * Determines the type of error based on the error message
 */
function getErrorType(error: Error | null): "network" | "server" | "generic" {
  if (!error) return "generic";

  const message = error.message.toLowerCase();

  if (
    message.includes("network") ||
    message.includes("offline") ||
    message.includes("fetch") ||
    message.includes("failed to fetch")
  ) {
    return "network";
  }

  if (message.includes("500") || message.includes("503") || message.includes("server")) {
    return "server";
  }

  return "generic";
}

/**
 * QueryError component for displaying TanStack Query errors with retry capability
 *
 * @example
 * // Basic usage
 * if (query.isError) {
 *   return <QueryError error={query.error} onRetry={() => query.refetch()} />;
 * }
 *
 * // Compact variant
 * <QueryError error={error} onRetry={refetch} variant="compact" />
 */
export function QueryError({
  error,
  onRetry,
  isRetrying = false,
  title,
  showMessage = true,
  variant = "default",
  className,
}: QueryErrorProps) {
  const errorType = getErrorType(error);

  const Icon = {
    network: WifiOff,
    server: ServerCrash,
    generic: AlertTriangle,
  }[errorType];

  const defaultTitles = {
    network: "Connection Error",
    server: "Server Error",
    generic: "Something went wrong",
  };

  const defaultMessages = {
    network: "Please check your internet connection and try again.",
    server: "The server is temporarily unavailable. Please try again later.",
    generic: "An unexpected error occurred. Please try again.",
  };

  if (variant === "compact") {
    return (
      <div
        className={cn(
          "flex items-center gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-800 dark:bg-red-950/50 dark:text-red-400",
          className
        )}
        role="alert"
      >
        <Icon className="h-4 w-4 flex-shrink-0" />
        <span className="flex-1 truncate">{error?.message || defaultMessages[errorType]}</span>
        {onRetry && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onRetry}
            disabled={isRetrying}
            className="h-7 px-2 text-red-700 hover:bg-red-100 hover:text-red-800 dark:text-red-400 dark:hover:bg-red-900 dark:hover:text-red-300"
          >
            <RefreshCw className={cn("h-3.5 w-3.5", isRetrying && "animate-spin")} />
          </Button>
        )}
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-4 p-6 text-center",
        variant === "full" && "min-h-[300px]",
        className
      )}
      role="alert"
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/50">
        <Icon className="h-6 w-6 text-red-600 dark:text-red-400" />
      </div>
      <div className="space-y-2">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{title || defaultTitles[errorType]}</h3>
        {showMessage && (
          <p className="max-w-sm text-sm text-gray-600 dark:text-gray-400">
            {error?.message || defaultMessages[errorType]}
          </p>
        )}
      </div>
      {onRetry && (
        <Button
          onClick={onRetry}
          disabled={isRetrying}
          variant="outline"
          className="mt-2"
        >
          <RefreshCw className={cn("mr-2 h-4 w-4", isRetrying && "animate-spin")} />
          {isRetrying ? "Retrying..." : "Try Again"}
        </Button>
      )}
    </div>
  );
}

interface QueryErrorInlineProps {
  error: Error | null;
  onRetry?: () => void;
  isRetrying?: boolean;
}

/**
 * Inline error message with retry link
 * For use in cards, list items, etc.
 */
export function QueryErrorInline({ error, onRetry, isRetrying }: QueryErrorInlineProps) {
  return (
    <span className="inline-flex items-center gap-1 text-sm text-red-600 dark:text-red-400">
      <AlertTriangle className="h-3.5 w-3.5" />
      <span>{error?.message || "Error loading data"}</span>
      {onRetry && (
        <button
          onClick={onRetry}
          disabled={isRetrying}
          className="ml-1 underline hover:no-underline focus:ring-2 focus:ring-red-500 focus:ring-offset-1 focus:outline-none"
        >
          {isRetrying ? <RefreshCw className="inline h-3 w-3 animate-spin" /> : "Retry"}
        </button>
      )}
    </span>
  );
}
