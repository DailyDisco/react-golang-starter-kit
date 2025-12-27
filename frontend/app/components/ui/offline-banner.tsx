import { useEffect, useState } from "react";

import { WifiOff } from "lucide-react";

/**
 * OfflineBanner displays a notification when the user loses network connectivity.
 * It automatically shows/hides based on the browser's online/offline status.
 */
export function OfflineBanner() {
  const [isOffline, setIsOffline] = useState(false);

  useEffect(() => {
    // Check initial status (only in browser)
    if (typeof window !== "undefined") {
      setIsOffline(!navigator.onLine);
    }

    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  if (!isOffline) return null;

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="bg-destructive text-destructive-foreground fixed left-0 right-0 top-0 z-50 flex items-center justify-center gap-2 p-2 text-center text-sm font-medium shadow-md"
    >
      <WifiOff className="h-4 w-4" aria-hidden="true" />
      <span>You are offline. Some features may not work until you reconnect.</span>
    </div>
  );
}
