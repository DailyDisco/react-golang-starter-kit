import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import {
  FileQuestion,
  FolderOpen,
  Inbox,
  Search,
  Users,
  type LucideIcon,
} from 'lucide-react';
import * as React from 'react';
import { Button } from './button';

type EmptyStateVariant = 'default' | 'search' | 'users' | 'files' | 'folder';

interface EmptyStateAction {
  label: string;
  onClick?: () => void;
  href?: string;
  variant?: 'default' | 'outline' | 'secondary';
}

interface EmptyStateProps {
  /** Predefined variant with matching icon */
  variant?: EmptyStateVariant;
  /** Custom icon (overrides variant icon) */
  icon?: React.ReactNode;
  /** Main heading */
  title: string;
  /** Supporting description text */
  description?: string;
  /** Primary action button */
  action?: EmptyStateAction;
  /** Secondary action button */
  secondaryAction?: EmptyStateAction;
  /** Additional content below actions */
  children?: React.ReactNode;
  /** Container className */
  className?: string;
  /** Size variant */
  size?: 'sm' | 'default' | 'lg';
}

const variantIcons: Record<EmptyStateVariant, LucideIcon> = {
  default: Inbox,
  search: Search,
  users: Users,
  files: FileQuestion,
  folder: FolderOpen,
};

const sizeClasses = {
  sm: {
    container: 'py-8',
    iconWrapper: 'mb-4 h-14 w-14',
    icon: 'h-7 w-7',
    title: 'text-base',
    description: 'text-sm max-w-xs',
  },
  default: {
    container: 'py-16',
    iconWrapper: 'mb-6 h-20 w-20',
    icon: 'h-10 w-10',
    title: 'text-lg',
    description: 'text-sm max-w-sm',
  },
  lg: {
    container: 'py-24',
    iconWrapper: 'mb-8 h-24 w-24',
    icon: 'h-12 w-12',
    title: 'text-xl',
    description: 'text-base max-w-md',
  },
};

/**
 * Empty state component for when there's no data to display
 *
 * @example
 * ```tsx
 * <EmptyState
 *   variant="users"
 *   title="No users yet"
 *   description="Get started by inviting your first team member."
 *   action={{
 *     label: "Invite User",
 *     onClick: () => setIsInviteOpen(true)
 *   }}
 * />
 * ```
 */
function EmptyState({
  variant = 'default',
  icon,
  title,
  description,
  action,
  secondaryAction,
  children,
  className,
  size = 'default',
}: EmptyStateProps) {
  const IconComponent = variantIcons[variant];
  const sizes = sizeClasses[size];

  const displayIcon = icon || (
    <IconComponent className={sizes.icon} strokeWidth={1.5} />
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: [0.25, 0.1, 0.25, 1] }}
      className={cn(
        'flex flex-col items-center justify-center text-center',
        sizes.container,
        className
      )}
    >
      <motion.div
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{
          delay: 0.1,
          type: 'spring',
          stiffness: 200,
          damping: 20,
        }}
        className={cn(
          'flex items-center justify-center rounded-full bg-muted text-muted-foreground',
          sizes.iconWrapper
        )}
      >
        {displayIcon}
      </motion.div>

      <motion.h3
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.2 }}
        className={cn('font-semibold text-foreground', sizes.title)}
      >
        {title}
      </motion.h3>

      {description && (
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3 }}
          className={cn('mt-2 text-muted-foreground', sizes.description)}
        >
          {description}
        </motion.p>
      )}

      {(action || secondaryAction) && (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="mt-6 flex flex-wrap items-center justify-center gap-3"
        >
          {action && (
            <Button
              onClick={action.onClick}
              variant={action.variant ?? 'default'}
            >
              {action.label}
            </Button>
          )}
          {secondaryAction && (
            <Button
              onClick={secondaryAction.onClick}
              variant={secondaryAction.variant ?? 'outline'}
            >
              {secondaryAction.label}
            </Button>
          )}
        </motion.div>
      )}

      {children && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="mt-6"
        >
          {children}
        </motion.div>
      )}
    </motion.div>
  );
}

/**
 * Simplified empty state for inline use (e.g., in tables, lists)
 */
function EmptyStateInline({
  icon,
  message,
  className,
}: {
  icon?: React.ReactNode;
  message: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        'flex items-center justify-center gap-2 py-8 text-muted-foreground',
        className
      )}
    >
      {icon || <Inbox className="h-4 w-4" />}
      <span className="text-sm">{message}</span>
    </div>
  );
}

export { EmptyState, EmptyStateInline };
