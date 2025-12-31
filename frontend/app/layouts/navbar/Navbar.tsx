import { useState } from "react";

import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Link, useLocation } from "@tanstack/react-router";

import { KeyboardShortcutsHelp } from "../../components/help";
import { NotificationBell } from "../../components/notifications";
import { useAuth } from "../../hooks/auth/useAuth";
import { AuthButtons } from "./AuthButtons";
import { DesktopNav } from "./DesktopNav";
import { MobileNav } from "./MobileNav";
import { UserMenu } from "./UserMenu";

export function Navbar() {
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);
  const { user, isAuthenticated, logout } = useAuth();

  return (
    <nav className="sticky top-0 z-50 border-b border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-gray-900">
      <div className="mx-auto max-w-screen-2xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          {/* Logo and Desktop Navigation */}
          <div className="flex">
            <div className="flex flex-shrink-0 items-center">
              <Link
                to="/"
                search={{}}
                className="text-xl font-bold text-gray-900 dark:text-white"
              >
                React + Go
              </Link>
            </div>
            <DesktopNav pathname={location.pathname} />
          </div>

          {/* Desktop User Controls */}
          <div className="hidden md:ml-6 md:flex md:items-center md:space-x-4">
            <ThemeToggle />
            {isAuthenticated && user && (
              <>
                <KeyboardShortcutsHelp />
                <NotificationBell />
              </>
            )}
            {isAuthenticated && user ? (
              <UserMenu
                user={user}
                onLogout={logout}
              />
            ) : (
              <AuthButtons />
            )}
          </div>

          {/* Mobile Menu */}
          <div className="-mr-2 flex items-center md:hidden">
            <ThemeToggle />
            <MobileNav
              isOpen={isOpen}
              setIsOpen={setIsOpen}
              pathname={location.pathname}
              isAuthenticated={isAuthenticated}
              user={user}
              onLogout={logout}
            />
          </div>
        </div>
      </div>
    </nav>
  );
}
