import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { Loader2 } from 'lucide-react';
import * as React from 'react';

interface PageLoadingProps {
  /** Loading message to display */
  message?: string;
  /** Size variant */
  size?: 'sm' | 'default' | 'lg';
  /** Additional className */
  className?: string;
  /** Minimum height */
  minHeight?: string;
}

const sizeConfig = {
  sm: {
    spinner: 'h-5 w-5',
    text: 'text-xs',
    gap: 'gap-2',
    minHeight: 'min-h-[200px]',
  },
  default: {
    spinner: 'h-8 w-8',
    text: 'text-sm',
    gap: 'gap-4',
    minHeight: 'min-h-[400px]',
  },
  lg: {
    spinner: 'h-12 w-12',
    text: 'text-base',
    gap: 'gap-6',
    minHeight: 'min-h-[600px]',
  },
};

/**
 * Full page/section loading indicator with optional message
 *
 * @example
 * ```tsx
 * <PageLoading message="Loading your dashboard..." />
 * ```
 */
function PageLoading({
  message = 'Loading...',
  size = 'default',
  className,
  minHeight,
}: PageLoadingProps) {
  const config = sizeConfig[size];

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className={cn(
        'flex flex-col items-center justify-center',
        config.gap,
        minHeight || config.minHeight,
        className
      )}
    >
      <motion.div
        animate={{ rotate: 360 }}
        transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
      >
        <Loader2
          className={cn(config.spinner, 'text-primary')}
          aria-hidden="true"
        />
      </motion.div>
      {message && (
        <motion.p
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className={cn(config.text, 'text-muted-foreground')}
        >
          {message}
        </motion.p>
      )}
    </motion.div>
  );
}

interface InlineLoadingProps {
  /** Loading message */
  message?: string;
  /** Additional className */
  className?: string;
}

/**
 * Inline loading indicator for smaller areas
 *
 * @example
 * ```tsx
 * <InlineLoading message="Saving..." />
 * ```
 */
function InlineLoading({ message, className }: InlineLoadingProps) {
  return (
    <div
      className={cn('flex items-center gap-2 text-muted-foreground', className)}
    >
      <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
      {message && <span className="text-sm">{message}</span>}
    </div>
  );
}

interface LoadingOverlayProps {
  /** Whether the overlay is visible */
  isLoading: boolean;
  /** Loading message */
  message?: string;
  /** Children to render behind overlay */
  children: React.ReactNode;
  /** Additional className for overlay */
  className?: string;
}

/**
 * Loading overlay that covers its children
 *
 * @example
 * ```tsx
 * <LoadingOverlay isLoading={isSaving} message="Saving changes...">
 *   <Form>...</Form>
 * </LoadingOverlay>
 * ```
 */
function LoadingOverlay({
  isLoading,
  message,
  children,
  className,
}: LoadingOverlayProps) {
  return (
    <div className="relative">
      {children}
      {isLoading && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className={cn(
            'absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm',
            className
          )}
        >
          <div className="flex flex-col items-center gap-3">
            <Loader2
              className="h-6 w-6 animate-spin text-primary"
              aria-hidden="true"
            />
            {message && (
              <span className="text-sm text-muted-foreground">{message}</span>
            )}
          </div>
        </motion.div>
      )}
    </div>
  );
}

/**
 * Dots loading animation (alternative to spinner)
 */
function LoadingDots({ className }: { className?: string }) {
  return (
    <div className={cn('flex items-center gap-1', className)}>
      {[0, 1, 2].map((i) => (
        <motion.span
          key={i}
          className="h-2 w-2 rounded-full bg-primary"
          animate={{ opacity: [0.3, 1, 0.3] }}
          transition={{
            duration: 1,
            repeat: Infinity,
            delay: i * 0.2,
          }}
        />
      ))}
    </div>
  );
}

export { PageLoading, InlineLoading, LoadingOverlay, LoadingDots };
