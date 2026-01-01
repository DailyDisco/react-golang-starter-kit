import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Announcement, AnnouncementCategory } from "@/services/admin/adminService";
import { Bug, ExternalLink, Rocket, Sparkles, X } from "lucide-react";

interface AnnouncementBannerProps {
  announcement: Announcement;
  onDismiss?: (id: number) => void;
}

const categoryConfig: Record<AnnouncementCategory, { icon: typeof Sparkles; label: string; className: string }> = {
  feature: {
    icon: Sparkles,
    label: "New Feature",
    className: "bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-950/50 dark:border-blue-800 dark:text-blue-200",
  },
  bugfix: {
    icon: Bug,
    label: "Bug Fix",
    className:
      "bg-amber-50 border-amber-200 text-amber-800 dark:bg-amber-950/50 dark:border-amber-800 dark:text-amber-200",
  },
  update: {
    icon: Rocket,
    label: "Update",
    className:
      "bg-purple-50 border-purple-200 text-purple-800 dark:bg-purple-950/50 dark:border-purple-800 dark:text-purple-200",
  },
};

export function AnnouncementBanner({ announcement, onDismiss }: AnnouncementBannerProps) {
  const config = categoryConfig[announcement.category] || categoryConfig.update;
  const Icon = config.icon;

  return (
    <div
      role="alert"
      aria-live="polite"
      className={cn("relative flex items-center justify-between gap-4 border-b px-4 py-2.5 text-sm", config.className)}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <div className="flex shrink-0 items-center gap-2">
          <Icon
            className="size-4"
            aria-hidden="true"
          />
          <span className="text-xs font-medium tracking-wide uppercase">{config.label}</span>
        </div>
        <span className="truncate font-medium">{announcement.title}</span>
        <span className="hidden truncate text-current/80 sm:inline">{announcement.message}</span>
        {announcement.link_url && (
          <a
            href={announcement.link_url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex shrink-0 items-center gap-1 font-medium underline underline-offset-2 hover:no-underline"
          >
            {announcement.link_text || "Learn more"}
            <ExternalLink
              className="size-3"
              aria-hidden="true"
            />
          </a>
        )}
      </div>
      {announcement.is_dismissible && onDismiss && (
        <Button
          variant="ghost"
          size="icon"
          className="size-6 shrink-0 hover:bg-current/10"
          onClick={() => onDismiss(announcement.id)}
          aria-label="Dismiss announcement"
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  );
}
