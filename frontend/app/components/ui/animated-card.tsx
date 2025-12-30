import { cn } from '@/lib/utils';
import { motion, type HTMLMotionProps } from 'framer-motion';
import * as React from 'react';

interface AnimatedCardProps extends HTMLMotionProps<'div'> {
  /** Delay before animation starts (in seconds) */
  delay?: number;
  /** Whether to show hover lift effect */
  hoverEffect?: boolean;
  /** Whether to show hover shadow effect */
  hoverShadow?: boolean;
  /** Animation variant */
  animation?: 'fadeInUp' | 'fadeIn' | 'scaleIn' | 'slideInRight' | 'none';
}

const animationVariants = {
  fadeInUp: {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
  },
  fadeIn: {
    initial: { opacity: 0 },
    animate: { opacity: 1 },
  },
  scaleIn: {
    initial: { opacity: 0, scale: 0.95 },
    animate: { opacity: 1, scale: 1 },
  },
  slideInRight: {
    initial: { opacity: 0, x: 20 },
    animate: { opacity: 1, x: 0 },
  },
  none: {
    initial: {},
    animate: {},
  },
};

/**
 * Card component with entrance animation and optional hover effects
 *
 * @example
 * ```tsx
 * <AnimatedCard delay={0.1} hoverEffect>
 *   <CardHeader>...</CardHeader>
 *   <CardContent>...</CardContent>
 * </AnimatedCard>
 * ```
 */
function AnimatedCard({
  children,
  className,
  delay = 0,
  hoverEffect = true,
  hoverShadow = true,
  animation = 'fadeInUp',
  ...props
}: AnimatedCardProps) {
  const variant = animationVariants[animation];

  return (
    <motion.div
      initial={variant.initial}
      animate={variant.animate}
      transition={{
        duration: 0.4,
        delay,
        ease: [0.25, 0.1, 0.25, 1],
      }}
      whileHover={
        hoverEffect
          ? {
              y: -2,
              transition: { duration: 0.2 },
            }
          : undefined
      }
      className={cn(
        'bg-card text-card-foreground rounded-xl border shadow-sm',
        hoverShadow && 'transition-shadow hover:shadow-md',
        className
      )}
      {...props}
    >
      {children}
    </motion.div>
  );
}

interface AnimatedCardHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Whether to add bottom border */
  bordered?: boolean;
}

/**
 * Header section for AnimatedCard
 */
function AnimatedCardHeader({
  className,
  bordered = false,
  ...props
}: AnimatedCardHeaderProps) {
  return (
    <div
      className={cn(
        'flex flex-col space-y-1.5 p-6',
        bordered && 'border-b',
        className
      )}
      {...props}
    />
  );
}

/**
 * Content section for AnimatedCard
 */
function AnimatedCardContent({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('p-6 pt-0', className)} {...props} />;
}

/**
 * Footer section for AnimatedCard
 */
function AnimatedCardFooter({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('flex items-center p-6 pt-0', className)}
      {...props}
    />
  );
}

interface StatCardProps {
  /** Card title/label */
  title: string;
  /** Main value to display */
  value: string | number;
  /** Optional description or subtext */
  description?: string;
  /** Optional icon */
  icon?: React.ReactNode;
  /** Trend indicator */
  trend?: {
    value: number;
    label?: string;
  };
  /** Animation delay */
  delay?: number;
  /** Additional className */
  className?: string;
}

/**
 * Pre-styled stat card for dashboards
 *
 * @example
 * ```tsx
 * <StatCard
 *   title="Total Users"
 *   value={1234}
 *   description="Active this month"
 *   trend={{ value: 12, label: "vs last month" }}
 *   icon={<Users className="h-4 w-4" />}
 * />
 * ```
 */
function StatCard({
  title,
  value,
  description,
  icon,
  trend,
  delay = 0,
  className,
}: StatCardProps) {
  const isPositive = trend && trend.value >= 0;

  return (
    <AnimatedCard delay={delay} className={cn('p-6', className)}>
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">
          {title}
        </span>
        {icon && (
          <div className="text-muted-foreground">{icon}</div>
        )}
      </div>
      <div className="mt-2">
        <span className="text-2xl font-bold">{value}</span>
        {trend && (
          <span
            className={cn(
              'ml-2 text-xs font-medium',
              isPositive ? 'text-success' : 'text-destructive'
            )}
          >
            {isPositive ? '+' : ''}
            {trend.value}%
            {trend.label && (
              <span className="text-muted-foreground"> {trend.label}</span>
            )}
          </span>
        )}
      </div>
      {description && (
        <p className="mt-1 text-xs text-muted-foreground">{description}</p>
      )}
    </AnimatedCard>
  );
}

export {
  AnimatedCard,
  AnimatedCardHeader,
  AnimatedCardContent,
  AnimatedCardFooter,
  StatCard,
};
