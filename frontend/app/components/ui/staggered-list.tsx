import { cn } from '@/lib/utils';
import { motion, type Variants, AnimatePresence } from 'framer-motion';
import * as React from 'react';

interface StaggeredListProps {
  /** List items to animate */
  children: React.ReactNode;
  /** Container className */
  className?: string;
  /** Delay between each item animation (in seconds) */
  staggerDelay?: number;
  /** Initial delay before first item animates (in seconds) */
  initialDelay?: number;
  /** Animation variant for items */
  variant?: 'fadeInUp' | 'fadeIn' | 'slideInRight' | 'scaleIn';
  /** Whether to animate items on exit */
  animateOnExit?: boolean;
  /** Render as different element type */
  as?: 'div' | 'ul' | 'ol';
}

const itemVariants: Record<string, Variants> = {
  fadeInUp: {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] },
    },
    exit: { opacity: 0, y: -10 },
  },
  fadeIn: {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: { duration: 0.3 },
    },
    exit: { opacity: 0 },
  },
  slideInRight: {
    hidden: { opacity: 0, x: 20 },
    visible: {
      opacity: 1,
      x: 0,
      transition: { duration: 0.4, ease: [0.25, 0.1, 0.25, 1] },
    },
    exit: { opacity: 0, x: -20 },
  },
  scaleIn: {
    hidden: { opacity: 0, scale: 0.95 },
    visible: {
      opacity: 1,
      scale: 1,
      transition: { duration: 0.3, ease: [0.25, 0.1, 0.25, 1] },
    },
    exit: { opacity: 0, scale: 0.95 },
  },
};

/**
 * Animates children with staggered entrance animation
 *
 * @example
 * ```tsx
 * <StaggeredList staggerDelay={0.1}>
 *   {items.map(item => (
 *     <Card key={item.id}>{item.title}</Card>
 *   ))}
 * </StaggeredList>
 * ```
 */
function StaggeredList({
  children,
  className,
  staggerDelay = 0.1,
  initialDelay = 0,
  variant = 'fadeInUp',
  animateOnExit = false,
  as: Component = 'div',
}: StaggeredListProps) {
  const containerVariants: Variants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: staggerDelay,
        delayChildren: initialDelay,
      },
    },
    exit: {
      opacity: 0,
      transition: {
        staggerChildren: staggerDelay / 2,
        staggerDirection: -1,
      },
    },
  };

  const MotionComponent = motion[Component];
  const childVariant = itemVariants[variant];

  return (
    <MotionComponent
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      exit={animateOnExit ? 'exit' : undefined}
      className={className}
    >
      {React.Children.map(children, (child, index) => {
        if (!React.isValidElement(child)) return child;

        return (
          <motion.div key={child.key ?? index} variants={childVariant}>
            {child}
          </motion.div>
        );
      })}
    </MotionComponent>
  );
}

interface StaggeredGridProps extends Omit<StaggeredListProps, 'as'> {
  /** Number of columns */
  columns?: 1 | 2 | 3 | 4;
  /** Gap between items */
  gap?: 'sm' | 'md' | 'lg';
}

const columnClasses = {
  1: 'grid-cols-1',
  2: 'grid-cols-1 sm:grid-cols-2',
  3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
  4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
};

const gapClasses = {
  sm: 'gap-3',
  md: 'gap-4',
  lg: 'gap-6',
};

/**
 * Grid layout with staggered animation
 *
 * @example
 * ```tsx
 * <StaggeredGrid columns={3} gap="md">
 *   {products.map(product => (
 *     <ProductCard key={product.id} {...product} />
 *   ))}
 * </StaggeredGrid>
 * ```
 */
function StaggeredGrid({
  children,
  className,
  columns = 3,
  gap = 'md',
  ...props
}: StaggeredGridProps) {
  return (
    <StaggeredList
      className={cn('grid', columnClasses[columns], gapClasses[gap], className)}
      {...props}
    >
      {children}
    </StaggeredList>
  );
}

interface AnimatedListItemProps {
  /** Content to animate */
  children: React.ReactNode;
  /** Additional className */
  className?: string;
  /** Animation delay (for manual control) */
  delay?: number;
  /** Animation variant */
  variant?: 'fadeInUp' | 'fadeIn' | 'slideInRight' | 'scaleIn';
}

/**
 * Individual animated list item (for custom list implementations)
 *
 * @example
 * ```tsx
 * <AnimatedListItem delay={0.2}>
 *   <UserRow user={user} />
 * </AnimatedListItem>
 * ```
 */
function AnimatedListItem({
  children,
  className,
  delay = 0,
  variant = 'fadeInUp',
}: AnimatedListItemProps) {
  const variantConfig = itemVariants[variant];

  return (
    <motion.div
      variants={variantConfig}
      initial="hidden"
      animate="visible"
      exit="exit"
      transition={{ delay }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

interface AnimateListPresenceProps {
  /** Unique key for the current list state */
  listKey?: string;
  /** List items */
  children: React.ReactNode;
  /** Whether to show empty state when no children */
  emptyState?: React.ReactNode;
}

/**
 * Wrapper for AnimatePresence with list-specific defaults
 *
 * @example
 * ```tsx
 * <AnimateListPresence emptyState={<EmptyState />}>
 *   {filteredItems.map(item => (
 *     <AnimatedListItem key={item.id}>{item.name}</AnimatedListItem>
 *   ))}
 * </AnimateListPresence>
 * ```
 */
function AnimateListPresence({
  listKey,
  children,
  emptyState,
}: AnimateListPresenceProps) {
  const childCount = React.Children.count(children);

  return (
    <AnimatePresence mode="wait">
      {childCount === 0 && emptyState ? (
        <motion.div
          key="empty"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
        >
          {emptyState}
        </motion.div>
      ) : (
        <motion.div key={listKey}>{children}</motion.div>
      )}
    </AnimatePresence>
  );
}

export {
  StaggeredList,
  StaggeredGrid,
  AnimatedListItem,
  AnimateListPresence,
};
