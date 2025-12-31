import { useState } from "react";

import { MobileSettingsSidebar, SettingsSidebar } from "./SettingsSidebar";

interface SettingsLayoutProps {
  children: React.ReactNode;
}

export function SettingsLayout({ children }: SettingsLayoutProps) {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div className="flex h-[calc(100vh-theme(spacing.16)-theme(spacing.12))] overflow-hidden">
      {/* Desktop Sidebar */}
      <div className="hidden md:flex">
        <SettingsSidebar
          collapsed={collapsed}
          onCollapsedChange={setCollapsed}
        />
      </div>

      {/* Main Content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Mobile Header with Sidebar Toggle */}
        <div className="flex h-12 items-center border-b px-4 md:hidden">
          <MobileSettingsSidebar />
          <span className="ml-2 font-semibold">Settings</span>
        </div>

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
