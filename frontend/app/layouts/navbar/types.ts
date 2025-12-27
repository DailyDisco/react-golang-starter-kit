import { API_BASE_URL } from "../../services";

export interface NavItem {
  name: string;
  href: string;
  external?: boolean;
}

export const navigation: NavItem[] = [
  { name: "Home", href: "/" },
  { name: "Demo", href: "/demo" },
  { name: "Pricing", href: "/pricing" },
  {
    name: "API Docs",
    href: `${API_BASE_URL}/swagger/`,
    external: true,
  },
];

export function isActive(pathname: string, href: string): boolean {
  if (href === "/" && pathname === "/") return true;
  if (href !== "/" && pathname.startsWith(href)) return true;
  return false;
}

export function getUserInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase();
}
