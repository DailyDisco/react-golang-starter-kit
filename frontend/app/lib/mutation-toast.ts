import { toast } from "sonner";

import type { ApiError } from "../services/api/client";
import { categorizeError, type CategorizedError, type ErrorCategory } from "./error-utils";

// Store active countdown intervals to clean up on dismiss
const countdownIntervals = new Map<string | number, ReturnType<typeof setInterval>>();

interface MutationToastOptions {
  /** The error to display */
  error: Error | ApiError | unknown;
  /** Callback to retry the failed operation */
  onRetry?: () => void;
  /** Additional context for the error message */
  context?: string;
  /** Override the default duration (ms) */
  duration?: number;
}

interface MutationSuccessOptions {
  /** The success message to display */
  message: string;
  /** Optional description */
  description?: string;
  /** Duration in ms (default: 3000) */
  duration?: number;
}

const CATEGORY_ICONS: Record<ErrorCategory, string> = {
  auth: "🔒",
  validation: "⚠️",
  network: "📡",
  server: "🔧",
  unknown: "❌",
};

/**
 * Show a standardized error toast for mutation failures.
 *
 * Features:
 * - Categorizes errors automatically
 * - Shows retry button for retryable errors
 * - Consistent styling across the app
 *
 * @example
 * // Basic usage
 * onError: (error) => showMutationError({ error })
 *
 * @example
 * // With retry
 * onError: (error, variables) => showMutationError({
 *   error,
 *   onRetry: () => mutate(variables),
 * })
 */
export function showMutationError({ error, onRetry, context, duration }: MutationToastOptions): CategorizedError {
  const categorized = categorizeError(error);

  // Build the toast message
  let title = categorized.message;
  if (context) {
    title = `${context}: ${title}`;
  }

  // Handle rate limit with countdown
  if (categorized.retryAfter && categorized.retryAfter > 0) {
    showRateLimitCountdown({
      initialSeconds: categorized.retryAfter,
      onRetry,
      context,
    });
    return categorized;
  }

  // Determine duration based on whether retry is available
  const toastDuration = duration ?? (categorized.retryable && onRetry ? 8000 : 5000);

  toast.error(title, {
    description: categorized.details,
    duration: toastDuration,
    action:
      categorized.retryable && onRetry
        ? {
            label: "Retry",
            onClick: onRetry,
          }
        : undefined,
  });

  return categorized;
}

/**
 * Show a countdown toast for rate limit errors.
 * Updates every second until the countdown reaches 0.
 */
function showRateLimitCountdown({
  initialSeconds,
  onRetry,
  context,
}: {
  initialSeconds: number;
  onRetry?: () => void;
  context?: string;
}): void {
  let remaining = initialSeconds;
  const toastId = `rate-limit-${Date.now()}`;

  const getMessage = (seconds: number) => {
    const base = `Too many requests. Retry in ${seconds}s`;
    return context ? `${context}: ${base}` : base;
  };

  // Show initial toast
  toast.error(getMessage(remaining), {
    id: toastId,
    duration: Infinity, // We'll dismiss manually
    action: undefined, // No retry until countdown complete
  });

  // Start countdown
  const interval = setInterval(() => {
    remaining -= 1;

    if (remaining <= 0) {
      // Countdown complete - show retry option
      clearInterval(interval);
      countdownIntervals.delete(toastId);

      toast.error(context ? `${context}: Rate limit expired` : "Rate limit expired", {
        id: toastId,
        duration: 8000,
        action: onRetry
          ? {
              label: "Retry now",
              onClick: onRetry,
            }
          : undefined,
      });
    } else {
      // Update countdown
      toast.error(getMessage(remaining), {
        id: toastId,
        duration: Infinity,
        action: undefined,
      });
    }
  }, 1000);

  countdownIntervals.set(toastId, interval);
}

/**
 * Show a success toast for mutations.
 *
 * @example
 * onSuccess: () => showMutationSuccess({ message: "User created" })
 */
export function showMutationSuccess({ message, description, duration = 3000 }: MutationSuccessOptions): void {
  toast.success(message, {
    description,
    duration,
  });
}

/**
 * Create an error handler function for mutation onError callbacks.
 * Useful when you want to create a reusable handler with retry support.
 *
 * @example
 * const deleteUser = useMutation({
 *   mutationFn: (id) => UserService.deleteUser(id),
 *   onError: createMutationErrorHandler((id) => deleteUser.mutate(id)),
 * });
 *
 * @example
 * // Without retry
 * onError: createMutationErrorHandler(),
 */
export function createMutationErrorHandler<TVariables = void>(retryFn?: (variables: TVariables) => void) {
  return (error: Error | ApiError | unknown, variables: TVariables): CategorizedError => {
    return showMutationError({
      error,
      onRetry: retryFn ? () => retryFn(variables) : undefined,
    });
  };
}

/**
 * Convenience function to create handlers for both success and error.
 *
 * @example
 * const { onSuccess, onError } = createMutationHandlers({
 *   successMessage: "User deleted",
 *   onRetry: (id) => deleteUser.mutate(id),
 * });
 *
 * const deleteUser = useMutation({
 *   mutationFn: UserService.deleteUser,
 *   onSuccess,
 *   onError,
 * });
 */
export function createMutationHandlers<TVariables = void>({
  successMessage,
  successDescription,
  onRetry,
}: {
  successMessage: string;
  successDescription?: string;
  onRetry?: (variables: TVariables) => void;
}) {
  return {
    onSuccess: () => {
      showMutationSuccess({
        message: successMessage,
        description: successDescription,
      });
    },
    onError: createMutationErrorHandler(onRetry),
  };
}
