import { Link } from "react-router";

export function meta() {
  return [
    { title: "Custom Layout Demo - React + Go Starter Kit" },
    {
      name: "description",
      content: "Demonstrating custom layouts with React Router",
    },
  ];
}

export default function CustomLayoutDemo() {
  return (
    <div className="p-8 space-y-6">
      <div className="text-center space-y-4">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Custom Layout Demo
        </h1>
        <p className="text-gray-600 dark:text-gray-300">
          This page uses a custom layout component instead of the root layout!
        </p>
      </div>

      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6">
        <h2 className="text-lg font-semibold text-blue-900 dark:text-blue-100 mb-2">
          🎯 Layout Difference
        </h2>
        <p className="text-blue-800 dark:text-blue-200 text-sm">
          Notice: This page uses the custom Layout component, while other pages
          use the root layout with navbar and footer.
        </p>
      </div>

      <div className="flex gap-4 justify-center">
        <Link
          to="/"
          className="bg-gray-600 hover:bg-gray-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
        >
          Back to Home
        </Link>
        <Link
          to="/demo"
          className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
        >
          View Demo
        </Link>
      </div>
    </div>
  );
}
