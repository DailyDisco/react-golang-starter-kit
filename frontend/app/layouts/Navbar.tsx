import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Link, useLocation } from "@tanstack/react-router";
import { CreditCard, DollarSign, LogOut, Menu, Settings, Shield, User } from "lucide-react";

import { useAuth } from "../hooks/auth/useAuth";
import { API_BASE_URL } from "../services";

export function Navbar() {
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);
  const { user, isAuthenticated, logout } = useAuth();
  const isAdmin = user?.role === "admin" || user?.role === "super_admin";

  const navigation = [
    { name: "Home", href: "/" },
    { name: "Demo", href: "/demo" },
    { name: "Pricing", href: "/pricing" },
    {
      name: "API Docs",
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
    <nav className="sticky top-0 z-50 border-b border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-gray-900">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
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
            <div className="hidden md:ml-6 md:flex md:space-x-1">
              {navigation.map((item) =>
                item.external ? (
                  <a
                    key={item.name}
                    href={item.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center rounded-md border-transparent px-3 py-2 text-sm font-medium text-gray-600 transition-colors duration-200 hover:bg-gray-100 hover:text-gray-900 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100"
                  >
                    {item.name}
                    <span className="ml-1 text-xs opacity-60">↗</span>
                  </a>
                ) : (
                  <Link
                    key={item.name}
                    to={item.href}
                    search={{}}
                    className={`inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none ${
                      isActive(item.href)
                        ? "border border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100"
                    }`}
                  >
                    {item.name}
                  </Link>
                )
              )}
              {/* Authenticated nav links */}
              {isAuthenticated && (
                <>
                  <Link
                    to="/settings"
                    search={{}}
                    className={`inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none ${
                      isActive("/settings")
                        ? "border border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100"
                    }`}
                  >
                    Settings
                  </Link>
                  {isAdmin && (
                    <Link
                      to="/admin"
                      search={{}}
                      className={`inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none ${
                        isActive("/admin")
                          ? "border border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300"
                          : "text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100"
                      }`}
                    >
                      Admin
                    </Link>
                  )}
                </>
              )}
            </div>
          </div>
          {/* User Controls */}
          <div className="hidden md:ml-6 md:flex md:items-center md:space-x-4">
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
                      <span>Profile</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      to="/settings"
                      search={{}}
                      className="cursor-pointer"
                    >
                      <Settings className="mr-2 h-4 w-4" />
                      <span>Settings</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      to="/billing"
                      search={{}}
                      className="cursor-pointer"
                    >
                      <CreditCard className="mr-2 h-4 w-4" />
                      <span>Billing</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link
                      to="/pricing"
                      search={{}}
                      className="cursor-pointer"
                    >
                      <DollarSign className="mr-2 h-4 w-4" />
                      <span>Pricing</span>
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
                        <span>Admin Panel</span>
                      </Link>
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={handleLogout}
                    className="cursor-pointer"
                  >
                    <LogOut className="mr-2 h-4 w-4" />
                    <span>Log out</span>
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
                    Sign in
                  </Link>
                </Button>
                <Button asChild>
                  <Link
                    to="/register"
                    search={{}}
                  >
                    Sign up
                  </Link>
                </Button>
              </div>
            )}
          </div>
          <div className="-mr-2 flex items-center md:hidden">
            <ThemeToggle />
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
                      onClick={() => setIsOpen(false)}
                    >
                      React + Go
                    </Link>
                  </div>

                  {/* Navigation */}
                  <div className="flex-1">
                    <div className="mb-8 space-y-1">
                      <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">
                        Navigation
                      </p>
                      {navigation.map((item) => (
                        <Link
                          key={item.name}
                          to={item.href}
                          target={item.external ? "_blank" : undefined}
                          rel={item.external ? "noopener noreferrer" : undefined}
                          onClick={() => setIsOpen(false)}
                          role="menuitem"
                          className={`mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none ${
                            isActive(item.href)
                              ? "bg-blue-600 text-white shadow-sm"
                              : "text-gray-700 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                          }`}
                        >
                          {item.name}
                          {item.external && <span className="ml-auto text-xs opacity-70">↗</span>}
                        </Link>
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
                              {user.name
                                .split(" ")
                                .map((n) => n[0])
                                .join("")
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div className="min-w-0 flex-1">
                            <p className="truncate text-sm font-medium text-gray-900 dark:text-white">{user.name}</p>
                            <p className="text-muted-foreground truncate text-xs">{user.email}</p>
                          </div>
                        </div>

                        <div className="space-y-1">
                          <p className="text-muted-foreground mb-4 text-xs font-semibold tracking-wider uppercase">
                            Account
                          </p>
                          <Link
                            to="/settings/profile"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                          >
                            <User className="mr-3 h-4 w-4" />
                            Profile
                          </Link>
                          <Link
                            to="/settings"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                          >
                            <Settings className="mr-3 h-4 w-4" />
                            Settings
                          </Link>
                          <Link
                            to="/billing"
                            search={{}}
                            onClick={() => setIsOpen(false)}
                            role="menuitem"
                            className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                          >
                            <CreditCard className="mr-3 h-4 w-4" />
                            Billing
                          </Link>
                          <Link
                            to="/pricing"
                            search={{}}
                            onClick={() => setIsOpen(false)}
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
                              onClick={() => setIsOpen(false)}
                              role="menuitem"
                              className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                            >
                              <Shield className="mr-3 h-4 w-4" />
                              Admin Panel
                            </Link>
                          )}
                          <button
                            onClick={() => {
                              handleLogout();
                              setIsOpen(false);
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
                          onClick={() => setIsOpen(false)}
                          role="menuitem"
                          className="mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
                        >
                          Sign In
                        </Link>
                        <Link
                          to="/register"
                          search={{}}
                          onClick={() => setIsOpen(false)}
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
          </div>
        </div>
      </div>
    </nav>
  );
}
