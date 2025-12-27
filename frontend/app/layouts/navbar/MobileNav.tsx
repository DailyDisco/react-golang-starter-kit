import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Link } from "@tanstack/react-router";
import { CreditCard, DollarSign, LogOut, Menu, Settings, Shield, User } from "lucide-react";

import type { User as UserType } from "../../services";
import { NavLink } from "./NavLink";
import { getUserInitials, navigation } from "./types";

interface MobileNavProps {
  isOpen: boolean;
  setIsOpen: (open: boolean) => void;
  pathname: string;
  isAuthenticated: boolean;
  user: UserType | null;
  onLogout: () => void;
}

export function MobileNav({ isOpen, setIsOpen, pathname, isAuthenticated, user, onLogout }: MobileNavProps) {
  const closeMenu = () => setIsOpen(false);
  const isAdmin = user?.role === "admin" || user?.role === "super_admin";

  return (
    <Sheet
      open={isOpen}
      onOpenChange={setIsOpen}
    >
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="ml-2 transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
          aria-label={isOpen ? "Close main menu" : "Open main menu"}
          aria-expanded={isOpen}
          aria-controls="mobile-menu"
        >
          <Menu className="h-6 w-6" />
          <span className="sr-only">{isOpen ? "Close main menu" : "Open main menu"}</span>
        </Button>
      </SheetTrigger>
      <SheetContent
        side="right"
        className="w-[320px] border-l border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-900"
        id="mobile-menu"
      >
        <div
          className="flex h-full flex-col p-4"
          role="menu"
        >
          {/* Header */}
          <div className="mb-6 border-b border-gray-200 pb-6 dark:border-gray-700">
            <Link
              to="/"
              search={{}}
              className="text-xl font-bold text-gray-900 transition-colors hover:text-blue-600 dark:text-white dark:hover:text-blue-400"
              onClick={closeMenu}
            >
              React + Go
            </Link>
          </div>

          {/* Navigation */}
          <div className="flex-1">
            <div className="mb-8 space-y-1">
              <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">Navigation</p>
              {navigation.map((item) => (
                <NavLink
                  key={item.name}
                  item={item}
                  pathname={pathname}
                  onClick={closeMenu}
                  variant="mobile"
                />
              ))}
            </div>
          </div>

          {/* User Section */}
          <div className="border-t border-gray-200 pt-6 dark:border-gray-700">
            {isAuthenticated && user ? (
              <>
                <div className="mb-6 flex items-center rounded-lg border border-blue-100 bg-gradient-to-r from-blue-50 to-indigo-50 p-4 shadow-sm dark:border-blue-900/50 dark:from-blue-950/50 dark:to-indigo-950/50">
                  <Avatar className="mr-3 h-10 w-10 ring-2 ring-blue-200 dark:ring-blue-800">
                    <AvatarImage
                      src=""
                      alt={user.name}
                    />
                    <AvatarFallback className="bg-blue-600 text-sm text-white">
                      {getUserInitials(user.name)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-gray-900 dark:text-white">{user.name}</p>
                    <p className="text-muted-foreground truncate text-xs">{user.email}</p>
                  </div>
                </div>

                <div className="space-y-1">
                  <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">Account</p>
                  <Link
                    to="/settings/profile"
                    search={{}}
                    onClick={closeMenu}
                    role="menuitem"
                    className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                  >
                    <User className="mr-3 h-4 w-4" />
                    Profile
                  </Link>
                  <Link
                    to="/settings"
                    search={{}}
                    onClick={closeMenu}
                    role="menuitem"
                    className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                  >
                    <Settings className="mr-3 h-4 w-4" />
                    Settings
                  </Link>
                  <Link
                    to="/billing"
                    search={{}}
                    onClick={closeMenu}
                    role="menuitem"
                    className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                  >
                    <CreditCard className="mr-3 h-4 w-4" />
                    Billing
                  </Link>
                  <Link
                    to="/pricing"
                    search={{}}
                    onClick={closeMenu}
                    role="menuitem"
                    className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                  >
                    <DollarSign className="mr-3 h-4 w-4" />
                    Pricing
                  </Link>
                  {isAdmin && (
                    <Link
                      to="/admin"
                      search={{}}
                      onClick={closeMenu}
                      role="menuitem"
                      className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                    >
                      <Shield className="mr-3 h-4 w-4" />
                      Admin Panel
                    </Link>
                  )}
                  <button
                    onClick={() => {
                      onLogout();
                      closeMenu();
                    }}
                    role="menuitem"
                    className="mx-2 flex w-full items-center rounded-lg px-4 py-3 text-sm font-medium text-red-600 transition-all duration-200 hover:bg-red-50 hover:text-red-700 hover:shadow-sm focus:ring-2 focus:ring-red-500 focus:ring-offset-2 focus:outline-none dark:text-red-400 dark:hover:bg-red-950/50 dark:hover:text-red-300"
                  >
                    <LogOut className="mr-3 h-4 w-4" />
                    Sign Out
                  </button>
                </div>
              </>
            ) : (
              <div className="space-y-1">
                <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">
                  Authentication
                </p>
                <Link
                  to="/login"
                  search={{}}
                  onClick={closeMenu}
                  role="menuitem"
                  className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                >
                  Sign In
                </Link>
                <Link
                  to="/register"
                  search={{}}
                  onClick={closeMenu}
                  role="menuitem"
                  className="mx-2 flex items-center rounded-lg bg-gradient-to-r from-blue-600 to-blue-700 px-4 py-3 text-sm font-medium text-white shadow-sm transition-all duration-200 hover:from-blue-700 hover:to-blue-800 hover:shadow-md focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
                >
                  Sign Up
                </Link>
              </div>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
