import { memo, useMemo, useState } from "react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ActivityEmptyState } from "@/components/ui/empty-state";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import {
  AlertCircle,
  CheckCircle,
  CreditCard,
  FileText,
  Key,
  LogIn,
  LogOut,
  Settings,
  Shield,
  User,
  type LucideIcon,
} from "lucide-react";

export type ActivityType =
  | "login"
  | "logout"
  | "profile_update"
  | "password_change"
  | "settings_change"
  | "payment"
  | "subscription"
  | "security"
  | "file_upload"
  | "success"
  | "warning"
  | "error";

export interface ActivityItem {
  id: string;
  type: ActivityType;
  title: string;
  description?: string;
  timestamp: Date | string;
  user?: {
    name: string;
    avatar?: string;
  };
  metadata?: Record<string, unknown>;
}

interface ActivityFeedProps {
  /** Activity items to display */
  activities: ActivityItem[];
  /** Whether data is loading */
  isLoading?: boolean;
  /** Maximum number of items to show */
  maxItems?: number;
  /** Whether to show user avatars */
  showAvatars?: boolean;
  /** Title for the card */
  title?: string;
  /** Additional CSS classes */
  className?: string;
}

const activityConfig: Record<ActivityType, { icon: LucideIcon; color: string; bgColor: string }> = {
  login: {
    icon: LogIn,
    color: "text-blue-600 dark:text-blue-400",
    bgColor: "bg-blue-100 dark:bg-blue-900/30",
  },
  logout: {
    icon: LogOut,
    color: "text-gray-600 dark:text-gray-400",
    bgColor: "bg-gray-100 dark:bg-gray-800/50",
  },
  profile_update: {
    icon: User,
    color: "text-purple-600 dark:text-purple-400",
    bgColor: "bg-purple-100 dark:bg-purple-900/30",
  },
  password_change: {
    icon: Key,
    color: "text-orange-600 dark:text-orange-400",
    bgColor: "bg-orange-100 dark:bg-orange-900/30",
  },
  settings_change: {
    icon: Settings,
    color: "text-gray-600 dark:text-gray-400",
    bgColor: "bg-gray-100 dark:bg-gray-800/50",
  },
  payment: {
    icon: CreditCard,
    color: "text-green-600 dark:text-green-400",
    bgColor: "bg-green-100 dark:bg-green-900/30",
  },
  subscription: {
    icon: CreditCard,
    color: "text-indigo-600 dark:text-indigo-400",
    bgColor: "bg-indigo-100 dark:bg-indigo-900/30",
  },
  security: {
    icon: Shield,
    color: "text-red-600 dark:text-red-400",
    bgColor: "bg-red-100 dark:bg-red-900/30",
  },
  file_upload: {
    icon: FileText,
    color: "text-cyan-600 dark:text-cyan-400",
    bgColor: "bg-cyan-100 dark:bg-cyan-900/30",
  },
  success: {
    icon: CheckCircle,
    color: "text-green-600 dark:text-green-400",
    bgColor: "bg-green-100 dark:bg-green-900/30",
  },
  warning: {
    icon: AlertCircle,
    color: "text-amber-600 dark:text-amber-400",
    bgColor: "bg-amber-100 dark:bg-amber-900/30",
  },
  error: {
    icon: AlertCircle,
    color: "text-red-600 dark:text-red-400",
    bgColor: "bg-red-100 dark:bg-red-900/30",
  },
};

const ActivityItemRow = memo(function ActivityItemRow({
  activity,
  showAvatar,
}: {
  activity: ActivityItem;
  showAvatar?: boolean;
}) {
  const config = activityConfig[activity.type] || activityConfig.success;
  const Icon = config.icon;
  const timestamp = typeof activity.timestamp === "string" ? new Date(activity.timestamp) : activity.timestamp;

  return (
    <div className="group hover:bg-muted/50 flex gap-3 rounded-lg p-3 transition-colors">
      {/* Icon or Avatar */}
      {showAvatar && activity.user ? (
        <Avatar className="h-9 w-9">
          <AvatarFallback className="text-xs">
            {activity.user.name
              .split(" ")
              .map((n) => n[0])
              .join("")
              .toUpperCase()}
          </AvatarFallback>
        </Avatar>
      ) : (
        <div className={cn("flex h-9 w-9 shrink-0 items-center justify-center rounded-full", config.bgColor)}>
          <Icon className={cn("h-4 w-4", config.color)} />
        </div>
      )}

      {/* Content */}
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm leading-tight font-medium">{activity.title}</p>
          <span className="text-muted-foreground shrink-0 text-xs">
            {formatDistanceToNow(timestamp, { addSuffix: true })}
          </span>
        </div>
        {activity.description && (
          <p className="text-muted-foreground mt-0.5 text-sm leading-tight">{activity.description}</p>
        )}
      </div>
    </div>
  );
});

function ActivityFeedSkeleton() {
  return (
    <div className="space-y-3 p-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <div
          key={i}
          className="flex gap-3"
        >
          <Skeleton className="h-9 w-9 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-1/2" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function ActivityFeed({
  activities,
  isLoading = false,
  maxItems = 10,
  showAvatars = false,
  title = "Recent Activity",
  className,
}: ActivityFeedProps) {
  const displayActivities = useMemo(() => {
    return activities.slice(0, maxItems);
  }, [activities, maxItems]);

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          {title}
          {!isLoading && activities.length > 0 && (
            <span className="bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-xs font-normal">
              {activities.length}
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        {isLoading ? (
          <ActivityFeedSkeleton />
        ) : displayActivities.length === 0 ? (
          <ActivityEmptyState className="py-8" />
        ) : (
          <ScrollArea className="h-[320px]">
            <div className="space-y-1 px-3 pb-3">
              {displayActivities.map((activity) => (
                <ActivityItemRow
                  key={activity.id}
                  activity={activity}
                  showAvatar={showAvatars}
                />
              ))}
            </div>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  );
}

/**
 * Creates mock activity timestamps relative to a base time
 */
function createMockTimestamps(baseTime: number) {
  return {
    fiveMinsAgo: new Date(baseTime - 1000 * 60 * 5),
    oneHourAgo: new Date(baseTime - 1000 * 60 * 60),
    threeHoursAgo: new Date(baseTime - 1000 * 60 * 60 * 3),
    oneDayAgo: new Date(baseTime - 1000 * 60 * 60 * 24),
    twoDaysAgo: new Date(baseTime - 1000 * 60 * 60 * 24 * 2),
    oneWeekAgo: new Date(baseTime - 1000 * 60 * 60 * 24 * 7),
  };
}

/**
 * Hook to generate mock activity data for demo purposes
 */
export function useMockActivities(): ActivityItem[] {
  const [activities] = useState<ActivityItem[]>(() => {
    const timestamps = createMockTimestamps(Date.now());
    return [
      {
        id: "1",
        type: "login",
        title: "Signed in successfully",
        description: "From Chrome on macOS",
        timestamp: timestamps.fiveMinsAgo,
      },
      {
        id: "2",
        type: "profile_update",
        title: "Profile updated",
        description: "Changed display name",
        timestamp: timestamps.oneHourAgo,
      },
      {
        id: "3",
        type: "security",
        title: "Two-factor authentication enabled",
        timestamp: timestamps.threeHoursAgo,
      },
      {
        id: "4",
        type: "payment",
        title: "Payment successful",
        description: "$29.00 for Pro Plan",
        timestamp: timestamps.oneDayAgo,
      },
      {
        id: "5",
        type: "settings_change",
        title: "Notification preferences updated",
        timestamp: timestamps.twoDaysAgo,
      },
      {
        id: "6",
        type: "password_change",
        title: "Password changed",
        timestamp: timestamps.oneWeekAgo,
      },
    ];
  });
  return activities;
}
