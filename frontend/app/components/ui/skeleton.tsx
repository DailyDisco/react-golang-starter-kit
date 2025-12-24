import { cn } from "@/lib/utils";

interface SkeletonProps extends React.ComponentProps<"div"> {
  /**
   * Animation variant:
   * - pulse: Simple opacity animation (default)
   * - shimmer: Wave effect animation
   * - none: No animation
   */
  variant?: "pulse" | "shimmer" | "none";
}

function Skeleton({ className, variant = "shimmer", ...props }: SkeletonProps) {
  return (
    <div
      data-slot="skeleton"
      className={cn(
        "rounded-md bg-muted",
        variant === "pulse" && "animate-pulse",
        variant === "shimmer" && "relative overflow-hidden before:absolute before:inset-0 before:-translate-x-full before:animate-shimmer before:bg-linear-to-r before:from-transparent before:via-white/20 before:to-transparent",
        className
      )}
      {...props}
    />
  );
}

export { Skeleton };
