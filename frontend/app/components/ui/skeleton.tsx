import { cn } from "@/lib/utils";

interface SkeletonProps extends React.ComponentProps<"div"> {
  /**
   * Animation variant:
   * - pulse: Simple opacity animation
   * - shimmer: Wave effect animation (default)
   * - wave: Alternative wave with foreground color
   * - none: No animation
   */
  variant?: "pulse" | "shimmer" | "wave" | "none";
  /**
   * Corner radius:
   * - none: No rounding
   * - sm: Small radius
   * - md: Medium radius (default)
   * - lg: Large radius
   * - full: Circular/pill shape
   */
  rounded?: "none" | "sm" | "md" | "lg" | "full";
}

const roundedClasses = {
  none: "rounded-none",
  sm: "rounded-sm",
  md: "rounded-md",
  lg: "rounded-lg",
  full: "rounded-full",
};

function Skeleton({
  className,
  variant = "shimmer",
  rounded = "md",
  ...props
}: SkeletonProps) {
  return (
    <div
      data-slot="skeleton"
      className={cn(
        "bg-muted",
        roundedClasses[rounded],
        variant === "pulse" && "animate-pulse",
        variant === "shimmer" &&
          "relative overflow-hidden before:absolute before:inset-0 before:-translate-x-full before:animate-shimmer before:bg-linear-to-r before:from-transparent before:via-white/20 before:to-transparent",
        variant === "wave" &&
          "relative overflow-hidden after:absolute after:inset-0 after:-translate-x-full after:animate-shimmer after:bg-linear-to-r after:from-transparent after:via-foreground/5 after:to-transparent",
        className
      )}
      {...props}
    />
  );
}

// ==========================================================================
// PRESET SKELETON COMPONENTS
// ==========================================================================

/** Skeleton for text lines */
function SkeletonText({
  lines = 3,
  className,
  ...props
}: { lines?: number } & Omit<SkeletonProps, "children">) {
  return (
    <div className={cn("space-y-2", className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          className={cn(
            "h-4",
            i === lines - 1 && lines > 1 ? "w-3/4" : "w-full"
          )}
          {...props}
        />
      ))}
    </div>
  );
}

/** Skeleton for avatar/profile pictures */
function SkeletonAvatar({
  size = "md",
  ...props
}: { size?: "sm" | "md" | "lg" } & Omit<SkeletonProps, "rounded">) {
  const sizeClasses = {
    sm: "h-8 w-8",
    md: "h-10 w-10",
    lg: "h-12 w-12",
  };

  return <Skeleton className={sizeClasses[size]} rounded="full" {...props} />;
}

/** Skeleton for cards with image and text */
function SkeletonCard({ className, ...props }: SkeletonProps) {
  return (
    <div className={cn("space-y-3", className)} {...props}>
      <Skeleton className="h-32 w-full" rounded="lg" />
      <div className="space-y-2">
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
      </div>
    </div>
  );
}

/** Skeleton for list items with avatar */
function SkeletonListItem({ className, ...props }: SkeletonProps) {
  return (
    <div className={cn("flex items-center gap-3", className)} {...props}>
      <SkeletonAvatar />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-1/3" />
        <Skeleton className="h-3 w-1/2" />
      </div>
    </div>
  );
}

/** Skeleton for table rows */
function SkeletonTableRow({
  columns = 4,
  className,
  ...props
}: { columns?: number } & SkeletonProps) {
  return (
    <div className={cn("flex items-center gap-4 py-3", className)} {...props}>
      {Array.from({ length: columns }).map((_, i) => (
        <Skeleton
          key={i}
          className={cn("h-4 flex-1", i === 0 && "max-w-[200px]")}
        />
      ))}
    </div>
  );
}

export {
  Skeleton,
  SkeletonText,
  SkeletonAvatar,
  SkeletonCard,
  SkeletonListItem,
  SkeletonTableRow,
};
