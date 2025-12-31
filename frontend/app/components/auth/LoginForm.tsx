import { useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useLocation, useNavigate } from "@tanstack/react-router";
import { AlertCircle, Eye, EyeOff } from "lucide-react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";

import { useLogin } from "../../hooks/mutations/use-auth-mutations";
import { ApiError } from "../../services/api/client";
import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { AuthErrorGuidance } from "./AuthErrorGuidance";
import { OAuthButtons, OAuthDivider } from "./OAuthButtons";

type LoginFormData = {
  email: string;
  password: string;
};

export function LoginForm() {
  const { t } = useTranslation(["auth", "validation", "common"]);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const loginMutation = useLogin();
  const navigate = useNavigate();
  const location = useLocation();

  const from = (location.state as { from?: { pathname?: string } } | undefined)?.from?.pathname ?? "/";

  // Create schema with translated messages
  const loginSchema = useMemo(
    () =>
      z.object({
        email: z.string().email(t("validation:email.invalid")),
        password: z.string().min(1, t("validation:password.required")),
      }),
    [t]
  );

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = (data: LoginFormData) => {
    setError(null);
    loginMutation.mutate(data, {
      onSuccess: (authData) => {
        toast.success(t("auth:login.success"), {
          description: t("auth:login.successDescription"),
        });

        // If user was trying to access a specific page, honor that
        if (from && from !== "/") {
          void navigate({ to: from, replace: true });
          return;
        }

        // Otherwise, redirect based on role
        const isAdmin = authData.user.role === "admin" || authData.user.role === "super_admin";
        void navigate({ to: isAdmin ? "/admin" : "/dashboard", replace: true });
      },
      onError: (err) => {
        setError(err);
        const errorMessage = err instanceof Error ? err.message : t("auth:login.error");
        toast.error(t("auth:login.error"), {
          description: errorMessage,
        });
      },
    });
  };

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-center text-2xl">{t("auth:login.title")}</CardTitle>
          <CardDescription className="text-center">{t("auth:login.subtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          <OAuthButtons
            mode="login"
            disabled={loginMutation.isPending}
            onError={(message) => setError(new Error(message))}
          />
          <OAuthDivider text={t("auth:oauth.continueWithEmail")} />
          <form
            onSubmit={handleSubmit(onSubmit)}
            className="space-y-4"
            role="form"
          >
            {error && (
              <>
                <Alert variant="destructive">
                  <AlertDescription>{error.message}</AlertDescription>
                </Alert>
                <AuthErrorGuidance
                  error={error}
                  context="login"
                />
              </>
            )}

            <div className="space-y-2">
              <Label htmlFor="email">{t("common:labels.email")}</Label>
              <Input
                id="email"
                type="email"
                placeholder={t("auth:login.emailPlaceholder")}
                autoFocus
                error={!!errors.email}
                aria-invalid={!!errors.email}
                aria-describedby={errors.email ? "email-error" : undefined}
                {...register("email")}
                disabled={loginMutation.isPending}
              />
              {errors.email && (
                <p
                  id="email-error"
                  role="alert"
                  className="text-destructive animate-fade-in-up flex items-center gap-1.5 text-sm"
                >
                  <AlertCircle
                    className="h-3.5 w-3.5 shrink-0"
                    aria-hidden="true"
                  />
                  {errors.email.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">{t("common:labels.password")}</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder={t("auth:login.passwordPlaceholder")}
                  error={!!errors.password}
                  aria-invalid={!!errors.password}
                  aria-describedby={errors.password ? "password-error" : undefined}
                  {...register("password")}
                  disabled={loginMutation.isPending}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                  onClick={() => setShowPassword(!showPassword)}
                  disabled={loginMutation.isPending}
                  aria-label={showPassword ? t("auth:login.hidePassword") : t("auth:login.showPassword")}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
              {errors.password && (
                <p
                  id="password-error"
                  role="alert"
                  className="text-destructive animate-fade-in-up flex items-center gap-1.5 text-sm"
                >
                  <AlertCircle
                    className="h-3.5 w-3.5 shrink-0"
                    aria-hidden="true"
                  />
                  {errors.password.message}
                </p>
              )}
            </div>

            <Button
              type="submit"
              className="w-full"
              loading={loginMutation.isPending}
            >
              {t("auth:login.submitButton")}
            </Button>
          </form>

          <div className="mt-4 text-center text-sm">
            {t("auth:login.noAccount")}{" "}
            <Link
              to="/register"
              className="text-primary hover:underline"
            >
              {t("auth:login.signUpLink")}
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
