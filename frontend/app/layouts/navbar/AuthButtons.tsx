import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

export function AuthButtons() {
  const { t } = useTranslation("common");

  return (
    <div className="flex items-center space-x-2">
      <Button
        variant="ghost"
        asChild
      >
        <Link
          to="/login"
          search={{}}
        >
          {t("auth.signIn")}
        </Link>
      </Button>
      <Button asChild>
        <Link
          to="/register"
          search={{}}
        >
          {t("auth.signUp")}
        </Link>
      </Button>
    </div>
  );
}
