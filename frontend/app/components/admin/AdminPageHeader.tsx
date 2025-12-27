import { Link } from "@tanstack/react-router";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "../ui/breadcrumb";

interface BreadcrumbItem {
  label: string;
  href?: string;
}

interface AdminPageHeaderProps {
  title: string;
  description?: string;
  breadcrumbs?: BreadcrumbItem[];
  actions?: React.ReactNode;
}

export function AdminPageHeader({ title, description, breadcrumbs = [], actions }: AdminPageHeaderProps) {
  const allBreadcrumbs = [{ label: "Admin", href: "/admin" }, ...breadcrumbs];

  return (
    <div className="space-y-4">
      {/* Breadcrumbs */}
      <Breadcrumb>
        <BreadcrumbList>
          {allBreadcrumbs.map((item, index) => (
            <BreadcrumbItem key={item.label}>
              {index > 0 && <BreadcrumbSeparator />}
              {index === allBreadcrumbs.length - 1 ? (
                <BreadcrumbPage>{item.label}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <Link to={item.href || "#"}>{item.label}</Link>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
          ))}
        </BreadcrumbList>
      </Breadcrumb>

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{title}</h2>
          {description && <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{description}</p>}
        </div>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
    </div>
  );
}
