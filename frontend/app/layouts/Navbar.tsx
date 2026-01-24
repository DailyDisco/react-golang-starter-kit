import { useState } from "react";

import { OrgSwitcher } from "@/components/navigation/OrgSwitcher";
import { PrefetchLink } from "@/components/navigation/PrefetchLink";
import { NotificationCenter } from "@/components/notifications/NotificationCenter";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { LanguageToggle } from "@/components/ui/language-toggle";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Link, useLocation } from "@tanstack/react-router";
import { CreditCard, DollarSign, LogOut, Menu, Settings, Shield, User } from "lucide-react";
import { useTranslation } from "react-i18next";

import { useAuth } from "../hooks/auth/useAuth";
import { usePrefetchSettingsData } from "../lib/prefetch";
import { API_BASE_URL } from "../services";

export function Navbar() {
  const { t } = useTranslation("common");
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);
  const { user, isAuthenticated, logout } = useAuth();
  const isAdmin = user?.role === "admin" || user?.role === "super_admin";
  const prefetchSettingsData = usePrefetchSettingsData();

  const navigation = [
    { name: t("navigation.home"), href: "/" },
    { name: t("navigation.demo"), href: "/demo" },
    { name: t("navigation.pricing"), href: "/pricing" },
    {
      name: t("navigation.apiDocs"),
      href: `${API_BASE_URL}/swagger/`,
      external: true,
    },
  ];

  const handleLogout = () => {
    logout();
  };

  const isActive = (href: string) => {
    if (href === "/" && location.pathname === "/") return true;
    if (href !== "/" && location.pathname.startsWith(href)) return true;
    return false;
  };

  return (
    <nav className="border-border bg-background sticky top-0 z-50 border-b shadow-sm">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <div className="flex">
            <div className="flex flex-shrink-0 items-center">
              <Link
                to="/"
                search={{}}
                className="text-foreground text-xl font-bold"
              >
                {t("appName")}
              </Link>
            </div>
            <div className="hidden md:ml-6 md:flex md:items-center md:space-x-1">
              {navigation.map((item) =>
                item.external ? (
                  <a
                    key={item.name}
                    href={item.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-muted-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary inline-flex items-center rounded-md border-transparent px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-offset-2 focus:outline-none"
                  >
                    {item.name}
                    <span className="ml-1 text-xs opacity-60">↗</span>
                  </a>
                ) : (
                  <Link
                    key={item.name}
                    to={item.href}
                    search={{}}
                    className={`focus:ring-primary inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-offset-2 focus:outline-none ${
                      isActive(item.href)
                        ? "border-primary/20 bg-primary/10 text-primary border"
                        : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    {item.name}
                  </Link>
                )
              )}
              {/* Separator before admin link */}
              {isAuthenticated && isAdmin && (
                <Separator
                  orientation="vertical"
                  className="mx-2 h-5 self-center"
                />
              )}
              {/* Authenticated nav links */}
              {isAuthenticated && isAdmin && (
                <Link
                  to="/admin"
                  search={{}}
                  className={`focus:ring-primary inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-offset-2 focus:outline-none ${
                    isActive("/admin")
                      ? "border-primary/20 bg-primary/10 text-primary border"
                      : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                  }`}
                >
                  {t("navigation.admin")}
                </Link>
              )}
            </div>
          </div>
          {/* User Controls */}
          <div className="hidden md:ml-6 md:flex md:items-center md:space-x-4">
            {isAuthenticated && <OrgSwitcher />}
            {isAuthenticated && <NotificationCenter />}
            <LanguageToggle />
            <ThemeToggle />

            {isAuthenticated && user ? (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    className="relative h-8 w-8 rounded-full"
                  >
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src=""
                        alt={user.name}
                      />
                      <AvatarFallback>
                        {user.name
                          .split(" ")
                          .map((n) => n[0])
                          .join("")
                          .toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  className="w-56"
                  align="end"
                  forceMount
                >
                  <div className="flex items-center justify-start gap-2 p-2">
                    <div className="flex flex-col space-y-1 leading-none">
                      <p className="font-medium">{user.name}</p>
                      <p className="text-muted-foreground w-[200px] truncate text-sm">{user.email}</p>
                    </div>
                  </div>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem asChild>
                    <Link
                      to="/settings/profile"
                      search={{}}
                      className="cursor-pointer"
                    >
                      <User className="mr-2 h-4 w-4" />
                      <span>{t("navigation.profile")}</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <PrefetchLink
                      to="/settings"
                      search={{}}
                      onPrefetch={prefetchSettingsData}
                      className="cursor-pointer"
                    >
                      <Settings className="mr-2 h-4 w-4" />
                      <span>{t("navigation.settings")}</span>
                    </PrefetchLink>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      to="/billing"
                      search={{}}
                      className="cursor-pointer"
                    >
                      <CreditCard className="mr-2 h-4 w-4" />
                      <span>{t("navigation.billing")}</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      to="/pricing"
                      search={{}}
                      className="cursor-pointer"
                    >
                      <DollarSign className="mr-2 h-4 w-4" />
                      <span>{t("navigation.pricing")}</span>
                    </Link>
                  </DropdownMenuItem>
                  {isAdmin && (
                    <DropdownMenuItem asChild>
                      <Link
                        to="/admin"
                        search={{}}
                        className="cursor-pointer"
                      >
                        <Shield className="mr-2 h-4 w-4" />
                        <span>{t("navigation.adminPanel")}</span>
                      </Link>
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={handleLogout}
                    className="cursor-pointer"
                  >
                    <LogOut className="mr-2 h-4 w-4" />
                    <span>{t("auth.logOut")}</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            ) : (
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
            )}
          </div>
          <div className="-mr-2 flex items-center md:hidden">
            <LanguageToggle />
            <ThemeToggle />
            <Sheet
              open={isOpen}
              onOpenChange={setIsOpen}
            >
              <SheetTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="hover:bg-accent ml-2 transition-colors"
                  aria-label={isOpen ? t("menu.closeMainMenu") : t("menu.openMainMenu")}
                  aria-expanded={isOpen}
                  aria-controls="mobile-menu"
                >
                  <Menu className="h-6 w-6" />
                  <span className="sr-only">{isOpen ? t("menu.closeMainMenu") : t("menu.openMainMenu")}</span>
                </Button>
              </SheetTrigger>
              <SheetContent
                side="right"
                className="border-border bg-background w-[320px] border-l"
                id="mobile-menu"
              >
                <div
                  className="flex h-full flex-col p-4"
                  role="menu"
                >
                  {/* Header */}
                  <div className="border-border mb-6 border-b pb-6">
                    <Link
                      to="/"
                      search={{}}
                      className="text-foreground hover:text-primary text-xl font-bold transition-colors"
                      onClick={() => setIsOpen(false)}
                    >
                      {t("appName")}
                    </Link>
                  </div>

                  {/* Navigation */}
                  <div className="flex-1">
                    <div className="mb-8 space-y-1">
                      <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">
                        {t("menu.navigation")}
                      </p>
                      {navigation.map((item) => (
                        <Link
                          key={item.name}
                          to={item.href}
                          target={item.external ? "_blank" : undefined}
                          rel={item.external ? "noopener noreferrer" : undefined}
                          onClick={() => setIsOpen(false)}
                          role="menuitem"
                          className={`focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 focus:ring-2 focus:ring-offset-2 focus:outline-none ${
                            isActive(item.href)
                              ? "bg-primary text-primary-foreground shadow-sm"
                              : "text-foreground hover:bg-accent hover:text-accent-foreground hover:shadow-sm"
                          }`}
                        >
                          {item.name}
                          {item.external && <span className="ml-auto text-xs opacity-70">↗</span>}
                        </Link>
                      ))}
                    </div>
                  </div>

                  {/* User Section */}
                  <div className="border-border border-t pt-6">
                    {isAuthenticated && user ? (
                      <>
                        <div className="border-primary/20 bg-primary/5 mb-6 flex items-center rounded-lg border p-4 shadow-sm">
                          <Avatar className="ring-primary/20 mr-3 h-10 w-10 ring-2">
                            <AvatarImage
                              src=""
                              alt={user.name}
                            />
                            <AvatarFallback className="bg-primary text-primary-foreground text-sm">
                              {user.name
                                .split(" ")
                                .map((n) => n[0])
                                .join("")
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div className="min-w-0 flex-1">
                            <p className="text-foreground truncate text-sm font-medium">{user.name}</p>
                            <p className="text-muted-foreground truncate text-xs">{user.email}</p>
                          </div>
                        </div>

                        {/* Organization Switcher */}
                        <div className="mb-4">
                          <OrgSwitcher className="w-full" />
                        </div>

                        <div className="space-y-1">
                          <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">
                            {t("menu.account")}
                          </p>
                          <Link
                            to="/settings/profile"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="text-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                          >
                            <User className="mr-3 h-4 w-4" />
                            {t("navigation.profile")}
                          </Link>
                          <Link
                            to="/settings"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="text-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                          >
                            <Settings className="mr-3 h-4 w-4" />
                            {t("navigation.settings")}
                          </Link>
                          <Link
                            to="/billing"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="text-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                          >
                            <CreditCard className="mr-3 h-4 w-4" />
                            {t("navigation.billing")}
                          </Link>
                          <Link
                            to="/pricing"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="text-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                          >
                            <DollarSign className="mr-3 h-4 w-4" />
                            {t("navigation.pricing")}
                          </Link>
                          {isAdmin && (
                            <Link
                              to="/admin"
                              search={{}}
                              onClick={() => setIsOpen(false)}
                              role="menuitem"
                              className="text-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                            >
                              <Shield className="mr-3 h-4 w-4" />
                              {t("navigation.adminPanel")}
                            </Link>
                          )}
                          <button
                            onClick={() => {
                              handleLogout();
                              setIsOpen(false);
                            }}
                            role="menuitem"
                            className="text-destructive hover:bg-destructive/10 focus:ring-destructive mx-2 flex w-full items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                          >
                            <LogOut className="mr-3 h-4 w-4" />
                            {t("auth.signOut")}
                          </button>
                        </div>
                      </>
                    ) : (
                      <div className="space-y-1">
                        <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">
                          {t("menu.authentication")}
                        </p>
                        <Link
                          to="/login"
                          search={{}}
                          onClick={() => setIsOpen(false)}
                          role="menuitem"
                          className="text-foreground hover:bg-accent hover:text-accent-foreground focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 hover:shadow-sm focus:ring-2 focus:ring-offset-2 focus:outline-none"
                        >
                          {t("auth.signIn")}
                        </Link>
                        <Link
                          to="/register"
                          search={{}}
                          onClick={() => setIsOpen(false)}
                          role="menuitem"
                          className="bg-primary text-primary-foreground hover:bg-primary/90 focus:ring-primary mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium shadow-sm transition-all duration-200 hover:shadow-md focus:ring-2 focus:ring-offset-2 focus:outline-none"
                        >
                          {t("auth.signUp")}
                        </Link>
                      </div>
                    )}
                  </div>
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </nav>
  );
}
