import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { X } from "lucide-react";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface TourStep {
  /** Unique identifier for the step */
  id: string;
  /** Target element selector (e.g., "[data-tour='dashboard']") */
  target: string;
  /** Title of the step */
  title: string;
  /** Description/content of the step */
  content: string;
  /** Position of the tooltip relative to target */
  placement?: "top" | "bottom" | "left" | "right";
  /** Optional action button */
  action?: {
    label: string;
    onClick: () => void;
  };
  /** Whether to highlight the target element */
  spotlight?: boolean;
}

interface OnboardingState {
  /** Whether the tour has been completed */
  completedTours: string[];
  /** Currently active tour */
  activeTour: string | null;
  /** Current step index */
  currentStep: number;
  /** Mark a tour as completed */
  completeTour: (tourId: string) => void;
  /** Start a tour */
  startTour: (tourId: string) => void;
  /** End the current tour */
  endTour: () => void;
  /** Go to next step */
  nextStep: () => void;
  /** Go to previous step */
  prevStep: () => void;
  /** Go to specific step */
  goToStep: (step: number) => void;
  /** Reset all tours (for testing) */
  resetTours: () => void;
}

export const useOnboardingStore = create<OnboardingState>()(
  persist(
    (set) => ({
      completedTours: [],
      activeTour: null,
      currentStep: 0,

      completeTour: (tourId) =>
        set((state) => ({
          completedTours: [...new Set([...state.completedTours, tourId])],
          activeTour: null,
          currentStep: 0,
        })),

      startTour: (tourId) =>
        set({
          activeTour: tourId,
          currentStep: 0,
        }),

      endTour: () =>
        set({
          activeTour: null,
          currentStep: 0,
        }),

      nextStep: () =>
        set((state) => ({
          currentStep: state.currentStep + 1,
        })),

      prevStep: () =>
        set((state) => ({
          currentStep: Math.max(0, state.currentStep - 1),
        })),

      goToStep: (step) =>
        set({
          currentStep: step,
        }),

      resetTours: () =>
        set({
          completedTours: [],
          activeTour: null,
          currentStep: 0,
        }),
    }),
    {
      name: "onboarding-storage",
    }
  )
);

interface OnboardingTourProps {
  /** Unique ID for this tour */
  tourId: string;
  /** Steps in the tour */
  steps: TourStep[];
  /** Auto-start if not completed */
  autoStart?: boolean;
  /** Callback when tour completes */
  onComplete?: () => void;
  /** Callback when tour is skipped */
  onSkip?: () => void;
}

/**
 * Onboarding tour component that guides users through features
 *
 * @example
 * <OnboardingTour
 *   tourId="dashboard-intro"
 *   steps={[
 *     { id: "welcome", target: "[data-tour='dashboard']", title: "Welcome!", content: "..." },
 *     { id: "sidebar", target: "[data-tour='sidebar']", title: "Navigation", content: "..." },
 *   ]}
 *   autoStart
 * />
 */
export function OnboardingTour({ tourId, steps, autoStart = false, onComplete, onSkip }: OnboardingTourProps) {
  const { completedTours, activeTour, currentStep, startTour, endTour, completeTour, nextStep, prevStep } =
    useOnboardingStore();

  const [targetRect, setTargetRect] = useState<DOMRect | null>(null);
  const isActive = activeTour === tourId;
  const hasCompleted = completedTours.includes(tourId);
  const currentStepData = steps[currentStep];
  const isLastStep = currentStep === steps.length - 1;

  // Auto-start tour if enabled and not completed
  useEffect(() => {
    if (autoStart && !hasCompleted && !activeTour) {
      // Small delay to ensure DOM is ready
      const timer = setTimeout(() => startTour(tourId), 500);
      return () => clearTimeout(timer);
    }
  }, [autoStart, hasCompleted, activeTour, tourId, startTour]);

  // Find and track target element
  useEffect(() => {
    if (!isActive || !currentStepData) return;

    const findTarget = () => {
      const target = document.querySelector(currentStepData.target);
      if (target) {
        setTargetRect(target.getBoundingClientRect());

        // Scroll target into view if needed
        target.scrollIntoView({ behavior: "smooth", block: "center" });
      } else {
        setTargetRect(null);
      }
    };

    findTarget();

    // Re-find on resize/scroll
    const handleResize = () => findTarget();
    window.addEventListener("resize", handleResize);
    window.addEventListener("scroll", handleResize, true);

    return () => {
      window.removeEventListener("resize", handleResize);
      window.removeEventListener("scroll", handleResize, true);
    };
  }, [isActive, currentStepData]);

  const handleNext = useCallback(() => {
    if (isLastStep) {
      completeTour(tourId);
      onComplete?.();
    } else {
      nextStep();
    }
  }, [isLastStep, completeTour, tourId, onComplete, nextStep]);

  const handleSkip = useCallback(() => {
    endTour();
    onSkip?.();
  }, [endTour, onSkip]);

  const handleComplete = useCallback(() => {
    completeTour(tourId);
    onComplete?.();
  }, [completeTour, tourId, onComplete]);

  if (!isActive || !currentStepData) return null;

  // Calculate tooltip position
  const getTooltipPosition = () => {
    if (!targetRect) {
      return { top: "50%", left: "50%", transform: "translate(-50%, -50%)" };
    }

    const padding = 16;
    const tooltipWidth = 320;
    const tooltipHeight = 200;

    const placement = currentStepData.placement || "bottom";

    switch (placement) {
      case "top":
        return {
          top: `${targetRect.top - tooltipHeight - padding}px`,
          left: `${targetRect.left + targetRect.width / 2 - tooltipWidth / 2}px`,
        };
      case "bottom":
        return {
          top: `${targetRect.bottom + padding}px`,
          left: `${targetRect.left + targetRect.width / 2 - tooltipWidth / 2}px`,
        };
      case "left":
        return {
          top: `${targetRect.top + targetRect.height / 2 - tooltipHeight / 2}px`,
          left: `${targetRect.left - tooltipWidth - padding}px`,
        };
      case "right":
        return {
          top: `${targetRect.top + targetRect.height / 2 - tooltipHeight / 2}px`,
          left: `${targetRect.right + padding}px`,
        };
      default:
        return {
          top: `${targetRect.bottom + padding}px`,
          left: `${targetRect.left + targetRect.width / 2 - tooltipWidth / 2}px`,
        };
    }
  };

  const tooltipPosition = getTooltipPosition();

  return (
    <>
      {/* Backdrop overlay */}
      <div
        role="button"
        tabIndex={0}
        aria-label="Skip tour"
        className="fixed inset-0 z-50 bg-black/50"
        onClick={handleSkip}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " " || e.key === "Escape") {
            handleSkip();
          }
        }}
      />

      {/* Spotlight on target */}
      {currentStepData.spotlight !== false && targetRect && (
        <div
          className="ring-primary/50 pointer-events-none fixed z-50 rounded-lg ring-4"
          style={{
            top: targetRect.top - 4,
            left: targetRect.left - 4,
            width: targetRect.width + 8,
            height: targetRect.height + 8,
            boxShadow: "0 0 0 9999px rgba(0, 0, 0, 0.5)",
          }}
        />
      )}

      {/* Tooltip card */}
      <Card
        className="animate-in fade-in slide-in-from-bottom-4 fixed z-50 w-80"
        style={tooltipPosition}
      >
        <CardHeader className="pb-2">
          <div className="flex items-start justify-between">
            <CardTitle className="text-base">{currentStepData.title}</CardTitle>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={handleSkip}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="pb-4">
          <p className="text-muted-foreground text-sm">{currentStepData.content}</p>
          {currentStepData.action && (
            <Button
              variant="link"
              size="sm"
              className="mt-2 h-auto p-0"
              onClick={currentStepData.action.onClick}
            >
              {currentStepData.action.label}
            </Button>
          )}
        </CardContent>
        <CardFooter className="flex items-center justify-between pt-0">
          {/* Progress dots */}
          <div className="flex gap-1">
            {steps.map((_, index) => (
              <div
                key={index}
                className={cn(
                  "h-1.5 w-1.5 rounded-full transition-colors",
                  index === currentStep ? "bg-primary" : "bg-muted"
                )}
              />
            ))}
          </div>

          {/* Navigation */}
          <div className="flex gap-2">
            {currentStep > 0 && (
              <Button
                variant="outline"
                size="sm"
                onClick={prevStep}
              >
                Back
              </Button>
            )}
            <Button
              size="sm"
              onClick={handleNext}
            >
              {isLastStep ? "Finish" : "Next"}
            </Button>
          </div>
        </CardFooter>
      </Card>
    </>
  );
}

/**
 * Hook to trigger onboarding tour programmatically
 */
export function useOnboarding(tourId: string) {
  const { completedTours, activeTour, startTour, endTour, completeTour, resetTours } = useOnboardingStore();

  return {
    isCompleted: completedTours.includes(tourId),
    isActive: activeTour === tourId,
    start: () => startTour(tourId),
    end: endTour,
    complete: () => completeTour(tourId),
    reset: resetTours,
  };
}

/**
 * Helper to add data-tour attribute to elements
 */
export function tourTarget(id: string) {
  return { "data-tour": id };
}
