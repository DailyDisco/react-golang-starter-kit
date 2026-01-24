import { z } from "zod";

/**
 * Shared validation schemas for authentication forms.
 * Use these with react-hook-form and zodResolver.
 *
 * @example
 * const schema = z.object({
 *   email: emailSchema(t),
 *   password: passwordSchema(t, { minLength: 8 }),
 * });
 */

type TranslationFn = (key: string, defaultValue?: string) => string;

/**
 * Email validation schema.
 * Uses i18n for error messages.
 */
export function emailSchema(t: TranslationFn) {
  return z.string().email(t("validation:email.invalid", "Invalid email address"));
}

/**
 * Password validation schema with configurable minimum length.
 * For login forms, use minLength: 1.
 * For registration, use minLength: 8.
 */
export function passwordSchema(t: TranslationFn, options: { minLength?: number } = {}) {
  const minLength = options.minLength ?? 8;

  if (minLength === 1) {
    return z.string().min(1, t("validation:password.required", "Password is required"));
  }

  return z
    .string()
    .min(minLength, t("validation:password.minLength", `Password must be at least ${minLength} characters`));
}

/**
 * Name validation schema.
 */
export function nameSchema(t: TranslationFn) {
  return z.string().min(2, t("validation:name.minLength", "Name must be at least 2 characters"));
}

/**
 * Create a login form schema.
 */
export function createLoginSchema(t: TranslationFn) {
  return z.object({
    email: emailSchema(t),
    password: passwordSchema(t, { minLength: 1 }),
  });
}

/**
 * Create a registration form schema with password confirmation.
 */
export function createRegisterSchema(t: TranslationFn) {
  return z
    .object({
      name: nameSchema(t),
      email: emailSchema(t),
      password: passwordSchema(t, { minLength: 8 }),
      confirmPassword: z.string(),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t("validation:password.mismatch", "Passwords do not match"),
      path: ["confirmPassword"],
    });
}

export type LoginFormData = z.infer<ReturnType<typeof createLoginSchema>>;
export type RegisterFormData = z.infer<ReturnType<typeof createRegisterSchema>>;
