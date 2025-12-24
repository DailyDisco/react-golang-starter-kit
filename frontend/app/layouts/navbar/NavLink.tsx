import { Link } from "@tanstack/react-router";

import { isActive, type NavItem } from "./types";

interface NavLinkProps {
  item: NavItem;
  pathname: string;
  onClick?: () => void;
  variant?: "desktop" | "mobile";
}

export function NavLink({ item, pathname, onClick, variant = "desktop" }: NavLinkProps) {
  const active = isActive(pathname, item.href);

  if (variant === "mobile") {
    return (
      <Link
        to={item.href}
        target={item.external ? "_blank" : undefined}
        rel={item.external ? "noopener noreferrer" : undefined}
        onClick={onClick}
        role="menuitem"
        className={`mx-2 flex items-center rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none ${
          active
            ? "bg-blue-600 text-white shadow-sm"
            : "text-gray-700 hover:bg-gray-50 hover:text-gray-900 hover:shadow-sm dark:text-gray-200 dark:hover:bg-gray-800/50 dark:hover:text-gray-100"
        }`}
      >
        {item.name}
        {item.external && <span className="ml-auto text-xs opacity-70">↗</span>}
      </Link>
    );
  }

  if (item.external) {
    return (
      <a
        href={item.href}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center rounded-md border-transparent px-3 py-2 text-sm font-medium text-gray-600 transition-colors duration-200 hover:bg-gray-100 hover:text-gray-900 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100"
      >
        {item.name}
        <span className="ml-1 text-xs opacity-60">↗</span>
      </a>
    );
  }

  return (
    <Link
      to={item.href}
      search={{}}
      className={`inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none ${
        active
          ? "border border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300"
          : "text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100"
      }`}
    >
      {item.name}
    </Link>
  );
}
