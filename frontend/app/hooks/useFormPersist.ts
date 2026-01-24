import { useCallback, useEffect, useRef } from "react";

import type { FieldValues, Path, UseFormReturn } from "react-hook-form";

interface FormPersistOptions<T extends FieldValues> {
  /** Unique key for localStorage */
  key: string;
  /** Debounce delay in ms (default: 500) */
  debounceMs?: number;
  /** Field names to exclude from persistence (e.g., ["password"]) */
  exclude?: Path<T>[];
  /** Storage backend (default: localStorage) */
  storage?: Storage;
  /** Called when storage operations fail (e.g., quota exceeded) */
  onError?: (error: Error, operation: "save" | "restore") => void;
}

interface FormPersistReturn {
  /** Clear the saved draft */
  clearDraft: () => void;
  /** Check if a draft exists */
  hasDraft: () => boolean;
  /** Get the timestamp of when draft was saved */
  getDraftTimestamp: () => number | null;
}

/**
 * Persist form data to localStorage with automatic restore.
 *
 * Features:
 * - Debounced saves to avoid excessive writes
 * - Excludes sensitive fields (like passwords)
 * - Restores on mount
 * - Clear draft on successful submit
 *
 * @example
 * function CreateUserForm() {
 *   const form = useForm<CreateUserData>({
 *     resolver: zodResolver(schema),
 *   });
 *
 *   const { clearDraft, hasDraft } = useFormPersist(form, {
 *     key: "create-user",
 *     exclude: ["password", "confirmPassword"],
 *   });
 *
 *   const onSubmit = async (data: CreateUserData) => {
 *     await createUser(data);
 *     clearDraft(); // Clear on successful submit
 *   };
 *
 *   return (
 *     <form onSubmit={form.handleSubmit(onSubmit)}>
 *       {hasDraft() && (
 *         <Alert>
 *           <p>You have unsaved changes from a previous session.</p>
 *           <Button variant="link" onClick={clearDraft}>
 *             Discard draft
 *           </Button>
 *         </Alert>
 *       )}
 *       ...
 *     </form>
 *   );
 * }
 */
export function useFormPersist<T extends FieldValues>(
  form: UseFormReturn<T>,
  options: FormPersistOptions<T>
): FormPersistReturn {
  const { key, debounceMs = 500, exclude = [], storage = localStorage, onError } = options;

  const storageKey = `form_draft_${key}`;
  const timestampKey = `form_draft_${key}_timestamp`;
  const isRestored = useRef(false);
  const debounceTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Restore on mount
  useEffect(() => {
    if (isRestored.current) return;

    try {
      const stored = storage.getItem(storageKey);
      if (stored) {
        const parsed = JSON.parse(stored) as Partial<T>;

        // Remove excluded fields
        for (const field of exclude) {
          delete parsed[field as keyof T];
        }

        // Reset form with restored values, keeping default values for excluded fields
        form.reset(parsed as T, { keepDefaultValues: true });
        isRestored.current = true;
      }
    } catch (error) {
      // Invalid JSON or storage error, clear it
      storage.removeItem(storageKey);
      storage.removeItem(timestampKey);
      onError?.(error instanceof Error ? error : new Error(String(error)), "restore");
    }
  }, [form, storageKey, timestampKey, storage, exclude, onError]);

  // Persist on change
  useEffect(() => {
    const subscription = form.watch((values) => {
      // Debounce the save
      if (debounceTimeout.current) {
        clearTimeout(debounceTimeout.current);
      }

      debounceTimeout.current = setTimeout(() => {
        try {
          const toStore = { ...values } as Record<string, unknown>;

          // Remove excluded fields
          for (const field of exclude) {
            delete toStore[field as string];
          }

          storage.setItem(storageKey, JSON.stringify(toStore));
          storage.setItem(timestampKey, Date.now().toString());
        } catch (error) {
          // Storage full or unavailable - notify caller if handler provided
          onError?.(error instanceof Error ? error : new Error(String(error)), "save");
        }
      }, debounceMs);
    });

    return () => {
      subscription.unsubscribe();
      if (debounceTimeout.current) {
        clearTimeout(debounceTimeout.current);
      }
    };
  }, [form, storageKey, timestampKey, storage, debounceMs, exclude, onError]);

  const clearDraft = useCallback(() => {
    storage.removeItem(storageKey);
    storage.removeItem(timestampKey);
  }, [storageKey, timestampKey, storage]);

  const hasDraft = useCallback(() => {
    return storage.getItem(storageKey) !== null;
  }, [storageKey, storage]);

  const getDraftTimestamp = useCallback(() => {
    const timestamp = storage.getItem(timestampKey);
    return timestamp ? parseInt(timestamp, 10) : null;
  }, [timestampKey, storage]);

  return {
    clearDraft,
    hasDraft,
    getDraftTimestamp,
  };
}
