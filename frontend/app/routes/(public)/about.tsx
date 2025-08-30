import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/(public)/about')({
  component: AboutPage,
  // Add metadata for SEO
  meta: () => [
    { title: 'About Us - React Go Starter' },
    { description: 'Learn more about our modern full-stack starter kit' }
  ],
});

function AboutPage() {
  return (
    <div className='max-w-4xl mx-auto py-12 px-4'>
      <h1 className='text-3xl font-bold mb-6'>About This Project</h1>
      <p className='text-gray-600 dark:text-gray-300 leading-relaxed'>
        This is a custom about page demonstrating TanStack Router file-based routing.
      </p>
    </div>
  );
}