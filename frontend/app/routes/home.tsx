import { Link } from "react-router";

export function meta() {
  return [
    { title: "About - React + Go Starter Kit" },
    { name: "description", content: "Learn more about this full-stack starter kit built with React and Go" },
  ];
}

const HomePage = () => {
  return (
    <main className="min-h-screen bg-gray-50 dark:bg-gray-900 py-12 px-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <header className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
            About This Project
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-300">
            A modern full-stack starter kit
          </p>
        </header>

        {/* Demo Button */}
        <div className="text-center mb-12">
          <Link
            to="/demo"
            className="inline-block bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-8 rounded-lg transition-colors duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5"
          >
            🚀 Try the Demo
          </Link>
        </div>

        {/* Main Content */}
        <div className="space-y-8">
          {/* Project Overview */}
          <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
              🚀 What is this?
            </h2>
            <p className="text-gray-700 dark:text-gray-300 leading-relaxed mb-4">
              This is a production-ready starter kit that combines the power of React on the frontend
              with Go on the backend. It's designed to help developers quickly bootstrap modern
              web applications with best practices built-in.
            </p>
            <p className="text-gray-700 dark:text-gray-300 leading-relaxed">
              Whether you're building a SaaS product, API service, or full-stack web app,
              this starter kit provides a solid foundation to build upon.
            </p>
          </section>

          {/* Technology Stack */}
          <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
              🛠️ Technology Stack
            </h2>
            <div className="grid md:grid-cols-2 gap-6">
              <div>
                <h3 className="text-lg font-medium text-blue-600 dark:text-blue-400 mb-3">
                  Frontend
                </h3>
                <ul className="space-y-2 text-gray-700 dark:text-gray-300">
                  <li>• React 18 with TypeScript</li>
                  <li>• React Router for navigation</li>
                  <li>• TailwindCSS for styling</li>
                  <li>• ShadCN/UI components</li>
                  <li>• Vite for fast development</li>
                </ul>
              </div>
              <div>
                <h3 className="text-lg font-medium text-green-600 dark:text-green-400 mb-3">
                  Backend
                </h3>
                <ul className="space-y-2 text-gray-700 dark:text-gray-300">
                  <li>• Go 1.24 with Chi router</li>
                  <li>• PostgreSQL database</li>
                  <li>• GORM ORM</li>
                  <li>• RESTful API design</li>
                  <li>• Docker & Docker Compose</li>
                </ul>
              </div>
            </div>
          </section>

          {/* Features */}
          <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
              ✨ Key Features
            </h2>
            <div className="grid md:grid-cols-2 gap-4">
              <div className="space-y-3">
                <div className="flex items-center space-x-2">
                  <span className="text-green-500">✓</span>
                  <span className="text-gray-700 dark:text-gray-300">Hot reload development</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-500">✓</span>
                  <span className="text-gray-700 dark:text-gray-300">Database migrations</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-500">✓</span>
                  <span className="text-gray-700 dark:text-gray-300">CORS configured</span>
                </div>
              </div>
              <div className="space-y-3">
                <div className="flex items-center space-x-2">
                  <span className="text-green-500">✓</span>
                  <span className="text-gray-700 dark:text-gray-300">Docker support</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-500">✓</span>
                  <span className="text-gray-700 dark:text-gray-300">TypeScript ready</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-green-500">✓</span>
                  <span className="text-gray-700 dark:text-gray-300">Dark mode support</span>
                </div>
              </div>
            </div>
          </section>

          {/* Getting Started */}
          <section className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
              🎯 Getting Started
            </h2>
            <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4">
              <p className="text-gray-700 dark:text-gray-300 mb-3">
                Ready to start building? Here's how to get up and running:
              </p>
              <div className="space-y-2 text-sm">
                <p className="font-medium text-gray-900 dark:text-white">1. Clone the repository</p>
                <p className="font-medium text-gray-900 dark:text-white">2. Set up your environment variables</p>
                <p className="font-medium text-gray-900 dark:text-white">3. Run <code className="bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded">docker-compose up</code></p>
                <p className="font-medium text-gray-900 dark:text-white">4. Start developing!</p>
              </div>
            </div>
          </section>

          {/* Footer */}
          <footer className="text-center pt-8 border-t border-gray-200 dark:border-gray-700">
            <p className="text-gray-600 dark:text-gray-400">
              Built with ❤️ using React & Go
            </p>
          </footer>
        </div>
      </div>
    </main>
  );
}

export default function Home() {
  return <HomePage />;
}
