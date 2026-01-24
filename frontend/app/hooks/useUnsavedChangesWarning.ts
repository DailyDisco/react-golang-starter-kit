import { useEffect } from "react";

import { useBlocker } from "@tanstack/react-router";

interface UnsavedChangesOptions {
  /** Whether the form has unsaved changes */
  isDirty: boolean;
  /** Custom confirmation message */
  message?: string;
  /** Optional callback to determine if navigation should be allowed */
  shouldBlock?: () => boolean;
}

/**
 * Warn users before navigating away from a page with unsaved changes.
 *
 * Features:
 * - Blocks TanStack Router navigation with confirmation dialog
 * - Blocks browser refresh/close with beforeunload event
 * - Configurable message
 *
 * @example
 * function EditProfileForm() {
 *   const form = useForm<ProfileData>();
 *
 *   useUnsavedChangesWarning({
 *     isDirty: form.formState.isDirty,
 *     message: "Your profile changes haven't been saved. Leave anyway?",
 *   });
 *
 *   return <form>{...}</form>;
 * }
 *
 * @example
 * // With custom blocking logic
 * useUnsavedChangesWarning({
 *   isDirty: form.formState.isDirty,
 *   shouldBlock: () => {
 *     // Only block if certain fields are dirty
 *     return form.formState.dirtyFields.email !== undefined;
 *   },
 * });
 */
export function useUnsavedChangesWarning({
  isDirty,
  message = "You have unsaved changes. Are you sure you want to leave?",
  shouldBlock,
}: UnsavedChangesOptions): void {
  // TanStack Router blocker for SPA navigation
  useBlocker({
    shouldBlockFn: () => {
      // Check if we should block
      const shouldBlockNavigation = shouldBlock ? shouldBlock() : isDirty;

      if (!shouldBlockNavigation) {
        return false;
      }

      // Show confirmation dialog
      return !window.confirm(message);
    },
  });

  // Browser beforeunload for closing tab/refresh
  useEffect(() => {
    if (!isDirty) return;

    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      // Check custom shouldBlock if provided
      if (shouldBlock && !shouldBlock()) {
        return;
      }

      // Standard way to trigger the browser's confirmation dialog
      event.preventDefault();
      // For older browsers
      event.returnValue = message;
      return message;
    };

    window.addEventListener("beforeunload", handleBeforeUnload);

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [isDirty, message, shouldBlock]);
}

/**
 * Hook that tracks if the user is attempting to leave with unsaved changes.
 * Useful for showing a custom dialog instead of the browser default.
 *
 * @example
 * function EditForm() {
 *   const form = useForm();
 *   const { isBlocked, proceed, cancel } = useUnsavedChangesDialog({
 *     isDirty: form.formState.isDirty,
 *   });
 *
 *   return (
 *     <>
 *       <form>...</form>
 *       <Dialog open={isBlocked}>
 *         <DialogContent>
 *           <p>Unsaved changes will be lost.</p>
 *           <Button onClick={cancel}>Stay</Button>
 *           <Button onClick={proceed}>Leave</Button>
 *         </DialogContent>
 *       </Dialog>
 *     </>
 *   );
 * }
 */
export function useUnsavedChangesDialog(options: Omit<UnsavedChangesOptions, "message">) {
  const { isDirty, shouldBlock } = options;

  // withResolver: true is required to get the blocker state object for custom dialogs
  const blocker = useBlocker({
    shouldBlockFn: () => {
      const shouldBlockNavigation = shouldBlock ? shouldBlock() : isDirty;
      return shouldBlockNavigation;
    },
    withResolver: true,
  });

  return {
    /** Whether navigation is currently blocked */
    isBlocked: blocker.status === "blocked",
    /** Allow the navigation to proceed */
    proceed: () => {
      if (blocker.status === "blocked") {
        blocker.proceed();
      }
    },
    /** Cancel the navigation and stay on page */
    cancel: () => {
      if (blocker.status === "blocked") {
        blocker.reset();
      }
    },
    /** The blocker state */
    blocker,
  };
}
