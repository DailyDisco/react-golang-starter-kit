/**
 * Animation utilities and Framer Motion variants
 * Use these for consistent animations across the application
 */

import type { Transition, Variants } from "framer-motion";

// ==========================================================================
// DURATION CONSTANTS
// ==========================================================================

export const duration = {
  fast: 0.15,
  normal: 0.3,
  slow: 0.5,
} as const;

// ==========================================================================
// EASING CURVES
// ==========================================================================

export const easing = {
  /** Standard ease-out for most transitions */
  smooth: [0.25, 0.1, 0.25, 1] as const,
  /** Bouncy easing for playful interactions */
  bounce: [0.68, -0.55, 0.265, 1.55] as const,
  /** Ease-in-out for symmetric animations */
  inOut: [0.4, 0, 0.2, 1] as const,
} as const;

// ==========================================================================
// BASIC FADE VARIANTS
// ==========================================================================

/** Simple fade in/out */
export const fadeIn: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
};

/** Fade with upward movement - great for content appearing */
export const fadeInUp: Variants = {
  initial: { opacity: 0, y: 20 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: duration.normal, ease: easing.smooth },
  },
  exit: { opacity: 0, y: -10 },
};

/** Fade with downward movement */
export const fadeInDown: Variants = {
  initial: { opacity: 0, y: -20 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: duration.normal, ease: easing.smooth },
  },
  exit: { opacity: 0, y: 10 },
};

/** Fade with rightward movement - for slide-in panels */
export const fadeInRight: Variants = {
  initial: { opacity: 0, x: 20 },
  animate: {
    opacity: 1,
    x: 0,
    transition: { duration: duration.normal, ease: easing.smooth },
  },
  exit: { opacity: 0, x: -20 },
};

/** Fade with leftward movement */
export const fadeInLeft: Variants = {
  initial: { opacity: 0, x: -20 },
  animate: {
    opacity: 1,
    x: 0,
    transition: { duration: duration.normal, ease: easing.smooth },
  },
  exit: { opacity: 0, x: 20 },
};

// ==========================================================================
// SCALE VARIANTS
// ==========================================================================

/** Scale in from smaller size - good for modals, popovers */
export const scaleIn: Variants = {
  initial: { scale: 0.95, opacity: 0 },
  animate: {
    scale: 1,
    opacity: 1,
    transition: { duration: duration.fast, ease: easing.smooth },
  },
  exit: { scale: 0.95, opacity: 0 },
};

/** Pop in with slight overshoot - for attention-grabbing elements */
export const popIn: Variants = {
  initial: { scale: 0.8, opacity: 0 },
  animate: {
    scale: 1,
    opacity: 1,
    transition: { type: "spring", stiffness: 400, damping: 25 },
  },
  exit: { scale: 0.8, opacity: 0 },
};

// ==========================================================================
// STAGGER CONTAINER VARIANTS
// ==========================================================================

/** Container for staggered children animations */
export const staggerContainer: Variants = {
  initial: {},
  animate: {
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.05,
    },
  },
  exit: {
    transition: {
      staggerChildren: 0.05,
      staggerDirection: -1,
    },
  },
};

/** Faster stagger for lists with many items */
export const staggerContainerFast: Variants = {
  initial: {},
  animate: {
    transition: {
      staggerChildren: 0.05,
      delayChildren: 0.02,
    },
  },
};

/** Child item for stagger containers */
export const staggerItem: Variants = {
  initial: { opacity: 0, y: 10 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: duration.normal, ease: easing.smooth },
  },
  exit: { opacity: 0, y: -10 },
};

// ==========================================================================
// PAGE TRANSITION VARIANTS
// ==========================================================================

/** Page transition - subtle fade and slide */
export const pageTransition: Variants = {
  initial: { opacity: 0, y: 8 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: duration.normal, ease: easing.smooth },
  },
  exit: {
    opacity: 0,
    y: -8,
    transition: { duration: duration.fast },
  },
};

// ==========================================================================
// HOVER ANIMATIONS (for whileHover prop)
// ==========================================================================

/** Subtle lift effect on hover */
export const hoverLift = {
  y: -2,
  transition: { duration: duration.fast },
};

/** Scale up slightly on hover */
export const hoverScale = {
  scale: 1.02,
  transition: { duration: duration.fast },
};

/** Glow effect on hover (combine with CSS shadow) */
export const hoverGlow = {
  scale: 1.01,
  transition: { duration: duration.fast },
};

// ==========================================================================
// TAP ANIMATIONS (for whileTap prop)
// ==========================================================================

/** Press down effect */
export const tapScale = {
  scale: 0.98,
};

/** Stronger press for larger buttons */
export const tapScaleStrong = {
  scale: 0.95,
};

// ==========================================================================
// SPRING TRANSITIONS
// ==========================================================================

export const springTransition: Transition = {
  type: "spring",
  stiffness: 400,
  damping: 30,
};

export const springTransitionSoft: Transition = {
  type: "spring",
  stiffness: 200,
  damping: 20,
};

// ==========================================================================
// UTILITY FUNCTIONS
// ==========================================================================

/**
 * Create a delayed variant
 * @param variants - The base variants
 * @param delay - Delay in seconds
 */
export function withDelay(variants: Variants, delay: number): Variants {
  return {
    ...variants,
    animate: {
      ...(typeof variants.animate === "object" ? variants.animate : {}),
      transition: {
        ...(typeof variants.animate === "object" && "transition" in variants.animate
          ? (variants.animate.transition as object)
          : {}),
        delay,
      },
    },
  };
}

/**
 * Create stagger container with custom timing
 * @param staggerDelay - Delay between children
 * @param initialDelay - Delay before first child
 */
export function createStaggerContainer(staggerDelay = 0.1, initialDelay = 0): Variants {
  return {
    initial: {},
    animate: {
      transition: {
        staggerChildren: staggerDelay,
        delayChildren: initialDelay,
      },
    },
  };
}
