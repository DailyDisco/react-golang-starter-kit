import { Check, Copy } from "lucide-react";
import * as React from "react";

import { logger } from "@/lib/logger";
import { cn } from "@/lib/utils";

import { Button, type buttonVariants } from "./button";
import { Tooltip, TooltipContent, TooltipTrigger } from "./tooltip";

import type { VariantProps } from "class-variance-authority";

interface CopyButtonProps
  extends Omit<React.ComponentProps<"button">, "onClick">,
    VariantProps<typeof buttonVariants> {
  /** The text value to copy to clipboard */
  value: string;
  /** Callback fired after successful copy */
  onCopy?: () => void;
  /** Duration in ms to show success state (default: 2000) */
  successDuration?: number;
  /** Tooltip text for copy action (default: "Copy to clipboard") */
  label?: string;
  /** Tooltip text after copy (default: "Copied!") */
  successLabel?: string;
}

/**
 * A button that copies text to clipboard with visual feedback.
 *
 * Features:
 * - Shows Copy icon, changes to Check icon on success
 * - Tooltip shows current state
 * - Keyboard accessible
 * - Auto-reverts after configurable duration
 *
 * @example
 * <CopyButton value={user.id} />
 *
 * @example
 * <CopyButton
 *   value={apiKey}
 *   label="Copy API key"
 *   successLabel="API key copied!"
 *   onCopy={() => trackEvent("api_key_copied")}
 * />
 */
function CopyButton({
  value,
  onCopy,
  successDuration = 2000,
  label = "Copy to clipboard",
  successLabel = "Copied!",
  className,
  variant = "ghost",
  size = "icon",
  ...props
}: CopyButtonProps) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = React.useCallback(async () => {
    if (!value) return;

    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      onCopy?.();

      setTimeout(() => {
        setCopied(false);
      }, successDuration);
    } catch (err) {
      logger.error("Failed to copy to clipboard", err);
    }
  }, [value, onCopy, successDuration]);

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant={variant}
          size={size}
          className={cn("h-8 w-8", className)}
          onClick={handleCopy}
          aria-label={copied ? successLabel : label}
          data-copied={copied}
          {...props}
        >
          {copied ? (
            <Check
              className="h-4 w-4 text-green-500"
              aria-hidden="true"
            />
          ) : (
            <Copy className="h-4 w-4" aria-hidden="true" />
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent>
        <p>{copied ? successLabel : label}</p>
      </TooltipContent>
    </Tooltip>
  );
}

export { CopyButton };
export type { CopyButtonProps };
