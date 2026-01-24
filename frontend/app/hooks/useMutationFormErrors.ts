import { useEffect } from "react";

import type { UseMutationResult } from "@tanstack/react-query";
import type { FieldValues, Path, UseFormReturn } from "react-hook-form";

import { categorizeError } from "../lib/error-utils";
import type { ApiError } from "../services/api/client";

interface FieldError {
  field: string;
  message: string;
}

interface MutationFormErrorsOptions<T extends FieldValues> {
  /** Map API field names to form field names */
  fieldMapping?: Partial<Record<string, Path<T>>>;
  /** Clear server errors when user starts typing (default: true) */
  clearOnChange?: boolean;
}

/**
 * Consolidate mutation errors into react-hook-form field errors.
 *
 * This hook maps API errors to form fields, avoiding duplicate error messages
 * (one in toast, one in form). When an API returns a validation error like
 * "email: already exists", it sets that error on the email field.
 *
 * Features:
 * - Parses field-level errors from API responses
 * - Sets errors on form fields using react-hook-form's setError
 * - Falls back to root error for non-field errors
 * - Clears server errors when user changes the field
 *
 * @example
 * function RegisterForm() {
 *   const form = useForm<RegisterFormData>();
 *   const registerMutation = useRegister();
 *
 *   useMutationFormErrors(form, registerMutation, {
 *     fieldMapping: {
 *       email: "email",
 *       password: "password",
 *     },
 *   });
 *
 *   return (
 *     <form onSubmit={form.handleSubmit(onSubmit)}>
 *       {form.formState.errors.root && (
 *         <Alert variant="destructive">
 *           {form.formState.errors.root.message}
 *         </Alert>
 *       )}
 *       <Input {...form.register("email")} />
 *       {form.formState.errors.email && (
 *         <span>{form.formState.errors.email.message}</span>
 *       )}
 *     </form>
 *   );
 * }
 */
export function useMutationFormErrors<T extends FieldValues, TData = unknown, TVariables = unknown>(
  form: UseFormReturn<T>,
  mutation: UseMutationResult<TData, Error | ApiError, TVariables>,
  options?: MutationFormErrorsOptions<T>
): void {
  const { fieldMapping = {}, clearOnChange = true } = options ?? {};

  // Set errors when mutation fails
  useEffect(() => {
    if (!mutation.error) return;

    const categorized = categorizeError(mutation.error);

    // For validation errors, try to extract field-level errors
    if (categorized.category === "validation") {
      const fieldErrors = parseFieldErrors(mutation.error);

      if (fieldErrors.length > 0) {
        let hasSetFieldError = false;

        for (const { field, message } of fieldErrors) {
          // Look up the mapped field name, or use the API field name directly
          const formField = (fieldMapping[field] ?? field) as Path<T>;

          // Only set if the field exists in the form
          try {
            form.setError(formField, {
              type: "server",
              message,
            });
            hasSetFieldError = true;
          } catch {
            // Field doesn't exist in form, will fall through to root error
          }
        }

        // If we set at least one field error, don't also set root error
        if (hasSetFieldError) {
          return;
        }
      }
    }

    // For non-field errors or if no field errors were extracted, set root error
    form.setError("root", {
      type: "server",
      message: categorized.message,
    });
  }, [mutation.error, form, fieldMapping]);

  // Clear server errors when user changes fields
  useEffect(() => {
    if (!clearOnChange) return;

    const subscription = form.watch((_, { name }) => {
      // Clear specific field's server error if it exists
      if (name) {
        const fieldError = form.formState.errors[name as keyof typeof form.formState.errors];
        if (fieldError && (fieldError as { type?: string }).type === "server") {
          form.clearErrors(name as Path<T>);
        }
      }

      // Clear root error if it's a server error
      if (form.formState.errors.root?.type === "server") {
        form.clearErrors("root");
      }
    });

    return () => subscription.unsubscribe();
  }, [form, clearOnChange]);
}

/**
 * Parse field-level errors from an API error.
 *
 * Supports common patterns:
 * - "field: message" format
 * - JSON error details with field property
 */
function parseFieldErrors(error: Error | ApiError): FieldError[] {
  const errors: FieldError[] = [];

  // Try to parse "field: message" format from error message
  // e.g., "email: already exists" or "password: must be at least 8 characters"
  const fieldMessageMatch = error.message?.match(/^(\w+):\s*(.+)$/);
  if (fieldMessageMatch) {
    errors.push({
      field: fieldMessageMatch[1].toLowerCase(),
      message: fieldMessageMatch[2],
    });
    return errors;
  }

  // Try common field-specific message patterns
  const message = error.message?.toLowerCase() || "";

  // Email-related errors
  if (message.includes("email") && (message.includes("exists") || message.includes("taken"))) {
    errors.push({ field: "email", message: "This email is already registered" });
    return errors;
  }

  // Password-related errors
  if (message.includes("password")) {
    if (message.includes("uppercase")) {
      errors.push({
        field: "password",
        message: "Password must contain at least one uppercase letter",
      });
      return errors;
    }
    if (message.includes("8 character") || message.includes("too short")) {
      errors.push({
        field: "password",
        message: "Password must be at least 8 characters",
      });
      return errors;
    }
    if (message.includes("number") || message.includes("digit")) {
      errors.push({
        field: "password",
        message: "Password must contain at least one number",
      });
      return errors;
    }
    if (message.includes("special")) {
      errors.push({
        field: "password",
        message: "Password must contain at least one special character",
      });
      return errors;
    }
  }

  // Username-related errors
  if (message.includes("username") && (message.includes("exists") || message.includes("taken"))) {
    errors.push({ field: "username", message: "This username is already taken" });
    return errors;
  }

  return errors;
}

export type { FieldError, MutationFormErrorsOptions };
