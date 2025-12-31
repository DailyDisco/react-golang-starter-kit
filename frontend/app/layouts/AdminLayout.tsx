import { useState } from "react";

import { AdminSidebar, MobileAdminSidebar } from "./AdminSidebar";

interface AdminLayoutProps {
  children: React.ReactNode;
}

export function AdminLayout({ children }: AdminLayoutProps) {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className="flex h-[calc(100vh-theme(spacing.16)-theme(spacing.12))] overflow-hidden">
      {/* Desktop Sidebar */}
      <div className="hidden md:flex">
        <AdminSidebar
          collapsed={collapsed}
          onCollapsedChange={setCollapsed}
        />
      </div>

      {/* Main Content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Mobile Header with Sidebar Toggle */}
        <div className="flex h-12 items-center border-b px-4 md:hidden">
          <MobileAdminSidebar />
          <span className="ml-2 font-semibold">Admin</span>
        </div>

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
