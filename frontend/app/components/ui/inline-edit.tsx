import { useCallback, useEffect, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { Check, Pencil, X } from "lucide-react";

import { Button } from "./button";
import { Input } from "./input";

interface InlineEditProps {
  /** Current value */
  value: string;
  /** Callback when value is saved */
  onSave: (value: string) => void | Promise<void>;
  /** Callback when edit is cancelled */
  onCancel?: () => void;
  /** Placeholder text when empty */
  placeholder?: string;
  /** Whether the field is currently saving */
  isSaving?: boolean;
  /** Validation function */
  validate?: (value: string) => string | null;
  /** CSS class for the display text */
  className?: string;
  /** CSS class for the input */
  inputClassName?: string;
  /** Render custom display */
  renderDisplay?: (value: string) => React.ReactNode;
  /** Make the whole cell clickable to edit */
  clickToEdit?: boolean;
  /** Show edit icon on hover */
  showEditIcon?: boolean;
}

/**
 * Inline edit component for editing text directly in place
 *
 * @example
 * <InlineEdit
 *   value={user.name}
 *   onSave={(value) => updateUser({ name: value })}
 *   placeholder="Enter name"
 * />
 */
export function InlineEdit({
  value,
  onSave,
  onCancel,
  placeholder = "Click to edit",
  isSaving = false,
  validate,
  className,
  inputClassName,
  renderDisplay,
  clickToEdit = true,
  showEditIcon = true,
}: InlineEditProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Sync edit value when value prop changes
  useEffect(() => {
    if (!isEditing) {
      setEditValue(value);
    }
  }, [value, isEditing]);

  // Focus input when entering edit mode
  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleStartEdit = useCallback(() => {
    setIsEditing(true);
    setEditValue(value);
    setError(null);
  }, [value]);

  const handleCancel = useCallback(() => {
    setIsEditing(false);
    setEditValue(value);
    setError(null);
    onCancel?.();
  }, [value, onCancel]);

  const handleSave = useCallback(async () => {
    // Validate if validation function provided
    if (validate) {
      const validationError = validate(editValue);
      if (validationError) {
        setError(validationError);
        return;
      }
    }

    // Don't save if value hasn't changed
    if (editValue === value) {
      setIsEditing(false);
      return;
    }

    try {
      await onSave(editValue);
      setIsEditing(false);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save");
    }
  }, [editValue, value, validate, onSave]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        e.preventDefault();
        handleSave();
      } else if (e.key === "Escape") {
        e.preventDefault();
        handleCancel();
      }
    },
    [handleSave, handleCancel]
  );

  if (isEditing) {
    return (
      <div className="flex items-center gap-1">
        <div className="flex-1">
          <Input
            ref={inputRef}
            value={editValue}
            onChange={(e) => {
              setEditValue(e.target.value);
              setError(null);
            }}
            onKeyDown={handleKeyDown}
            onBlur={() => {
              // Small delay to allow button clicks to register
              setTimeout(() => {
                if (document.activeElement !== inputRef.current) {
                  handleCancel();
                }
              }, 150);
            }}
            disabled={isSaving}
            className={cn("h-8", error && "border-destructive", inputClassName)}
            aria-invalid={!!error}
            aria-describedby={error ? "inline-edit-error" : undefined}
          />
          {error && (
            <p
              id="inline-edit-error"
              className="text-destructive mt-1 text-xs"
            >
              {error}
            </p>
          )}
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 text-green-600 hover:text-green-700"
          onClick={handleSave}
          disabled={isSaving}
        >
          {isSaving ? (
            <div className="border-current h-4 w-4 animate-spin rounded-full border-2 border-t-transparent" />
          ) : (
            <Check className="h-4 w-4" />
          )}
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="text-muted-foreground hover:text-foreground h-8 w-8"
          onClick={handleCancel}
          disabled={isSaving}
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    );
  }

  const displayContent = renderDisplay ? renderDisplay(value) : value || <span className="text-muted-foreground">{placeholder}</span>;

  return (
    <div
      className={cn(
        "group flex items-center gap-2",
        clickToEdit && "cursor-pointer rounded px-1 py-0.5 -mx-1 hover:bg-muted/50",
        className
      )}
      onClick={clickToEdit ? handleStartEdit : undefined}
      onKeyDown={clickToEdit ? (e) => e.key === "Enter" && handleStartEdit() : undefined}
      tabIndex={clickToEdit ? 0 : undefined}
      role={clickToEdit ? "button" : undefined}
    >
      <span className="flex-1">{displayContent}</span>
      {showEditIcon && (
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 opacity-0 transition-opacity group-hover:opacity-100"
          onClick={(e) => {
            e.stopPropagation();
            handleStartEdit();
          }}
        >
          <Pencil className="h-3 w-3" />
        </Button>
      )}
    </div>
  );
}

/**
 * Inline edit for select/dropdown values
 */
interface InlineSelectProps<T extends string> {
  value: T;
  options: { value: T; label: string }[];
  onSave: (value: T) => void | Promise<void>;
  isSaving?: boolean;
  className?: string;
  renderValue?: (value: T) => React.ReactNode;
}

export function InlineSelect<T extends string>({
  value,
  options,
  onSave,
  isSaving = false,
  className,
  renderValue,
}: InlineSelectProps<T>) {
  const [isOpen, setIsOpen] = useState(false);

  const handleSelect = async (newValue: T) => {
    if (newValue === value) {
      setIsOpen(false);
      return;
    }

    try {
      await onSave(newValue);
      setIsOpen(false);
    } catch {
      // Error handling done by parent
    }
  };

  const currentOption = options.find((o) => o.value === value);
  const displayContent = renderValue ? renderValue(value) : currentOption?.label || value;

  return (
    <div className={cn("relative", className)}>
      <button
        type="button"
        className={cn(
          "hover:bg-muted/50 flex items-center gap-2 rounded px-2 py-1 text-left transition-colors",
          isOpen && "bg-muted/50"
        )}
        onClick={() => setIsOpen(!isOpen)}
        disabled={isSaving}
      >
        {isSaving ? (
          <div className="border-current h-4 w-4 animate-spin rounded-full border-2 border-t-transparent" />
        ) : (
          displayContent
        )}
      </button>

      {isOpen && (
        <>
          <div
            role="button"
            tabIndex={0}
            aria-label="Close dropdown"
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " " || e.key === "Escape") {
                setIsOpen(false);
              }
            }}
          />
          <div className="bg-popover absolute top-full left-0 z-20 mt-1 min-w-[120px] rounded-md border py-1 shadow-md">
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                className={cn(
                  "hover:bg-muted w-full px-3 py-1.5 text-left text-sm",
                  option.value === value && "bg-muted font-medium"
                )}
                onClick={() => handleSelect(option.value)}
              >
                {option.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
