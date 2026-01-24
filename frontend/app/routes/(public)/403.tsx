import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { createFileRoute, Link, useSearch } from "@tanstack/react-router";
import { Home, Lock, LogIn } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(public)/403")({
  validateSearch: (search: Record<string, unknown>) => ({
    from: typeof search.from === "string" ? search.from : undefined,
  }),
  component: UnauthorizedPage,
});

function UnauthorizedPage() {
  const { t } = useTranslation();
  const { from } = useSearch({ from: "/(public)/403" });

  return (
    <div className="flex min-h-[80vh] flex-col items-center justify-center px-4">
      <Card className="w-full max-w-md text-center">
        <CardHeader className="pb-4">
          <div className="bg-destructive/10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
            <Lock className="text-destructive h-8 w-8" />
          </div>
          <CardTitle className="text-2xl">{t("errors.403.title", "Access Denied")}</CardTitle>
          <CardDescription className="text-base">
            {t("errors.403.description", "You don't have permission to access this resource.")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground text-sm">
            {t(
              "errors.403.explanation",
              "This might be because your account doesn't have the required role, or you're not a member of the requested organization."
            )}
          </p>

          {from && (
            <p className="text-muted-foreground text-xs">
              {t("errors.403.attemptedPath", "Attempted to access:")}{" "}
              <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-xs">{from}</code>
            </p>
          )}

          <div className="flex flex-col gap-2 pt-4 sm:flex-row sm:justify-center">
            <Button
              variant="outline"
              asChild
            >
              <Link to="/">
                <Home className="mr-2 h-4 w-4" />
                {t("common.goHome", "Go Home")}
              </Link>
            </Button>
            <Button asChild>
              <Link
                to="/login"
                search={from ? { redirect: from } : undefined}
              >
                <LogIn className="mr-2 h-4 w-4" />
                {t("errors.403.tryDifferentAccount", "Try Different Account")}
              </Link>
            </Button>
          </div>

          <p className="text-muted-foreground pt-2 text-xs">
            {t(
              "errors.403.contactAdmin",
              "If you believe this is a mistake, please contact your organization administrator."
            )}
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
