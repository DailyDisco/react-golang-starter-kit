import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/layout-demo/')({
  component: CustomLayoutDemo,
});

function CustomLayoutDemo() {
  return (
    <div className='space-y-6 p-8'>
      <div className='space-y-4 text-center'>
        <h1 className='text-3xl font-bold text-gray-900 dark:text-white'>
          Custom Layout Demo
        </h1>
        <p className='text-gray-600 dark:text-gray-300'>
          This page uses a custom layout component instead of the root layout!
        </p>
      </div>

      <div className='rounded-lg border border-blue-200 bg-blue-50 p-6 dark:border-blue-800 dark:bg-blue-900/20'>
        <h2 className='mb-2 text-lg font-semibold text-blue-900 dark:text-blue-100'>
          🎯 Layout Difference
        </h2>
        <p className='text-sm text-blue-800 dark:text-blue-200'>
          Notice: This page uses the custom Layout component, while other pages
          use the root layout with navbar and footer.
        </p>
      </div>

      <div className='flex justify-center gap-4'>
        <Link
          to='/'
          className='rounded-lg bg-gray-600 px-4 py-2 font-medium text-white transition-colors hover:bg-gray-700 dark:bg-gray-700 dark:hover:bg-gray-600'
        >
          Back to Home
        </Link>
        <Link
          to='/demo'
          className='rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-600'
        >
          View Demo
        </Link>
      </div>
    </div>
  );
}
