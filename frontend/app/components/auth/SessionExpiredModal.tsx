import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { useAuthStore } from "../../stores/auth-store";

/**
 * Modal displayed when user's session expires.
 * Listens for 'session-expired' custom events and shows a friendly message
 * before redirecting to login.
 */
export function SessionExpiredModal() {
  const { t } = useTranslation("auth");
  const [isOpen, setIsOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const handleSessionExpired = () => {
      // A 401 from any anonymous request (e.g. the public demo page probing
      // /api/files) also fires this event. Showing "your session has
      // expired" to a visitor who never signed in is wrong and it blocks the
      // page behind a modal - only surface it when this tab actually had a
      // backend-validated session.
      if (!useAuthStore.getState().hasValidatedSession) {
        return;
      }
      setIsOpen(true);
    };

    window.addEventListener("session-expired", handleSessionExpired);

    return () => {
      window.removeEventListener("session-expired", handleSessionExpired);
    };
  }, []);

  const handleSignIn = () => {
    setIsOpen(false);
    // Navigate to login - auth cookies are httpOnly and cleared by backend on logout
    void navigate({ to: "/login" });
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={setIsOpen}
    >
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          {/* @ts-expect-error - Deep type instantiation with react-i18next TFunction */}
          <DialogTitle>{t("session.expired")}</DialogTitle>
          <DialogDescription>{t("session.expiredDescription")}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button onClick={handleSignIn}>{t("session.signInAgain")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
