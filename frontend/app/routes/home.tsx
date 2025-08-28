import { Link } from 'react-router';
import { API_BASE_URL } from '../services';

export function meta() {
  return [
    { title: 'About - React + Go Starter Kit' },
    {
      name: 'description',
      content:
        'Learn more about this full-stack starter kit built with React and Go',
    },
  ];
}

const HomePage = () => {
  return (
    <main className='bg-gray-50 dark:bg-gray-900 py-12 px-4'>
      <div className='max-w-4xl mx-auto'>
        {/* Header */}
        <header className='mb-12'>
          <div className='text-center'>
            <h1 className='text-4xl font-bold text-gray-900 dark:text-white mb-4'>
              About This Project
            </h1>
            <p className='text-xl text-gray-600 dark:text-gray-300'>
              A modern full-stack starter kit
            </p>
          </div>
        </header>

        {/* Action Buttons */}
        <div className='flex flex-col sm:flex-row gap-4 sm:gap-6 md:gap-8 justify-center items-center mb-12 px-4'>
          <Link
            to='/demo'
            className='w-full sm:w-auto inline-flex items-center justify-center bg-blue-600 hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-600 text-white font-semibold py-3 px-6 sm:px-8 rounded-lg transition-colors duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 min-w-[180px] text-center'
          >
            🚀 Try the Demo
          </Link>
          <a
            href={`${API_BASE_URL}/swagger/`}
            target='_blank'
            rel='noopener noreferrer'
            className='w-full sm:w-auto inline-flex items-center justify-center bg-green-600 hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-600 text-white font-semibold py-3 px-6 sm:px-8 rounded-lg transition-colors duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 min-w-[180px] text-center'
          >
            📚 API Docs
          </a>
        </div>

        {/* Main Content */}
        <div className='space-y-8'>
          {/* Project Overview */}
          <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-8'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
              🚀 What is this?
            </h2>
            <p className='text-gray-700 dark:text-gray-300 leading-relaxed mb-4'>
              This is a production-ready starter kit that combines the power of
              React on the frontend with Go on the backend. It's designed to
              help developers quickly bootstrap modern web applications with
              best practices built-in.
            </p>
            <p className='text-gray-700 dark:text-gray-300 leading-relaxed'>
              Whether you're building a SaaS product, API service, or full-stack
              web app, this starter kit provides a solid foundation to build
              upon.
            </p>
          </section>

          {/* Technology Stack */}
          <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-8'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
              🛠️ Technology Stack
            </h2>
            <div className='grid md:grid-cols-2 gap-6'>
              <div>
                <h3 className='text-lg font-medium text-blue-600 dark:text-blue-400 mb-3'>
                  Frontend
                </h3>
                <ul className='space-y-2 text-gray-700 dark:text-gray-300'>
                  <li>• React 19 with TypeScript</li>
                  <li>• React Router for navigation</li>
                  <li>• TailwindCSS for styling</li>
                  <li>• ShadCN/UI components</li>
                  <li>• Vite for fast development</li>
                </ul>
              </div>
              <div>
                <h3 className='text-lg font-medium text-green-600 dark:text-green-400 mb-3'>
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
          <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-8'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
              ✨ Key Features
            </h2>
            <div className='grid md:grid-cols-2 gap-4'>
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
          <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-8'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
              🎯 Perfect For
            </h2>
            <div className='grid md:grid-cols-2 gap-4'>
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
          <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-8'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
              🎯 Getting Started
            </h2>
            <div className='bg-gray-100 dark:bg-gray-700 rounded-lg p-4'>
              <p className='text-gray-700 dark:text-gray-300 mb-3'>
                Ready to start building? Here's how to get up and running:
              </p>
              <div className='space-y-2 text-sm'>
                <p className='font-medium text-gray-900 dark:text-white'>
                  1. Clone the repository
                </p>
                <p className='font-medium text-gray-900 dark:text-white'>
                  2. Copy{' '}
                  <code className='bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded'>
                    .env.example
                  </code>{' '}
                  to{' '}
                  <code className='bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded'>
                    .env
                  </code>{' '}
                  and configure database settings
                </p>
                <p className='font-medium text-gray-900 dark:text-white'>
                  3. Run{' '}
                  <code className='bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded'>
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
          <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-8'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
              ⚡ Development Workflow
            </h2>
            <div className='grid md:grid-cols-2 gap-6'>
              <div>
                <h3 className='text-lg font-medium text-blue-600 dark:text-blue-400 mb-3'>
                  🚀 Quick Commands
                </h3>
                <div className='space-y-2 text-sm'>
                  <div className='flex justify-between'>
                    <span className='text-gray-700 dark:text-gray-300'>
                      Start all services:
                    </span>
                    <code className='bg-gray-200 dark:bg-gray-600 px-2 py-1 rounded'>
                      docker-compose up
                    </code>
                  </div>
                  <div className='flex justify-between'>
                    <span className='text-gray-700 dark:text-gray-300'>
                      View logs:
                    </span>
                    <code className='bg-gray-200 dark:bg-gray-600 px-2 py-1 rounded'>
                      docker-compose logs -f
                    </code>
                  </div>
                  <div className='flex justify-between'>
                    <span className='text-gray-700 dark:text-gray-300'>
                      Stop services:
                    </span>
                    <code className='bg-gray-200 dark:bg-gray-600 px-2 py-1 rounded'>
                      docker-compose down
                    </code>
                  </div>
                </div>
              </div>
              <div>
                <h3 className='text-lg font-medium text-green-600 dark:text-green-400 mb-3'>
                  🔗 Useful Links
                </h3>
                <div className='space-y-2 text-sm'>
                  <div className='text-gray-700 dark:text-gray-300'>
                    📚{' '}
                    <a
                      href={`${API_BASE_URL}/swagger/`}
                      className='text-blue-600 dark:text-blue-400 hover:underline'
                    >
                      API Documentation
                    </a>
                  </div>
                  <div className='text-gray-700 dark:text-gray-300'>
                    🎮{' '}
                    <a
                      href='/demo'
                      className='text-blue-600 dark:text-blue-400 hover:underline'
                    >
                      Try Live Demo
                    </a>
                  </div>
                  <div className='text-gray-700 dark:text-gray-300'>
                    🐙{' '}
                    <a
                      href='https://github.com'
                      className='text-blue-600 dark:text-blue-400 hover:underline'
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
};

export default function Home() {
  return <HomePage />;
}
