import {
  useMutation,
  useQueryClient,
  type QueryKey,
  type UseMutationOptions,
  type UseMutationResult,
} from "@tanstack/react-query";

import { showMutationError, showMutationSuccess } from "./mutation-toast";

/**
 * Options for creating a standardized mutation hook.
 */
export interface CreateMutationOptions<TData, TVariables> {
  /** The mutation function to execute */
  mutationFn: (variables: TVariables) => Promise<TData>;
  /** Success message to display */
  successMessage?: string;
  /** Success message description */
  successDescription?: string;
  /** Query keys to invalidate on success */
  invalidateKeys?: QueryKey[];
  /** Additional onSuccess handler */
  onSuccess?: (data: TData, variables: TVariables) => void;
  /** Additional onError handler */
  onError?: (error: Error, variables: TVariables) => void;
  /** Whether to show success toast (default: true if successMessage provided) */
  showSuccessToast?: boolean;
  /** Whether to show error toast (default: true) */
  showErrorToast?: boolean;
}

/**
 * Factory function to create a standardized mutation hook with common patterns.
 *
 * Reduces boilerplate by handling:
 * - Query invalidation on success
 * - Success/error toast notifications
 * - Consistent error handling
 *
 * @example
 * // Simple mutation with invalidation
 * export function useUpdateProfile() {
 *   return useCreateMutation({
 *     mutationFn: (data) => SettingsService.updateProfile(data),
 *     successMessage: "Profile updated",
 *     invalidateKeys: [queryKeys.auth.user],
 *   });
 * }
 *
 * @example
 * // Without success toast
 * export function useSetup2FA() {
 *   return useCreateMutation({
 *     mutationFn: () => SettingsService.setup2FA(),
 *     showSuccessToast: false,
 *   });
 * }
 */
export function useCreateMutation<TData = unknown, TVariables = void>({
  mutationFn,
  successMessage,
  successDescription,
  invalidateKeys,
  onSuccess: customOnSuccess,
  onError: customOnError,
  showSuccessToast = !!successMessage,
  showErrorToast = true,
}: CreateMutationOptions<TData, TVariables>): UseMutationResult<TData, Error, TVariables> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn,
    onSuccess: (data, variables) => {
      // Invalidate specified queries
      if (invalidateKeys?.length) {
        for (const key of invalidateKeys) {
          queryClient.invalidateQueries({ queryKey: key });
        }
      }

      // Show success toast
      if (showSuccessToast && successMessage) {
        showMutationSuccess({
          message: successMessage,
          description: successDescription,
        });
      }

      // Call custom handler
      customOnSuccess?.(data, variables);
    },
    onError: (error, variables) => {
      // Show error toast
      if (showErrorToast) {
        showMutationError({ error });
      }

      // Call custom handler
      customOnError?.(error, variables);
    },
  });
}

/**
 * Options for creating a mutation hook with optimistic updates.
 */
export interface CreateOptimisticMutationOptions<TData, TVariables, TContext> extends Omit<
  CreateMutationOptions<TData, TVariables>,
  "onSuccess" | "onError"
> {
  /** Query key for the list/detail to update optimistically */
  queryKey: QueryKey;
  /** Function to update the cache optimistically */
  optimisticUpdate: (old: unknown, variables: TVariables) => unknown;
  /** Additional onSuccess handler */
  onSuccess?: (data: TData, variables: TVariables, context: TContext | undefined) => void;
  /** Additional onError handler */
  onError?: (error: Error, variables: TVariables, context: TContext | undefined) => void;
}

/**
 * Factory function to create a mutation hook with optimistic updates.
 *
 * Handles:
 * - Canceling outgoing refetches
 * - Snapshotting previous data
 * - Optimistic cache update
 * - Rollback on error
 *
 * @example
 * export function useRevokeSession() {
 *   return useCreateOptimisticMutation({
 *     mutationFn: (sessionId) => SettingsService.revokeSession(sessionId),
 *     queryKey: queryKeys.settings.sessions(),
 *     optimisticUpdate: (old, sessionId) => {
 *       if (!old || !Array.isArray(old)) return old;
 *       return old.filter((session) => session.id !== sessionId);
 *     },
 *     successMessage: "Session revoked",
 *   });
 * }
 */
export function useCreateOptimisticMutation<TData = unknown, TVariables = void, TContext = { previousData: unknown }>({
  mutationFn,
  queryKey,
  optimisticUpdate,
  successMessage,
  successDescription,
  invalidateKeys,
  onSuccess: customOnSuccess,
  onError: customOnError,
  showSuccessToast = !!successMessage,
  showErrorToast = true,
}: CreateOptimisticMutationOptions<TData, TVariables, TContext>): UseMutationResult<
  TData,
  Error,
  TVariables,
  TContext
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn,
    onMutate: async (variables) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey });

      // Snapshot the previous data
      const previousData = queryClient.getQueryData(queryKey);

      // Optimistically update the cache
      queryClient.setQueryData(queryKey, (old: unknown) => optimisticUpdate(old, variables));

      // Return context with previous data for rollback
      return { previousData } as TContext;
    },
    onSuccess: (data, variables, context) => {
      // Invalidate specified queries
      if (invalidateKeys?.length) {
        for (const key of invalidateKeys) {
          queryClient.invalidateQueries({ queryKey: key });
        }
      }

      // Show success toast
      if (showSuccessToast && successMessage) {
        showMutationSuccess({
          message: successMessage,
          description: successDescription,
        });
      }

      // Call custom handler
      customOnSuccess?.(data, variables, context);
    },
    onError: (error, variables, context) => {
      // Rollback on error
      if (context && typeof context === "object" && "previousData" in context) {
        queryClient.setQueryData(queryKey, (context as { previousData: unknown }).previousData);
      }

      // Show error toast
      if (showErrorToast) {
        showMutationError({ error });
      }

      // Call custom handler
      customOnError?.(error, variables, context);
    },
  });
}

/**
 * Type helper for mutation options without the mutationFn.
 * Useful for creating typed mutation configurations.
 */
export type MutationConfig<TData, TVariables> = Omit<CreateMutationOptions<TData, TVariables>, "mutationFn">;

/**
 * Type helper for optimistic mutation options without the mutationFn.
 */
export type OptimisticMutationConfig<TData, TVariables, TContext = { previousData: unknown }> = Omit<
  CreateOptimisticMutationOptions<TData, TVariables, TContext>,
  "mutationFn"
>;
