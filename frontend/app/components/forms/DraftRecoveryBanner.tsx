import { useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { FileText, X } from "lucide-react";
import { useTranslation } from "react-i18next";

interface DraftRecoveryBannerProps {
  /** Timestamp of when draft was saved (from getDraftTimestamp) */
  timestamp: number | null;
  /** Callback when user dismisses the banner (draft is kept) */
  onDismiss?: () => void;
  /** Callback when user discards the draft */
  onDiscard: () => void;
}

/**
 * Format a timestamp as relative time (e.g., "5 minutes ago")
 */
function formatRelativeTime(timestamp: number): string {
  const now = Date.now();
  const diff = now - timestamp;

  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return days === 1 ? "1 day ago" : `${days} days ago`;
  }
  if (hours > 0) {
    return hours === 1 ? "1 hour ago" : `${hours} hours ago`;
  }
  if (minutes > 0) {
    return minutes === 1 ? "1 minute ago" : `${minutes} minutes ago`;
  }
  return "just now";
}

/**
 * Banner shown when a form draft is restored from localStorage.
 * Shows relative time and allows user to dismiss or discard the draft.
 *
 * @example
 * ```tsx
 * const { clearDraft, hasDraft, getDraftTimestamp } = useFormPersist(form, {
 *   key: "create-user",
 * });
 *
 * return (
 *   <form>
 *     {hasDraft() && (
 *       <DraftRecoveryBanner
 *         timestamp={getDraftTimestamp()}
 *         onDiscard={clearDraft}
 *       />
 *     )}
 *     ...
 *   </form>
 * );
 * ```
 */
export function DraftRecoveryBanner({ timestamp, onDismiss, onDiscard }: DraftRecoveryBannerProps) {
  const { t } = useTranslation("common");
  const [dismissed, setDismissed] = useState(false);

  if (dismissed) {
    return null;
  }

  const relativeTime = timestamp ? formatRelativeTime(timestamp) : null;

  const handleDismiss = () => {
    setDismissed(true);
    onDismiss?.();
  };

  const handleDiscard = () => {
    onDiscard();
    setDismissed(true);
  };

  return (
    <Alert className="bg-info/5 border-info/20 relative mb-4">
      <FileText className="text-info h-4 w-4" />
      <AlertDescription className="flex items-center justify-between gap-4">
        <span className="text-sm">
          {t("form.draftRestored", "Draft restored")}
          {relativeTime && (
            <span className="text-muted-foreground ml-1">
              ({t("form.savedTimeAgo", `saved ${relativeTime}`, { time: relativeTime })})
            </span>
          )}
        </span>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleDiscard}
            className="text-destructive hover:text-destructive h-auto px-2 py-1 text-xs"
          >
            {t("form.discardDraft", "Discard draft")}
          </Button>
          {onDismiss && (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={handleDismiss}
              aria-label={t("common.dismiss", "Dismiss")}
            >
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>
      </AlertDescription>
    </Alert>
  );
}
