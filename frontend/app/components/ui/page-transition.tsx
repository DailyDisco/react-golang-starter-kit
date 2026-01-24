import { useReducedMotion } from '@/lib/animations';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence, type Variants } from 'framer-motion';
import * as React from 'react';

type TransitionVariant = 'fade' | 'fadeUp' | 'fadeDown' | 'slideRight' | 'slideLeft' | 'scale';

/** No-motion variant for accessibility */
const noMotionVariant: Variants = {
  initial: { opacity: 1 },
  enter: { opacity: 1 },
  exit: { opacity: 1 },
};

interface PageTransitionProps {
  /** Content to animate */
  children: React.ReactNode;
  /** Unique key for the page (usually route path) */
  pageKey?: string;
  /** Animation variant */
  variant?: TransitionVariant;
  /** Duration in seconds */
  duration?: number;
  /** Additional className */
  className?: string;
  /** Whether to animate on initial mount */
  animateOnMount?: boolean;
}

const variants: Record<TransitionVariant, Variants> = {
  fade: {
    initial: { opacity: 0 },
    enter: { opacity: 1 },
    exit: { opacity: 0 },
  },
  fadeUp: {
    initial: { opacity: 0, y: 8 },
    enter: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: -8 },
  },
  fadeDown: {
    initial: { opacity: 0, y: -8 },
    enter: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: 8 },
  },
  slideRight: {
    initial: { opacity: 0, x: -20 },
    enter: { opacity: 1, x: 0 },
    exit: { opacity: 0, x: 20 },
  },
  slideLeft: {
    initial: { opacity: 0, x: 20 },
    enter: { opacity: 1, x: 0 },
    exit: { opacity: 0, x: -20 },
  },
  scale: {
    initial: { opacity: 0, scale: 0.98 },
    enter: { opacity: 1, scale: 1 },
    exit: { opacity: 0, scale: 1.02 },
  },
};

/**
 * Wrapper component for page/route transitions
 *
 * @example
 * ```tsx
 * // In your route component
 * function DashboardPage() {
 *   return (
 *     <PageTransition>
 *       <DashboardContent />
 *     </PageTransition>
 *   );
 * }
 * ```
 */
function PageTransition({
  children,
  pageKey,
  variant = 'fadeUp',
  duration = 0.3,
  className,
  animateOnMount = true,
}: PageTransitionProps) {
  const prefersReducedMotion = useReducedMotion();
  const selectedVariant = prefersReducedMotion ? noMotionVariant : variants[variant];
  const transitionDuration = prefersReducedMotion ? 0 : duration;

  return (
    <motion.div
      key={pageKey}
      initial={animateOnMount ? 'initial' : false}
      animate="enter"
      exit="exit"
      variants={selectedVariant}
      transition={{
        duration: transitionDuration,
        ease: [0.25, 0.1, 0.25, 1],
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

interface PageTransitionProviderProps {
  /** Page content with unique key */
  children: React.ReactNode;
  /** Unique key for current page */
  pageKey: string;
  /** Animation variant */
  variant?: TransitionVariant;
  /** Duration in seconds */
  duration?: number;
  /** Animation mode for AnimatePresence */
  mode?: 'wait' | 'sync' | 'popLayout';
  /** Additional className for wrapper */
  className?: string;
}

/**
 * Provider for animating between different pages
 * Use with router's location/pathname as key
 *
 * @example
 * ```tsx
 * // In your layout component
 * const { pathname } = useLocation();
 *
 * return (
 *   <PageTransitionProvider pageKey={pathname}>
 *     <Outlet />
 *   </PageTransitionProvider>
 * );
 * ```
 */
function PageTransitionProvider({
  children,
  pageKey,
  variant = 'fadeUp',
  duration = 0.3,
  mode = 'wait',
  className,
}: PageTransitionProviderProps) {
  return (
    <AnimatePresence mode={mode}>
      <PageTransition
        key={pageKey}
        pageKey={pageKey}
        variant={variant}
        duration={duration}
        className={className}
      >
        {children}
      </PageTransition>
    </AnimatePresence>
  );
}

interface SectionTransitionProps {
  /** Content to animate */
  children: React.ReactNode;
  /** Delay before animation (in seconds) */
  delay?: number;
  /** Duration in seconds */
  duration?: number;
  /** Additional className */
  className?: string;
  /** Animation variant */
  variant?: 'fadeUp' | 'fadeIn' | 'slideIn';
}

/**
 * Transition wrapper for page sections
 * Useful for animating content blocks within a page
 *
 * @example
 * ```tsx
 * <SectionTransition delay={0.2}>
 *   <StatsGrid />
 * </SectionTransition>
 * <SectionTransition delay={0.4}>
 *   <RecentActivity />
 * </SectionTransition>
 * ```
 */
function SectionTransition({
  children,
  delay = 0,
  duration = 0.4,
  className,
  variant = 'fadeUp',
}: SectionTransitionProps) {
  const prefersReducedMotion = useReducedMotion();

  const sectionVariants: Record<string, Variants> = {
    fadeUp: {
      hidden: { opacity: 0, y: 20 },
      visible: { opacity: 1, y: 0 },
    },
    fadeIn: {
      hidden: { opacity: 0 },
      visible: { opacity: 1 },
    },
    slideIn: {
      hidden: { opacity: 0, x: -20 },
      visible: { opacity: 1, x: 0 },
    },
  };

  const noMotionSectionVariant: Variants = {
    hidden: { opacity: 1 },
    visible: { opacity: 1 },
  };

  return (
    <motion.section
      initial="hidden"
      animate="visible"
      variants={prefersReducedMotion ? noMotionSectionVariant : sectionVariants[variant]}
      transition={{
        duration: prefersReducedMotion ? 0 : duration,
        delay: prefersReducedMotion ? 0 : delay,
        ease: [0.25, 0.1, 0.25, 1],
      }}
      className={className}
    >
      {children}
    </motion.section>
  );
}

interface ContentRevealProps {
  /** Content to reveal */
  children: React.ReactNode;
  /** Whether content should be visible */
  isVisible?: boolean;
  /** Additional className */
  className?: string;
}

/**
 * Animated content reveal (show/hide with animation)
 *
 * @example
 * ```tsx
 * <ContentReveal isVisible={showDetails}>
 *   <DetailsPanel />
 * </ContentReveal>
 * ```
 */
function ContentReveal({
  children,
  isVisible = true,
  className,
}: ContentRevealProps) {
  const prefersReducedMotion = useReducedMotion();

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          initial={prefersReducedMotion ? { opacity: 1, height: 'auto' } : { opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={prefersReducedMotion ? { opacity: 1, height: 'auto' } : { opacity: 0, height: 0 }}
          transition={{ duration: prefersReducedMotion ? 0 : 0.3, ease: [0.25, 0.1, 0.25, 1] }}
          className={cn('overflow-hidden', className)}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  );
}

export {
  PageTransition,
  PageTransitionProvider,
  SectionTransition,
  ContentReveal,
};
