import type { QueryClient, QueryKey } from "@tanstack/react-query";

import { logger } from "./logger";
import { showMutationError, showMutationSuccess } from "./mutation-toast";

/**
 * Context returned by optimistic delete operations
 */
export interface OptimisticDeleteContext<T> {
  previousData: T[] | undefined;
}

/**
 * Options for creating optimistic delete handlers
 */
export interface OptimisticDeleteOptions<T, TId> {
  /** Query client instance */
  queryClient: QueryClient;
  /** Query key for the list */
  listQueryKey: QueryKey;
  /** Function to get the ID from an item */
  getId: (item: T) => TId;
  /** Optional query key for the detail (will be removed on success) */
  detailQueryKey?: (id: TId) => QueryKey;
  /** Success message */
  successMessage: string;
  /** Error log prefix */
  errorLogPrefix?: string;
  /** Retry function (receives the mutation function to wrap) */
  onRetry?: (id: TId) => void;
}

/**
 * Creates handlers for optimistic delete mutations.
 * Handles: cancel queries, snapshot, filter, rollback, invalidate.
 *
 * @example
 * ```ts
 * const handlers = createOptimisticDeleteHandlers<User, number>({
 *   queryClient,
 *   listQueryKey: queryKeys.users.lists(),
 *   getId: (user) => user.id,
 *   detailQueryKey: (id) => queryKeys.users.detail(id),
 *   successMessage: "User deleted",
 * });
 *
 * const mutation = useMutation({
 *   mutationFn: UserService.deleteUser,
 *   ...handlers,
 * });
 * ```
 */
export function createOptimisticDeleteHandlers<T, TId>({
  queryClient,
  listQueryKey,
  getId,
  detailQueryKey,
  successMessage,
  errorLogPrefix = "Delete error",
  onRetry,
}: OptimisticDeleteOptions<T, TId>) {
  return {
    onMutate: async (id: TId): Promise<OptimisticDeleteContext<T>> => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: listQueryKey });

      // Snapshot previous value
      const previousData = queryClient.getQueryData<T[]>(listQueryKey);

      // Optimistically remove from cache
      queryClient.setQueryData<T[]>(listQueryKey, (old) => old?.filter((item) => getId(item) !== id));

      return { previousData };
    },

    onSuccess: (_: unknown, id: TId) => {
      // Remove detail query if provided
      if (detailQueryKey) {
        queryClient.removeQueries({ queryKey: detailQueryKey(id) });
      }
      showMutationSuccess({ message: successMessage });
    },

    onError: (error: Error, id: TId, context: OptimisticDeleteContext<T> | undefined) => {
      logger.error(errorLogPrefix, error);

      // Rollback optimistic update
      if (context?.previousData) {
        queryClient.setQueryData(listQueryKey, context.previousData);
      }

      showMutationError({
        error,
        onRetry: onRetry ? () => onRetry(id) : undefined,
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: listQueryKey });
    },
  };
}

/**
 * Context returned by optimistic update operations
 */
export interface OptimisticUpdateContext<T> {
  previousListData: T[] | undefined;
  previousDetailData: T | undefined;
}

/**
 * Options for creating optimistic update handlers
 */
export interface OptimisticUpdateOptions<T, TId> {
  /** Query client instance */
  queryClient: QueryClient;
  /** Query key for the list */
  listQueryKey: QueryKey;
  /** Function to get the ID from an item */
  getId: (item: T) => TId;
  /** Query key for the detail */
  detailQueryKey: (id: TId) => QueryKey;
  /** Success message */
  successMessage: string;
  /** Error log prefix */
  errorLogPrefix?: string;
  /** Retry function */
  onRetry?: (item: T) => void;
}

/**
 * Creates handlers for optimistic update mutations.
 * Handles: cancel queries, snapshot, update both caches, rollback, invalidate.
 *
 * @example
 * ```ts
 * const handlers = createOptimisticUpdateHandlers<User, number>({
 *   queryClient,
 *   listQueryKey: queryKeys.users.lists(),
 *   getId: (user) => user.id,
 *   detailQueryKey: (id) => queryKeys.users.detail(id),
 *   successMessage: "User updated",
 * });
 *
 * const mutation = useMutation({
 *   mutationFn: UserService.updateUser,
 *   ...handlers,
 * });
 * ```
 */
export function createOptimisticUpdateHandlers<T, TId>({
  queryClient,
  listQueryKey,
  getId,
  detailQueryKey,
  successMessage,
  errorLogPrefix = "Update error",
  onRetry,
}: OptimisticUpdateOptions<T, TId>) {
  return {
    onMutate: async (updatedItem: T): Promise<OptimisticUpdateContext<T>> => {
      const id = getId(updatedItem);

      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: listQueryKey });
      await queryClient.cancelQueries({ queryKey: detailQueryKey(id) });

      // Snapshot previous values
      const previousListData = queryClient.getQueryData<T[]>(listQueryKey);
      const previousDetailData = queryClient.getQueryData<T>(detailQueryKey(id));

      // Optimistically update both caches
      queryClient.setQueryData(detailQueryKey(id), updatedItem);
      queryClient.setQueryData<T[]>(listQueryKey, (old) =>
        old?.map((item) => (getId(item) === id ? updatedItem : item))
      );

      return { previousListData, previousDetailData };
    },

    onSuccess: (serverResponse: T) => {
      // Update caches with server response
      const id = getId(serverResponse);
      queryClient.setQueryData(detailQueryKey(id), serverResponse);
      queryClient.setQueryData<T[]>(listQueryKey, (old) =>
        old?.map((item) => (getId(item) === id ? serverResponse : item))
      );
      showMutationSuccess({ message: successMessage });
    },

    onError: (error: Error, updatedItem: T, context: OptimisticUpdateContext<T> | undefined) => {
      const id = getId(updatedItem);
      logger.error(errorLogPrefix, error);

      // Rollback optimistic updates
      if (context?.previousDetailData !== undefined) {
        queryClient.setQueryData(detailQueryKey(id), context.previousDetailData);
      }
      if (context?.previousListData) {
        queryClient.setQueryData(listQueryKey, context.previousListData);
      }

      showMutationError({
        error,
        onRetry: onRetry ? () => onRetry(updatedItem) : undefined,
      });
    },

    onSettled: (_: unknown, __: unknown, updatedItem: T) => {
      const id = getId(updatedItem);
      queryClient.invalidateQueries({ queryKey: detailQueryKey(id) });
      queryClient.invalidateQueries({ queryKey: listQueryKey });
    },
  };
}
