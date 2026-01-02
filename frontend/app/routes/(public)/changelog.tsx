import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { queryKeys } from "@/lib/query-keys";
import { cn } from "@/lib/utils";
import { ChangelogService, type AnnouncementCategory, type ChangelogEntry } from "@/services/admin/adminService";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { format } from "date-fns";
import { Bug, ExternalLink, Loader2, Rocket, Sparkles } from "lucide-react";

export const Route = createFileRoute("/(public)/changelog")({
  component: ChangelogPage,
});

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

function ChangelogPage() {
  const [page, setPage] = useState(1);
  const [selectedCategory, setSelectedCategory] = useState<AnnouncementCategory | undefined>();
  const limit = 10;

  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.changelog.entries(page, limit, selectedCategory),
    queryFn: () => ChangelogService.getChangelog(page, limit, selectedCategory),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const entries = data?.data || [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages || 1;

  return (
    <div className="mx-auto max-w-3xl px-4 py-16">
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="mb-4 text-4xl font-bold">Changelog</h1>
        <p className="text-muted-foreground mx-auto max-w-2xl text-lg">
          Stay up to date with the latest features, improvements, and bug fixes.
        </p>
      </div>

      {/* Category Filter */}
      <div className="mb-8 flex flex-wrap justify-center gap-2">
        <Button
          variant={selectedCategory === undefined ? "default" : "outline"}
          size="sm"
          onClick={() => {
            setSelectedCategory(undefined);
            setPage(1);
          }}
        >
          All
        </Button>
        {(Object.keys(categoryConfig) as AnnouncementCategory[]).map((category) => {
          const config = categoryConfig[category];
          const Icon = config.icon;
          return (
            <Button
              key={category}
              variant={selectedCategory === category ? "default" : "outline"}
              size="sm"
              onClick={() => {
                setSelectedCategory(category);
                setPage(1);
              }}
              className="gap-1.5"
            >
              <Icon className="size-3.5" />
              {config.label}
            </Button>
          );
        })}
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="text-muted-foreground size-8 animate-spin" />
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="border-destructive/50 bg-destructive/10 rounded-lg border p-6 text-center">
          <p className="text-destructive">Failed to load changelog. Please try again later.</p>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && entries.length === 0 && (
        <div className="bg-muted/50 rounded-lg border p-12 text-center">
          <Rocket className="text-muted-foreground/50 mx-auto size-12" />
          <h3 className="mt-4 text-lg font-medium">No updates yet</h3>
          <p className="text-muted-foreground mt-2">Check back soon for new features and improvements.</p>
        </div>
      )}

      {/* Changelog Entries */}
      {!isLoading && !error && entries.length > 0 && (
        <div className="space-y-8">
          {entries.map((entry) => (
            <ChangelogCard
              key={entry.id}
              entry={entry}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {meta && totalPages > 1 && (
        <div className="mt-12 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            Previous
          </Button>
          <span className="text-muted-foreground px-4 text-sm">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
          >
            Next
          </Button>
        </div>
      )}

      {/* Back to Home */}
      <div className="mt-16 text-center">
        <Link
          to="/"
          className="text-primary hover:underline"
        >
          &larr; Back to home
        </Link>
      </div>
    </div>
  );
}

interface ChangelogCardProps {
  entry: ChangelogEntry;
}

function ChangelogCard({ entry }: ChangelogCardProps) {
  const config = categoryConfig[entry.category] || categoryConfig.update;
  const Icon = config.icon;
  const publishedDate = entry.published_at ? new Date(entry.published_at) : null;

  return (
    <article className="group bg-card rounded-xl border p-6 transition-all hover:shadow-md">
      <div className="mb-4 flex items-center justify-between gap-4">
        <Badge
          variant={config.variant}
          className="gap-1.5"
        >
          <Icon
            className="size-3"
            aria-hidden="true"
          />
          {config.label}
        </Badge>
        {publishedDate && (
          <time
            dateTime={entry.published_at}
            className="text-muted-foreground text-sm"
          >
            {format(publishedDate, "MMMM d, yyyy")}
          </time>
        )}
      </div>
      <h2 className="mb-2 text-xl font-semibold">{entry.title}</h2>
      <p className="text-muted-foreground whitespace-pre-wrap">{entry.message}</p>
      {entry.link_url && (
        <a
          href={entry.link_url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary mt-4 inline-flex items-center gap-1.5 text-sm font-medium hover:underline"
        >
          {entry.link_text || "Learn more"}
          <ExternalLink
            className="size-3.5"
            aria-hidden="true"
          />
        </a>
      )}
    </article>
  );
}
