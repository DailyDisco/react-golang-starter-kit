import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import type { Announcement, AnnouncementCategory } from "@/services/admin/adminService";
import { Bug, ExternalLink, Rocket, Sparkles } from "lucide-react";

interface AnnouncementModalProps {
  announcement: Announcement | null;
  open: boolean;
  onClose: () => void;
  onMarkRead?: (id: number) => void;
}

const categoryConfig: Record<
  AnnouncementCategory,
  { icon: typeof Sparkles; label: string; variant: "default" | "secondary" | "outline" }
> = {
  feature: {
    icon: Sparkles,
    label: "New Feature",
    variant: "default",
  },
  bugfix: {
    icon: Bug,
    label: "Bug Fix",
    variant: "secondary",
  },
  update: {
    icon: Rocket,
    label: "Update",
    variant: "outline",
  },
};

export function AnnouncementModal({ announcement, open, onClose, onMarkRead }: AnnouncementModalProps) {
  if (!announcement) return null;

  const config = categoryConfig[announcement.category] || categoryConfig.update;
  const Icon = config.icon;

  const handleClose = () => {
    if (onMarkRead) {
      onMarkRead(announcement.id);
    }
    onClose();
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => !isOpen && handleClose()}
    >
      <DialogContent
        className="sm:max-w-md"
        aria-describedby="announcement-description"
      >
        <DialogHeader className="space-y-3">
          <div className="flex items-center gap-2">
            <Badge
              variant={config.variant}
              className="gap-1"
            >
              <Icon
                className="size-3"
                aria-hidden="true"
              />
              {config.label}
            </Badge>
          </div>
          <DialogTitle className="text-xl">{announcement.title}</DialogTitle>
        </DialogHeader>
        <DialogDescription
          id="announcement-description"
          className="text-foreground/80 text-base whitespace-pre-wrap"
        >
          {announcement.message}
        </DialogDescription>
        <DialogFooter className="flex-col gap-2 sm:flex-row sm:gap-0">
          {announcement.link_url && (
            <Button
              variant="outline"
              asChild
              className="w-full sm:w-auto"
            >
              <a
                href={announcement.link_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2"
              >
                {announcement.link_text || "Learn more"}
                <ExternalLink
                  className="size-4"
                  aria-hidden="true"
                />
              </a>
            </Button>
          )}
          <Button
            onClick={handleClose}
            className="w-full sm:w-auto"
          >
            Got it
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
