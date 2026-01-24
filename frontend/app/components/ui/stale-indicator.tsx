import { RefreshCw } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";

import { Badge } from "./badge";
import { Button } from "./button";
import { Tooltip, TooltipContent, TooltipTrigger } from "./tooltip";

interface StaleIndicatorProps {
  /** Whether the data is considered stale */
  isStale: boolean;
  /** Whether a refetch is currently in progress */
  isFetching: boolean;
  /** Callback to trigger a refresh */
  onRefresh: () => void;
  /** Timestamp of when data was last updated (ms since epoch) */
  dataUpdatedAt?: number;
  /** Custom className */
  className?: string;
  /** Show the badge even when not stale (just the refresh button) */
  showRefreshAlways?: boolean;
}

/**
 * Indicator component that shows when data is stale and allows refresh.
 *
 * Uses TanStack Query's state properties:
 * - `isStale`: Data is older than staleTime
 * - `isFetching`: A refetch is in progress
 * - `dataUpdatedAt`: Timestamp of last successful fetch
 *
 * @example
 * const { data, isStale, isFetching, dataUpdatedAt, refetch } = useUsers();
 *
 * return (
 *   <div className="flex items-center justify-between">
 *     <h2>Users</h2>
 *     <StaleIndicator
 *       isStale={isStale}
 *       isFetching={isFetching}
 *       dataUpdatedAt={dataUpdatedAt}
 *       onRefresh={() => refetch()}
 *     />
 *   </div>
 * );
 */
function StaleIndicator({
  isStale,
  isFetching,
  onRefresh,
  dataUpdatedAt,
  className,
  showRefreshAlways = false,
}: StaleIndicatorProps) {
  // Don't render if not stale, not fetching, and not showing always
  if (!isStale && !isFetching && !showRefreshAlways) {
    return null;
  }

  const timeAgo = dataUpdatedAt ? formatTimeAgo(new Date(dataUpdatedAt)) : "unknown";

  const tooltipText = isFetching
    ? "Refreshing data..."
    : isStale
      ? `Data from ${timeAgo}. Click to refresh.`
      : `Last updated ${timeAgo}`;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          onClick={onRefresh}
          disabled={isFetching}
          className={cn("gap-1.5 h-8 px-2", className)}
          aria-label={isFetching ? "Refreshing" : "Refresh data"}
        >
          <RefreshCw
            className={cn("h-3.5 w-3.5", isFetching && "animate-spin")}
            aria-hidden="true"
          />
          {isStale && !isFetching && (
            <Badge variant="secondary" className="text-xs px-1.5 py-0">
              Stale
            </Badge>
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent>
        <p>{tooltipText}</p>
      </TooltipContent>
    </Tooltip>
  );
}

/**
 * Format a timestamp as a relative time string
 */
function formatTimeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);

  if (seconds < 5) return "just now";
  if (seconds < 60) return `${seconds}s ago`;

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;

  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export { StaleIndicator, formatTimeAgo };
export type { StaleIndicatorProps };
