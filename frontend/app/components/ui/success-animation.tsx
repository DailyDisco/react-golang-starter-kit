import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";
import { Check, PartyPopper, Sparkles } from "lucide-react";

interface SuccessAnimationProps {
  /** Whether the animation should be visible */
  show: boolean;
  /** Animation variant */
  variant?: "checkmark" | "confetti" | "sparkle";
  /** Duration in ms before auto-hiding (0 = no auto-hide) */
  duration?: number;
  /** Callback when animation completes */
  onComplete?: () => void;
  /** Size of the animation */
  size?: "sm" | "md" | "lg";
  /** Additional CSS classes */
  className?: string;
}

const sizeConfig = {
  sm: {
    container: "h-12 w-12",
    icon: "h-6 w-6",
  },
  md: {
    container: "h-16 w-16",
    icon: "h-8 w-8",
  },
  lg: {
    container: "h-24 w-24",
    icon: "h-12 w-12",
  },
};

/**
 * Success animation component with multiple variants
 */
export function SuccessAnimation({
  show,
  variant = "checkmark",
  duration = 2000,
  onComplete,
  size = "md",
  className,
}: SuccessAnimationProps) {
  const [isVisible, setIsVisible] = useState(show);
  const sizes = sizeConfig[size];

  useEffect(() => {
    if (show) {
      setIsVisible(true);

      if (duration > 0) {
        const timer = setTimeout(() => {
          setIsVisible(false);
          onComplete?.();
        }, duration);
        return () => clearTimeout(timer);
      }
    } else {
      setIsVisible(false);
    }
  }, [show, duration, onComplete]);

  if (!isVisible) return null;

  return (
    <div
      className={cn(
        "pointer-events-none fixed inset-0 z-50 flex items-center justify-center",
        "animate-in fade-in zoom-in duration-300",
        className
      )}
    >
      {variant === "checkmark" && <CheckmarkAnimation sizes={sizes} />}
      {variant === "confetti" && <ConfettiAnimation sizes={sizes} />}
      {variant === "sparkle" && <SparkleAnimation sizes={sizes} />}
    </div>
  );
}

function CheckmarkAnimation({ sizes }: { sizes: (typeof sizeConfig)[keyof typeof sizeConfig] }) {
  return (
    <div
      className={cn(
        "flex items-center justify-center rounded-full bg-green-500 text-white shadow-lg",
        "animate-bounce",
        sizes.container
      )}
    >
      <Check
        className={cn(sizes.icon, "animate-in zoom-in duration-300")}
        strokeWidth={3}
      />
    </div>
  );
}

function ConfettiAnimation({ sizes }: { sizes: (typeof sizeConfig)[keyof typeof sizeConfig] }) {
  return (
    <div className="relative">
      {/* Confetti particles */}
      {Array.from({ length: 12 }).map((_, i) => (
        <div
          key={i}
          className="absolute animate-confetti"
          style={{
            left: "50%",
            top: "50%",
            animationDelay: `${i * 50}ms`,
            transform: `rotate(${i * 30}deg)`,
          }}
        >
          <div
            className={cn(
              "h-2 w-2 rounded-full",
              ["bg-yellow-400", "bg-pink-400", "bg-blue-400", "bg-green-400", "bg-purple-400", "bg-orange-400"][i % 6]
            )}
          />
        </div>
      ))}

      {/* Center icon */}
      <div
        className={cn(
          "flex items-center justify-center rounded-full bg-gradient-to-br from-yellow-400 to-orange-500 text-white shadow-lg",
          "animate-bounce",
          sizes.container
        )}
      >
        <PartyPopper className={sizes.icon} />
      </div>
    </div>
  );
}

function SparkleAnimation({ sizes }: { sizes: (typeof sizeConfig)[keyof typeof sizeConfig] }) {
  return (
    <div className="relative">
      {/* Sparkle rays */}
      {Array.from({ length: 8 }).map((_, i) => (
        <div
          key={i}
          className="absolute left-1/2 top-1/2 animate-sparkle-ray"
          style={{
            transform: `rotate(${i * 45}deg)`,
            animationDelay: `${i * 75}ms`,
          }}
        >
          <div className="h-8 w-0.5 -translate-x-1/2 bg-gradient-to-t from-transparent via-yellow-400 to-transparent" />
        </div>
      ))}

      {/* Center icon */}
      <div
        className={cn(
          "flex items-center justify-center rounded-full bg-gradient-to-br from-purple-500 to-pink-500 text-white shadow-lg",
          "animate-pulse",
          sizes.container
        )}
      >
        <Sparkles className={sizes.icon} />
      </div>
    </div>
  );
}

/**
 * Hook for triggering success animations
 */
export function useSuccessAnimation() {
  const [animationState, setAnimationState] = useState<{
    show: boolean;
    variant: "checkmark" | "confetti" | "sparkle";
  }>({
    show: false,
    variant: "checkmark",
  });

  const triggerSuccess = (variant: "checkmark" | "confetti" | "sparkle" = "checkmark") => {
    setAnimationState({ show: true, variant });
  };

  const hideSuccess = () => {
    setAnimationState((prev) => ({ ...prev, show: false }));
  };

  return {
    ...animationState,
    triggerSuccess,
    hideSuccess,
  };
}

// Add these styles to your global CSS or Tailwind config
// @keyframes confetti {
//   0% { transform: translateY(0) rotate(0deg); opacity: 1; }
//   100% { transform: translateY(-100px) rotate(720deg); opacity: 0; }
// }
// .animate-confetti { animation: confetti 1s ease-out forwards; }
