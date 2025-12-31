import type { QueryClient } from "@tanstack/react-query";

/**
 * Helper types for optimistic update context
 */
export interface OptimisticContext<T> {
  previousData: T | undefined;
}

/**
 * Create optimistic update handlers for a mutation
 *
 * @example
 * const updateUser = useMutation({
 *   mutationFn: updateUserApi,
 *   ...createOptimisticUpdate({
 *     queryClient,
 *     queryKey: ['users', userId],
 *     updateFn: (oldData, newData) => ({ ...oldData, ...newData }),
 *   }),
 * });
 */
export function createOptimisticUpdate<TData, TVariables>({
  queryClient,
  queryKey,
  updateFn,
}: {
  queryClient: QueryClient;
  queryKey: unknown[];
  updateFn: (oldData: TData | undefined, variables: TVariables) => TData;
}) {
  return {
    onMutate: async (variables: TVariables): Promise<OptimisticContext<TData>> => {
      // Cancel outgoing refetches to prevent overwriting optimistic update
      await queryClient.cancelQueries({ queryKey });

      // Snapshot the previous value
      const previousData = queryClient.getQueryData<TData>(queryKey);

      // Optimistically update
      queryClient.setQueryData<TData>(queryKey, (old) => updateFn(old, variables));

      // Return context with snapshotted value
      return { previousData };
    },

    onError: (_error: Error, _variables: TVariables, context: OptimisticContext<TData> | undefined) => {
      // Roll back to previous value on error
      if (context?.previousData !== undefined) {
        queryClient.setQueryData(queryKey, context.previousData);
      }
    },

    onSettled: () => {
      // Always refetch after error or success to ensure server state
      queryClient.invalidateQueries({ queryKey });
    },
  };
}

/**
 * Create optimistic update handlers for list mutations (add/remove/update items)
 *
 * @example
 * const deleteUser = useMutation({
 *   mutationFn: deleteUserApi,
 *   ...createListOptimisticUpdate({
 *     queryClient,
 *     queryKey: ['users'],
 *     type: 'remove',
 *     getId: (variables) => variables.userId,
 *   }),
 * });
 */
export function createListOptimisticUpdate<TItem extends { id: string | number }, TVariables>({
  queryClient,
  queryKey,
  type,
  getId,
  getNewItem,
  updateItem,
}: {
  queryClient: QueryClient;
  queryKey: unknown[];
  type: "add" | "remove" | "update";
  getId?: (variables: TVariables) => string | number;
  getNewItem?: (variables: TVariables) => TItem;
  updateItem?: (item: TItem, variables: TVariables) => TItem;
}) {
  return createOptimisticUpdate<TItem[], TVariables>({
    queryClient,
    queryKey,
    updateFn: (oldData, variables) => {
      if (!oldData) return [] as TItem[];

      switch (type) {
        case "add":
          if (!getNewItem) throw new Error("getNewItem required for add type");
          return [...oldData, getNewItem(variables)];

        case "remove":
          if (!getId) throw new Error("getId required for remove type");
          return oldData.filter((item) => item.id !== getId(variables));

        case "update":
          if (!getId || !updateItem) throw new Error("getId and updateItem required for update type");
          return oldData.map((item) => (item.id === getId(variables) ? updateItem(item, variables) : item));

        default:
          return oldData;
      }
    },
  });
}

/**
 * Create optimistic toggle handler (for boolean fields like isActive, isFavorite, etc.)
 *
 * @example
 * const toggleFavorite = useMutation({
 *   mutationFn: toggleFavoriteApi,
 *   ...createToggleOptimisticUpdate({
 *     queryClient,
 *     queryKey: ['item', itemId],
 *     field: 'isFavorite',
 *   }),
 * });
 */
export function createToggleOptimisticUpdate<TData extends Record<string, unknown>>({
  queryClient,
  queryKey,
  field,
}: {
  queryClient: QueryClient;
  queryKey: unknown[];
  field: keyof TData;
}) {
  return createOptimisticUpdate<TData, void>({
    queryClient,
    queryKey,
    updateFn: (oldData) => {
      if (!oldData) return undefined as unknown as TData;
      return {
        ...oldData,
        [field]: !oldData[field],
      };
    },
  });
}

/**
 * Optimistic counter update (increment/decrement)
 *
 * @example
 * const incrementLikes = useMutation({
 *   mutationFn: incrementLikesApi,
 *   ...createCounterOptimisticUpdate({
 *     queryClient,
 *     queryKey: ['post', postId],
 *     field: 'likeCount',
 *     delta: 1,
 *   }),
 * });
 */
export function createCounterOptimisticUpdate<TData extends Record<string, unknown>>({
  queryClient,
  queryKey,
  field,
  delta,
}: {
  queryClient: QueryClient;
  queryKey: unknown[];
  field: keyof TData;
  delta: number;
}) {
  return createOptimisticUpdate<TData, void>({
    queryClient,
    queryKey,
    updateFn: (oldData) => {
      if (!oldData) return undefined as unknown as TData;
      const currentValue = oldData[field];
      if (typeof currentValue !== "number") return oldData;
      return {
        ...oldData,
        [field]: currentValue + delta,
      };
    },
  });
}
