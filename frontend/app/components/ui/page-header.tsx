import { cn } from '@/lib/utils';
import * as React from 'react';

interface PageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Optional icon displayed before the title */
  icon?: React.ReactNode;
  /** Page title (required) */
  title: string;
  /** Optional description text below the title */
  description?: string;
  /** Optional action buttons displayed on the right */
  actions?: React.ReactNode;
  /** Optional badge displayed after the title */
  badge?: React.ReactNode;
}

/**
 * Consistent page header with icon, title, description, and actions
 *
 * @example
 * ```tsx
 * <PageHeader
 *   icon={<Users className="h-6 w-6" />}
 *   title="Users"
 *   description="Manage user accounts and permissions"
 *   actions={
 *     <Button>
 *       <Plus className="h-4 w-4" />
 *       Add User
 *     </Button>
 *   }
 * />
 * ```
 */
function PageHeader({
  icon,
  title,
  description,
  actions,
  badge,
  className,
  ...props
}: PageHeaderProps) {
  return (
    <header
      className={cn(
        'mb-8 flex flex-col gap-4 animate-fade-in-up sm:flex-row sm:items-start sm:justify-between',
        className
      )}
      {...props}
    >
      <div className="flex items-start gap-4">
        {icon && (
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
            {icon}
          </div>
        )}
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold tracking-tight text-foreground md:text-3xl">
              {title}
            </h1>
            {badge}
          </div>
          {description && (
            <p className="text-muted-foreground max-w-2xl">{description}</p>
          )}
        </div>
      </div>
      {actions && (
        <div className="flex shrink-0 items-center gap-2 animation-delay-100 animate-fade-in">
          {actions}
        </div>
      )}
    </header>
  );
}

interface PageHeaderSkeletonProps {
  /** Whether to show an icon placeholder */
  hasIcon?: boolean;
  /** Whether to show action button placeholders */
  hasActions?: boolean;
}

/**
 * Skeleton loading state for PageHeader
 */
function PageHeaderSkeleton({
  hasIcon = false,
  hasActions = false,
}: PageHeaderSkeletonProps) {
  return (
    <header className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div className="flex items-start gap-4">
        {hasIcon && (
          <div className="h-12 w-12 shrink-0 animate-pulse rounded-lg bg-muted" />
        )}
        <div className="space-y-2">
          <div className="h-8 w-48 animate-pulse rounded bg-muted" />
          <div className="h-4 w-64 animate-pulse rounded bg-muted" />
        </div>
      </div>
      {hasActions && (
        <div className="flex gap-2">
          <div className="h-9 w-24 animate-pulse rounded-md bg-muted" />
        </div>
      )}
    </header>
  );
}

export { PageHeader, PageHeaderSkeleton };
