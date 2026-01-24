import { useCallback, useState } from "react";

import { Link, type LinkProps } from "@tanstack/react-router";

interface PrefetchLinkProps extends Omit<LinkProps, "children"> {
  /** Function to call for prefetching data */
  onPrefetch?: () => void;
  /** Delay in ms before prefetch triggers (prevents accidental prefetch) */
  prefetchDelay?: number;
  /** Children to render */
  children: React.ReactNode;
  /** Additional class names */
  className?: string;
}

/**
 * A Link component that prefetches data on mouse enter/focus.
 * Use this for navigation links where you want to prefetch route data
 * before the user actually clicks.
 *
 * @example
 * const prefetchSettings = usePrefetchSettingsData();
 *
 * <PrefetchLink to="/settings" onPrefetch={prefetchSettings}>
 *   Settings
 * </PrefetchLink>
 */
export function PrefetchLink({
  onPrefetch,
  prefetchDelay = 100,
  children,
  className,
  ...linkProps
}: PrefetchLinkProps) {
  const [timeoutId, setTimeoutId] = useState<ReturnType<typeof setTimeout> | null>(null);

  const handleMouseEnter = useCallback(() => {
    if (!onPrefetch) return;

    // Use a small delay to avoid prefetching on accidental hover
    const id = setTimeout(() => {
      onPrefetch();
    }, prefetchDelay);
    setTimeoutId(id);
  }, [onPrefetch, prefetchDelay]);

  const handleMouseLeave = useCallback(() => {
    if (timeoutId) {
      clearTimeout(timeoutId);
      setTimeoutId(null);
    }
  }, [timeoutId]);

  const handleFocus = useCallback(() => {
    // Prefetch immediately on focus (keyboard navigation)
    if (onPrefetch) {
      onPrefetch();
    }
  }, [onPrefetch]);

  return (
    <Link
      {...linkProps}
      className={className}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      onFocus={handleFocus}
    >
      {children}
    </Link>
  );
}
