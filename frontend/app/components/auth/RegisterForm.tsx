import { useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useNavigate } from "@tanstack/react-router";
import { Eye, EyeOff, Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";

import { useRegister } from "../../hooks/mutations/use-auth-mutations";
import { ApiError } from "../../services/api/client";
import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { AuthErrorGuidance } from "./AuthErrorGuidance";
import { OAuthButtons, OAuthDivider } from "./OAuthButtons";
import { PasswordStrengthMeter } from "./PasswordStrengthMeter";

type RegisterFormData = {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
};

export function RegisterForm() {
  const { t } = useTranslation(["auth", "validation", "common"]);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const registerMutation = useRegister();
  const navigate = useNavigate();

  // Create schema with translated messages
  const registerSchema = useMemo(
    () =>
      z
        .object({
          name: z.string().min(2, t("validation:name.minLength")),
          email: z.string().email(t("validation:email.invalid")),
          password: z.string().min(8, t("validation:password.minLength")),
          confirmPassword: z.string(),
        })
        .refine((data) => data.password === data.confirmPassword, {
          message: t("validation:password.mismatch"),
          path: ["confirmPassword"],
        }),
    [t]
  );

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  const passwordValue = watch("password", "");

  const onSubmit = (data: RegisterFormData) => {
    setError(null);
    registerMutation.mutate(
      {
        name: data.name,
        email: data.email,
        password: data.password,
      },
      {
        onSuccess: () => {
          toast.success(t("auth:register.success"), {
            description: t("auth:register.successDescription"),
          });
          void navigate({ to: "/", search: undefined });
        },
        onError: (err) => {
          setError(err);
          const errorMessage = err instanceof Error ? err.message : t("auth:register.error");
          toast.error(t("auth:register.error"), {
            description: errorMessage,
          });
        },
      }
    );
  };

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-center text-2xl">{t("auth:register.title")}</CardTitle>
          <CardDescription className="text-center">{t("auth:register.subtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          <OAuthButtons
            mode="register"
            disabled={registerMutation.isPending}
            onError={(message) => setError(new Error(message))}
          />
          <OAuthDivider text={t("auth:oauth.registerWithEmail")} />
          <form
            onSubmit={handleSubmit(onSubmit)}
            className="space-y-4"
          >
            {error && (
              <>
                <Alert variant="destructive">
                  <AlertDescription>{error.message}</AlertDescription>
                </Alert>
                <AuthErrorGuidance
                  error={error}
                  context="register"
                />
              </>
            )}

            <div className="space-y-2">
              <Label htmlFor="name">{t("common:labels.fullName")}</Label>
              <Input
                id="name"
                type="text"
                placeholder={t("auth:register.namePlaceholder")}
                autoFocus
                {...register("name")}
                disabled={registerMutation.isPending}
              />
              {errors.name && <p className="text-sm text-red-500">{errors.name.message}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">{t("common:labels.email")}</Label>
              <Input
                id="email"
                type="email"
                placeholder={t("auth:register.emailPlaceholder")}
                {...register("email")}
                disabled={registerMutation.isPending}
              />
              {errors.email && <p className="text-sm text-red-500">{errors.email.message}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">{t("common:labels.password")}</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder={t("auth:register.passwordPlaceholder")}
                  {...register("password")}
                  disabled={registerMutation.isPending}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                  onClick={() => setShowPassword(!showPassword)}
                  disabled={registerMutation.isPending}
                  aria-label={showPassword ? t("auth:login.hidePassword") : t("auth:login.showPassword")}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              {errors.password && <p className="text-sm text-red-500">{errors.password.message}</p>}
              <PasswordStrengthMeter password={passwordValue} />
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirmPassword">{t("common:labels.confirmPassword")}</Label>
              <div className="relative">
                <Input
                  id="confirmPassword"
                  type={showConfirmPassword ? "text" : "password"}
                  placeholder={t("auth:register.confirmPasswordPlaceholder")}
                  {...register("confirmPassword")}
                  disabled={registerMutation.isPending}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  disabled={registerMutation.isPending}
                  aria-label={showConfirmPassword ? t("auth:login.hidePassword") : t("auth:login.showPassword")}
                >
                  {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              {errors.confirmPassword && <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>}
            </div>

            <Button
              type="submit"
              className="w-full"
              disabled={registerMutation.isPending}
            >
              {registerMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {t("auth:register.submitButton")}
            </Button>
          </form>

          <div className="mt-4 text-center text-sm">
            {t("auth:register.hasAccount")}{" "}
            <Link
              to="/login"
              className="text-primary hover:underline"
            >
              {t("auth:register.signInLink")}
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
