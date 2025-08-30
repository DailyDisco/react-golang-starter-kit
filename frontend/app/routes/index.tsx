import { createFileRoute, Link } from '@tanstack/react-router';
import { API_BASE_URL } from '../services';

export const Route = createFileRoute('/')({
  component: Home,
});

function Home() {
  return (
    <main className='bg-gray-50 px-4 py-12 dark:bg-gray-900'>
      <div className='mx-auto max-w-4xl'>
        {/* Header */}
        <header className='mb-12'>
          <div className='text-center'>
            <h1 className='mb-4 text-4xl font-bold text-gray-900 dark:text-white'>
              About This Project
            </h1>
            <p className='text-xl text-gray-600 dark:text-gray-300'>
              A modern full-stack starter kit
            </p>
          </div>
        </header>

        {/* Action Buttons */}
        <div className='mb-12 flex flex-col items-center justify-center gap-4 px-4 sm:flex-row sm:gap-6 md:gap-8'>
          <Link
            to='/demo'
            search={{}}
            className='inline-flex w-full min-w-[180px] transform items-center justify-center rounded-lg bg-blue-600 px-6 py-3 text-center font-semibold text-white shadow-lg transition-colors duration-200 hover:-translate-y-0.5 hover:bg-blue-700 hover:shadow-xl sm:w-auto sm:px-8 dark:bg-blue-700 dark:hover:bg-blue-600'
          >
            🚀 Try the Demo
          </Link>
          <a
            href={`${API_BASE_URL}/swagger/`}
            target='_blank'
            rel='noopener noreferrer'
            className='inline-flex w-full min-w-[180px] transform items-center justify-center rounded-lg bg-green-600 px-6 py-3 text-center font-semibold text-white shadow-lg transition-colors duration-200 hover:-translate-y-0.5 hover:bg-green-700 hover:shadow-xl sm:w-auto sm:px-8 dark:bg-green-700 dark:hover:bg-green-600'
          >
            📚 API Docs
          </a>
        </div>

        {/* Main Content */}
        <div className='space-y-8'>
          {/* Project Overview */}
          <section className='rounded-lg bg-white p-8 shadow-md dark:bg-gray-800'>
            <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
              🚀 What is this?
            </h2>
            <p className='mb-4 leading-relaxed text-gray-700 dark:text-gray-300'>
              This is a production-ready starter kit that combines the power of
              React on the frontend with Go on the backend. It's designed to
              help developers quickly bootstrap modern web applications with
              best practices built-in.
            </p>
            <p className='leading-relaxed text-gray-700 dark:text-gray-300'>
              Whether you're building a SaaS product, API service, or full-stack
              web app, this starter kit provides a solid foundation to build
              upon.
            </p>
          </section>

          {/* Technology Stack */}
          <section className='rounded-lg bg-white p-8 shadow-md dark:bg-gray-800'>
            <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
              🛠️ Technology Stack
            </h2>
            <div className='grid gap-6 md:grid-cols-2'>
              <div>
                <h3 className='mb-3 text-lg font-medium text-blue-600 dark:text-blue-400'>
                  Frontend
                </h3>
                <ul className='space-y-2 text-gray-700 dark:text-gray-300'>
                  <li>• React 19 with TypeScript</li>
                  <li>• TanStack Router for navigation</li>
                  <li>• TailwindCSS for styling</li>
                  <li>• ShadCN/UI components</li>
                  <li>• Vite for fast development</li>
                </ul>
              </div>
              <div>
                <h3 className='mb-3 text-lg font-medium text-green-600 dark:text-green-400'>
                  Backend
                </h3>
                <ul className='space-y-2 text-gray-700 dark:text-gray-300'>
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
          <section className='rounded-lg bg-white p-8 shadow-md dark:bg-gray-800'>
            <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
              ✨ Key Features
            </h2>
            <div className='grid gap-4 md:grid-cols-2'>
              <div className='space-y-3'>
                <div className='flex items-center space-x-2'>
                  <span className='text-green-500'>✓</span>
                  <span className='text-gray-700 dark:text-gray-300'>
                    Hot reload development
                  </span>
                </div>
                <div className='flex items-center space-x-2'>
                  <span className='text-green-500'>✓</span>
                  <span className='text-gray-700 dark:text-gray-300'>
                    Database migrations
                  </span>
                </div>
                <div className='flex items-center space-x-2'>
                  <span className='text-green-500'>✓</span>
                  <span className='text-gray-700 dark:text-gray-300'>
                    CORS configured
                  </span>
                </div>
              </div>
              <div className='space-y-3'>
                <div className='flex items-center space-x-2'>
                  <span className='text-green-500'>✓</span>
                  <span className='text-gray-700 dark:text-gray-300'>
                    Docker support
                  </span>
                </div>
                <div className='flex items-center space-x-2'>
                  <span className='text-green-500'>✓</span>
                  <span className='text-gray-700 dark:text-gray-300'>
                    TypeScript ready
                  </span>
                </div>
                <div className='flex items-center space-x-2'>
                  <span className='text-green-500'>✓</span>
                  <span className='text-gray-700 dark:text-gray-300'>
                    Dark mode support
                  </span>
                </div>
              </div>
            </div>
          </section>

          {/* Use Cases */}
          <section className='rounded-lg bg-white p-8 shadow-md dark:bg-gray-800'>
            <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
              🎯 Perfect For
            </h2>
            <div className='grid gap-4 md:grid-cols-2'>
              <div className='space-y-3'>
                <div className='text-gray-700 dark:text-gray-300'>
                  • <strong>SaaS Applications</strong> - User management,
                  billing, dashboards
                </div>
                <div className='text-gray-700 dark:text-gray-300'>
                  • <strong>API Services</strong> - REST APIs with documentation
                </div>
                <div className='text-gray-700 dark:text-gray-300'>
                  • <strong>Admin Panels</strong> - CRUD interfaces, user
                  management
                </div>
              </div>
              <div className='space-y-3'>
                <div className='text-gray-700 dark:text-gray-300'>
                  • <strong>Prototyping</strong> - Fast development and testing
                </div>
                <div className='text-gray-700 dark:text-gray-300'>
                  • <strong>MVPs</strong> - Quick validation with
                  production-ready code
                </div>
                <div className='text-gray-700 dark:text-gray-300'>
                  • <strong>Full-Stack Projects</strong> - From idea to
                  deployment
                </div>
              </div>
            </div>
          </section>

          {/* Getting Started */}
          <section className='rounded-lg bg-white p-8 shadow-md dark:bg-gray-800'>
            <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
              🎯 Getting Started
            </h2>
            <div className='rounded-lg bg-gray-100 p-4 dark:bg-gray-700'>
              <p className='mb-3 text-gray-700 dark:text-gray-300'>
                Ready to start building? Here's how to get up and running:
              </p>
              <div className='space-y-2 text-sm'>
                <p className='font-medium text-gray-900 dark:text-white'>
                  1. Clone the repository
                </p>
                <p className='font-medium text-gray-900 dark:text-white'>
                  2. Copy{' '}
                  <code className='rounded bg-gray-200 px-1 py-0.5 dark:bg-gray-600'>
                    .env.example
                  </code>{' '}
                  to{' '}
                  <code className='rounded bg-gray-200 px-1 py-0.5 dark:bg-gray-600'>
                    .env
                  </code>{' '}
                  and configure database settings
                </p>
                <p className='font-medium text-gray-900 dark:text-white'>
                  3. Run{' '}
                  <code className='rounded bg-gray-200 px-1 py-0.5 dark:bg-gray-600'>
                    docker-compose up
                  </code>
                </p>
                <p className='font-medium text-gray-900 dark:text-white'>
                  4. Start developing!
                </p>
              </div>
              <div className='mt-4 text-xs text-gray-600 dark:text-gray-400'>
                💡 <strong>Tip:</strong> The project uses PostgreSQL, Docker,
                and includes hot reload for both frontend and backend.
              </div>
            </div>
          </section>

          {/* Development Workflow */}
          <section className='rounded-lg bg-white p-8 shadow-md dark:bg-gray-800'>
            <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
              ⚡ Development Workflow
            </h2>
            <div className='grid gap-6 md:grid-cols-2'>
              <div>
                <h3 className='mb-3 text-lg font-medium text-blue-600 dark:text-blue-400'>
                  🚀 Quick Commands
                </h3>
                <div className='space-y-2 text-sm'>
                  <div className='flex justify-between'>
                    <span className='text-gray-700 dark:text-gray-300'>
                      Start all services:
                    </span>
                    <code className='rounded bg-gray-200 px-2 py-1 dark:bg-gray-600'>
                      docker-compose up
                    </code>
                  </div>
                  <div className='flex justify-between'>
                    <span className='text-gray-700 dark:text-gray-300'>
                      View logs:
                    </span>
                    <code className='rounded bg-gray-200 px-2 py-1 dark:bg-gray-600'>
                      docker-compose logs -f
                    </code>
                  </div>
                  <div className='flex justify-between'>
                    <span className='text-gray-700 dark:text-gray-300'>
                      Stop services:
                    </span>
                    <code className='rounded bg-gray-200 px-2 py-1 dark:bg-gray-600'>
                      docker-compose down
                    </code>
                  </div>
                </div>
              </div>
              <div>
                <h3 className='mb-3 text-lg font-medium text-green-600 dark:text-green-400'>
                  🔗 Useful Links
                </h3>
                <div className='space-y-2 text-sm'>
                  <div className='text-gray-700 dark:text-gray-300'>
                    📚{' '}
                    <a
                      href={`${API_BASE_URL}/swagger/`}
                      className='text-blue-600 hover:underline dark:text-blue-400'
                    >
                      API Documentation
                    </a>
                  </div>
                  <div className='text-gray-700 dark:text-gray-300'>
                    🎮{' '}
                    <Link
                      to='/demo'
                      search={{}}
                      className='text-blue-600 hover:underline dark:text-blue-400'
                    >
                      Try Live Demo
                    </Link>
                  </div>
                  <div className='text-gray-700 dark:text-gray-300'>
                    🐙{' '}
                    <a
                      href='https://github.com'
                      className='text-blue-600 hover:underline dark:text-blue-400'
                    >
                      View Source Code
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </main>
  );
}
