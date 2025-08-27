import { Outlet } from "react-router";

export default function Layout() {
  return (
    <div className="min-h-screen flex flex-col bg-slate-50">
      {/* Custom Layout Header */}
      <div className="bg-slate-800 text-white p-4 text-center">
        <div className="text-lg font-medium">Custom Demo Layout</div>
        <div className="text-sm text-slate-300 mt-1">
          This page uses a different layout structure
        </div>
      </div>

      <main className="flex-1 p-6">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-8">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
