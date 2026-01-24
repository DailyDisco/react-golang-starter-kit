import * as React from "react";

import { cn } from "@/lib/utils";

import { CopyButton } from "../copy-button";

interface CopyableCellProps {
  /** The value to copy to clipboard */
  value: string | number;
  /** Optional display value (if different from copy value) */
  displayValue?: React.ReactNode;
  /** Custom className for the container */
  className?: string;
  /** Whether to use monospace font (default: true for IDs) */
  mono?: boolean;
  /** Truncate long values to this max width */
  maxWidth?: string;
  /** Custom copy button label */
  copyLabel?: string;
}

/**
 * A table cell component that shows a copy button on hover.
 *
 * Use this for ID columns, API keys, or other values users might want to copy.
 *
 * @example
 * // In DataTable column definition
 * {
 *   accessorKey: "id",
 *   header: "ID",
 *   cell: ({ row }) => <CopyableCell value={row.original.id} />,
 * }
 *
 * @example
 * // With custom display
 * {
 *   accessorKey: "apiKey",
 *   header: "API Key",
 *   cell: ({ row }) => (
 *     <CopyableCell
 *       value={row.original.apiKey}
 *       displayValue={`${row.original.apiKey.slice(0, 8)}...`}
 *       copyLabel="Copy API key"
 *     />
 *   ),
 * }
 */
function CopyableCell({
  value,
  displayValue,
  className,
  mono = true,
  maxWidth,
  copyLabel,
}: CopyableCellProps) {
  const stringValue = String(value);

  return (
    <div
      className={cn(
        "flex items-center gap-1 group",
        className
      )}
    >
      <span
        className={cn(
          "text-sm",
          mono && "font-mono",
          maxWidth && "truncate"
        )}
        style={maxWidth ? { maxWidth } : undefined}
        title={maxWidth ? stringValue : undefined}
      >
        {displayValue ?? stringValue}
      </span>
      <CopyButton
        value={stringValue}
        label={copyLabel}
        className="opacity-0 group-hover:opacity-100 group-focus-within:opacity-100 transition-opacity h-6 w-6"
        size="sm"
        variant="ghost"
      />
    </div>
  );
}

/**
 * Shorthand for ID columns - common use case
 */
function IdCell({ id }: { id: string | number }) {
  return <CopyableCell value={id} copyLabel="Copy ID" />;
}

/**
 * Shorthand for truncated values like API keys
 */
function TruncatedCopyableCell({
  value,
  visibleChars = 8,
  copyLabel,
}: {
  value: string;
  visibleChars?: number;
  copyLabel?: string;
}) {
  const truncated = value.length > visibleChars * 2
    ? `${value.slice(0, visibleChars)}...${value.slice(-visibleChars)}`
    : value;

  return (
    <CopyableCell
      value={value}
      displayValue={truncated}
      copyLabel={copyLabel}
    />
  );
}

export { CopyableCell, IdCell, TruncatedCopyableCell };
export type { CopyableCellProps };
