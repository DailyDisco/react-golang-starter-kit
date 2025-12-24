import { RefreshCw, Wifi, WifiOff } from "lucide-react";

import { useOnlineStatus } from "../../hooks/useOnlineStatus";
import { cn } from "../../lib/utils";

interface OfflineBannerProps {
  className?: string;
}

/**
 * Banner that shows when the user is offline
 * Automatically hides when back online with a brief "reconnected" message
 */
export function OfflineBanner({ className }: OfflineBannerProps) {
  const { isOnline, wasOffline } = useOnlineStatus();

  if (isOnline && !wasOffline) {
    return null;
  }

  return (
    <div
      role="alert"
      aria-live="polite"
      className={cn("fixed bottom-4 left-1/2 z-50 -translate-x-1/2 transform", className)}
    >
      <div
        className={cn(
          "flex items-center gap-3 rounded-lg px-4 py-3 shadow-lg transition-all duration-300",
          isOnline ? "bg-green-600 text-white" : "bg-yellow-600 text-white"
        )}
      >
        {isOnline ? (
          <>
            <Wifi className="h-5 w-5" />
            <span className="text-sm font-medium">Back online</span>
            <RefreshCw className="h-4 w-4 animate-spin" />
          </>
        ) : (
          <>
            <WifiOff className="h-5 w-5" />
            <span className="text-sm font-medium">You&apos;re offline. Some features may be unavailable.</span>
          </>
        )}
      </div>
    </div>
  );
}
