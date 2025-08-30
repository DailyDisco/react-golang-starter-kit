import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/(public)/about')({
  component: AboutPage,
  // Add metadata for SEO
  meta: () => [
    { title: 'About Us - React Go Starter' },
    { description: 'Learn more about our modern full-stack starter kit' },
  ],
});

function AboutPage() {
  return (
    <div className='mx-auto max-w-4xl px-4 py-12'>
      <h1 className='mb-6 text-3xl font-bold'>About This Project</h1>
      <p className='leading-relaxed text-gray-600 dark:text-gray-300'>
        This is a custom about page demonstrating TanStack Router file-based
        routing.
      </p>
    </div>
  );
}
