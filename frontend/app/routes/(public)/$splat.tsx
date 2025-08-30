import { createFileRoute, Link } from '@tanstack/react-router';
import { Home, ArrowLeft, FileQuestion } from 'lucide-react';

export const Route = createFileRoute('/(public)/$splat')({
  component: NotFoundPage,
});

function NotFoundPage() {
  return (
    <main className='flex-1 bg-gradient-to-br from-slate-50 via-gray-50 to-zinc-50 dark:from-slate-950 dark:via-gray-950 dark:to-zinc-950'>
      <div className='flex min-h-[60vh] items-center justify-center px-4 py-12'>
        <div className='mx-auto w-full max-w-lg text-center'>
          {/* Icon and 404 Number */}
          <div className='mb-8'>
            <div className='mb-6 inline-flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-blue-100 to-indigo-100 dark:from-blue-900/30 dark:to-indigo-900/30'>
              <FileQuestion className='h-10 w-10 text-blue-600 dark:text-blue-400' />
            </div>
            <h1 className='bg-gradient-to-r from-slate-600 to-slate-800 bg-clip-text text-6xl font-bold text-transparent select-none md:text-7xl dark:from-slate-300 dark:to-slate-500'>
              404
            </h1>
          </div>

          {/* Main Content Card */}
          <div className='rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm md:p-10 dark:border-gray-800/50 dark:bg-gray-900/80'>
            <div className='space-y-6'>
              <div>
                <h2 className='mb-3 text-2xl font-semibold text-gray-900 md:text-3xl dark:text-white'>
                  Page Not Found
                </h2>
                <p className='leading-relaxed text-gray-600 dark:text-gray-300'>
                  The page you're looking for doesn't exist or may have been
                  moved.
                </p>
              </div>

              {/* Action Buttons */}
              <div className='flex flex-col gap-3 pt-2 sm:flex-row'>
                <Link
                  to='/'
                  search={{}}
                  className='inline-flex flex-1 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-3 font-medium text-white shadow-lg transition-all duration-200 hover:scale-[1.02] hover:from-blue-700 hover:to-blue-800 hover:shadow-xl'
                >
                  <Home className='h-4 w-4' />
                  Go Home
                </Link>
                <button
                  onClick={() => window.history.back()}
                  className='inline-flex flex-1 items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white px-6 py-3 font-medium text-gray-700 shadow-sm transition-all duration-200 hover:scale-[1.02] hover:bg-gray-50 hover:shadow-md dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700'
                >
                  <ArrowLeft className='h-4 w-4' />
                  Go Back
                </button>
              </div>

              {/* Additional Help */}
              <div className='border-t border-gray-200 pt-4 dark:border-gray-700'>
                <p className='text-sm text-gray-500 dark:text-gray-400'>
                  Try checking the URL or return to our{' '}
                  <Link
                    to='/demo'
                    search={{}}
                    className='font-medium text-blue-600 transition-colors hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300'
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
}
