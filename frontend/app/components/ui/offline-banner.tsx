import { WifiOff } from "lucide-react";

import { useOnlineStatus } from "../../hooks/useOnlineStatus";

/**
 * Banner that displays when the user is offline
 * Shows at the top of the page with a warning message
 */
export function OfflineBanner() {
  const { isOnline, wasOffline } = useOnlineStatus();

  // Show reconnected message briefly
  if (wasOffline && isOnline) {
    return (
      <div className="bg-green-600 px-4 py-2 text-center text-sm text-white">
        <span className="font-medium">Back online!</span> Your connection has been restored.
      </div>
    );
  }

  // Show offline banner when not connected
  if (!isOnline) {
    return (
      <div className="bg-yellow-600 px-4 py-2 text-center text-sm text-white">
        <span className="inline-flex items-center gap-2">
          <WifiOff className="h-4 w-4" />
          <span className="font-medium">You're offline.</span> Some features may not be available.
        </span>
      </div>
    );
  }

  return null;
}
