import { Link } from "react-router";
import { Home, ArrowLeft, FileQuestion } from "lucide-react";

export function meta() {
  return [
    { title: "Page Not Found - React + Go Starter Kit" },
    {
      name: "description",
      content: "The page you're looking for doesn't exist.",
    },
  ];
}

const NotFoundPage = () => {
  return (
    <main className="flex-1 bg-gradient-to-br from-slate-50 via-gray-50 to-zinc-50 dark:from-slate-950 dark:via-gray-950 dark:to-zinc-950">
      <div className="flex items-center justify-center px-4 py-12 min-h-[60vh]">
        <div className="w-full max-w-lg mx-auto text-center">
          {/* Icon and 404 Number */}
          <div className="mb-8">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-gradient-to-br from-blue-100 to-indigo-100 dark:from-blue-900/30 dark:to-indigo-900/30 mb-6">
              <FileQuestion className="w-10 h-10 text-blue-600 dark:text-blue-400" />
            </div>
            <h1 className="text-6xl md:text-7xl font-bold bg-gradient-to-r from-slate-600 to-slate-800 dark:from-slate-300 dark:to-slate-500 bg-clip-text text-transparent select-none">
              404
            </h1>
          </div>

          {/* Main Content Card */}
          <div className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm rounded-2xl shadow-xl border border-gray-200/50 dark:border-gray-800/50 p-8 md:p-10">
            <div className="space-y-6">
              <div>
                <h2 className="text-2xl md:text-3xl font-semibold text-gray-900 dark:text-white mb-3">
                  Page Not Found
                </h2>
                <p className="text-gray-600 dark:text-gray-300 leading-relaxed">
                  The page you're looking for doesn't exist or may have been
                  moved.
                </p>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-col sm:flex-row gap-3 pt-2">
                <Link
                  to="/"
                  className="flex-1 inline-flex items-center justify-center gap-2 bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-medium py-3 px-6 rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl hover:scale-[1.02]"
                >
                  <Home className="w-4 h-4" />
                  Go Home
                </Link>
                <button
                  onClick={() => window.history.back()}
                  className="flex-1 inline-flex items-center justify-center gap-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 font-medium py-3 px-6 rounded-xl border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all duration-200 shadow-sm hover:shadow-md hover:scale-[1.02]"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Go Back
                </button>
              </div>

              {/* Additional Help */}
              <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Try checking the URL or return to our{" "}
                  <Link
                    to="/demo"
                    className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
                  >
                    demo page
                  </Link>
                  .
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default function NotFound() {
  return <NotFoundPage />;
}
