import type { resources } from "@/i18n";
import { useTranslation } from "react-i18next";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./alert-dialog";

interface ConfirmDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when dialog open state changes */
  onOpenChange: (open: boolean) => void;
  /** Dialog title */
  title: string;
  /** Dialog description/message */
  description: string;
  /** Custom confirm button label */
  confirmLabel?: string;
  /** Custom cancel button label */
  cancelLabel?: string;
  /** Visual style - "destructive" for dangerous actions */
  variant?: "default" | "destructive";
  /** Callback when confirm is clicked */
  onConfirm: () => void;
  /** Whether the action is in progress (disables buttons) */
  loading?: boolean;
  /** i18n namespace for button labels (defaults to "common") */
  i18nNamespace?: keyof (typeof resources)["en"];
}

/**
 * Reusable confirmation dialog for destructive or important actions.
 *
 * @example
 * ```tsx
 * const [open, setOpen] = useState(false);
 *
 * <Button onClick={() => setOpen(true)}>Delete</Button>
 * <ConfirmDialog
 *   open={open}
 *   onOpenChange={setOpen}
 *   title="Delete item?"
 *   description="This action cannot be undone."
 *   variant="destructive"
 *   onConfirm={() => deleteItem()}
 * />
 * ```
 */
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  cancelLabel,
  variant = "default",
  onConfirm,
  loading = false,
  i18nNamespace = "common",
}: ConfirmDialogProps) {
  const { t } = useTranslation(i18nNamespace);

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={loading}>
            {cancelLabel ?? t("actions.cancel", "Cancel")}
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault();
              onConfirm();
            }}
            disabled={loading}
            className={
              variant === "destructive"
                ? "bg-destructive text-destructive-foreground hover:bg-destructive/90 focus:ring-destructive"
                : ""
            }
          >
            {loading
              ? t("actions.loading", "Loading...")
              : (confirmLabel ?? t("actions.confirm", "Confirm"))}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
