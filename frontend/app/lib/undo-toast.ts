import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

interface UndoableDeleteOptions<TItem extends { id: string | number }> {
  /** TanStack Query client for cache operations */
  queryClient: QueryClient;
  /** Query key for the list containing the item */
  queryKey: unknown[];
  /** The item being deleted */
  item: TItem;
  /** Human-readable label for the item (shown in toast) */
  itemLabel: string;
  /** Callback when user clicks Undo (optional) */
  onUndo?: () => void;
  /** Async function to perform the actual delete after timeout */
  onConfirm: () => Promise<void>;
  /** Time in ms before delete is confirmed (default: 5000) */
  timeout?: number;
}

interface UndoableDeleteResult {
  /** Toast ID for programmatic dismissal */
  toastId: string | number;
  /** Cancel the pending delete (same as clicking Undo) */
  cancel: () => void;
}

/**
 * Show an undoable delete toast with optimistic cache update.
 *
 * The item is immediately removed from the cache (optimistic).
 * If the user clicks "Undo" within the timeout, the item is restored.
 * If the timeout expires without undo, the actual delete API is called.
 *
 * @example
 * // In a delete handler
 * const handleDelete = (user: User) => {
 *   showUndoableDelete({
 *     queryClient,
 *     queryKey: queryKeys.users.lists(),
 *     item: user,
 *     itemLabel: user.name,
 *     onConfirm: () => UserService.deleteUser(user.id),
 *   });
 * };
 *
 * @example
 * // With custom undo callback
 * showUndoableDelete({
 *   queryClient,
 *   queryKey: queryKeys.files.list(),
 *   item: file,
 *   itemLabel: file.name,
 *   onConfirm: () => FileService.deleteFile(file.id),
 *   onUndo: () => trackEvent("file_delete_undone"),
 *   timeout: 8000,
 * });
 */
export function showUndoableDelete<TItem extends { id: string | number }>({
  queryClient,
  queryKey,
  item,
  itemLabel,
  onUndo,
  onConfirm,
  timeout = 5000,
}: UndoableDeleteOptions<TItem>): UndoableDeleteResult {
  // Snapshot the previous data for potential rollback
  const previousData = queryClient.getQueryData<TItem[]>(queryKey);

  // Optimistically remove from cache immediately
  queryClient.setQueryData<TItem[]>(queryKey, (old) => old?.filter((i) => i.id !== item.id));

  let undone = false;
  let confirmTimeoutId: ReturnType<typeof setTimeout> | null = null;

  const cancel = () => {
    undone = true;
    if (confirmTimeoutId) {
      clearTimeout(confirmTimeoutId);
      confirmTimeoutId = null;
    }

    // Restore the item to cache
    queryClient.setQueryData<TItem[]>(queryKey, previousData);
    onUndo?.();

    toast.success(`${itemLabel} restored`);
  };

  const toastId = toast(`${itemLabel} deleted`, {
    duration: timeout,
    action: {
      label: "Undo",
      onClick: cancel,
    },
    onDismiss: () => {
      // If toast was dismissed (not undone), perform the actual delete
      if (!undone) {
        performDelete();
      }
    },
    onAutoClose: () => {
      // If toast auto-closed (timeout), perform the actual delete
      if (!undone) {
        performDelete();
      }
    },
  });

  const performDelete = async () => {
    if (undone) return;

    try {
      await onConfirm();
      // Invalidate to ensure server state is reflected
      queryClient.invalidateQueries({ queryKey });
    } catch (error) {
      // Rollback on error
      queryClient.setQueryData<TItem[]>(queryKey, previousData);
      toast.error(`Failed to delete ${itemLabel}`, {
        description: error instanceof Error ? error.message : "An error occurred",
      });
    }
  };

  return {
    toastId,
    cancel,
  };
}

/**
 * Options for bulk undoable delete operations
 */
interface UndoableBulkDeleteOptions<TItem extends { id: string | number }> {
  queryClient: QueryClient;
  queryKey: unknown[];
  items: TItem[];
  itemsLabel: string;
  onUndo?: () => void;
  onConfirm: () => Promise<void>;
  timeout?: number;
}

/**
 * Show an undoable delete toast for multiple items.
 *
 * @example
 * showUndoableBulkDelete({
 *   queryClient,
 *   queryKey: queryKeys.users.lists(),
 *   items: selectedUsers,
 *   itemsLabel: `${selectedUsers.length} users`,
 *   onConfirm: () => UserService.deleteUsers(selectedUsers.map(u => u.id)),
 * });
 */
export function showUndoableBulkDelete<TItem extends { id: string | number }>({
  queryClient,
  queryKey,
  items,
  itemsLabel,
  onUndo,
  onConfirm,
  timeout = 5000,
}: UndoableBulkDeleteOptions<TItem>): UndoableDeleteResult {
  const previousData = queryClient.getQueryData<TItem[]>(queryKey);
  const itemIds = new Set(items.map((i) => i.id));

  // Optimistically remove all items
  queryClient.setQueryData<TItem[]>(queryKey, (old) => old?.filter((i) => !itemIds.has(i.id)));

  let undone = false;

  const cancel = () => {
    undone = true;
    queryClient.setQueryData<TItem[]>(queryKey, previousData);
    onUndo?.();
    toast.success(`${itemsLabel} restored`);
  };

  const toastId = toast(`${itemsLabel} deleted`, {
    duration: timeout,
    action: {
      label: "Undo",
      onClick: cancel,
    },
    onDismiss: () => {
      if (!undone) performDelete();
    },
    onAutoClose: () => {
      if (!undone) performDelete();
    },
  });

  const performDelete = async () => {
    if (undone) return;

    try {
      await onConfirm();
      queryClient.invalidateQueries({ queryKey });
    } catch (error) {
      queryClient.setQueryData<TItem[]>(queryKey, previousData);
      toast.error(`Failed to delete ${itemsLabel}`, {
        description: error instanceof Error ? error.message : "An error occurred",
      });
    }
  };

  return { toastId, cancel };
}
