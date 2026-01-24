import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useNavigate } from "@tanstack/react-router";
import { AlertTriangle, ArrowRight, X } from "lucide-react";

interface UsageLimitExceededDetail {
  alertType: string;
  usageType: string;
  currentUsage: number;
  limit: number;
  percentageUsed: number;
  message: string;
  canUpgrade: boolean;
  currentPlan?: string;
  suggestedPlan?: string;
  upgradeUrl?: string;
}

interface UpgradePromptProps {
  className?: string;
}

/**
 * UpgradePrompt listens for usage-limit-exceeded events and displays
 * a banner prompting the user to upgrade their plan.
 */
export function UpgradePrompt({ className }: UpgradePromptProps) {
  const navigate = useNavigate();
  const [alert, setAlert] = useState<UsageLimitExceededDetail | null>(null);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    const handleUsageLimitExceeded = (event: CustomEvent<UsageLimitExceededDetail>) => {
      setAlert(event.detail);
      setDismissed(false);
    };

    window.addEventListener("usage-limit-exceeded", handleUsageLimitExceeded as EventListener);

    return () => {
      window.removeEventListener("usage-limit-exceeded", handleUsageLimitExceeded as EventListener);
    };
  }, []);

  if (!alert || dismissed || !alert.canUpgrade) {
    return null;
  }

  const handleUpgrade = () => {
    const url = alert.upgradeUrl || "/settings/billing";
    const searchParams = alert.suggestedPlan ? `?plan=${encodeURIComponent(alert.suggestedPlan.toLowerCase())}` : "";
    navigate({ to: `${url}${searchParams}` });
  };

  const handleDismiss = () => {
    setDismissed(true);
  };

  const usageTypeLabels: Record<string, string> = {
    api_call: "API calls",
    storage: "storage",
    compute: "compute time",
    file_upload: "file uploads",
  };

  const usageLabel = usageTypeLabels[alert.usageType] || alert.usageType;

  return (
    <div
      className={cn(
        "flex items-center gap-4 rounded-lg border border-amber-200 bg-amber-50 p-4 shadow-sm dark:border-amber-800 dark:bg-amber-950/30",
        className
      )}
      role="alert"
    >
      <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/50">
        <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400" />
      </div>

      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
          You&apos;ve exceeded your {usageLabel} limit
        </p>
        <p className="mt-0.5 text-sm text-amber-600 dark:text-amber-400">
          Upgrade to {alert.suggestedPlan} for higher limits and more features
        </p>
      </div>

      <div className="flex flex-shrink-0 items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={handleDismiss}
          className="text-amber-600 hover:bg-amber-100 hover:text-amber-700 dark:text-amber-400 dark:hover:bg-amber-900/50 dark:hover:text-amber-300"
        >
          Later
        </Button>
        <Button
          size="sm"
          onClick={handleUpgrade}
          className="bg-amber-600 text-white hover:bg-amber-700"
        >
          View Plans
          <ArrowRight className="ml-1 h-4 w-4" />
        </Button>
        <button
          onClick={handleDismiss}
          className="p-1 text-amber-400 hover:text-amber-600 dark:hover:text-amber-300"
          aria-label="Dismiss"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
