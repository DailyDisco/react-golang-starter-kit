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

/**
 * Modal displayed when user's session expires.
 * Listens for 'session-expired' custom events and shows a friendly message
 * before redirecting to login.
 */
export function SessionExpiredModal() {
  const [isOpen, setIsOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const handleSessionExpired = () => {
      setIsOpen(true);
    };

    window.addEventListener("session-expired", handleSessionExpired);

    return () => {
      window.removeEventListener("session-expired", handleSessionExpired);
    };
  }, []);

  const handleSignIn = () => {
    setIsOpen(false);
    // Clear any stored auth data
    localStorage.removeItem("auth_user");
    localStorage.removeItem("refresh_token");
    // Navigate to login
    void navigate({ to: "/login" });
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={setIsOpen}
    >
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>Session Expired</DialogTitle>
          <DialogDescription>
            Your session has expired due to inactivity. Please sign in again to continue.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button onClick={handleSignIn}>Sign In Again</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
